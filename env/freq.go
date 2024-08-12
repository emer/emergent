// Copyright (c) 2019, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package env

import (
	"fmt"
	"log"
	"math"

	"cogentcore.org/core/base/randx"
	"cogentcore.org/core/tensor"
	"cogentcore.org/core/tensor/table"
	"github.com/emer/emergent/v2/etime"
)

// FreqTable is an Env that manages patterns from an table.Table with frequency
// information so that items are presented according to their associated frequencies
// which are effectively probabilities of presenting any given input -- must have
// a Freq column with these numbers in the table (actual col name in FreqCol).
// Either sequential or permuted random ordering is supported, with std Trial / Epoch
// TimeScale counters to record progress and iterations through the table.
// It also records the outer loop of Run as provided by the model.
// It uses an IndexView indexed view of the Table, so a single shared table
// can be used across different environments, with each having its own unique view.
type FreqTable struct {

	// name of this environment
	Name string

	// this is an indexed view of the table with the set of patterns to output -- the indexes are used for the *sequential* view so you can easily sort / split / filter the patterns to be presented using this view -- we then add the random permuted Order on top of those if !sequential
	Table *table.IndexView

	// number of samples to use in constructing the list of items to present according to frequency -- number per epoch ~ NSamples * Freq -- see RandSamp option
	NSamples float64

	// if true, use random sampling of items NSamples times according to given Freq probability value -- otherwise just directly add NSamples * Freq items to the list
	RandSamp bool

	// present items from the table in sequential order (i.e., according to the indexed view on the Table)?  otherwise permuted random order.  All repetitions of given item will be sequential if Sequential
	Sequential bool

	// list of items to present, with repetitions -- updated every time through the list
	Order []int

	// current ordinal item in Table -- if Sequential then = row number in table, otherwise is index in Order list that then gives row number in Table
	Trial Counter `display:"inline"`

	// if Table has a Name column, this is the contents of that
	TrialName CurPrvString

	// if Table has a Group column, this is contents of that
	GroupName CurPrvString

	// name of the Name column -- defaults to 'Name'
	NameCol string

	// name of the Group column -- defaults to 'Group'
	GroupCol string

	// name of the Freq column -- defaults to 'Freq'
	FreqCol string
}

func (ft *FreqTable) Validate() error {
	if ft.Table == nil || ft.Table.Table == nil {
		return fmt.Errorf("env.FreqTable: %v has no Table set", ft.Name)
	}
	if ft.Table.Table.NumColumns() == 0 {
		return fmt.Errorf("env.FreqTable: %v Table has no columns -- Outputs will be invalid", ft.Name)
	}
	_, err := ft.Table.Table.ColumnByNameTry(ft.FreqCol)
	if err != nil {
		return err
	}
	return nil
}

func (ft *FreqTable) Label() string { return ft.Name }

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
	ft.Trial.Scale = etime.Trial
	ft.Trial.Init()
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
	frqs := ft.Table.Table.ColumnByName(ft.FreqCol)

	for ri := 0; ri < np; ri++ {
		ti := ft.Table.Indexes[ri]
		frq := frqs.Float1D(ti)
		if ft.RandSamp {
			n := int(ft.NSamples)
			for i := 0; i < n; i++ {
				if randx.BoolP(frq) {
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
		randx.PermuteInts(ft.Order)
	}
}

// Row returns the current row number in table, based on Sequential / perumuted Order and
// already de-referenced through the IndexView's indexes to get the actual row in the table.
func (ft *FreqTable) Row() int {
	return ft.Table.Indexes[ft.Order[ft.Trial.Cur]]
}

func (ft *FreqTable) SetTrialName() {
	if nms := ft.Table.Table.ColumnByName(ft.NameCol); nms != nil {
		rw := ft.Row()
		if rw >= 0 && rw < nms.Len() {
			ft.TrialName.Set(nms.String1D(rw))
		}
	}
}

func (ft *FreqTable) SetGroupName() {
	if nms := ft.Table.Table.ColumnByName(ft.GroupCol); nms != nil {
		rw := ft.Row()
		if rw >= 0 && rw < nms.Len() {
			ft.GroupName.Set(nms.String1D(rw))
		}
	}
}

func (ft *FreqTable) Step() bool {
	if ft.Trial.Incr() { // if true, hit max, reset to 0
		ft.Sample()
		ft.Trial.Max = len(ft.Order)
	}
	ft.SetTrialName()
	ft.SetGroupName()
	return true
}

func (ft *FreqTable) State(element string) tensor.Tensor {
	et := ft.Table.Table.Tensor(element, ft.Row())
	if et == nil {
		log.Println("FreqTable.State -- could not find element:", element)
	}
	return et
}

func (ft *FreqTable) Action(element string, input tensor.Tensor) {
	// nop
}

// Compile-time check that implements Env interface
var _ Env = (*FreqTable)(nil)
