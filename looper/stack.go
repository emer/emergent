// Copyright (c) 2022, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package looper

import "github.com/emer/emergent/etime"

// Stack contains one stack of nested loops
type Stack struct {
	Scope     etime.ScopeKey           `desc:"top, entry level of the loops (i.e., first element in Order)"`
	Order     []etime.ScopeKey         `desc:"order of the loops"`
	StepScope etime.ScopeKey           `desc:"stepping level"`
	StepN     int                      `desc:"number of times to iterate at StepScope level, no stepping if 0"`
	Loops     map[etime.ScopeKey]*Loop `desc:"the loops by scope"`
}

func NewStack(mode etime.EvalModes, times ...etime.Times) *Stack {
	ord := make([]etime.ScopeKey, len(times))
	for i, t := range times {
		ord[i] = etime.Scope(mode, t)
	}
	return NewStackScope(ord...)
}

func NewStackScope(scopes ...etime.ScopeKey) *Stack {
	st := &Stack{}
	st.Order = scopes
	st.Scope = st.Order[0]
	st.Loops = make(map[etime.ScopeKey]*Loop, len(st.Order))
	for _, sc := range st.Order {
		st.Loops[sc] = NewLoop(sc)
	}
	return st
}

func (st *Stack) Loop(time etime.Times) *Loop {
	md, _ := st.Scope.ModesAndTimes()
	sc := etime.ScopeStr(md[0], time.String())
	return st.Loops[sc]
}

func (st *Stack) Level(lev int) *Loop {
	if lev < 0 || lev >= len(st.Order) {
		return nil
	}
	return st.Loops[st.Order[lev]]
}

func (st *Stack) Run(set *Set) {
	lev := 0
	lp := st.Level(lev)
	lp.Start.Run()
	var nlp *Loop
	for {
		lp.Pre.Run()
		lev++
		nlp = st.Level(lev)
		if nlp != nil {
			lp = nlp
			lp.Start.Run()
			continue
		}
		lev--
	post:
		lp.Post.Run()
		stop := lp.Stop.Run()
		if stop || set.StopFlag {
			lp.End.Run()
			lev--
			nlp = st.Level(lev)
			if nlp == nil || set.StopFlag {
				break
			}
			lp = nlp
			goto post
		}
	}
}
