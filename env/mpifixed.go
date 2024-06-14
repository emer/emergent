// Copyright (c) 2020, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package env

import (
	"fmt"
	"log"
	"math/rand"

	"cogentcore.org/core/base/randx"
	"cogentcore.org/core/tensor"
	"cogentcore.org/core/tensor/table"
	"cogentcore.org/core/tensor/tensormpi"
)

// MPIFixedTable is an MPI-enabled version of the FixedTable, which is
// a basic Env that manages patterns from an table.Table, with
// either sequential or permuted random ordering, and uses standard Trial / Epoch
// TimeScale counters to record progress and iterations through the table.
// It also records the outer loop of Run as provided by the model.
// It uses an IndexView indexed view of the Table, so a single shared table
// can be used across different environments, with each having its own unique view.
// The MPI version distributes trials across MPI procs, in the Order list.
// It is ESSENTIAL that the number of trials (rows) in Table is
// evenly divisible by number of MPI procs!
// If all nodes start with the same seed, it should remain synchronized.
type MPIFixedTable struct {

	// name of this environment
	Nm string

	// description of this environment
	Dsc string

	// this is an indexed view of the table with the set of patterns to output -- the indexes are used for the *sequential* view so you can easily sort / split / filter the patterns to be presented using this view -- we then add the random permuted Order on top of those if !sequential
	Table *table.IndexView

	// present items from the table in sequential order (i.e., according to the indexed view on the Table)?  otherwise permuted random order
	Sequential bool

	// permuted order of items to present if not sequential -- updated every time through the list
	Order []int

	// current run of model as provided during Init
	Run Ctr `display:"inline"`

	// number of times through entire set of patterns
	Epoch Ctr `display:"inline"`

	// current ordinal item in Table -- if Sequential then = row number in table, otherwise is index in Order list that then gives row number in Table
	Trial Ctr `display:"inline"`

	// if Table has a Name column, this is the contents of that
	TrialName CurPrvString

	// if Table has a Group column, this is contents of that
	GroupName CurPrvString

	// name of the Name column -- defaults to 'Name'
	NameCol string

	// name of the Group column -- defaults to 'Group'
	GroupCol string

	// for MPI, trial we start each epoch on, as index into Order
	TrialSt int

	// for MPI, trial number we end each epoch before (i.e., when ctr gets to Ed, restarts)
	TrialEd int
}

func (ft *MPIFixedTable) Name() string { return ft.Nm }
func (ft *MPIFixedTable) Desc() string { return ft.Dsc }

func (ft *MPIFixedTable) Validate() error {
	if ft.Table == nil || ft.Table.Table == nil {
		return fmt.Errorf("MPIFixedTable: %v has no Table set", ft.Nm)
	}
	if ft.Table.Table.NumColumns() == 0 {
		return fmt.Errorf("MPIFixedTable: %v Table has no columns -- Outputs will be invalid", ft.Nm)
	}
	return nil
}

func (ft *MPIFixedTable) Init(run int) {
	if ft.NameCol == "" {
		ft.NameCol = "Name"
	}
	if ft.GroupCol == "" {
		ft.GroupCol = "Group"
	}
	ft.Run.Scale = Run
	ft.Epoch.Scale = Epoch
	ft.Trial.Scale = Trial
	ft.Run.Init()
	ft.Epoch.Init()
	ft.Trial.Init()
	ft.Run.Cur = run
	ft.NewOrder()
	ft.Trial.Cur = ft.TrialSt - 1 // init state -- key so that first Step() = ft.TrialSt
}

// NewOrder sets a new random Order based on number of rows in the table.
func (ft *MPIFixedTable) NewOrder() {
	np := ft.Table.Len()
	ft.Order = rand.Perm(np) // always start with new one so random order is identical
	// and always maintain Order so random number usage is same regardless, and if
	// user switches between Sequential and random at any point, it all works..
	ft.TrialSt, ft.TrialEd, _ = tensormpi.AllocN(np)
	ft.Trial.Max = ft.TrialEd
}

// PermuteOrder permutes the existing order table to get a new random sequence of inputs
// just calls: randx.PermuteInts(ft.Order)
func (ft *MPIFixedTable) PermuteOrder() {
	randx.PermuteInts(ft.Order)
}

// Row returns the current row number in table, based on Sequential / perumuted Order and
// already de-referenced through the IndexView's indexes to get the actual row in the table.
func (ft *MPIFixedTable) Row() int {
	if ft.Sequential {
		return ft.Table.Indexes[ft.Trial.Cur]
	}
	return ft.Table.Indexes[ft.Order[ft.Trial.Cur]]
}

func (ft *MPIFixedTable) SetTrialName() {
	if nms := ft.Table.Table.ColumnByName(ft.NameCol); nms != nil {
		rw := ft.Row()
		if rw >= 0 && rw < nms.Len() {
			ft.TrialName.Set(nms.String1D(rw))
		}
	}
}

func (ft *MPIFixedTable) SetGroupName() {
	if nms := ft.Table.Table.ColumnByName(ft.GroupCol); nms != nil {
		rw := ft.Row()
		if rw >= 0 && rw < nms.Len() {
			ft.GroupName.Set(nms.String1D(rw))
		}
	}
}

func (ft *MPIFixedTable) Step() bool {
	ft.Epoch.Same() // good idea to just reset all non-inner-most counters at start

	if ft.Trial.Incr() { // if true, hit max, reset to 0
		ft.Trial.Cur = ft.TrialSt // key to reset always to start
		ft.PermuteOrder()
		ft.Epoch.Incr()
	}
	ft.SetTrialName()
	ft.SetGroupName()
	return true
}

func (ft *MPIFixedTable) Counters() []TimeScales {
	return []TimeScales{Run, Epoch, Trial}
}

func (ft *MPIFixedTable) Counter(scale TimeScales) (cur, prv int, chg bool) {
	switch scale {
	case Run:
		return ft.Run.Query()
	case Epoch:
		return ft.Epoch.Query()
	case Trial:
		return ft.Trial.Query()
	}
	return -1, -1, false
}

func (ft *MPIFixedTable) States() Elements {
	// els := Elements{}
	// els.FromSchema(ft.Table.Table.Schema())
	// return els
	return nil
}

func (ft *MPIFixedTable) State(element string) tensor.Tensor {
	et := ft.Table.Table.Tensor(element, ft.Row())
	if et == nil {
		log.Println("MPIFixedTable.State -- could not find element:", element)
	}
	return et
}

func (ft *MPIFixedTable) Actions() Elements {
	return nil
}

func (ft *MPIFixedTable) Action(element string, input tensor.Tensor) {
	// nop
}

// Compile-time check that implements Env interface
var _ Env = (*MPIFixedTable)(nil)
