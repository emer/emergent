// Copyright (c) 2019, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package esg

import (
	"fmt"
)

// Rules is a collection of rules
type Rules struct {
	Name      string           `desc:"name of this rule collection"`
	Desc      string           `desc:"description of this rule collection"`
	Trace     bool             `desc:"if true, will print out a trace during generation"`
	Top       *Rule            `desc:"top-level rule -- this is where to start generating"`
	Map       map[string]*Rule `desc:"map of each rule"`
	Fired     map[string]bool  `desc:"map of names of all the rules that have fired"`
	Output    []string         `desc:"array of output strings -- appended as the rules generate output"`
	States    State            `desc:"user-defined state map optionally created during generation"`
	ParseErrs []error          `desc:"errors from parsing"`
	ParseLn   int              `desc:"current line number during parsing"`
}

// Gen generates one expression according to the rules.
// returns the token output, which is also avail as rls.Output
func (rls *Rules) Gen() []string {
	rls.Fired = make(map[string]bool)
	rls.States = make(State)
	rls.Output = nil
	if rls.Trace {
		fmt.Printf("\n#########################\nRules: %v starting Gen\n", rls.Name)
	}
	rls.Top.Gen(rls)
	return rls.Output
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

// Init initializes rule order state
func (rls *Rules) Init() {
	rls.Top.Init()
	for _, rl := range rls.Map {
		rl.Init()
	}
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

// HasOutput returns true if given token is in the output string
// strips ' ' delimiters if present in out string
func (rls *Rules) HasOutput(out string) bool {
	if out[0] == '\'' {
		out = out[1 : len(out)-1]
	}
	for _, o := range rls.Output {
		if o == out {
			return true
		}
	}
	return false
}

// SetFired adds given rule name to map of those that fired this round
func (rls *Rules) SetFired(name string) {
	rls.Fired[name] = true
}

// AddOutput adds given string to Output array
func (rls *Rules) AddOutput(out string) {
	rls.Output = append(rls.Output, out)
}

// Adds given rule
func (rls *Rules) Add(rl *Rule) {
	if rls.Map == nil {
		rls.Map = make(map[string]*Rule)
		rls.Top = rl
	}
	rls.Map[rl.Name] = rl
}
