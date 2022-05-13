package agent

import "github.com/emer/etable/etensor"

// The SpaceSpec of an action or observation.
// It should *either* be continuous with a SpaceSpec or discrete, but not both. If both are set, treat it as continuous, using SpaceSpec.
type SpaceSpec struct {
	// Continuous
	ContinuousShape []int   `desc:"The dimensions of an array. For example, [3,2] would be a 3 by 2 array. [1] would be a single value."`
	Stride          []int   `desc:"TODO Replace Shape and Stride with one Shape object, like an etensor.Tensor has."`
	Min             float64 `desc:"The minimum continuous value."`
	Max             float64 `desc:"The maximum continuous value."`

	// Discrete
	DiscreteLabels []string `desc:"The names of the discrete possibilities, such as ['left', 'right']. The length of this is the number of discrete possibilities that this shape encapsulates."`
}

// An Action describes what the agent is doing at a given timestep. It should contain either a continuous vector or a discrete option, as specified by its shape.
type Action struct {
	ActionShape    *SpaceSpec     `desc:"Optional description of the action."`
	Vector         etensor.Tensor `desc:"A vector describing the action. For example, this might be joint positions or forces applied to actuators."`
	DiscreteOption int            `desc:"Choice from among the DiscreteLabels in Continuous."`
}

// AgentInterface allows the Agent to provide actions given observations. This allows the agent to be embedded within a world.
type AgentInterface interface {
	// Init passes variables to the Agent: Action space, Observation space, and initial Observation. It receives any specification in the form of a string which the agent chooses to provide. Agent should reinitialize the network for the beginning of a new run.
	// details lets the agent request that the world be configured to spec.
	Init(actionSpace map[string]SpaceSpec, observationSpace map[string]SpaceSpec) (details map[string]string)

	// Step takes in a map of named Observations. It returns a map of named Actions. The observation can be expected to conform to the shape given in Init, and the Action should conform to the action specification given there. The debug string is for debug information from the environment and should not be used for real training or evaluation.
	Step(observations map[string]etensor.Tensor, debug string) (actions map[string]Action)
}
