Docs: [GoDoc](https://pkg.go.dev/github.com/emer/emergent/envlp)

See [Wiki Env](https://github.com/emer/emergent/wiki/Env) page for detailed docs.

Package `envlp` defines the `Env` interface for environments, using the [looper](https://github.com/emer/emergent/tree/master/looper) control framework to manage the incrementing of counters on the Env, instead of the Env automatically incrementing counters on its own, which is the behavior of the original `env.Env` environments.

The Env determines the nature and sequence of States that can be used as inputs to a model and it can also accept Action responses from the model that affect how the environment evolves in the future.

By adhering to this interface, it is then easier to mix-and-match environments with models.

![Env / Agent](agent_env_interface.png?raw=true "Logical interface between the agent and the environment: the Environment supplies State to the Agent, and receives Actions from the Agent.")

Env uses a *post increment* logic for stepping the state, consistent with the [looper](https://github.com/emer/emergent/tree/master/looper) framework: current State of the Env should be ready to use after the `Init()` call, and `Step()` is called *after* the current State is used, to advance so its ready to be used for the next iteration, while also incrementing the lowest-level counter that tracks Step updates (e.g., the Trial level).  No other counters should be updated.

The Counters are associated with [etime](https://github.com/emer/emergent/tree/master/etime) Times, as used in [looper](https://github.com/emer/emergent/tree/master/looper), which manages the counter updating based on `Max` values specified in the Env, or other relevant stopping criteria in the `Stop()` functions on Loops.  Each `looper.Stack` can be associated with an Env, which it manages, and both the Stack and the Env have an associated `etime.Mode` evaluation mode (Train, Test, etc) that should be consistent.

Thus, multiple different environments will typically be used in a model, for each different evaluation mode.  Nevertheless, these different Envs can share a common database of patterns, e.g., using the `etable.IdxView` to present different indexed views into a shared common `etable.Table` (e.g., train / test splits). The basic `FixedTable` env implementation uses this.

Particular paradigms of environments must establish naming conventions for these state elements which then allow the model to use the information appropriately -- e.g., "Input", "Output" are widely used, but more realistic and multimodal models have many different types of state.  The Env interface only provides the most basic framework for establishing these paradigms, and ultimately a given model will only work within a particular paradigm of environments following specific conventions.

Typically each specific implementation of this Env interface will have a `Config()` function that at least takes the eval mode for the env as an arg, and multiple parameters etc that can be modified to control env behavior -- all of this is paradigm-specific and outside the scope of this basic interface.

The `States32` type is convenient for managing named env states.

# Standard boilerplate code

Here's some standard boilerplate code used in most Env implementations:

```Go
// MyEnv is a skeleton environment to use as a starting point for custom envs
type MyEnv struct {
	Nm    string     `desc:"name of this environment"`
	Dsc   string     `desc:"description of this environment"`
	EMode string     `desc:"eval mode for this env"`
	Ctrs  envlp.Ctrs `desc:"counters for this environment"`
}

func (ev *MyEnv) Name() string          { return ev.Nm }
func (ev *MyEnv) Desc() string          { return ev.Dsc }
func (ev *MyEnv) Mode() string          { return ev.EMode }
func (ev *MyEnv) Counters() *envlp.Ctrs { return &ev.Ctrs }
func (ev *MyEnv) Counter(time etime.Times) *envlp.Ctr {
	return ev.Ctrs.ByScope(etime.ScopeStr(ev.EMode, time.String()))
}
func (ev *MyEnv) String() string { return "" }
func (ev *MyEnv) CtrsToStats(stats *estats.Stats) {
	ev.Ctrs.CtrsToStats(stats)
}

func (ev *MyEnv) Config(mode string) {
	ev.EMode = mode
	ev.Ctrs.SetTimes(mode, etime.Run, etime.Epoch, etime.Trial)
}

func (ev *MyEnv) Validate() error {
	return nil
}

func (ev *MyEnv) Init() {
	run := 0
	rc := ev.Counter(etime.Run)
	if rc != nil {
		run = rc.Cur
	}
	ev.Ctrs.Init()
	if rc != nil {
		rc.Set(run)
	}
}

func (ev *MyEnv) Step() {
	tc := ev.Counter(etime.Trial)
	tc.Incr()
}

func (ev *MyEnv) State(element string) etensor.Tensor {
	return nil
}

func (ev *MyEnv) Action(element string, input etensor.Tensor) {
	// nop
}

// Compile-time check that implements Env interface
var _ envlp.Env = (*MyEnv)(nil)
```

