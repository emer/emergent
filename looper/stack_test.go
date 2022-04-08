// Copyright (c) 2022, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package looper

import (
	"fmt"
	"testing"

	"github.com/emer/emergent/etime"
)

func TestStack(t *testing.T) {
	set := NewSet()
	trn := NewStack(etime.Train, etime.Run, etime.Epoch, etime.Trial)
	set.AddStack(trn)

	run := 0
	lp := trn.Loop(etime.Run)
	lp.OnStart.Add(func() {
		run = 0
		fmt.Printf("Run Start: %d\n", run)
	})
	lp.RunPre.Add(func() {
		fmt.Printf("Run Pre: %d\n", run)
	})
	lp.RunPost.Add(func() {
		run++
		fmt.Printf("Run Post: %d\n", run)
	})
	lp.OnEnd.Add(func() {
		fmt.Printf("Run End: %d\n", run)
	})
	lp.Stop.Add(func() bool {
		if run >= 3 {
			fmt.Printf("Run Stop: %d\n", run)
			return true
		}
		return false
	})

	epoch := 0
	lp = trn.Loop(etime.Epoch)
	lp.OnStart.Add(func() {
		epoch = 0
		fmt.Printf("Epoch Start: %d\n", epoch)
	})
	lp.RunPre.Add(func() {
		fmt.Printf("Epoch Pre: %d\n", epoch)
	})
	lp.RunPost.Add(func() {
		epoch++
		fmt.Printf("Epoch Post: %d\n", epoch)
	})
	lp.OnEnd.Add(func() {
		fmt.Printf("Epoch End: %d\n", epoch)
	})
	lp.Stop.Add(func() bool {
		if epoch >= 3 {
			fmt.Printf("Epoch Stop: %d\n", epoch)
			return true
		}
		return false
	})

	trial := 0
	lp = trn.Loop(etime.Trial)
	lp.OnStart.Add(func() {
		trial = 0
		fmt.Printf("Trial Start: %d\n", trial)
	})
	lp.RunPre.Add(func() {
		fmt.Printf("Trial Pre: %d\n", trial)
	})
	lp.RunPost.Add(func() {
		trial++
		fmt.Printf("Trial Post: %d\n", trial)
	})
	lp.OnEnd.Add(func() {
		fmt.Printf("Trial End: %d\n", trial)
	})
	lp.Stop.Add(func() bool {
		if trial >= 3 {
			fmt.Printf("Trial Stop: %d\n", trial)
			return true
		}
		return false
	})

	set.Run(etime.Train, etime.Run)
}
