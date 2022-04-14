// Copyright (c) 2022, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package looper

import (
	"fmt"
	"strings"

	"github.com/emer/emergent/envlp"
	"github.com/emer/emergent/etime"
	"github.com/goki/ki/indent"
)

// IndentSize is number of spaces to indent for output
var IndentSize = 4

// Stack contains one stack of nested loops,
// associated with one evaluation Mode, and, optionally, envlp.Env.
// If the Env is set, then counters at corresponding Scope levels
// are incremented, checked for stopping, and reset to control looping.
type Stack struct {
	Mode  string                   `desc:"eval mode for this stack"`
	Env   envlp.Env                `desc:"environment used by default for loop iteration, stopping, if set"`
	Order []etime.ScopeKey         `desc:"ordered list of the loops, from outer-most (highest) to inner-most (lowest)"`
	Loops map[etime.ScopeKey]*Loop `desc:"the loops by scope"`
	Ctxt  map[string]interface{}   `desc:"named context data that can hold state relevant for this stack (e.g., Time struct that holds counters for algorithm inner loops)"`
	Step  Step                     `desc:"stepping state"`
	Set   *Set                     `desc:"Set of Stacks that we belong to"`
}

// NewStack returns new stack for given mode and times
func NewStack(mode string, times ...etime.Times) *Stack {
	ord := make([]etime.ScopeKey, len(times))
	for i, t := range times {
		ord[i] = etime.ScopeStr(mode, t.String())
	}
	return NewStackScope(ord...)
}

// NewStack returns new stack for given list of scopes
func NewStackScope(scopes ...etime.ScopeKey) *Stack {
	st := &Stack{}
	st.Order = etime.CloneScopeSlice(scopes)
	md, _ := st.Order[0].ModeAndTimeStr()
	st.Mode = md
	st.Loops = make(map[etime.ScopeKey]*Loop, len(st.Order))
	st.Ctxt = make(map[string]interface{})
	for _, sc := range st.Order {
		st.Loops[sc] = NewLoop(sc, st)
	}
	return st
}

// NewStackEnv returns new stack with loops matching those in
// the given environment. Adds standard Env counter funcs to
// manage updating of counters, with Step at the lowest level.
func NewStackEnv(ev envlp.Env) *Stack {
	ctrs := ev.Counters()
	st := NewStackScope(ctrs.Order...)
	st.Env = ev
	st.AddEnvFuncs()
	return st
}

// AddLevels adds given levels to stack.
// For algorithms to add mechanism inner loops.
func (st *Stack) AddLevels(times ...etime.Times) {
	for _, tm := range times {
		sc := etime.ScopeStr(st.Mode, tm.String())
		st.Order = append(st.Order, sc)
		st.Loops[sc] = NewLoop(sc, st)
	}
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

// MainRun runs Main functions on loop
func (st *Stack) MainRun(lp *Loop, level int) {
	if st.Step.LoopTrace || st.Step.FuncTrace {
		fmt.Println(lp.StageString("Main", level))
	}
	if st.Step.FuncTrace {
		lp.Main.RunTrace(level + 1)
	} else {
		lp.Main.Run()
	}
}

// StopCheck checks if it is time to stop, based on loop Stop functions.
func (st *Stack) StopCheck(lp *Loop, level int) bool {
	if st.Step.FuncTrace {
		fmt.Println(lp.StageString("Stop", level))
		return lp.Stop.RunTrace(level + 1)
	}
	stp := lp.Stop.Run()
	if stp && st.Step.LoopTrace {
		fmt.Println(lp.StageString("Stop", level))
	}
	return stp
}

// StepCheck checks if it is time to stop based on stepping
func (st *Stack) StepCheck(lp *Loop, level int) bool {
	stp := st.Step.StopCheck(lp.Scope)
	if stp && st.Step.LoopTrace {
		fmt.Println(lp.StageString("Step", level))
	}
	return stp
}

// StepIsScope returns true if stepping is happening at scope level of given loop
func (st *Stack) StepIsScope(lp *Loop) bool {
	return st.Step.IsScope(lp.Scope)
}

// EndRun runs End functions on loop, and then resets Env counter
// at same Scope level (if Env)
func (st *Stack) EndRun(lp *Loop, level int) {
	if st.Step.LoopTrace || st.Step.FuncTrace {
		fmt.Println(lp.StageString("End", level))
	}
	if st.Step.FuncTrace {
		lp.End.RunTrace(level + 1)
	} else {
		lp.End.Run()
	}
}

// Init runs End functions for all levels in the Stack,
// to reset state for a fresh Run.
func (st *Stack) Init() {
	for i, sc := range st.Order {
		lp := st.Loops[sc]
		if st.Step.LoopTrace || st.Step.FuncTrace {
			fmt.Println(lp.StageString("Init", i))
		}
		if st.Step.FuncTrace {
			lp.End.RunTrace(i + 1)
		} else {
			lp.End.Run()
		}
	}
}

// Run runs the stack of looping functions.  It will stop at any existing
// Step settings -- call StepClear to clear those.
func (st *Stack) Run() {
	st.Set.StopFlag = false
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
		st.MainRun(lp, lev)
		stop := st.StopCheck(lp, lev)
		if stop {
			if st.StepCheck(lp, lev) {
				stepStopNext = true // can't stop now, do it next time..
			}
			st.EndRun(lp, lev)
			lev--
			nlp = st.Level(lev)
			if nlp == nil {
				break
			}
			lp = nlp
			goto main
		} else {
			if st.StepCheck(lp, lev) {
				break
			}
			if st.Set.StopFlag {
				break
			}
		}
	}
}

// SetStep sets the stepping scope and n -- 0 = no stepping
// resets counter.
func (st *Stack) SetStep(time etime.Times, n int) {
	st.SetStepTime(time.String(), n)
}

// SetStepTime sets the stepping time and n -- 0 = no stepping
// resets counter.
func (st *Stack) SetStepTime(time string, n int) {
	st.Step.Set(time, n)
}

// StepClear resets stepping
func (st *Stack) StepClear() {
	st.Step.Clear()
}

// DocString returns an indented summary of the loops
// and functions in the stack
func (st *Stack) DocString() string {
	var sb strings.Builder
	sb.WriteString("Stack: " + st.Mode + "\n")
	for i, sc := range st.Order {
		lp := st.Loops[sc]
		sb.WriteString(indent.Spaces(i, IndentSize) + string(lp.Scope) + ":\n")
		sb.WriteString(indent.Spaces(i+1, IndentSize) + "  Main: " + lp.Main.String() + "\n")
		sb.WriteString(indent.Spaces(i+1, IndentSize) + "  Stop: " + lp.Stop.String() + "\n")
		sb.WriteString(indent.Spaces(i+1, IndentSize) + "  End:  " + lp.End.String() + "\n")
	}
	return sb.String()
}

// Times returns a list of Times strings for loops
func (st *Stack) Times() []string {
	tms := make([]string, len(st.Order))
	for i, sc := range st.Order {
		_, tm := sc.ModeAndTimeStr()
		tms[i] = tm
	}
	return tms
}
