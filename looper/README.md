# Looper: A flexible steppable control hierarchy

Docs: [GoDoc](https://pkg.go.dev/github.com/emer/emergent/looper)

Looper implements a fully generic looping control system, with a `Stack` of `Loop` elements that iterate over different time scales of processing, where the processing performed is provided by function closures on the Loop elements.

Critically, the loop logic supports _reentrant stepping_, such that you can iteratively `Step` through the loop processing and accomplish exactly the same outcomes as if you did a complete `Run` from the start.

Each loop implements the following logic, where it is key to understand that the time scale associated with the loop _runs the full iterations over that time scale_.  For example, a `Trial` loop _iterates over Trials_ -- it is _not_ a single trial, but the whole sequence (loop) of trials.

```go
for {
	Events[Counter == AtCounter] // run events at counter
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

Each `Stack` of loops has an associated `Mode` enum, e.g., `Train` or `Test`, and each `Loop` has an associated `Time` level, e.g., `Run`, `Epoch`, `Trial`.

The collection of `Stacks` has a high-level API for configuring and controlling the set of `Stack` elements, and has the logic for running everything, in the form of `Run`, `Step`, `Stop` methods, etc.

# Examples

The following examples use the [etime](../etime) `Modes` and `Times` enums. It is recommended that you define your own `Modes` enums if not using the basic `Train` and `Test` cases, to provide a better [egui](../egui) representation of the loop stack.

## Configuration

From `step_test.go` `ExampleStacks`:

```Go
	stacks := NewStacks()
	stacks.AddStack(etime.Train, etime.Trial).
		AddTime(etime.Epoch, 3).
		AddTime(etime.Trial, 2)

	// add function closures:         
	stacks.Loop(etime.Train, etime.Epoch).OnStart.Add("Epoch Start", func() { fmt.Println("Epoch Start") })
	stacks.Loop(etime.Train, etime.Epoch).OnEnd.Add("Epoch End", func() { fmt.Println("Epoch End") })
	stacks.Loop(etime.Train, etime.Trial).OnStart.Add("Trial Run", func() { fmt.Println("  Trial Run") })

   // add events:
	stacks.Loop(etime.Train, etime.Epoch).AddEvent("EpochTwoEvent", 2, func() { fmt.Println("Epoch==2") })
	stacks.Loop(etime.Train, etime.Trial).AddEvent("TrialOneEvent", 1, func() { fmt.Println("  Trial==1") })
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
stacks.Run(etime.Train)
```

Reset first and Run, ensures that the full sequence is run even if it might have been stopped or stepped previously:
```Go
stacks.ResetAndRun(etime.Train)
```

Step by 1 Trial:
```Go
stacks.Step(etime.Train, 1, etime.Trial)
```


