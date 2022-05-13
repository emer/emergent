package agent

import (
	"github.com/emer/etable/etensor"
)

// WorldInterface is a minimal interface for the world. It allows the agent to initialize it, step it with actions, and get observations from it.
type WorldInterface interface {
	// InitWorld Initializes or reinitialize the world. This blocks until it hears from the world that it has been initialized. It returns the specifications for the action and observation spaces as its two return arguments.
	// The details allow the Agent to request that the world configure itself in some way.
	// The Action Space describes the shape and names of what the model can send as outputs. This will be constant across the run.
	// The Observation Space describes the shape and names of what the model can expect as inputs. This will be constant across the run.
	// InitWorld and StepWorld are so named to allow a single object to implement both this interface and the AgentInterface.
	InitWorld(details map[string]string) (actionSpace map[string]SpaceSpec, observationSpace map[string]SpaceSpec)

	// StepWorld the environment. It takes in a set of actions and returns observations and a debug string. // TODO Recomment
	// The actions should conform to the action space specification.
	// The observations can be expected to conform to the observation space specification. The observations will be cached such that a separate function can get them before the next time Step is called.
	// The debug string should not be used for actual training.
	StepWorld(actions map[string]Action, agentDone bool) (done bool, debug string)

	// Observe returns a cached tensor for the named modality. E.g. “x” or “vision” or “reward”. This just returns a cached entry into the map gotten the last time Step was called.
	Observe(name string) etensor.Tensor
}
