// Copyright (c) 2019, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package esg

//go:generate core generate -add-types

import (
	"fmt"
	"math/rand"
	"strings"

	"cogentcore.org/lab/base/randx"
)

// RuleTypes are different types of rules (i.e., how the items are selected)
type RuleTypes int32 //enums:enum

const (
	// UniformItems is the default mutually exclusive items chosen at uniform random
	UniformItems RuleTypes = iota

	// ProbItems has specific probabilities for each item
	ProbItems

	// CondItems has conditionals for each item, indicated by ?
	CondItems

	// SequentialItems progresses through items in sequential order, indicated by |
	SequentialItems

	// PermutedItems progresses through items in permuted order, indicated by $
	PermutedItems
)

// Rule is one rule containing some number of items
type Rule struct { //git:add

	// name of rule
	Name string

	// description / notes on rule
	Desc string

	// type of rule -- how to choose the items
	Type RuleTypes

	// items in rule
	Items []*Item

	// state update for rule
	State State

	// previously selected item (from perspective of current rule)
	PrevIndex int

	// current index in Items (what will be used next)
	CurIndex int

	// probability of repeating same item -- signaled by =%p
	RepeatP float32

	// permuted order if doing that
	Order []int
}

// Init initializes the rules -- only relevant for ordered rules (restarts at start)
func (rl *Rule) Init() {
	rl.CurIndex = 0
	rl.PrevIndex = -1
	if rl.Type == PermutedItems {
		rl.Order = rand.Perm(len(rl.Items))
	}
}

// Gen generates expression according to the rule, appending output
// to the rls.Output array
func (rl *Rule) Gen(rls *Rules) {
	rls.SetFired(rl.Name)
	rl.State.Set(rls, rl.Name)
	if rls.Trace {
		fmt.Printf("Fired Rule: %v\n", rl.Name)
	}
	if rl.RepeatP > 0 && rl.PrevIndex >= 0 {
		rpt := randx.BoolP32(rl.RepeatP)
		if rpt {
			if rls.Trace {
				fmt.Printf("Selected item: %v due to RepeatP = %v\n", rl.PrevIndex, rl.RepeatP)
			}
			rl.Items[rl.PrevIndex].Gen(rl, rls)
			return
		}
	}
	switch rl.Type {
	case UniformItems:
		no := len(rl.Items)
		opt := rand.Intn(no)
		if rls.Trace {
			fmt.Printf("Selected item: %v from: %v uniform random\n", opt, no)
		}
		rl.PrevIndex = opt
		rl.Items[opt].Gen(rl, rls)
	case ProbItems:
		pv := rand.Float32()
		sum := float32(0)
		for ii, it := range rl.Items {
			sum += it.Prob
			if pv < sum { // note: lower values already excluded
				if rls.Trace {
					fmt.Printf("Selected item: %v using rnd val: %v sum: %v\n", ii, pv, sum)
				}
				rl.PrevIndex = ii
				it.Gen(rl, rls)
				return
			}
		}
		rl.PrevIndex = -1
		if rls.Trace {
			fmt.Printf("No items selected using rnd val: %v sum: %v\n", pv, sum)
		}
	case CondItems:
		var copts []int
		for ii, it := range rl.Items {
			if it.CondEval(rl, rls) {
				copts = append(copts, ii)
			}
		}
		no := len(copts)
		if no == 0 {
			if rls.Trace {
				fmt.Printf("No items match Conds\n")
			}
			return
		}
		opt := rand.Intn(no)
		if rls.Trace {
			fmt.Printf("Selected item: %v from: %v matching Conds\n", copts[opt], no)
		}
		rl.PrevIndex = copts[opt]
		rl.Items[copts[opt]].Gen(rl, rls)
	case SequentialItems:
		no := len(rl.Items)
		if no == 0 {
			return
		}
		if rl.CurIndex >= no {
			rl.CurIndex = 0
		}
		opt := rl.CurIndex
		if rls.Trace {
			fmt.Printf("Selected item: %v sequentially\n", opt)
		}
		rl.PrevIndex = opt
		rl.CurIndex++
		rl.Items[opt].Gen(rl, rls)
	case PermutedItems:
		no := len(rl.Items)
		if no == 0 {
			return
		}
		if len(rl.Order) != no {
			rl.Order = rand.Perm(no)
			rl.CurIndex = 0
		}
		if rl.CurIndex >= no {
			randx.PermuteInts(rl.Order)
			rl.CurIndex = 0
		}
		opt := rl.Order[rl.CurIndex]
		if rls.Trace {
			fmt.Printf("Selected item: %v sequentially\n", opt)
		}
		rl.PrevIndex = opt
		rl.CurIndex++
		rl.Items[opt].Gen(rl, rls)
	}
}

// String generates string representation of rule
func (rl *Rule) String() string {
	if strings.HasSuffix(rl.Name, "SubRule") {
		str := " {\n"
		for _, it := range rl.Items {
			str += "\t\t" + it.String() + "\n"
		}
		str += "\t}\n"
		return str
	} else {
		str := "\n\n"
		if rl.Desc != "" {
			str += "// " + rl.Desc + "\n"
		}
		str += rl.Name
		switch rl.Type {
		case CondItems:
			str += " ? "
		case SequentialItems:
			str += " | "
		case PermutedItems:
			str += " $ "
		}
		str += " {\n"
		for _, it := range rl.Items {
			str += "\t" + it.String() + "\n"
		}
		str += "}\n"
		return str
	}
}

// Validate checks for config errors
func (rl *Rule) Validate(rls *Rules) []error {
	nr := len(rl.Items)
	if nr == 0 {
		err := fmt.Errorf("Rule: %v has no items", rl.Name)
		return []error{err}
	}
	var errs []error
	for _, it := range rl.Items {
		if rl.Type == CondItems {
			if len(it.Cond) == 0 {
				errs = append(errs, fmt.Errorf("Rule: %v is CondItems, but Item: %v has no Cond", rl.Name, it.String()))
			}
			if it.SubRule == nil {
				errs = append(errs, fmt.Errorf("Rule: %v is CondItems, but Item: %v has nil SubRule", rl.Name, it.String()))
			}
		} else {
			if rl.Type == ProbItems && it.Prob == 0 {
				errs = append(errs, fmt.Errorf("Rule: %v is ProbItems, but Item: %v has 0 Prob", rl.Name, it.String()))
			} else if rl.Type == UniformItems && it.Prob > 0 {
				errs = append(errs, fmt.Errorf("Rule: %v is UniformItems, but Item: %v has > 0 Prob", rl.Name, it.String()))
			}
		}
		iterrs := it.Validate(rl, rls)
		if len(iterrs) > 0 {
			errs = append(errs, iterrs...)
		}
	}
	return errs
}
