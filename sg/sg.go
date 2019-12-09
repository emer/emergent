// Copyright (c) 2019, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package sg

import (
	"fmt"
	"math/rand"
	"strings"

	"github.com/goki/ki/kit"
)

// Rules is a collection of rules
type Rules struct {
	Name  string              `desc:"name of this rule collection"`
	Desc  string              `desc:"description of this rule collection"`
	Trace bool                `desc:"if true, will print out a trace during generation"`
	Top   *Rule               `desc:"top-level rule -- this is where to start generating"`
	Map   map[string]*Rule    `desc:"map of each rule"`
	Fired map[string]struct{} `desc:"map of names of all the rules that have fired"`
}

// Gen generates one expression according to the rules
func (rls *Rules) Gen() []string {
	rls.Fired = make(map[string]struct{}, 100)
	if rls.Trace {
		fmt.Printf("\n#########################\nRules: %v starting Gen\n", rls.Name)
	}
	return rls.Top.Gen(rls)
}

// String generates string representation of all rules
func (rls *Rules) String() string {
	str := "Rules: " + rls.Name
	if rls.Desc != "" {
		str += ": " + rls.Desc
	}
	str += "\n"
	for _, rl := range rls.Map {
		str += rl.String()
	}
	return str
}

// Validate checks for config errors
func (rls *Rules) Validate() []error {
	if len(rls.Map) == 0 {
		return []error{fmt.Errorf("Rules: %v has no Rules", rls.Name)}
	}
	var errs []error
	if rls.Top == nil {
		errs = append(errs, fmt.Errorf("Rules: %v Top is nil", rls.Name))
	}
	for _, rl := range rls.Map {
		ers := rl.Validate(rls)
		if len(ers) > 0 {
			errs = append(errs, ers...)
		}
	}
	if len(errs) > 0 {
		fmt.Printf("\nValidation errors:\n")
		for _, err := range errs {
			fmt.Println(err)
		}
	}
	return errs
}

// Rule returns rule of given name (nil if not found)
func (rls *Rules) Rule(name string) *Rule {
	return rls.Map[name]
}

// RuleTry returns rule of given name, and error if not found
func (rls *Rules) RuleTry(name string) (*Rule, error) {
	rl, ok := rls.Map[name]
	if !ok {
		return nil, fmt.Errorf("Rule: %v not found in Rules: %v", name, rls.Name)
	}
	return rl, nil
}

// HasFired returns true if rule of given name has fired
func (rls *Rules) HasFired(name string) bool {
	_, has := rls.Fired[name]
	return has
}

// SetFired adds given rule name to map of those that fired this round
func (rls *Rules) SetFired(name string) {
	rls.Fired[name] = struct{}{}
}

// Adds given rule
func (rls *Rules) Add(rl *Rule) {
	if rls.Map == nil {
		rls.Map = make(map[string]*Rule)
		rls.Top = rl
	}
	rls.Map[rl.Name] = rl
}

/////////////////////////////////////////////////////////////////////
// Rule

// Rule is one rule
type Rule struct {
	Name     string  `desc:"name of rule"`
	Desc     string  `desc:"description / notes on rule"`
	IsConds  bool    `desc:"items are conditionals -- choose first that fits"`
	HasProbs bool    `desc:"items have probabilities (else uninform random)"`
	Items    []*Item `desc:"items in rule"`
}

// Gen generates expression according to the rule
func (rl *Rule) Gen(rls *Rules) []string {
	rls.SetFired(rl.Name)
	if rls.Trace {
		fmt.Printf("Fired Rule: %v\n", rl.Name)
	}
	if rl.IsConds {
		var copts []int
		for ii, it := range rl.Items {
			if it.CondTrue(rl, rls) {
				copts = append(copts, ii)
			}
		}
		no := len(copts)
		if no == 0 {
			if rls.Trace {
				fmt.Printf("No items match Conds\n")
			}
			return nil
		}
		opt := rand.Intn(no)
		if rls.Trace {
			fmt.Printf("Selected item: %v from: %v matching Conds\n", copts[opt], no)
		}
		return rl.Items[copts[opt]].Gen(rl, rls)
	}
	if rl.HasProbs {
		pv := rand.Float32()
		sum := float32(0)
		for ii, it := range rl.Items {
			sum += it.Prob
			if pv < sum { // note: lower values already excluded
				if rls.Trace {
					fmt.Printf("Selected item: %v using rnd val: %v sum: %v\n", ii, pv, sum)
				}
				return it.Gen(rl, rls)
			}
		}
		if rls.Trace {
			fmt.Printf("No items selected using rnd val: %v sum: %v\n", pv, sum)
		}
		return nil
	} else {
		no := len(rl.Items)
		opt := rand.Intn(no)
		if rls.Trace {
			fmt.Printf("Selected item: %v from: %v uniform random\n", opt, no)
		}
		return rl.Items[opt].Gen(rl, rls)
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
		if rl.IsConds {
			str += " ? "
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
		if rl.IsConds {
			if len(it.Cond) == 0 {
				errs = append(errs, fmt.Errorf("Rule: %v IsConds, but Item: %v has no Cond", rl.Name, it.String()))
			}
			if it.SubRule == nil {
				errs = append(errs, fmt.Errorf("Rule: %v IsConds, but Item: %v has nil SubRule", rl.Name, it.String()))
			}
		} else {
			if rl.HasProbs && it.Prob == 0 {
				errs = append(errs, fmt.Errorf("Rule: %v HasProbs, but Item: %v has 0 Prob", rl.Name, it.String()))
			} else if !rl.HasProbs && it.Prob > 0 {
				errs = append(errs, fmt.Errorf("Rule: %v !HasProbs, but Item: %v has > 0 Prob", rl.Name, it.String()))
			}
		}
		iterrs := it.Validate(rl, rls)
		if len(iterrs) > 0 {
			errs = append(errs, iterrs...)
		}
	}
	return errs
}

/////////////////////////////////////////////////////////////////////
// Item

// Item is one item within a rule
type Item struct {
	Prob    float32 `desc:"probability for choosing this item -- 0 if uniform random"`
	Elems   []Elem  `desc:"elements of the rule -- for non-Cond rules"`
	Cond    Conds   `desc:"conditions for this item -- specified by ?"`
	SubRule *Rule   `desc:"for conditional, this is the sub-rule that is run with sub-items"`
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
func (it *Item) Gen(rl *Rule, rls *Rules) []string {
	if it.SubRule != nil {
		return it.SubRule.Gen(rls)
	}
	var gout []string
	for i := range it.Elems {
		el := &it.Elems[i]
		ov := el.Gen(rls)
		if len(ov) > 0 {
			gout = append(gout, ov...)
		}
	}
	if rls.Trace {
		fmt.Printf("Item generated tokens: %v\n", gout)
	}
	return gout
}

// CondTrue evalutes whether the condition is true
func (it *Item) CondTrue(rl *Rule, rls *Rules) bool {
	return it.Cond.True(rls)
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
func (el *Elem) Gen(rls *Rules) []string {
	switch el.El {
	case RuleEl:
		rl := rls.Rule(el.Value)
		return rl.Gen(rls)
	case TokenEl:
		return []string{el.Value}
	}
	return nil
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

var KiT_Elements = kit.Enums.AddEnum(ElementsN, false, nil)

func (ev Elements) MarshalJSON() ([]byte, error)  { return kit.EnumMarshalJSON(ev) }
func (ev *Elements) UnmarshalJSON(b []byte) error { return kit.EnumUnmarshalJSON(ev, b) }

const (
	// RuleEl means Value is name of a rule
	RuleEl Elements = iota

	// TokenEl means Value is a token to emit
	TokenEl

	ElementsN
)
