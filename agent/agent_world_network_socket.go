package agent

import (
	"context"
	"github.com/Astera-org/worlds/network"
	"github.com/Astera-org/worlds/network/gengo/env"
	"github.com/emer/emergent/etime"
	"github.com/emer/emergent/looper"
	"github.com/emer/etable/etensor"
)

// AgentProxyWithWorldCache represents the agent using the AgentInterface, so that the agent can be called with a Step function.
// At the same time, it implements the WorldInterface, allowing this same object to represent the world internally to the agent.
// It holds caches for observations and actions, acting like a mailbox between where these observations and actions can be stashed.
type AgentProxyWithWorldCacheSocket struct {
	AgentInterface
	WorldInterface
	loops *looper.Manager

	CachedObservations map[string]etensor.Tensor `desc:"Observations from the last step."`
	CachedActions      map[string]Action         `desc:"Actions the action wants to take this step."`
}

// this implements the thrift interface and serves as a proxy
// between the network world and the local types
type AgentHandler struct {
	agent *AgentProxyWithWorldCacheSocket
}

// Thrift Agent interface implementation (stuff that comes in over the network)
// we convert to and from the types we need locally

func (handler AgentHandler) Init(ctx context.Context, actionSpace env.Space,
	observationSpace env.Space) (string, error) {
	handler.agent.Init(transformSpace(actionSpace), transformSpace(observationSpace))
	// TODO: update thrift idl to return map[string]string
	// TODO: return something useful here and in the other init
	return "", nil
}

func (handler AgentHandler) Step(ctx context.Context, observations env.Observations, debug string) (env.Actions, error) {
	obs := transformObservations(observations)
	actions := handler.agent.Step(obs, debug)
	return transformActions(actions), nil
}

// helper functions

func transformActions(actions map[string]Action) env.Actions {
	res := make(env.Actions)
	for k, v := range actions {
		res[k] = toEnvAction(&v)
	}
	return res
}

func transformSpace(space env.Space) map[string]SpaceSpec {
	res := make(map[string]SpaceSpec)
	for k, v := range space {
		res[k] = toSpaceSpec(v)
	}
	return res
}

func transformObservations(observations env.Observations) map[string]etensor.Tensor {
	res := make(map[string]etensor.Tensor)
	for k, v := range observations {
		res[k] = toTensor(v)
	}
	return res
}

func toSpaceSpec(spec *env.SpaceSpec) SpaceSpec {
	return SpaceSpec{
		ContinuousShape: toInt(spec.Shape.Shape),
		Stride:          toInt(spec.Shape.Stride),
		Min:             spec.Min, Max: spec.Max,
		DiscreteLabels: spec.DiscreteLabels,
	}
}

func fromSpaceSpec(spec *SpaceSpec) *env.SpaceSpec {
	return &env.SpaceSpec{
		Shape: &env.Shape{Shape: toInt32(spec.ContinuousShape), Stride: toInt32(spec.Stride)},
		Min:   spec.Min,
		Max:   spec.Max,
	}
}

func toEnvAction(action *Action) *env.Action {
	return &env.Action{
		ActionShape:    fromSpaceSpec(action.ActionShape),
		Vector:         fromTensor(action.Vector),
		DiscreteOption: int32(action.DiscreteOption),
	}
}

func toTensor(envtensor *env.ETensor) etensor.Tensor {
	return etensor.NewFloat64Shape(toShape(envtensor.Shape), envtensor.Values)
}

func toShape(shape *env.Shape) *etensor.Shape {
	return &etensor.Shape{
		Shp:  toInt(shape.Shape),
		Strd: toInt(shape.Stride),
		Nms:  shape.Names,
	}
}

func fromShape(shape *etensor.Shape) *env.Shape {
	return &env.Shape{
		Shape:  toInt32(shape.Shp),
		Stride: toInt32(shape.Strd),
		Names:  shape.Nms,
	}
}

func fromTensor(tensor etensor.Tensor) *env.ETensor {
	res := &env.ETensor{
		Shape:  fromShape(tensor.ShapeObj()),
		Values: nil, // gets set in the next line
	}
	tensor.Floats(&res.Values)
	return res
}

func toInt(xs []int32) []int {
	res := make([]int, len(xs))
	for i := range xs {
		res[i] = int(xs[i])
	}
	return res
}

func toInt32(xs []int) []int32 {
	res := make([]int32, len(xs))
	for i := range xs {
		res[i] = int32(xs[i])
	}
	return res
}

/////////////////////
// Agent functions //

// StartServer blocks, waiting for calls from the environment
func (agent *AgentProxyWithWorldCacheSocket) StartServer() {
	handler := AgentHandler{agent}
	server := network.MakeServer(handler)
	server.Serve()
}

// GetServerFunc returns a function that can be used to start the server.
func (agent *AgentProxyWithWorldCacheSocket) GetServerFunc(loops *looper.Manager) func() {
	agent.loops = loops
	return agent.StartServer
}

// Init the agent. This tells the agent what shape input/output to use, and gets some suggestions from the agent about what the world should be like.
func (agent *AgentProxyWithWorldCacheSocket) Init(actionSpace map[string]SpaceSpec, observationSpace map[string]SpaceSpec) map[string]string {
	// TODO If you want, you could add a callback here to reconfigure the network based on the action and observation spaces.
	agent.loops.Init()
	return nil // Return agent name or type or requests for the environment or something.
}

// Step the agent. Internally, this calls looper.Manager.Step. It provides observations to the agent, and records what actions were taken, using the caches on the WorldInterface to move them in and out.
func (agent *AgentProxyWithWorldCacheSocket) Step(observations map[string]etensor.Tensor, debug string) map[string]Action {
	agent.CachedObservations = observations // Record observations for this timestep for the world to report.
	agent.loops.Step(1, etime.Trial)
	// After 1 trial has been stepped, a new action will be ready to return.
	return agent.CachedActions
}

/////////////////////
// World functions //

// Init sets up a server and waits for the agent to handshake with it for initiation.
func (world *AgentProxyWithWorldCacheSocket) InitWorld(details map[string]string) (map[string]SpaceSpec, map[string]SpaceSpec) {
	// This does nothing. The external world initializes itself.
	return nil, nil // Return action space and observation space.
}

// StepWorld steps the world one step. This function is called internally by the agent as a way of recording what actions it is taking that step.
func (world *AgentProxyWithWorldCacheSocket) StepWorld(actions map[string]Action, agentDone bool) (bool, string) {
	world.CachedActions = actions
	// world.CachedObservations = nil // These may be needed after an action is taken? E.g. for a teaching signal?
	return false, "" // Return observations, done, and debug string.
}

// Observe returns an observation from the cache.
func (world *AgentProxyWithWorldCacheSocket) Observe(name string) etensor.Tensor {
	if world.CachedObservations == nil {
		return nil
	}
	obs, ok := world.CachedObservations[name]
	if ok {
		return obs
	}
	return nil
}

// ObserveWithShape Returns a tensor for the named modality like Observe. This allows the agent to request an observation in a specific shape, which may involve downsampling. It should throw an error if the shap can't be satisfied.
func (world *AgentProxyWithWorldCacheSocket) ObserveWithShape(name string, shape SpaceSpec) etensor.Tensor {
	// TODO Actually call Observe and reshape it.
	return getRandomTensor(shape)
}
