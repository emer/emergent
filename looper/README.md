# Looper: A flexible steppable control hierarchy

Looper implements a fully generic looping control system with extensible functionality at each level of the loop, with logic that supports reentrant stepping so each time it is Run it advances at any specified step size, with results that are identical to running.

This steppability constraint requires that no code is run at the start of a loop, which is equivalent to a `do..while` loop with no initialization condition:

```C
do {
    Main()
} while !Stop()
End()
```

The `Loop` object has these three function lists: `Main(), Stop(), End()` where function closures can be added to perform any relevant functionality.

In Go syntax, including the running of sub-loops under a given level, this would be:

```Go
for {
   for { <subloops here> } // drills down levels for each subloop
   Main()                  // Main is called after subloops -- increment counters!
   if Stop() {
       break
   }
}
End()                      // Reset counters here so next pass starts over
```

To make this work, an initialization function must be run prior to starting, which puts the system in a ready-to-run state.  The `End()` function at each level must likewise ensure that it is ready to start again properly the next time through.

The [envlp](https://github.com/emer/emergent/tree/master/envlp) Env is designed to work in concert with the looper control, where the Env holds counter values, and looper automatically increments and uses these counters to stop looping at a given level.  Each `Stack` of loops is associated with a given `etime.Mode`, corresponding to that of the Env.

# Algorithm and Sim Integration

Specific algorithms use `AddLevels` to add inner, lower levels of loops to implement specific algorithm code (typically `Phase` and `Cycle`).  Leabra and Axon use the `Time` struct as a context for holding the relevant counters and mode, which is then accessed directly in the callback functions as needed.

In cases where something must be done prior to looping through cycles (e.g., `ApplyInputs` and new phase startup methods), trigger it on the first cycle, before calling other functions, using a provided `AddCycle0` function.

# Concrete Example of Looping Logic

The `stack_test.go` output shows the logic of the looping functions:

Here's the trace of a Run with 2 Run iterations, 3 Epoch iterations, and 3 Trials per epoch:

```
		Trial Main: 1
		Trial Main: 2
		Trial Main: 3
		Trial Stop: 3
		Trial End: 0
	Epoch Main: 1
		Trial Main: 1
		Trial Main: 2
		Trial Main: 3
		Trial Stop: 3
		Trial End: 0
	Epoch Main: 2
		Trial Main: 1
		Trial Main: 2
		Trial Main: 3
		Trial Stop: 3
		Trial End: 0
	Epoch Main: 3
	Epoch Stop: 3
	Epoch End: 0
Run Main: 1
		Trial Main: 1
   .... (repeat of above)
	Epoch Main: 3
	Epoch Stop: 3
	Epoch End: 0
Run Main: 2
Run Stop: 2
Run End: 0
```

Here is stepping 1 Trial at a time:

```
##############
Step Trial 1
		Trial Main: 1

##############
Step Trial 1
		Trial Main: 2

##############
Step Trial 1
		Trial Main: 3
		Trial Stop: 3
		Trial End: 0
	Epoch Main: 1

##############
Step Trial 1
		Trial Main: 1

##############
Step Trial 2
		Trial Main: 1
		Trial Main: 2
```

Here is stepping 2 Trials at a time:

```
##############
Step Trial 2
		Trial Main: 3
		Trial Stop: 3
		Trial End: 0
	Epoch Main: 1
		Trial Main: 1

##############
Step Trial 2
		Trial Main: 2
		Trial Main: 3
		Trial Stop: 3
		Trial End: 0
	Epoch Main: 2

##############
Step Trial 2
		Trial Main: 1
		Trial Main: 2
```

And here's stepping 1 Epoch at a time:

```
##############
Step Epoch 1
		Trial Main: 1
		Trial Main: 2
		Trial Main: 3
		Trial Stop: 3
		Trial End: 0
	Epoch Main: 1

##############
Step Epoch 1
		Trial Main: 1
		Trial Main: 2
		Trial Main: 3
		Trial Stop: 3
		Trial End: 0
	Epoch Main: 2

##############
Step Epoch 1
		Trial Main: 1
		Trial Main: 2
		Trial Main: 3
		Trial Stop: 3
		Trial End: 0
	Epoch Main: 3
	Epoch Stop: 3
	Epoch End: 0
Run Main: 1

##############
Step Epoch 1
		Trial Main: 1
		Trial Main: 2
		Trial Main: 3
		Trial Stop: 3
		Trial End: 0
	Epoch Main: 1
```


