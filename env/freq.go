// Copyright (c) 2019, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package env

import (
	"fmt"
	"log"
	"math"

	"github.com/emer/emergent/erand"
	"github.com/emer/etable/etable"
	"github.com/emer/etable/etensor"
)

// FreqTable is an Env that manages patterns from an etable.Table with frequency
// information so that items are presented according to their associated frequencies
// which are effectively probabilities of presenting any given input -- must have
// a Freq column with these numbers in the table (actual col name in FreqCol).
// Either sequential or permuted random ordering is supported, with std Trial / Epoch
// TimeScale counters to record progress and iterations through the table.
// It also records the outer loop of Run as provided by the model.
// It uses an IdxView indexed view of the Table, so a single shared table
// can be used across different environments, with each having its own unique view.
type FreqTable struct {
	Nm           string          `desc:"name of this environment"`
	Dsc          string          `desc:"description of this environment"`
	Table        *etable.IdxView `desc:"this is an indexed view of the table with the set of patterns to output -- the indexes are used for the *sequential* view so you can easily sort / split / filter the patterns to be presented using this view -- we then add the random permuted Order on top of those if !sequential"`
	NSamples     float64         `desc:"number of samples to use in constructing the list of items to present according to frequency -- number per epoch ~ NSamples * Freq -- see RndSamp option"`
	RndSamp      bool            `desc:"if true, use random sampling of items NSamples times according to given Freq probability value -- otherwise just directly add NSamples * Freq items to the list"`
	Sequential   bool            `desc:"present items from the table in sequential order (i.e., according to the indexed view on the Table)?  otherwise permuted random order.  All repetitions of given item will be sequential if Sequential"`
	Order        []int           `desc:"list of items to present, with repetitions -- updated every time through the list"`
	Run          Ctr             `view:"inline" desc:"current run of model as provided during Init"`
	Epoch        Ctr             `view:"inline" desc:"number of times through entire set of patterns"`
	Trial        Ctr             `view:"inline" desc:"current ordinal item in Table -- if Sequential then = row number in table, otherwise is index in Order list that then gives row number in Table"`
	TrialName    string          `desc:"if Table has a Name column, this is the contents of that for current trial"`
	PrvTrialName string          `desc:"if Table has a Name column, this is the contents of that for current trial"`
	GroupName    CurPrvString    `desc:"if Table has a Group column, this is contents of that"`
	NameCol      string          `desc:"name of the Name column -- defaults to 'Name'"`
	GroupCol     string          `desc:"name of the Group column -- defaults to 'Group'"`
	FreqCol      string          `desc:"name of the Freq column -- defaults to 'Freq'"`
}

func (ft *FreqTable) Name() string { return ft.Nm }
func (ft *FreqTable) Desc() string { return ft.Dsc }

func (ft *FreqTable) Validate() error {
	if ft.Table == nil || ft.Table.Table == nil {
		return fmt.Errorf("env.FreqTable: %v has no Table set", ft.Nm)
	}
	if ft.Table.Table.NumCols() == 0 {
		return fmt.Errorf("env.FreqTable: %v Table has no columns -- Outputs will be invalid", ft.Nm)
	}
	_, err := ft.Table.Table.ColByNameTry(ft.FreqCol)
	if err != nil {
		return err
	}
	return nil
}

func (ft *FreqTable) Init(run int) {
	if ft.NameCol == "" {
		ft.NameCol = "Name"
	}
	if ft.GroupCol == "" {
		ft.GroupCol = "Group"
	}
	if ft.FreqCol == "" {
		ft.FreqCol = "Freq"
	}
	ft.Run.Scale = Run
	ft.Epoch.Scale = Epoch
	ft.Trial.Scale = Trial
	ft.Run.Init()
	ft.Epoch.Init()
	ft.Trial.Init()
	ft.Run.Cur = run
	ft.Sample()
	ft.Trial.Max = len(ft.Order)
	ft.Trial.Cur = -1 // init state -- key so that first Step() = 0
}

// Sample generates a new sample of items
func (ft *FreqTable) Sample() {
	if ft.NSamples <= 0 {
		ft.NSamples = 1
	}
	np := ft.Table.Len()
	if ft.Order == nil {
		ft.Order = make([]int, 0, int(math.Round(float64(np)*ft.NSamples)))
	} else {
		ft.Order = ft.Order[:0]
	}
	frqs := ft.Table.Table.ColByName(ft.FreqCol)

	for ri := 0; ri < np; ri++ {
		ti := ft.Table.Idxs[ri]
		frq := frqs.FloatVal1D(ti)
		if ft.RndSamp {
			n := int(ft.NSamples)
			for i := 0; i < n; i++ {
				if erand.BoolProb(frq, -1) {
					ft.Order = append(ft.Order, ri)
				}
			}
		} else {
			n := int(math.Round(ft.NSamples * frq))
			for i := 0; i < n; i++ {
				ft.Order = append(ft.Order, ri)
			}
		}
	}
	if !ft.Sequential {
		erand.PermuteInts(ft.Order)
	}
}

// Row returns the current row number in table, based on Sequential / perumuted Order and
// already de-referenced through the IdxView's indexes to get the actual row in the table.
func (ft *FreqTable) Row() int {
	return ft.Table.Idxs[ft.Order[ft.Trial.Cur]]
}

func (ft *FreqTable) SetTrialName() {
	if nms := ft.Table.Table.ColByName(ft.NameCol); nms != nil {
		rw := ft.Row()
		if rw >= 0 && rw < nms.Len() {
			ft.TrialName = nms.StringVal1D(rw)
		}
	}
}

func (ft *FreqTable) SetGroupName() {
	if nms := ft.Table.Table.ColByName(ft.GroupCol); nms != nil {
		rw := ft.Row()
		if rw >= 0 && rw < nms.Len() {
			ft.GroupName.Set(nms.StringVal1D(rw))
		}
	}
}

func (ft *FreqTable) Step() bool {
	ft.Epoch.Same() // good idea to just reset all non-inner-most counters at start

	if ft.Trial.Incr() { // if true, hit max, reset to 0
		ft.Sample()
		ft.Trial.Max = len(ft.Order)
		ft.Epoch.Incr()
	}
	ft.PrvTrialName = ft.TrialName
	ft.SetTrialName()
	ft.SetGroupName()
	return true
}

func (ft *FreqTable) Counters() []TimeScales {
	return []TimeScales{Run, Epoch, Trial}
}

func (ft *FreqTable) Counter(scale TimeScales) (cur, prv int, chg bool) {
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

func (ft *FreqTable) States() Elements {
	els := Elements{}
	els.FromSchema(ft.Table.Table.Schema())
	return els
}

func (ft *FreqTable) State(element string) etensor.Tensor {
	et, err := ft.Table.Table.CellTensorTry(element, ft.Row())
	if err != nil {
		log.Println(err)
	}
	return et
}

func (ft *FreqTable) Actions() Elements {
	return nil
}

func (ft *FreqTable) Action(element string, input etensor.Tensor) {
	// nop
}

// Compile-time check that implements Env interface
var _ Env = (*FreqTable)(nil)
