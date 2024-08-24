// Copyright (c) 2019, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package env

import (
	"fmt"
	"log"
	"math/rand"

	"cogentcore.org/core/base/errors"
	"cogentcore.org/core/base/randx"
	"cogentcore.org/core/tensor"
	"cogentcore.org/core/tensor/table"
	"github.com/emer/emergent/v2/etime"
)

// FixedTable is a basic Env that manages patterns from an table.Table, with
// either sequential or permuted random ordering, with the Trial counters
// to record progress and iterations through the table.
// It uses an IndexView indexed view of the Table, so a single shared table
// can be used across different environments, with each having its own unique view.
type FixedTable struct {
	// name of this environment, usually Train vs. Test.
	Name string

	// this is an indexed view of the table with the set of patterns to output.
	// The indexes are used for the *sequential* view so you can easily
	// sort / split / filter the patterns to be presented using this view.
	// This adds the random permuted Order on top of those if !sequential.
	Table *table.IndexView

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
	TrialName CurPrvString

	// if Table has a Group column, this is contents of that.
	GroupName CurPrvString

	// name of the Name column -- defaults to 'Name'.
	NameCol string

	// name of the Group column -- defaults to 'Group'.
	GroupCol string
}

func (ft *FixedTable) Validate() error {
	if ft.Table == nil || ft.Table.Table == nil {
		return fmt.Errorf("env.FixedTable: %v has no Table set", ft.Name)
	}
	if ft.Table.Table.NumColumns() == 0 {
		return fmt.Errorf("env.FixedTable: %v Table has no columns -- Outputs will be invalid", ft.Name)
	}
	return nil
}

func (ft *FixedTable) Label() string { return ft.Name }

func (ft *FixedTable) Init(run int) {
	if ft.NameCol == "" {
		ft.NameCol = "Name"
	}
	if ft.GroupCol == "" {
		ft.GroupCol = "Group"
	}
	ft.Trial.Scale = etime.Trial
	ft.Trial.Init()
	ft.NewOrder()
	ft.Trial.Cur = -1 // init state -- key so that first Step() = 0
}

// Config configures the environment to use given table IndexView and
// evaluation mode (e.g., etime.Train.String()).  If mode is Train
// then a Run counter is added, otherwise just Epoch and Trial.
// NameCol and GroupCol are initialized to "Name" and "Group"
// so set these to something else after this if needed.
func (ft *FixedTable) Config(tbl *table.IndexView) {
	ft.Table = tbl
	ft.Init(0)
}

// NewOrder sets a new random Order based on number of rows in the table.
func (ft *FixedTable) NewOrder() {
	np := ft.Table.Len()
	ft.Order = rand.Perm(np) // always start with new one so random order is identical
	// and always maintain Order so random number usage is same regardless, and if
	// user switches between Sequential and random at any point, it all works..
	ft.Trial.Max = np
}

// PermuteOrder permutes the existing order table to get a new random sequence of inputs
// just calls: randx.PermuteInts(ft.Order)
func (ft *FixedTable) PermuteOrder() {
	randx.PermuteInts(ft.Order)
}

// Row returns the current row number in table, based on Sequential / perumuted Order and
// already de-referenced through the IndexView's indexes to get the actual row in the table.
func (ft *FixedTable) Row() int {
	if ft.Sequential {
		return ft.Table.Indexes[ft.Trial.Cur]
	}
	return ft.Table.Indexes[ft.Order[ft.Trial.Cur]]
}

func (ft *FixedTable) SetTrialName() {
	if nms := errors.Ignore1(ft.Table.Table.ColumnByName(ft.NameCol)); nms != nil {
		rw := ft.Row()
		if rw >= 0 && rw < nms.Len() {
			ft.TrialName.Set(nms.String1D(rw))
		}
	}
}

func (ft *FixedTable) SetGroupName() {
	if nms := errors.Ignore1(ft.Table.Table.ColumnByName(ft.GroupCol)); nms != nil {
		rw := ft.Row()
		if rw >= 0 && rw < nms.Len() {
			ft.GroupName.Set(nms.String1D(rw))
		}
	}
}

func (ft *FixedTable) Step() bool {
	if ft.Trial.Incr() { // if true, hit max, reset to 0
		ft.PermuteOrder()
	}
	ft.SetTrialName()
	ft.SetGroupName()
	return true
}

func (ft *FixedTable) State(element string) tensor.Tensor {
	et := ft.Table.Table.Tensor(element, ft.Row())
	if et == nil {
		log.Println("FixedTable.State -- could not find element:", element)
	}
	return et
}

func (ft *FixedTable) Action(element string, input tensor.Tensor) {
	// nop
}

// Compile-time check that implements Env interface
var _ Env = (*FixedTable)(nil)
