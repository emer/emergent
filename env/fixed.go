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

// FixedTable is a basic Env that outputs patterns from an etable.Table in
// either sequential or permuted random order, and uses standard Trial / Epoch
// TimeScale counters to record progress and iterations through the table.
// It also records the outer loop of Run as provided by the model
type FixedTable struct {
	Nm         string        `desc:"name of this environment"`
	Dsc        string        `desc:"description of this environment"`
	Table      *etable.Table `desc:"this is the table with the set of patterns to output"`
	Sequential bool          `desc:"present items from the table in sequential order?  otherwise permuted random order"`
	Order      []int         `desc:"permuted order of items to present if not sequential -- updated every time through the list"`
	Run        Ctr           `view:"inline" desc:"current run of model as provided during Init"`
	Epoch      Ctr           `view:"inline" desc:"number of times through entire set of patterns"`
	Trial      Ctr           `view:"inline" desc:"current ordinal item in Table -- if Sequential then = row number in table, otherwise is index in Order list that then gives row number in Table"`
	TrialName  string        `desc:"if Table has a Name column, this is the contents of that for current trial"`
}

func (ft *FixedTable) Name() string { return ft.Nm }
func (ft *FixedTable) Desc() string { return ft.Dsc }

func (ft *FixedTable) Validate() error {
	ft.Run.Scale = Run
	ft.Epoch.Scale = Epoch
	ft.Trial.Scale = Trial
	if ft.Table == nil {
		return fmt.Errorf("env.FixedTable: %v has no Table set", ft.Nm)
	}
	if ft.Table.NumCols() == 0 {
		return fmt.Errorf("env.FixedTable: %v Table has no columns -- Outputs will be invalid", ft.Nm)
	}
	return nil
}

func (ft *FixedTable) Counters() []TimeScales {
	return []TimeScales{Run, Epoch, Trial}
}

func (ft *FixedTable) Outputs() Channels {
	ch := Channels{}
	ch.FromSchema(ft.Table.Schema())
	return ch
}

func (ft *FixedTable) Inputs() Channels {
	return nil
}

func (ft *FixedTable) Init(run int) {
	ft.Run.Init()
	ft.Epoch.Init()
	ft.Trial.Init()
	ft.Run.Cur = run
	np := ft.Table.NumRows()
	ft.Order = rand.Perm(np) // always start with new one so random order is identical
	// and always maintain Order so random number usage is same regardless, and if
	// user switches between Sequential and random at any point, it all works..
	ft.Trial.Max = np
	ft.Trial.Cur = -1 // init state
}

// Row returns the current row number in table, based on Sequential / Order
func (ft *FixedTable) Row() int {
	if ft.Sequential {
		return ft.Trial.Cur
	}
	return ft.Order[ft.Trial.Cur]
}

func (ft *FixedTable) Next() bool {
	if ft.Trial.Incr() { // if true, hit max, reset to 0
		erand.PermuteInts(ft.Order)
		ft.Epoch.Incr()
	}
	if nms := ft.Table.ColByName("Name"); nms != nil {
		ft.TrialName = nms.StringVal1D(ft.Row())
	}
	return true
}

func (ft *FixedTable) Output(channel string) etensor.Tensor {
	et, err := ft.Table.CellTensorByNameTry(channel, ft.Row())
	if err != nil {
		log.Println(err)
	}
	return et
}

func (ft *FixedTable) Input(channel string, input etensor.Tensor) {
	// nop
}

func (ft *FixedTable) Counter(scale TimeScales) (int, bool) {
	switch scale {
	case Run:
		return ft.Run.Query()
	case Epoch:
		return ft.Epoch.Query()
	case Trial:
		return ft.Epoch.Query()
	}
	return -1, false
}

// Compile-time check that implements Env interface
var _ Env = (*FixedTable)(nil)
