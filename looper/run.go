// Copyright (c) 2022, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package looper

import (
	"fmt"
	"strings"
)

func indent(level int) string {
	return strings.Repeat("   ", level)
}

// runLevel implements nested for loops recursively.
// It is set up so that it can be stopped and resumed at any point.
func (ss *Stacks) runLevel(currentLevel int) bool {
	st := ss.Stacks[ss.Mode]
	if currentLevel >= len(st.Order) {
		return true // Stack overflow, expected at bottom of stack.
	}
	time := st.Order[currentLevel]
	loop := st.Loops[time]
	ctr := &loop.Counter

	for ctr.Cur < ctr.Max || ctr.Max <= 0 { // Loop forever for non-maxes
		stoplev := int64(-1)
		if st.StopLevel != nil {
			stoplev = st.StopLevel.Int64()
		}
		stopAtLevelOrLarger := st.Order[currentLevel].Int64() >= stoplev
		if st.StopFlag && stopAtLevelOrLarger {
			ss.internalStop = true
		}
		if ss.internalStop {
			// This should occur before ctr incrementing and before functions.
			st.StopFlag = false
			return false // Don't continue above, e.g. Stop functions
		}
		if st.StopNext && st.Order[currentLevel] == st.StopLevel {
			st.StopCount -= 1
			if st.StopCount <= 0 {
				st.StopNext = false
				st.StopFlag = true // Stop at the top of the next StopLevel
			}
		}

		// Don't ever Start the same iteration of the same level twice.
		lastCounter, ok := ss.lastStartedCounter[ToScope(ss.Mode, time)]
		if !ok || ctr.Cur > lastCounter {
			ss.lastStartedCounter[ToScope(ss.Mode, time)] = ctr.Cur
			if PrintControlFlow {
				fmt.Printf("%s%s: Start: %d\n", indent(currentLevel), time.String(), ctr.Cur)
			}
			for _, ev := range loop.Events {
				if ctr.Cur == ev.AtCounter {
					ev.OnEvent.Run()
				}
			}
			loop.OnStart.Run()
		} else if PrintControlFlow {
			fmt.Printf("%s%s: Skipping Start: %d\n", indent(currentLevel), time.String(), ctr.Cur)
		}

		// Recursion!
		runComplete := ss.runLevel(currentLevel + 1)

		if runComplete {
			if PrintControlFlow {
				fmt.Printf("%s%s: End: %d\n", indent(currentLevel), time.String(), ctr.Cur)
			}
			loop.OnEnd.Run()
			ctr.Incr()
			// Reset the counter at the next level.
			// Do this here so that the counter number is visible during loop.OnEnd.
			if currentLevel+1 < len(st.Order) {
				st.Level(currentLevel + 1).Counter.Cur = 0
				ss.lastStartedCounter[ToScope(ss.Mode, st.Order[currentLevel+1])] = -1
			}

			for _, fun := range loop.IsDone {
				if fun.Func() {
					if PrintControlFlow {
						fmt.Printf("%s%s: IsDone Stop at: %d from: %s\n", indent(currentLevel), time.String(), ctr.Cur, fun.Name)
					}
					goto exitLoop // Exit IsDone and Counter for-loops without flag variable.
				}
			}
		}
	}

exitLoop:
	// Only get to this point if this loop is done.
	return true
}
