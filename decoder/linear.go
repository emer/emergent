// Copyright (c) 2023, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package decoder

//go:generate core generate -add-types

import (
	"fmt"

	"cogentcore.org/core/math32"
	"cogentcore.org/lab/base/mpi"
	"github.com/emer/etensor/tensor"
)

type ActivationFunc func(float32) float32

// Linear is a linear neural network, which can be configured with a custom
// activation function. By default it will use the identity function.
// It learns using the delta rule for each output unit.
type Linear struct {

	// learning rate
	LRate float32 `default:"0.1"`

	// layers to decode
	Layers []Layer

	// unit values -- read this for decoded output
	Units []LinearUnit

	// number of inputs -- total sizes of layer inputs
	NInputs int

	// number of outputs -- total sizes of layer inputs
	NOutputs int

	// input values, copied from layers
	Inputs []float32

	// for holding layer values
	ValuesTsrs map[string]*tensor.Float32 `display:"-"`

	// synaptic weights: outer loop is units, inner loop is inputs
	Weights tensor.Float32

	// activation function
	ActivationFn ActivationFunc

	// which pool to use within a layer
	PoolIndex int

	// mpi communicator -- MPI users must set this to their comm -- do direct assignment
	Comm *mpi.Comm `display:"-"`

	// delta weight changes: only for MPI mode -- outer loop is units, inner loop is inputs
	MPIDWts tensor.Float32
}

// Layer is the subset of emer.Layer that is used by this code
type Layer interface {
	Name() string
	UnitValuesTensor(tsr tensor.Tensor, varNm string, di int) error
	Shape() *tensor.Shape
}

func IdentityFunc(x float32) float32 { return x }

// LogisticFunc implements the standard logistic function.
// Its outputs are in the range (0, 1).
// Also known as Sigmoid. See https://en.wikipedia.org/wiki/Logistic_function.
func LogisticFunc(x float32) float32 { return 1 / (1 + math32.FastExp(-x)) }

// LinearUnit has variables for Linear decoder unit
type LinearUnit struct {

	// target activation value -- typically 0 or 1 but can be within that range too
	Target float32

	// final activation = sum x * w -- this is the decoded output
	Act float32

	// net input = sum x * w
	Net float32
}

// InitLayer initializes detector with number of categories and layers
func (dec *Linear) InitLayer(nOutputs int, layers []Layer, activationFn ActivationFunc) {
	dec.Layers = layers
	nIn := 0
	for _, ly := range dec.Layers {
		nIn += ly.Shape().Len()
	}
	dec.Init(nOutputs, nIn, -1, activationFn)
}

// InitPool initializes detector with number of categories, 1 layer and
func (dec *Linear) InitPool(nOutputs int, layer Layer, poolIndex int, activationFn ActivationFunc) {
	dec.Layers = []Layer{layer}
	shape := layer.Shape()
	// TODO: assert that it's a 4D layer
	nIn := shape.DimSize(2) * shape.DimSize(3)
	dec.Init(nOutputs, nIn, poolIndex, activationFn)
}

// Init initializes detector with number of categories and number of inputs
func (dec *Linear) Init(nOutputs, nInputs int, poolIndex int, activationFn ActivationFunc) {
	dec.NInputs = nInputs
	dec.LRate = 0.1
	dec.NOutputs = nOutputs
	dec.Units = make([]LinearUnit, dec.NOutputs)
	dec.Inputs = make([]float32, dec.NInputs)
	dec.Weights.SetShape([]int{dec.NOutputs, dec.NInputs}, "Outputs", "Inputs")
	for i := range dec.Weights.Values {
		dec.Weights.Values[i] = 0.1
	}
	dec.PoolIndex = poolIndex
	dec.ActivationFn = activationFn
}

// Decode decodes the given variable name from layers (forward pass).
// Decoded values are in Units[i].Act -- see also Output to get into a []float32.
// di is a data parallel index di, for networks capable
// of processing input patterns in parallel.
func (dec *Linear) Decode(varNm string, di int) {
	dec.Input(varNm, di)
	dec.Forward()
}

// Output returns the resulting Decoded output activation values into given slice
// which is automatically resized if not of sufficient size.
func (dec *Linear) Output(acts *[]float32) {
	if cap(*acts) < dec.NOutputs {
		*acts = make([]float32, dec.NOutputs)
	} else if len(*acts) != dec.NOutputs {
		*acts = (*acts)[:dec.NOutputs]
	}
	for ui := range dec.Units {
		u := &dec.Units[ui]
		(*acts)[ui] = u.Act
	}
}

// Train trains the decoder with given target correct answers, as []float32 values.
// Returns SSE (sum squared error) of difference between targets and outputs.
// Also returns and prints an error if targets are not sufficient length for NOutputs.
func (dec *Linear) Train(targs []float32) (float32, error) {
	err := dec.SetTargets(targs)
	if err != nil {
		return 0, err
	}
	sse := dec.Back()
	return sse, nil
}

// TrainMPI trains the decoder with given target correct answers, as []float32 values.
// Returns SSE (sum squared error) of difference between targets and outputs.
// Also returns and prints an error if targets are not sufficient length for NOutputs.
// MPI version uses mpi to synchronize weight changes across parallel nodes.
func (dec *Linear) TrainMPI(targs []float32) (float32, error) {
	err := dec.SetTargets(targs)
	if err != nil {
		return 0, err
	}
	sse := dec.BackMPI()
	return sse, nil
}

// SetTargets sets given target correct answers, as []float32 values.
// Also returns and prints an error if targets are not sufficient length for NOutputs.
func (dec *Linear) SetTargets(targs []float32) error {
	if len(targs) < dec.NOutputs {
		err := fmt.Errorf("decoder.Linear: number of targets < NOutputs: %d < %d", len(targs), dec.NOutputs)
		fmt.Println(err)
		return err
	}
	for ui := range dec.Units {
		u := &dec.Units[ui]
		u.Target = targs[ui]
	}
	return nil
}

// ValuesTsr gets value tensor of given name, creating if not yet made
func (dec *Linear) ValuesTsr(name string) *tensor.Float32 {
	if dec.ValuesTsrs == nil {
		dec.ValuesTsrs = make(map[string]*tensor.Float32)
	}
	tsr, ok := dec.ValuesTsrs[name]
	if !ok {
		tsr = &tensor.Float32{}
		dec.ValuesTsrs[name] = tsr
	}
	return tsr
}

// Input grabs the input from given variable in layers
// di is a data parallel index di, for networks capable
// of processing input patterns in parallel.
func (dec *Linear) Input(varNm string, di int) {
	off := 0
	for _, ly := range dec.Layers {
		tsr := dec.ValuesTsr(ly.Name())
		ly.UnitValuesTensor(tsr, varNm, di)
		if dec.PoolIndex >= 0 {
			shape := ly.Shape()
			y := dec.PoolIndex / shape.DimSize(1)
			x := dec.PoolIndex % shape.DimSize(1)
			tsr = tsr.SubSpace([]int{y, x}).(*tensor.Float32)
		}
		for j, v := range tsr.Values {
			dec.Inputs[off+j] = v
		}
		off += ly.Shape().Len()
	}
}

// Forward compute the forward pass from input
func (dec *Linear) Forward() {
	for ui := range dec.Units {
		u := &dec.Units[ui]
		net := float32(0)
		off := ui * dec.NInputs
		for j, in := range dec.Inputs {
			net += dec.Weights.Values[off+j] * in
		}
		u.Net = net
		u.Act = dec.ActivationFn(net)
	}
}

// https://en.wikipedia.org/wiki/Delta_rule
// Delta rule: delta = learning rate * error * input
// We don't need the g' (derivative of activation function) term assuming:
// 1. Identity activation function with SSE loss (beecause it's 1), OR
// 2. Logistic activation function with Cross Entropy loss (because it cancels out, see
//    https://towardsdatascience.com/deriving-backpropagation-with-cross-entropy-loss-d24811edeaf9)
// The fact that we return SSE does not mean we're optimizing SSE.

// Back compute the backward error propagation pass
// Returns SSE (sum squared error) of difference between targets and outputs.
func (dec *Linear) Back() float32 {
	var sse float32
	for ui := range dec.Units {
		u := &dec.Units[ui]
		err := u.Target - u.Act
		sse += err * err
		del := dec.LRate * err
		off := ui * dec.NInputs
		for j, in := range dec.Inputs {
			dec.Weights.Values[off+j] += del * in
		}
	}
	return sse
}

// BackMPI compute the backward error propagation pass
// Returns SSE (sum squared error) of difference between targets and outputs.
func (dec *Linear) BackMPI() float32 {
	if dec.MPIDWts.Len() != dec.Weights.Len() {
		dec.MPIDWts.CopyShapeFrom(&dec.Weights)
	}
	var sse float32
	for ui := range dec.Units {
		u := &dec.Units[ui]
		err := u.Target - u.Act
		sse += err * err
		del := dec.LRate * err
		off := ui * dec.NInputs
		for j, in := range dec.Inputs {
			dec.MPIDWts.Values[off+j] = del * in
		}
	}
	dec.Comm.AllReduceF32(mpi.OpSum, dec.MPIDWts.Values, nil)

	for i, dw := range dec.MPIDWts.Values {
		dec.Weights.Values[i] += dw
	}

	return sse
}
