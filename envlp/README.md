Docs: [GoDoc](https://pkg.go.dev/github.com/emer/emergent/envlp)

See [Wiki Env](https://github.com/emer/emergent/wiki/Env) page for detailed docs.

Package `envlp` defines the `Env` interface for environments, using the `looper` control framework to manage the incrementing of counters on the Env, instead of the Env automatically incrementing counters on its own, which is the behavior of the original `env.Env` environments.

The Env determines the nature and sequence of States that can be used as inputs to a model and it can also accept Action responses from the model that affect how the environment evolves in the future.

By adhering to this interface, it is then easier to mix-and-match environments with models.

![Env / Agent](agent_env_interface.png?raw=true "Logical interface between the agent and the environment: the Environment supplies State to the Agent, and receives Actions from the Agent.")

The current State of the Env should be ready to use (based on `Init()`), and `Step()` should advance the State so its ready to be used for the next iteration.

The Counters are expected to be managed by the `looper` system -- each `looper.Stack` can be associated with an Env, that it manages.  

Multiple different environments will typically be used in a model, e.g., one for training and other(s) for testing.  Even if these envs all share a common database of patterns, a different Env should be used for each case where different counters and sequences of events etc are presented, which keeps them from interfering with each other.  Also, the etable.IdxView can be used to allow multiple different Env's to all present different indexed views into a shared common etable.Table (e.g., train / test splits). The basic `FixedTable` env implementation uses this.

The `EnvDesc` interface provides additional methods (originally included in `Env`) that describe the Counters, States, and Actions, of the Env.  Each `Element` of the overall `State` allows annotation about the different elements of state that are available in general.

Particular paradigms of environments must establish naming conventions for these state elements which then allow the model to use the information appropriately -- the Env interface only provides the most basic framework for establishing these paradigms, and ultimately a given model will only work within a particular paradigm of environments following specific conventions.

Typically each specific implementation of this Env interface will have a `Config()` function that at least takes the eval mode for the env as an arg, and multiple parameters etc that can be modified to control env behavior -- all of this is paradigm-specific and outside the scope of this basic interface.

