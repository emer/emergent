package looper

import (
	"fmt"
	"github.com/emer/emergent/envlp"
	"github.com/emer/emergent/etime"
	"strconv"
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

func (loopman *LoopManager) DocString() string {
	s := ""
	for evalMode, stack := range loopman.Stacks {
		s = s + "\nStack: " + evalMode.String()
		for _, time := range stack.Order {
			loop := stack.Loops[time]
			s = s + "\n\tScale: " + time.String() + "\tMax: " + strconv.Itoa(loop.Counter.Max)
			if len(loop.OnStart) > 0 {
				s = s + "\n\t\tOnStart: "
				for _, f := range loop.OnStart {
					s = s + f.Name + ", "
				}
			}
			if len(loop.OnEnd) > 0 {
				s = s + "\n\t\tOnEnd: "
				for _, f := range loop.OnEnd {
					s = s + f.Name + ", "
				}
			}
			if len(loop.IsDone) > 0 {
				s = s + "\n\t\tIsDone: "
				for name, _ := range loop.IsDone {
					s = s + name + ", "
				}
			}
			if len(loop.Phases) > 0 {
				s = s + "\n\t\tPhases: "
				for _, phase := range loop.Phases {
					s = s + phase.Name + ", "
				}
				// Also print out phase details
			}
		}
	}
	return s
}

func (loopman LoopManager) GetLooperStack() *Set {
	set := NewSet()

	for m, loops := range loopman.Stacks {
		scopes := []etime.ScopeKey{}
		for _, t := range loops.Order {
			scopes = append(scopes, etime.Scope(m, t))
		}
		st := NewStackScope(scopes...)
		set.Stacks[m.String()] = st
		st.Mode = m.String()
		// TODO Env
		for _, t := range loops.Order {
			//st.Order = append(st.Order, etime.Scope(m, t))
			//loop := Loop{}
			//st.Loops[etime.Scope(m, t)] = &loop
			//loop.Stack = st
			//loop.Scope = etime.Scope(m, t)
			loop := st.Loop(t)
			ourloop := loops.Loops[t]
			// TODO Putting these both in Main?
			// TODO Check time == 0 here
			for _, nf := range ourloop.OnStart {
				loop.Main.Add(nf.Name, nf.Func)
			}
			for _, nf := range ourloop.Main {
				loop.Main.Add(nf.Name, nf.Func)
			}
			for _, nf := range ourloop.OnEnd {
				loop.End.Add(nf.Name, nf.Func)
			}
			for nm, fn := range ourloop.IsDone {
				loop.Stop.Add(nm, fn)
			}
		}
		st.Step.Default = loops.Order[0].String()
		st.Set = set
	}

	return set
}

//////////////////////////////////////////////////////////////////////
// Running

type Stepper struct {
	StopFlag  bool        `desc:"If true, stop model ASAP."`
	StopNext  bool        `desc:"If true, stop model after next stop level."`
	StopLevel etime.Times `desc:"Time level to stop at the end of."`
	//currentLevel int          `desc:"An internal variable representing our place in the stack of loops."` // TODO Is this necessary?
	Loops *LoopManager `desc:"The information about loops."`
	Mode  etime.Modes  `desc:"The current evaluation mode."`

	lastStoppedLevel int `desc:"The level at which a stop interrupted flow."`
	internalStop     bool
}

func (stepper *Stepper) Init(loopman *LoopManager) {
	stepper.Loops = loopman
	stepper.StopLevel = etime.Run
	stepper.Mode = etime.Train
	stepper.lastStoppedLevel = -2 // -2 or less is necessary
}

func (stepper *Stepper) Run() {
	// Reset internal variables
	stepper.internalStop = false
	stepper.lastStoppedLevel = -2

	// 0 Means the top level loop, probably Run
	stepper.runLevel(0)
}

// runLevel implements nested for loops recursively. It is set up so that it can be stopped and resumed at any point.
func (stepper *Stepper) runLevel(currentLevel int) {
	//stepper.StopFlag = false // TODO Will this not work right?
	st := stepper.Loops.Stacks[stepper.Mode]
	if currentLevel >= len(st.Order) {
		return // Stack overflow
	}
	time := st.Order[currentLevel]
	loop := st.Loops[time]
	ctr := loop.Counter

	for ctr.Cur < ctr.Max || ctr.Max < 0 { // Loop forever for negative maxes
		stopAtLevel := st.Order[currentLevel] >= stepper.StopLevel // Based on conversion of etime.Times to int
		if stepper.internalStop || (stepper.StopFlag && stopAtLevel) {
			// This should occur before ctr incrementing and before functions.
			//fmt.Println("Stop! " + time.String()) // DO NOT SUBMIT Remove these
			stepper.lastStoppedLevel = currentLevel
			stepper.internalStop = true
			stepper.StopFlag = false
			return
		}
		if stepper.StopNext && st.Order[currentLevel] == stepper.StopLevel {
			stepper.StopNext = false
			stepper.StopFlag = true
		}

		if currentLevel > stepper.lastStoppedLevel+1 {
			// Loop flow was interrupted, and we should not start again.
			if time > etime.Trial {
				fmt.Println(time.String() + ":Start:" + strconv.Itoa(ctr.Cur))
			}
			for _, fun := range loop.OnStart {
				fun.Func()
			}
		}

		// Recursion!
		stepper.phaseLogic(loop)
		stepper.runLevel(currentLevel + 1)

		if currentLevel > stepper.lastStoppedLevel {
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
					goto exitLoop // Exit multiple for loops without flag variable.
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
		if ctr.Cur <= amount { //In between on Start and on End, inclusive
			for _, function := range phase.OnMillisecondEnd {
				function.Func()
			}
		}
		if ctr.Cur == (amount) { //if end of a phase
			for _, function := range phase.PhaseEnd {
				function.Func()
			}
		}
	}
}
