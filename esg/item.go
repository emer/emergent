// Copyright (c) 2019, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package esg

import (
	"fmt"
	"strings"

	"github.com/goki/ki/kit"
)

// Item is one item within a rule
type Item struct {
	Prob    float32 `desc:"probability for choosing this item -- 0 if uniform random"`
	Elems   []Elem  `desc:"elements of the rule -- for non-Cond rules"`
	Cond    Conds   `desc:"conditions for this item -- specified by ?"`
	SubRule *Rule   `desc:"for conditional, this is the sub-rule that is run with sub-items"`
	State   State   `desc:"state update name=value to set for rule"`
}

// String returns string rep
func (it *Item) String() string {
	if it.Cond != nil {
		return it.Cond.String() + it.SubRule.String()
	}
	sout := ""
	if it.Prob > 0 {
		sout = "%" + fmt.Sprintf("%g ", it.Prob)
	}
	for i := range it.Elems {
		el := &it.Elems[i]
		sout += el.String() + " "
	}
	return sout
}

// Gen generates expression according to the item
func (it *Item) Gen(rl *Rule, rls *Rules) {
	if it.SubRule != nil {
		it.State.Set(rls, "") // no value
		it.SubRule.Gen(rls)
	}
	if len(it.Elems) > 0 {
		it.State.Set(rls, it.Elems[0].Value)
		for i := range it.Elems {
			el := &it.Elems[i]
			el.Gen(rl, rls)
		}
	}
}

// CondTrue evalutes whether the condition is true
func (it *Item) CondEval(rl *Rule, rls *Rules) bool {
	return it.Cond.Eval(rls)
}

// Validate checks for config errors
func (it *Item) Validate(rl *Rule, rls *Rules) []error {
	if it.Cond != nil {
		ers := it.Cond.Validate(rl, it, rls)
		if it.SubRule == nil {
			ers = append(ers, fmt.Errorf("Rule: %v Item: %v IsCond but SubRule == nil", rl.Name, it.String()))
		} else {
			srs := it.SubRule.Validate(rls)
			if len(srs) > 0 {
				ers = append(ers, srs...)
			}
		}
		return ers
	}
	var errs []error
	for i := range it.Elems {
		el := &it.Elems[i]
		ers := el.Validate(it, rl, rls)
		if len(ers) > 0 {
			errs = append(errs, ers...)
		}
	}
	return errs
}

/////////////////////////////////////////////////////////////////////
// Elem

// Elem is one elemenent in a concrete Item: either rule or token
type Elem struct {
	El    Elements `desc:"type of element: Rule, Token, or SubItems"`
	Value string   `desc:"value of the token: name of Rule or Token"`
}

// String returns string rep
func (el *Elem) String() string {
	if el.El == TokenEl {
		return "'" + el.Value + "'"
	}
	return el.Value
}

// Gen generates expression according to the element
func (el *Elem) Gen(rl *Rule, rls *Rules) {
	switch el.El {
	case RuleEl:
		rl := rls.Rule(el.Value)
		rl.Gen(rls)
	case TokenEl:
		if rls.Trace {
			fmt.Printf("Rule: %v added Token output: %v\n", rl.Name, el.Value)
		}
		rls.AddOutput(el.Value)
	}
}

// Validate checks for config errors
func (el *Elem) Validate(it *Item, rl *Rule, rls *Rules) []error {
	switch el.El {
	case RuleEl:
		_, err := rls.RuleTry(el.Value)
		if err != nil {
			return []error{err}
		}
		return nil
	case TokenEl:
		if el.Value == "" {
			err := fmt.Errorf("Rule: %v Item: %v has empty Token element", rl.Name, it.String())
			return []error{err}
		}
	}
	return nil
}

// Elements are different types of elements
type Elements int32

//go:generate stringer -type=Elements

var KiT_Elements = kit.Enums.AddEnum(ElementsN, kit.NotBitFlag, nil)

func (ev Elements) MarshalJSON() ([]byte, error)  { return kit.EnumMarshalJSON(ev) }
func (ev *Elements) UnmarshalJSON(b []byte) error { return kit.EnumUnmarshalJSON(ev, b) }

const (
	// RuleEl means Value is name of a rule
	RuleEl Elements = iota

	// TokenEl means Value is a token to emit
	TokenEl

	ElementsN
)

/////////////////////////////////////////////////////////////////////
// State

// State holds the name=value state settings associated with rule or item
// as a string, string map
type State map[string]string

// Add adds give name, value to state
func (ss *State) Add(name, val string) {
	if *ss == nil {
		*ss = make(map[string]string)
	}
	(*ss)[name] = val
}

// Set sets state in rules States map, using given value for any items that have empty values
func (ss *State) Set(rls *Rules, val string) bool {
	if len(*ss) == 0 {
		return false
	}
	for k, v := range *ss {
		if v == "" {
			v = val
		}
		rls.States[k] = v
		if rls.Trace {
			fmt.Printf("Set State: %v = %v\n", k, v)
		}
	}
	return true
}

// TrimQualifiers removes any :X qualifiers after state values
func (ss *State) TrimQualifiers() {
	for k, v := range *ss {
		ci := strings.Index(v, ":")
		if ci > 0 {
			(*ss)[k] = v[:ci]
		}
	}
}
