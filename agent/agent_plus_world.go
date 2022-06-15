package agent

import (
	"github.com/emer/emergent/etime"
	"github.com/emer/emergent/looper"
	"github.com/emer/emergent/patgen"
	"github.com/emer/etable/etensor"
)

// AgentProxyWithWorldCache represents the agent using the AgentInterface, so that the agent can be called with a Step function.
// At the same time, it implements the WorldInterface, allowing this same object to represent the world internally to the agent.
// It holds caches for observations and actions, acting like a mailbox between where these observations and actions can be stashed.
type AgentProxyWithWorldCache struct {
	AgentInterface
	WorldInterface
	loops *looper.Manager

	CachedObservations map[string]etensor.Tensor `desc:"Observations from the last step."`
	CachedActions      map[string]Action         `desc:"Actions the action wants to take this step."`
}

/////////////////////
// Agent functions //

// StartServer blocks, and sometimes calls Init or Step.
func (agent *AgentProxyWithWorldCache) StartServer() {
	// TODO Set up server and replace logic below.

	// TODO Replace this logic which does not use the network at all.
	agent.Init(nil, nil)
	for {
		agent.Step(nil, "Everything is fine")
	}
}

// GetServerFunc returns a function that can be used to start the server.
func (agent *AgentProxyWithWorldCache) GetServerFunc(loops *looper.Manager) func() {
	agent.loops = loops
	return agent.StartServer
}

// Init the agent. This tells the agent what shape input/output to use, and gets some suggestions from the agent about what the world should be like.
func (agent *AgentProxyWithWorldCache) Init(actionSpace map[string]SpaceSpec, observationSpace map[string]SpaceSpec) map[string]string {
	// TODO If you want, you could add a callback here to reconfigure the network based on the action and observation spaces.
	agent.loops.ResetCounters() //todo this be moved or clarified about it's purpose
	return nil                  // Return agent name or type or requests for the environment or something.
}

// Step the agent. Internally, this calls looper.Manager.Step. It provides observations to the agent, and records what actions were taken, using the caches on the WorldInterface to move them in and out.
func (agent *AgentProxyWithWorldCache) Step(observations map[string]etensor.Tensor, debug string) map[string]Action {
	agent.CachedObservations = observations // Record observations for this timestep for the world to report.
	agent.loops.Step(agent.loops.Mode, 1, etime.Trial)
	// After 1 trial has been stepped, a new action will be ready to return.
	return agent.CachedActions
}

/////////////////////
// World functions //

// Init sets up a server and waits for the agent to handshake with it for initiation.
func (world *AgentProxyWithWorldCache) InitWorld(details map[string]string) (map[string]SpaceSpec, map[string]SpaceSpec) {
	// This does nothing. The external world initializes itself.
	return nil, nil // Return action space and observation space.
}

// StepWorld steps the world one step. This function is called internally by the agent as a way of recording what actions it is taking that step.
func (world *AgentProxyWithWorldCache) StepWorld(actions map[string]Action, agentDone bool) (bool, string) {
	world.CachedActions = actions
	// world.CachedObservations = nil // These may be needed after an action is taken? E.g. for a teaching signal?
	return false, "" // Return observations, done, and debug string.
}

// getRandomTensor is a helper method for generating random observations.
func getRandomTensor(shape SpaceSpec) etensor.Tensor {
	rt := etensor.NewFloat32(shape.ContinuousShape, shape.Stride, nil)
	patgen.PermutedBinaryRows(rt, 1, 1, 0)
	return rt
}

// Observe returns an observation from the cache.
func (world *AgentProxyWithWorldCache) Observe(name string) etensor.Tensor {
	if world.CachedObservations == nil {
		// Random observations.
		return getRandomTensor(SpaceSpec{
			ContinuousShape: []int{5, 5},
		})
	}
	obs, ok := world.CachedObservations[name]
	if ok {
		return obs
	}
	return nil
}

// ObserveWithShape Returns a tensor for the named modality like Observe. This allows the agent to request an observation in a specific shape, which may involve downsampling. It should throw an error if the shap can't be satisfied.
func (world *AgentProxyWithWorldCache) ObserveWithShape(name string, shape SpaceSpec) etensor.Tensor {
	// TODO Actually call Observe and reshape it.
	return getRandomTensor(shape)
}
