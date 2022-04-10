// Copyright (c) 2022, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package looper

import (
	"github.com/emer/emergent/envlp"
	"github.com/emer/emergent/etime"
)

// Stack contains one stack of nested loops,
// associated with one evaluation Mode, and, optionally, envlp.Env.
// If the Env is set, then counters at corresponding Scope levels
// are incremented, checked for stopping, and reset to control looping.
type Stack struct {
	Mode  string                   `desc:"eval mode for this stack"`
	Env   envlp.Env                `desc:"environment used by default for loop iteration, stopping, if set"`
	Order []etime.ScopeKey         `desc:"order of the loops"`
	Loops map[etime.ScopeKey]*Loop `desc:"the loops by scope"`
	Step  Step                     `desc:"stepping state"`
}

func NewStack(mode string, times ...etime.Times) *Stack {
	ord := make([]etime.ScopeKey, len(times))
	for i, t := range times {
		ord[i] = etime.ScopeStr(mode, t.String())
	}
	return NewStackScope(ord...)
}

func NewStackScope(scopes ...etime.ScopeKey) *Stack {
	st := &Stack{}
	st.Order = scopes
	md, _ := st.Order[0].ModesAndTimes()
	st.Mode = md[0]
	st.Loops = make(map[etime.ScopeKey]*Loop, len(st.Order))
	for _, sc := range st.Order {
		st.Loops[sc] = NewLoop(sc)
	}
	return st
}

// Scope returns the top-level scope for this stack
func (st *Stack) Scope() etime.ScopeKey {
	if len(st.Order) > 0 {
		return st.Order[0]
	}
	return etime.ScopeKey("")
}

// Loop returns loop for given time
func (st *Stack) Loop(time etime.Times) *Loop {
	sc := etime.ScopeStr(st.Mode, time.String())
	return st.Loops[sc]
}

// Level returns loop for given level in order
func (st *Stack) Level(lev int) *Loop {
	if lev < 0 || lev >= len(st.Order) {
		return nil
	}
	return st.Loops[st.Order[lev]]
}

// MainRun runs Main functions on loop, and then increments Env counter
// at same Scope level (if Env)
func (st *Stack) MainRun(lp *Loop) {
	lp.Main.Run()
	if st.Env != nil {
		if ctr, ok := st.Env.Counters()[lp.Scope]; ok {
			ctr.Incr()
		}
	}
}

// StopCheck checks if it is time to stop, based on Env counters (if Env),
// and loop Stop functions.
func (st *Stack) StopCheck(lp *Loop) bool {
	if st.Env != nil {
		if ctr, ok := st.Env.Counters()[lp.Scope]; ok {
			if ctr.IsOverMax() {
				return true
			}
		}
	}
	return lp.Stop.Run()
}

// StepCheck checks if it is time to stop based on stepping
func (st *Stack) StepCheck(lp *Loop) bool {
	return st.Step.StopCheck(lp.Scope)
}

// StepIsScope returns true if stepping is happening at scope level of given loop
func (st *Stack) StepIsScope(lp *Loop) bool {
	return st.Step.IsScope(lp.Scope)
}

// EndRun runs End functions on loop, and then resets Env counter
// at same Scope level (if Env)
func (st *Stack) EndRun(lp *Loop) {
	lp.End.Run()
	if st.Env != nil {
		if ctr, ok := st.Env.Counters()[lp.Scope]; ok {
			ctr.ResetIfOverMax()
		}
	}
}

// Run runs the stack of looping functions.  It will stop at any existing
// Step settings -- call StepClear to clear those.
func (st *Stack) Run(set *Set) {
	lev := 0
	lp := st.Level(lev)
	stepStopNext := false
	for {
		lev++
		nlp := st.Level(lev)
		if nlp != nil {
			if stepStopNext && st.StepIsScope(nlp) {
				stepStopNext = false
				break
			}
			lp = nlp
			continue
		}
		lev--
	main:
		st.MainRun(lp)
		stop := st.StopCheck(lp)
		if stop {
			if st.StepCheck(lp) {
				stepStopNext = true // can't stop now, do it next time..
			}
			st.EndRun(lp)
			lev--
			nlp = st.Level(lev)
			if nlp == nil {
				break
			}
			lp = nlp
			goto main
		} else {
			if st.StepCheck(lp) {
				break
			}
			if set.StopFlag {
				break
			}
		}
	}
}

// SetStep sets the stepping scope and n -- 0 = no stepping
// resets counter.
func (st *Stack) SetStep(time etime.Times, n int) {
	sc := etime.ScopeStr(st.Mode, time.String())
	st.SetStepScope(sc, n)
}

// SetStepScope sets the stepping scope and n -- 0 = no stepping
// resets counter.
func (st *Stack) SetStepScope(scope etime.ScopeKey, n int) {
	st.Step.Set(scope, n)
}

// StepClear resets stepping
func (st *Stack) StepClear() {
	st.Step.Clear()
}
