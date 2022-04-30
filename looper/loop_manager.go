package looper

import (
	"fmt"
	"github.com/emer/emergent/envlp"
	"github.com/emer/emergent/etime"
	"github.com/goki/ki/indent"
	"strconv"
	"strings"
)

type namedFunc struct {
	Name string
	Func func()
}

type orderedMapFuncs []namedFunc

func (funcs *orderedMapFuncs) Add(name string, fun func()) *orderedMapFuncs {
	*funcs = append(*funcs, namedFunc{Name: name, Func: fun})
	return funcs
}

func (funcs orderedMapFuncs) String() string {
	s := ""
	if len(funcs) > 0 {
		for _, f := range funcs {
			s = s + f.Name + " "
		}
	}
	return s
}

type Phase struct {
	Name             string // Might be plus or minus for example
	Duration         int
	IsPlusPhase      bool
	OnMillisecondEnd orderedMapFuncs
	PhaseStart       orderedMapFuncs
	PhaseEnd         orderedMapFuncs

	Counter *envlp.Ctr `desc:"Tracks time within the loop. Also tracks the maximum."`
}

type LoopStructure struct {
	OnStart orderedMapFuncs
	// Either Main or the inner loop occurs between OnStart and OnEnd
	Main   orderedMapFuncs
	OnEnd  orderedMapFuncs
	IsDone map[string]func() bool `desc:"If true, end loop. Maintained as an unordered map because they should not have side effects."`

	Phases []Phase `desc:"Only use Phases at the Theta Cycle timescale (200ms)."`
	// TODO Add an axon.time here but move it to etimes

	Counter *envlp.Ctr `desc:"Tracks time within the loop. Also tracks the maximum."`
}

func (loops *LoopStructure) AddPhases(phases ...Phase) {
	for _, phase := range phases {
		loops.Phases = append(loops.Phases, phase)
		phase.OnMillisecondEnd = orderedMapFuncs{}
		phase.PhaseStart = orderedMapFuncs{}
		phase.PhaseEnd = orderedMapFuncs{}
	}
}

type EvaluationModeLoops struct {
	Loops map[etime.Times]*LoopStructure
	Order []etime.Times // This should be managed internally.
}

func (loops *EvaluationModeLoops) Init() *EvaluationModeLoops {
	loops.Loops = map[etime.Times]*LoopStructure{}
	return loops
}

func (loops *EvaluationModeLoops) AddTimeScales(times ...etime.Times) *EvaluationModeLoops {
	if loops.Loops == nil {
		loops.Loops = map[etime.Times]*LoopStructure{}
	}
	for _, time := range times {
		loops.Loops[time] = &LoopStructure{}
		loops.Order = append(loops.Order, time)
	}
	return loops
}

func (loops *EvaluationModeLoops) AddTime(time etime.Times, max int) *EvaluationModeLoops {
	loops.Loops[time] = &LoopStructure{Counter: &envlp.Ctr{Max: max}, IsDone: map[string]func() bool{}}
	loops.Order = append(loops.Order, time)
	return loops
}

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

//////////////////////////////////////////////////////////////////////
// Running

type Stepper struct {
	StopFlag  bool        `desc:"If true, stop model ASAP."`
	StopNext  bool        `desc:"If true, stop model after next stop level."`
	StopLevel etime.Times `desc:"Time level to stop at the end of."`
	//currentLevel int          `desc:"An internal variable representing our place in the stack of loops."`
	StepIterations int          `desc:"How many steps to do."`
	Loops          *LoopManager `desc:"The information about loops."`
	Mode           etime.Modes  `desc:"The current evaluation mode."`

	lastStoppedLevel int `desc:"The level at which a stop interrupted flow."`
	internalStop     bool
}

func (stepper *Stepper) Init(loopman *LoopManager) {
	stepper.Loops = loopman
	stepper.StopLevel = etime.Run
	stepper.Mode = etime.Train
	stepper.lastStoppedLevel = -1
}

func (stepper *Stepper) Run() {
	// Reset internal variables
	stepper.internalStop = false

	// 0 Means the top level loop, probably Run
	stepper.runLevel(0)
}

// runLevel implements nested for loops recursively. It is set up so that it can be stopped and resumed at any point.
func (stepper *Stepper) runLevel(currentLevel int) bool {
	//stepper.StopFlag = false // TODO Will this not work right?
	st := stepper.Loops.Stacks[stepper.Mode]
	if currentLevel >= len(st.Order) {
		return true // Stack overflow, expected at bottom of stack.
	}
	time := st.Order[currentLevel]
	loop := st.Loops[time]
	ctr := loop.Counter

	for ctr.Cur < ctr.Max || ctr.Max < 0 { // Loop forever for negative maxes
		stopAtLevel := st.Order[currentLevel] == stepper.StopLevel // Based on conversion of etime.Times to int
		if stepper.StopFlag && stopAtLevel {
			stepper.internalStop = true
			stepper.lastStoppedLevel = currentLevel
		}
		if stepper.internalStop {
			// This should occur before ctr incrementing and before functions.
			stepper.StopFlag = false
			return false // Don't continue above, e.g. Stop functions
		}
		if stepper.StopNext && st.Order[currentLevel] == stepper.StopLevel {
			stepper.StepIterations -= 1
			if stepper.StepIterations <= 0 {
				stepper.StopNext = false
				stepper.StopFlag = true
				stepper.lastStoppedLevel = -1
			}
		}

		if currentLevel >= stepper.lastStoppedLevel {
			// Loop flow was interrupted, and we should not start again.
			stepper.lastStoppedLevel = -1
			if time > etime.Trial {
				fmt.Println(time.String() + ":Start:" + strconv.Itoa(ctr.Cur))
			}
			for _, fun := range loop.OnStart {
				fun.Func()
			}
		}

		// Recursion!
		stepper.phaseLogic(loop)
		runComplete := stepper.runLevel(currentLevel + 1)

		if runComplete {
			for _, fun := range loop.Main {
				fun.Func()
			}
			if time > etime.Trial {
				fmt.Println(time.String() + ":End:  " + strconv.Itoa(ctr.Cur))
			}
			for _, fun := range loop.OnEnd {
				fun.Func()
			}
			for name, fun := range loop.IsDone {
				if fun() {
					_ = name // For debugging
					ctr.Cur = 0
					goto exitLoop // Exit multiple for-loops without flag variable.
				}
			}
			ctr.Cur = ctr.Cur + 1 // Increment
		}
	}

exitLoop:
	// Only get to this point if this loop is done.
	if !stepper.internalStop {
		ctr.Cur = 0
	}
	return true
}

// phaseLogic a loop can be broken up into discrete segments, so in a certain window you may want distinct behavior
func (stepper *Stepper) phaseLogic(loop *LoopStructure) {
	ctr := loop.Counter
	amount := 0
	for _, phase := range loop.Phases {
		amount += phase.Duration
		if ctr.Cur == (amount - phase.Duration) { //if start of a phase
			for _, function := range phase.PhaseStart {
				function.Func()
			}
		}
		if ctr.Cur < amount { //In between on Start and on End, inclusive
			for _, function := range phase.OnMillisecondEnd {
				function.Func()
			}
		}
		if ctr.Cur == amount-1 { //if end of a phase
			for _, function := range phase.PhaseEnd {
				function.Func()
			}
		}
	}
}
