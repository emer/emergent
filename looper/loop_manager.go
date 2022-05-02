package looper

import (
	"github.com/emer/emergent/etime"
	"github.com/goki/ki/indent"
	"strconv"
	"strings"
)

type LoopManager struct {
	Stacks map[etime.Modes]*EvaluationModeLoops
	Steps  Stepper
}

func (loopman *LoopManager) GetLoop(modes etime.Modes, times etime.Times) *LoopStructure {
	return loopman.Stacks[modes].Loops[times]
}

func (loopman LoopManager) Init() *LoopManager {
	loopman.Stacks = map[etime.Modes]*EvaluationModeLoops{}
	return &loopman
}

func (loopman *LoopManager) Validate() *LoopManager {
	// TODO Make sure there are no duplicates.
	// TODO Print a note if there's a negative Max which will translate to looping forever.
	return loopman
}

// TODO Use this in ra25.go
func (loopman *LoopManager) AddAcrossAllModesAndTimes(fun func(etime.Modes, etime.Times)) {
	for _, m := range []etime.Modes{etime.Train, etime.Test} {
		curMode := m // For closures.
		for _, t := range []etime.Times{etime.Trial, etime.Epoch} {
			curTime := t
			fun(curMode, curTime)
		}
	}
}

func (loopman *LoopManager) AddPhaseAllModes(t etime.Times, phase Phase) {
	// Note that phase is copied
	for mode, _ := range loopman.Stacks {
		stack := loopman.Stacks[mode]
		stack.Loops[t].AddPhases(phase)
	}
}

// DocString returns an indented summary of the loops
// and functions in the stack
func (loopman LoopManager) DocString() string {
	var sb strings.Builder

	// indentSize is number of spaces to indent for output
	var indentSize = 4

	for evalMode, st := range loopman.Stacks {
		sb.WriteString("Stack: " + evalMode.String() + "\n")
		for i, t := range st.Order {
			lp := st.Loops[t]
			sb.WriteString(indent.Spaces(i, indentSize) + evalMode.String() + ":" + t.String() + ":\n")
			sb.WriteString(indent.Spaces(i+1, indentSize) + "  Start:  " + lp.OnStart.String() + "\n")
			sb.WriteString(indent.Spaces(i+1, indentSize) + "  Main:  " + lp.Main.String() + "\n")
			if len(lp.IsDone) > 0 {
				s := ""
				for nm, _ := range lp.IsDone {
					s = s + nm + " "
				}
				sb.WriteString(indent.Spaces(i+1, indentSize) + "  Stop:  " + s + "\n")
			}
			sb.WriteString(indent.Spaces(i+1, indentSize) + "  End:   " + lp.OnEnd.String() + "\n")
			if len(lp.Phases) > 0 {
				s := ""
				for _, ph := range lp.Phases {
					s = s + ph.Name + "(" + strconv.Itoa(ph.Duration) + ") "
				}
				sb.WriteString(indent.Spaces(i+1, indentSize) + "  Phases:" + s + "\n")
			}
		}
	}
	return sb.String()
}
