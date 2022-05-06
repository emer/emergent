// Copyright (c) 2022, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package looper

import (
	"github.com/emer/emergent/etime"
)

// Stack contains a list of Loops Ordered from top to bottom.
// For example, a Stack might be created like this:
//   myStack.Init().AddTime(etime.Run, 2).AddTime(etime.Trial, 3)
//   myStack.Loops[etime.Run].OnStart.Add("NewRun", initRunFunc)
//   myStack.Loops[etime.Trial].OnStart.Add("PresentTrial", trialFunc)
// When run, myStack will behave like this:
// initRunFunc, trialFunc, trialFunc, trialFunc, initRunFunc, trialFunc, trialFunc, trialFunc
type Stack struct {
	Loops map[etime.Times]*Loop `desc:"An ordered map of Loops, from the outer loop at the start to the inner loop at the end."`
	Order []etime.Times         `desc:"This should be managed internally. Ordered from top to bottom, so longer timescales like Run should be at the beginning and shorter timescales like Trial should be and the end."`
}

// Init makes sure data structures are initialized, and empties them if they are.
func (loops *Stack) Init() *Stack {
	loops.Loops = map[etime.Times]*Loop{}
	loops.Order = []etime.Times{}
	return loops
}

// AddTime adds a new timescale to this Stack with a given number of iterations. The order in which this method is invoked is important, as it adds loops in order from top to bottom.
func (loops *Stack) AddTime(time etime.Times, max int) *Stack {
	loops.Loops[time] = &Loop{Counter: Ctr{Max: max}, IsDone: map[string]func() bool{}}
	loops.Order = append(loops.Order, time)
	return loops
}
