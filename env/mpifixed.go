// Copyright (c) 2020, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package env

import (
	"fmt"
	"log/slog"

	"cogentcore.org/lab/base/randx"
	"cogentcore.org/lab/table"
	"cogentcore.org/lab/tensor"
	"cogentcore.org/lab/tensor/tensormpi"
)

// MPIFixedTable is an MPI-enabled version of the [FixedTable], which is
// a basic Env that manages patterns from a [table.Table[, with
// either sequential or permuted random ordering, and a Trial counter to
// record iterations through the table.
// Use [table.NewView] to provide a unique indexed view of a shared table.
// The MPI version distributes trials across MPI procs, in the Order list.
// It is ESSENTIAL that the number of trials (rows) in Table is
// evenly divisible by number of MPI procs!
// If all nodes start with the same seed, it should remain synchronized.
type MPIFixedTable struct {

	// name of this environment
	Name string

	// Table has the set of patterns to output.
	// The indexes are used for the *sequential* view so you can easily
	// sort / split / filter the patterns to be presented using this view.
	// This adds the random permuted Order on top of those if !sequential.
	Table *table.Table

	// present items from the table in sequential order (i.e., according to the indexed view on the Table)?  otherwise permuted random order
	Sequential bool

	// permuted order of items to present if not sequential -- updated every time through the list
	Order []int

	// current ordinal item in Table -- if Sequential then = row number in table, otherwise is index in Order list that then gives row number in Table
	Trial Counter `display:"inline"`

	// if Table has a Name column, this is the contents of that
	TrialName CurPrevString

	// if Table has a Group column, this is contents of that
	GroupName CurPrevString

	// name of the Name column -- defaults to 'Name'
	NameCol string

	// name of the Group column -- defaults to 'Group'
	GroupCol string

	// for MPI, trial we start each epoch on, as index into Order
	TrialSt int

	// for MPI, trial number we end each epoch before (i.e., when ctr gets to Ed, restarts)
	TrialEd int

	// Rand is the random number generator for the env.
	// Created in Init if not already there.
	Rand randx.Rand `display:"-"`

	// RunRandSeed is the random seed multiplier for run counter.
	// It is set to 173 if 0 at start for consistent results by default.
	RunRandSeed int64 `edit:"-"`
}

func (ev *MPIFixedTable) Validate() error {
	if ev.Table == nil {
		return fmt.Errorf("MPIFixedTable: %v has no Table set", ev.Name)
	}
	if ev.Table.NumColumns() == 0 {
		return fmt.Errorf("MPIFixedTable: %v Table has no columns -- Outputs will be invalid", ev.Name)
	}
	return nil
}

func (ev *MPIFixedTable) Label() string { return ev.Name }

func (ev *MPIFixedTable) String() string {
	s := ev.TrialName.Cur
	if ev.GroupName.Cur != "" {
		s = ev.GroupName.Cur + "_" + s
	}
	return s
}

func (ev *MPIFixedTable) Init(run int) {
	if ev.RunRandSeed == 0 {
		ev.RunRandSeed = 173
	}
	randx.InitSysRand(&ev.Rand, ev.RunRandSeed*(int64(run)+1))
	if ev.NameCol == "" {
		ev.NameCol = "Name"
	}
	if ev.GroupCol == "" {
		ev.GroupCol = "Group"
	}
	ev.Trial.Init()
	ev.NewOrder()
	ev.Trial.Cur = ev.TrialSt - 1 // init state -- key so that first Step() = ft.TrialSt
}

// NewOrder sets a new random Order based on number of rows in the table.
func (ev *MPIFixedTable) NewOrder() {
	np := ev.Table.NumRows()
	ev.Order = ev.Rand.Perm(np) // always start with new one so random order is identical
	// and always maintain Order so random number usage is same regardless, and if
	// user switches between Sequential and random at any point, it all works..
	ev.TrialSt, ev.TrialEd, _ = tensormpi.AllocN(np)
	ev.Trial.Max = ev.TrialEd
}

// PermuteOrder permutes the existing order table to get a new random sequence of inputs
// just calls: randx.PermuteInts(ft.Order)
func (ev *MPIFixedTable) PermuteOrder() {
	randx.PermuteInts(ev.Order, ev.Rand)
}

// Row returns the current row number in table, based on Sequential / perumuted Order.
func (ev *MPIFixedTable) Row() int {
	if ev.Sequential {
		return ev.Trial.Cur
	}
	return ev.Order[ev.Trial.Cur]
}

func (ev *MPIFixedTable) SetTrialName() {
	if nms := ev.Table.Column(ev.NameCol); nms != nil {
		rw := ev.Row()
		if rw >= 0 && rw < nms.Len() {
			ev.TrialName.Set(nms.StringRow(rw, 0))
		}
	}
}

func (ev *MPIFixedTable) SetGroupName() {
	if nms := ev.Table.Column(ev.GroupCol); nms != nil {
		rw := ev.Row()
		if rw >= 0 && rw < nms.Len() {
			ev.GroupName.Set(nms.StringRow(rw, 0))
		}
	}
}

func (ev *MPIFixedTable) Step() bool {
	if ev.Trial.Incr() { // if true, hit max, reset to 0
		ev.Trial.Cur = ev.TrialSt // key to reset always to start
		ev.PermuteOrder()
	}
	ev.SetTrialName()
	ev.SetGroupName()
	return true
}

func (ev *MPIFixedTable) State(element string) tensor.Values {
	et := ev.Table.Column(element).RowTensor(ev.Row())
	if et == nil {
		slog.Error("MPIFixedTable.State: could not find:", "element", element)
	}
	return et
}

// Compile-time check that implements Env interface
var _ Env = (*MPIFixedTable)(nil)
