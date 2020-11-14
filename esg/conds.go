// Copyright (c) 2019, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package esg

import (
	"fmt"

	"github.com/goki/ki/kit"
)

// Conds are conditionals
type Conds []*Cond

// String returns string rep
func (cs *Conds) String() string {
	str := ""
	for ci := range *cs {
		cd := (*cs)[ci]
		str += cd.String() + " "
	}
	return str
}

// True returns true if conditional expression is true
func (cs *Conds) Eval(rls *Rules) bool {
	cval := true
	hasCval := false
	lastBin := CondElsN // binary op
	lastNot := false
	for ci := range *cs {
		cd := (*cs)[ci]
		switch cd.El {
		case And, Or:
			lastBin = cd.El
		case Not:
			lastNot = true
		default:
			tst := cd.Eval(rls)
			if lastNot {
				tst = !tst
				lastNot = false
			}
			if !hasCval {
				cval = tst
				hasCval = true
				continue
			}
			hasCval = true
			switch lastBin {
			case And:
				cval = cval && tst
			case Or:
				cval = cval || tst
			}
			lastBin = CondElsN
		}
	}
	return cval
}

// Validate checks for errors in expression
func (cs *Conds) Validate(rl *Rule, it *Item, rls *Rules) []error {
	lastBin := CondElsN // binary op
	lastNot := false
	var errs []error
	ncd := len(*cs)
	for ci := range *cs {
		cd := (*cs)[ci]
		switch cd.El {
		case And, Or:
			if lastBin != CondElsN {
				errs = append(errs, fmt.Errorf("Rule: %v Item: %v Condition has two binary logical operators in a row", rl.Name, it.String()))
			}
			if ci == 0 || ci == ncd-1 {
				errs = append(errs, fmt.Errorf("Rule: %v Item: %v Condition has binary logical operator at start or end", rl.Name, it.String()))
			}
			lastBin = cd.El
		case Not:
			if lastNot {
				errs = append(errs, fmt.Errorf("Rule: %v Item: %v Condition has two Not operators in a row", rl.Name, it.String()))
			}
			if ci == 0 || ci == ncd-1 {
				errs = append(errs, fmt.Errorf("Rule: %v Item: %v Condition has Not operator at start or end", rl.Name, it.String()))
			}
			lastNot = true
		default:
			elers := cd.Validate(rl, it, rls)
			if elers != nil {
				errs = append(errs, elers...)
			}
			lastNot = false
			lastBin = CondElsN
		}
	}
	return errs
}

/////////////////////////////////////////////////////////////////////////
// Cond

// Cond is one element of a conditional
type Cond struct {
	El    CondEls `desc:"what type of conditional element is this"`
	Rule  string  `desc:"name of rule or token to evaluate for CRule"`
	Conds Conds   `desc:"sub-conditions for SubCond"`
}

// String returns string rep
func (cd *Cond) String() string {
	switch cd.El {
	case And:
		return "&&"
	case Or:
		return "||"
	case Not:
		return "!"
	case CRule:
		return cd.Rule
	case SubCond:
		return "(" + cd.Conds.String() + ")"
	}
	return ""
}

// True returns true if conditional expression is true
func (cd *Cond) Eval(rls *Rules) bool {
	if cd.El == CRule {
		if cd.Rule[0] == '\'' {
			return rls.HasOutput(cd.Rule)
		} else {
			return rls.HasFired(cd.Rule)
		}
	}
	if cd.El == SubCond && cd.Conds != nil {
		return cd.Conds.Eval(rls)
	}
	return false
}

// Validate checks for errors in expression
func (cd *Cond) Validate(rl *Rule, it *Item, rls *Rules) []error {
	if cd.El == CRule {
		if cd.Rule == "" {
			return []error{fmt.Errorf("Rule: %v Item: %v Rule Condition has empty Rule", rl.Name, it.String())}
		}
		if cd.Rule[0] != '\'' {
			_, err := rls.RuleTry(cd.Rule)
			if err != nil {
				return []error{err}
			}
		}
		return nil
	}
	if cd.El == SubCond {
		if len(cd.Conds) == 0 {
			return []error{fmt.Errorf("Rule: %v Item: %v Rule SubConds are empty", rl.Name, it.String())}
		}
		return cd.Conds.Validate(rl, it, rls)
	}
	return nil
}

// CondEls are different types of conditional elements
type CondEls int32

//go:generate stringer -type=CondEls

var KiT_CondEls = kit.Enums.AddEnum(CondElsN, kit.NotBitFlag, nil)

func (ev CondEls) MarshalJSON() ([]byte, error)  { return kit.EnumMarshalJSON(ev) }
func (ev *CondEls) UnmarshalJSON(b []byte) error { return kit.EnumUnmarshalJSON(ev, b) }

const (
	// CRule means Rule is name of a rule to evaluate truth value
	CRule CondEls = iota
	And
	Or
	Not

	// SubCond is a sub-condition expression
	SubCond

	CondElsN
)
