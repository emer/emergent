Docs: [GoDoc](https://pkg.go.dev/github.com/emer/emergent/env)

Package `env` defines the `Env` interface for environments, which determine the nature and sequence of States as inputs to a model. Action responses from the model can also drive state evolution.

By adhering to this interface, it is then easier to mix-and-match environments with models.

![Env / Agent](agent_env_interface.png?raw=true "Logical interface between the agent and the environment: the Environment supplies State to the Agent, and receives Actions from the Agent.")

Multiple different environments will typically be used in a model, e.g., one for training and other(s) for testing. Even if these envs all share a common database of patterns, a different Env should be used for each case where different counters and sequences of events etc are presented, which keeps them from interfering with each other. Also, `table.NewView` can be used to create new views on a common set of pattenrs, so different Envs can present different indexed views. The basic `FixedTable` env implementation uses this.

The standard `String() string` `fmt.Stringer` method must be defined to return a string description of the current environment state, e.g., as a TrialName. A `Label() string` method must be defined to return the Name of the environment, which is typically the Mode of usage (Train vs. Test).

There is also an `Envs` map that provides a basic container for managing multiple Envs, using a `string` key based on the `Label()` name.

The `Step` should update all relevant state elements as appropriate, so these can be queried by the user. Particular paradigms of environments must establish naming conventions for these state elements which then allow the model to use the information appropriately -- the Env interface only provides the most basic framework for establishing these paradigms, and ultimately a given model will only work within a particular paradigm of environments following specific conventions.

See e.g., env.FixedTable for particular implementation of a fixed Table of patterns, for one example of a widely used paradigm.

Typically each specific implementation of this Env interface will have multiple parameters etc that can be modified to control env behavior -- all of this is paradigm-specific and outside the scope of this basic interface.

