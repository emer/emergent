// Copyright (c) 2019, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package env

import (
	"fmt"
	"log/slog"

	"cogentcore.org/lab/base/randx"
	"cogentcore.org/lab/table"
	"cogentcore.org/lab/tensor"
)

// FixedTable is a basic Env that manages patterns from a [table.Table], with
// either sequential or permuted random ordering, with a Trial counter
// to record progress and iterations through the table.
// Use [table.NewView] to provide a unique indexed view of a shared table.
type FixedTable struct {
	// name of this environment, usually Train vs. Test.
	Name string

	// Table has the set of patterns to output.
	// The indexes are used for the *sequential* view so you can easily
	// sort / split / filter the patterns to be presented using this view.
	// This adds the random permuted Order on top of those if !sequential.
	Table *table.Table

	// present items from the table in sequential order (i.e., according to
	// the indexed view on the Table)?  otherwise permuted random order.
	Sequential bool

	// permuted order of items to present if not sequential.
	// updated every time through the list.
	Order []int

	// current ordinal item in Table. if Sequential then = row number in table,
	// otherwise is index in Order list that then gives row number in Table.
	Trial Counter `display:"inline"`

	// if Table has a Name column, this is the contents of that.
	TrialName CurPrevString

	// if Table has a Group column, this is contents of that.
	GroupName CurPrevString

	// name of the Name column -- defaults to 'Name'.
	NameCol string

	// name of the Group column -- defaults to 'Group'.
	GroupCol string

	// Rand is the random number generator for the env.
	// Created in Init if not already there.
	Rand randx.Rand `display:"-"`

	// RunRandSeed is the random seed multiplier for run counter.
	// It is set to 173 if 0 at start for consistent results by default.
	RunRandSeed int64 `edit:"-"`
}

func (ev *FixedTable) Validate() error {
	if ev.Table == nil {
		return fmt.Errorf("env.FixedTable: %v has no Table set", ev.Name)
	}
	if ev.Table.NumColumns() == 0 {
		return fmt.Errorf("env.FixedTable: %v Table has no columns -- Outputs will be invalid", ev.Name)
	}
	return nil
}

func (ev *FixedTable) Label() string { return ev.Name }

func (ev *FixedTable) String() string {
	s := ev.TrialName.Cur
	if ev.GroupName.Cur != "" {
		s = ev.GroupName.Cur + "_" + s
	}
	return s
}

func (ev *FixedTable) Init(run int) {
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
	ev.Trial.Cur = -1 // init state -- key so that first Step() = 0
}

// Config configures the environment to use given table IndexView and
// evaluation mode (e.g., etime.Train.String()).  If mode is Train
// then a Run counter is added, otherwise just Epoch and Trial.
// NameCol and GroupCol are initialized to "Name" and "Group"
// so set these to something else after this if needed.
func (ev *FixedTable) Config(tbl *table.Table) {
	ev.Table = tbl
	ev.Init(0)
}

// NewOrder sets a new random Order based on number of rows in the table.
func (ev *FixedTable) NewOrder() {
	np := ev.Table.NumRows()
	ev.Order = ev.Rand.Perm(np) // always start with new one so random order is identical
	// and always maintain Order so random number usage is same regardless, and if
	// user switches between Sequential and random at any point, it all works..
	ev.Trial.Max = np
}

// PermuteOrder permutes the existing order table to get a new random sequence of inputs
// just calls: randx.PermuteInts(ft.Order)
func (ev *FixedTable) PermuteOrder() {
	randx.PermuteInts(ev.Order, ev.Rand)
}

// Row returns the current row number in table, based on Sequential / perumuted Order.
func (ev *FixedTable) Row() int {
	if ev.Sequential {
		return ev.Trial.Cur
	}
	return ev.Order[ev.Trial.Cur]
}

func (ev *FixedTable) SetTrialName() {
	if nms := ev.Table.Column(ev.NameCol); nms != nil {
		rw := ev.Row()
		if rw >= 0 && rw < nms.Len() {
			ev.TrialName.Set(nms.StringRow(rw, 0))
		}
	}
}

func (ev *FixedTable) SetGroupName() {
	if nms := ev.Table.Column(ev.GroupCol); nms != nil {
		rw := ev.Row()
		if rw >= 0 && rw < nms.Len() {
			ev.GroupName.Set(nms.StringRow(rw, 0))
		}
	}
}

func (ev *FixedTable) Step() bool {
	if ev.Trial.Incr() { // if true, hit max, reset to 0
		ev.PermuteOrder()
	}
	ev.SetTrialName()
	ev.SetGroupName()
	return true
}

func (ev *FixedTable) State(element string) tensor.Values {
	et := ev.Table.Column(element).RowTensor(ev.Row())
	if et == nil {
		slog.Error("FixedTable.State: could not find", "element", element)
	}
	return et
}

// Compile-time check that implements Env interface
var _ Env = (*FixedTable)(nil)
