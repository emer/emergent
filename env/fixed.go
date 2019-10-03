// Copyright (c) 2019, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package env

import (
	"fmt"
	"log"
	"math/rand"

	"github.com/emer/emergent/erand"
	"github.com/emer/etable/etable"
	"github.com/emer/etable/etensor"
)

// FixedTable is a basic Env that manages patterns from an etable.Table, with
// either sequential or permuted random ordering, and uses standard Trial / Epoch
// TimeScale counters to record progress and iterations through the table.
// It also records the outer loop of Run as provided by the model.
// It uses an IdxView indexed view of the Table, so a single shared table
// can be used across different environments, with each having its own unique view.
type FixedTable struct {
	Nm           string          `desc:"name of this environment"`
	Dsc          string          `desc:"description of this environment"`
	Table        *etable.IdxView `desc:"this is an indexed view of the table with the set of patterns to output -- the indexes are used for the *sequential* view so you can easily sort / split / filter the patterns to be presented using this view -- we then add the random permuted Order on top of those if !sequential"`
	Sequential   bool            `desc:"present items from the table in sequential order (i.e., according to the indexed view on the Table)?  otherwise permuted random order"`
	Order        []int           `desc:"permuted order of items to present if not sequential -- updated every time through the list"`
	Run          Ctr             `view:"inline" desc:"current run of model as provided during Init"`
	Epoch        Ctr             `view:"inline" desc:"number of times through entire set of patterns"`
	Trial        Ctr             `view:"inline" desc:"current ordinal item in Table -- if Sequential then = row number in table, otherwise is index in Order list that then gives row number in Table"`
	TrialName    string          `desc:"if Table has a Name column, this is the contents of that for current trial"`
	PrvTrialName string          `desc:"previous trial name"`
}

func (ft *FixedTable) Name() string { return ft.Nm }
func (ft *FixedTable) Desc() string { return ft.Dsc }

func (ft *FixedTable) Validate() error {
	if ft.Table == nil || ft.Table.Table == nil {
		return fmt.Errorf("env.FixedTable: %v has no Table set", ft.Nm)
	}
	if ft.Table.Table.NumCols() == 0 {
		return fmt.Errorf("env.FixedTable: %v Table has no columns -- Outputs will be invalid", ft.Nm)
	}
	return nil
}

func (ft *FixedTable) Counters() []TimeScales {
	return []TimeScales{Run, Epoch, Trial}
}

func (ft *FixedTable) States() Elements {
	els := Elements{}
	els.FromSchema(ft.Table.Table.Schema())
	return els
}

func (ft *FixedTable) Actions() Elements {
	return nil
}

func (ft *FixedTable) Init(run int) {
	ft.Run.Scale = Run
	ft.Epoch.Scale = Epoch
	ft.Trial.Scale = Trial
	ft.Run.Init()
	ft.Epoch.Init()
	ft.Trial.Init()
	ft.Run.Cur = run
	np := ft.Table.Len()
	ft.Order = rand.Perm(np) // always start with new one so random order is identical
	// and always maintain Order so random number usage is same regardless, and if
	// user switches between Sequential and random at any point, it all works..
	ft.Trial.Max = np
	ft.Trial.Cur = -1 // init state -- key so that first Step() = 0
}

// Row returns the current row number in table, based on Sequential / perumuted Order and
// already de-referenced through the IdxView's indexes to get the actual row in the table.
func (ft *FixedTable) Row() int {
	if ft.Sequential {
		return ft.Table.Idxs[ft.Trial.Cur]
	}
	return ft.Table.Idxs[ft.Order[ft.Trial.Cur]]
}

func (ft *FixedTable) SetTrialName() {
	if nms := ft.Table.Table.ColByName("Name"); nms != nil {
		rw := ft.Row()
		if rw >= 0 && rw < nms.Len() {
			ft.TrialName = nms.StringVal1D(rw)
		}
	}
}

func (ft *FixedTable) Step() bool {
	ft.Epoch.Same() // good idea to just reset all non-inner-most counters at start

	if ft.Trial.Incr() { // if true, hit max, reset to 0
		erand.PermuteInts(ft.Order)
		ft.Epoch.Incr()
	}
	ft.PrvTrialName = ft.TrialName
	ft.SetTrialName()
	return true
}

func (ft *FixedTable) State(element string) etensor.Tensor {
	et, err := ft.Table.Table.CellTensorTry(element, ft.Row())
	if err != nil {
		log.Println(err)
	}
	return et
}

func (ft *FixedTable) Action(element string, input etensor.Tensor) {
	// nop
}

func (ft *FixedTable) Counter(scale TimeScales) (cur, prv int, chg bool) {
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

// Compile-time check that implements Env interface
var _ Env = (*FixedTable)(nil)
