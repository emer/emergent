package looper

import (
	"github.com/emer/emergent/etime"
	"github.com/goki/ki/indent"
	"strings"
)

type DataManager struct {
	Stacks map[etime.Modes]*Stack
	Steps  Stepper
}

func (loopman *DataManager) GetLoop(modes etime.Modes, times etime.Times) *Loop {
	return loopman.Stacks[modes].Loops[times]
}

func (loopman DataManager) Init() *DataManager {
	loopman.Stacks = map[etime.Modes]*Stack{}
	return &loopman
}

func (loopman *DataManager) AddAcrossAllModesAndTimes(fun func(etime.Modes, etime.Times)) {
	for _, m := range []etime.Modes{etime.Train, etime.Test} {
		curMode := m // For closures.
		for _, t := range []etime.Times{etime.Trial, etime.Epoch} {
			curTime := t
			fun(curMode, curTime)
		}
	}
}

func (loopman *DataManager) AddSegmentAllModes(t etime.Times, loopSegment Span) {
	// Note that phase is copied
	for mode, _ := range loopman.Stacks {
		stack := loopman.Stacks[mode]
		stack.Loops[t].AddSpans(loopSegment)
	}
}

// DocString returns an indented summary of the loops
// and functions in the stack
func (loopman DataManager) DocString() string {
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
			if len(lp.Spans) > 0 {
				sb.WriteString(indent.Spaces(i+1, indentSize) + "  Phases:\n")
				for _, ph := range lp.Spans {
					sb.WriteString(indent.Spaces(i+2, indentSize) + ph.String() + "\n")
				}
			}
		}
	}
	return sb.String()
}
