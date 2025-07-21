# Looper: A flexible steppable control hierarchy

Docs: [GoDoc](https://pkg.go.dev/github.com/emer/emergent/looper)

Looper implements a fully generic looping control system, with a `Stack` of `Loop` elements that iterate over different levels or time scales of processing, where the processing performed is provided by function closures on the Loop elements. Each Stack is defined by a `Mode` enum, e.g., Train vs. Test.

Thus, the looper structure is defined by two "coordinate" variables: `Mode` stack and loop `Level`, which should be provided by end-user defined [enums](https://github.com/cogentcore/core/tree/enums) values (the looper code uses the `enums.Enum` interface).

Critically, the loop logic supports _reentrant stepping_, such that you can iteratively `Step` through the loop processing and accomplish exactly the same outcomes as if you did a complete `Run` from the start.

Each loop implements the following logic, where it is key to understand that the level associated with the loop _runs the full iterations over that level_.  For example, a `Trial` loop _iterates over Trials_ -- it is _not_ a single trial, but the whole sequence (loop) of trials.

```go
for {
	Events[Counter == AtCounter] // run events for current counter value
	OnStart()
	    Run Sub-Loop to completion
	OnEnd()
	Counter += Inc
	if Counter >= Max || IsDone() {
	  break
	}
}
```

The `Loop` object has the above function lists (`OnStart`, `OnEnd`, and `IsDone`), where function closures can be added to perform any relevant functionality. `Events` have the trigger `AtCounter` and a list of functions to call.

Each level of loop holds a corresponding `Counter` value, which increments at each iteration, and its `Max` value determines when the loop iteration terminates.

The collection of `Stacks` has a high-level API for configuring and controlling the set of `Stack` elements, and has the logic for running everything, in the form of `Run`, `Step`, `Stop`, `Init` methods, etc.

# Examples

The following examples use the `Modes` and `Levels` enums defined in the [levels](levels) sub-package, which is intended for testing and example purposes: each use-case should define its own enums for better clarity and flexibility down the road.

```Go
type Modes int32 //enums:enum
const (
	Train Modes = iota
	Test
)

type Levels int32 //enums:enum
const (
	Cycle Levels = iota
	Trial
	Epoch
	Run
)
```

## Configuration

From `step_test.go` `ExampleStacks`:

```Go
	stacks := NewStacks()
	stacks.AddStack(levels.Train, levels.Trial).
		AddLevel(levels.Epoch, 3).
		AddLevel(levels.Trial, 2)

	// add function closures:
	stacks.Loop(levels.Train, levels.Epoch).OnStart.Add("Epoch Start", func() { fmt.Println("Epoch Start") })
	stacks.Loop(levels.Train, levels.Epoch).OnEnd.Add("Epoch End", func() { fmt.Println("Epoch End") })
	stacks.Loop(levels.Train, levels.Trial).OnStart.Add("Trial Run", func() { fmt.Println("  Trial Run") })

	// add events:
	stacks.Loop(levels.Train, levels.Epoch).AddEvent("EpochTwoEvent", 2, func() { fmt.Println("Epoch==2") })
	stacks.Loop(levels.Train, levels.Trial).AddEvent("TrialOneEvent", 1, func() { fmt.Println("  Trial==1") })
```

The `DocString` for this stack is:

```
Stack Train:
   Epoch[0 : 3]:
      Events:
         EpochTwoEvent: [at 2] Events: EpochTwoEvent 
      Start:  Epoch Start 
      Trial[0 : 2]:
         Events:
            TrialOneEvent: [at 1] Events: TrialOneEvent 
         Start:  Trial Run 
      End:    Epoch End 
```

and the output when run is:

```
Epoch Start
  Trial Run
  Trial==1
  Trial Run
Epoch End
Epoch Start
  Trial Run
  Trial==1
  Trial Run
Epoch End
Epoch==2
Epoch Start
  Trial Run
  Trial==1
  Trial Run
Epoch End
```

## Running, Stepping

Run a full stack:
```Go
stacks.Run(level.Train)
```

Reset first and Run, ensures that the full sequence is run even if it might have been stopped or stepped previously:
```Go
stacks.ResetAndRun(level.Train)
```

Step by 1 Trial:
```Go
stacks.Step(level.Train, 1, level.Trial)
```

## Stacks config API

Most configuration can be handled by these helper functions defined on the `Stacks` type:

```Go
// AddEventAllModes adds a new event for all modes at given loop level.
AddEventAllModes(level enums.Enum, name string, atCtr int, fun func())

// AddOnStartToAll adds given function taking mode and level args to OnStart in all stacks, loops
AddOnStartToAll(name string, fun func(mode, level enums.Enum))

// AddOnEndToAll adds given function taking mode and level args to OnEnd in all stacks, loops
AddOnEndToAll(name string, fun func(mode, level enums.Enum))

// AddOnStartToLoop adds given function taking mode arg to OnStart in all stacks for given loop.
AddOnStartToLoop(level enums.Enum, name string, fun func(mode enums.Enum))

// AddOnEndToLoop adds given function taking mode arg to OnEnd in all stacks for given loop.
AddOnEndToLoop(level enums.Enum, name string, fun func(mode enums.Enum))
```

