package agent

import (
	"context"

	"github.com/Astera-org/worlds/network"
	"github.com/Astera-org/worlds/network/gengo/env"
	"github.com/emer/emergent/looper"
	"github.com/emer/etable/etensor"
)

// AgentHandler implements the thrift interface and serves as a proxy
// between the network world and the local types
type AgentHandler struct {
	Agent *AgentProxyWithWorldCache
}

// GetServerFunc returns a function that blocks, waiting for calls from the environment.
func (handler *AgentHandler) GetServerFunc(loops *looper.Manager) func() {
	_ = handler.Agent.GetServerFunc(loops)
	return func() {
		server := network.MakeServer(handler)
		server.Serve()
	}
}

type Serverable interface {
	GetServerFunc(loops *looper.Manager) func()
}

// Thrift Agent interface implementation (stuff that comes in over the network)
// we convert to and from the types we need locally

func (handler *AgentHandler) Init(ctx context.Context, actionSpace env.Space,
	observationSpace env.Space) (map[string]string, error) {
	handler.Agent.Init(transformSpace(actionSpace), transformSpace(observationSpace))
	// TODO: update thrift idl to return map[string]string
	// TODO: return something useful here and in the other init
	return map[string]string{}, nil
}

func (handler *AgentHandler) Step(ctx context.Context, observations env.Observations, debug string) (env.Actions, error) {
	obs := transformObservations(observations)
	actions := handler.Agent.Step(obs, debug)
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

// Names starting with "to" mean that it's from network code to local code. Names starting with "from" mean it's a conversion from local code to network code.

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
	return etensor.NewShape(toInt(shape.Shape), toInt(shape.Stride), shape.Names)
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
	if xs == nil {
		return nil
	}
	res := make([]int, len(xs))
	for i := range xs {
		res[i] = int(xs[i])
	}
	return res
}

func toInt32(xs []int) []int32 {
	if xs == nil {
		return nil
	}
	res := make([]int32, len(xs))
	for i := range xs {
		res[i] = int32(xs[i])
	}
	return res
}
