// Copyright (c) 2022, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package looper

import "github.com/emer/emergent/envlp"

// this file has Env support functions

// EnvCtr returns the counter corresponding to this loop's scope, nil if not found
func (lp *Loop) EnvCtr() *envlp.Ctr {
	if lp.Stack.Env == nil {
		return nil
	}
	ctr, err := lp.Stack.Env.Counters().ByScopeTry(lp.Scope)
	if err == nil {
		return ctr
	}
	return nil
}

// EnvIncr is the Env:Incr Main function
func (lp *Loop) EnvIncr() {
	if ctr := lp.EnvCtr(); ctr != nil {
		ctr.Incr()
	}
}

// EnvStep is the Env:Step Main function
func (lp *Loop) EnvStep() {
	if lp.Stack.Env == nil {
		return
	}
	lp.Stack.Env.Step()
}

// EnvIsOverMax is the Env:IsOverMax Stop function
func (lp *Loop) EnvIsOverMax() bool {
	if ctr := lp.EnvCtr(); ctr != nil {
		return ctr.IsOverMax()
	}
	return false
}

// EnvResetIfOverMax is the Env:ResetIfOverMax End function
func (lp *Loop) EnvResetIfOverMax() {
	if ctr := lp.EnvCtr(); ctr != nil {
		ctr.ResetIfOverMax()
	}
}

// AddEnvFuncs adds Env: funcs for loop
func (lp *Loop) AddEnvFuncs() {
	ctr := lp.EnvCtr()
	if ctr == nil {
		return
	}
	ord := lp.Stack.Env.Counters().Order
	last := ord[len(ord)-1]
	if last == lp.Scope {
		lp.Main.Add("Env:Step", lp.EnvStep)
	} else {
		lp.Main.Add("Env:Incr", lp.EnvIncr)
	}
	lp.Stop.Add("Env:IsOverMax", lp.EnvIsOverMax)
	lp.End.Add("Env:ResetIfOverMax", lp.EnvResetIfOverMax)
}

// AddEnvFuncs adds Env: funcs for Stack
func (st *Stack) AddEnvFuncs() {
	if st.Env == nil {
		return
	}
	for _, lp := range st.Loops {
		lp.AddEnvFuncs()
	}
}
