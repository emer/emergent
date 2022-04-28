package looper

import (
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

type ThetaPhase struct {
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

	Phases []ThetaPhase `desc:"Only use Phases at the Theta Cycle timescale (200ms)."`
	// TODO Add an axon.time here but move it to etimes

	Counter *envlp.Ctr `desc:"Tracks time within the loop. Also tracks the maximum."`
}

func (loops *LoopStructure) AddPhases(phases ...ThetaPhase) {
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
	StopFlag     bool         `desc:"If true, stop model ASAP."`
	StopLevel    etime.Times  `desc:"Time level to stop at the end of."`
	currentLevel int          `desc:"An internal variable representing our place in the stack of loops."` // TODO Is this necessary?
	Loops        *LoopManager `desc:"The information about loops."`
	Mode         etime.Modes  `desc:"The current evaluation mode."`
}

func (stepper *Stepper) Init(loopman *LoopManager) {
	stepper.Loops = loopman
	stepper.StopLevel = etime.Run
	stepper.Mode = etime.Train
	stepper.currentLevel = 0
}

func (stepper *Stepper) Run() {
	stepper.StopFlag = false // TODO Will this not work right?
	st := stepper.Loops.Stacks[stepper.Mode]
	if stepper.currentLevel >= len(st.Order) {
		stepper.currentLevel -= 1
		return // Stack overflow
	}
	loop := st.Loops[st.Order[stepper.currentLevel]]
	ctr := loop.Counter

	for ctr.Cur < ctr.Max || ctr.Max < 0 { // Loop forever for negative maxes
		//stopAtLevel := st.Order[stepper.currentLevel] >= stepper.StopLevel // Based on conversion of etime.Times to int
		if stepper.StopFlag { //} || stopAtLevel {//DO NOT SUBMIT Whoops!
			// This should occur before ctr incrementing and before functions.
			return
		}

		for _, fun := range loop.OnStart {
			fun.Func()
		}

		// Recursion!
		stepper.currentLevel += 1 // Go down a level
		stepper.Run()
		ctr.Cur = ctr.Cur + 1 // Increment

		for _, fun := range loop.Main {
			fun.Func()
			// TODO Could do no recursion if there are Main functions
		}
		for _, fun := range loop.OnEnd {
			fun.Func()
		}
		for _, fun := range loop.IsDone {
			if fun() {
				goto exitLoop // Exit multiple for loops without flag variable.
			}
		}
	}
exitLoop:
	// Only get to this point if this loop is done.
	ctr.Cur = 0
	stepper.currentLevel -= 1 // Go up a level
}
