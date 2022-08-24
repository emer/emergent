# Looper: A flexible steppable control hierarchy

Docs: [GoDoc](https://pkg.go.dev/github.com/emer/emergent/looper)

Looper implements a fully generic looping control system with extensible functionality at each level of the loop, with logic that supports reentrant stepping so each time it is Run it advances at any specified step size, with results that are identical to running.

Each loop implements the following logic:

```go
OnStart()  // run at start of loop
do {
    Main() // run in loop
    <run subloops>
} while !IsDone() // test for completion
OnEnd()    // run after loop is done
```

The `Loop` object has the above function lists  where function closures can be added to perform any relevant functionality.

Each level of loop holds a corresponding counter value, and the looper automatically increments and uses these counters to stop looping at a given level.  Each `Stack` of loops is associated with a given `etime.Mode`, e.g., `etime.Train` or `etime.Test`.

# Algorithm and Sim Integration

Specific algorithms use `AddLevels` to add inner, lower levels of loops to implement specific algorithm code (typically `Phase` and `Cycle`).  Leabra and Axon use the `Time` struct as a context for holding the relevant counters and mode, which is then accessed directly in the callback functions as needed.

In cases where something must be done prior to looping through cycles (e.g., `ApplyInputs` and new phase startup methods), trigger it on the first cycle, before calling other functions, using a provided `AddCycle0` function.

# Concrete Example of Looping Logic

The `stack_test.go` can generate a trace the looping -- edit the if false to true to see.



