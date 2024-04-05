// Copyright (c) 2021, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package decoder

import (
	"fmt"
	"testing"

	"github.com/emer/etable/v2/etensor"
	"github.com/stretchr/testify/assert"
)

// TestLayer implements a Layer
type TestLayer struct {
	tensors map[string]etensor.Tensor
}

func (tl *TestLayer) Name() string {
	return "TestLayer"
}

func (tl *TestLayer) UnitValuesTensor(tsr etensor.Tensor, varNm string, di int) error {
	src, ok := tl.tensors[varNm]
	if !ok {
		return fmt.Errorf("bad key: %s", varNm)
	}
	tsr.CopyShapeFrom(src)
	tsr.CopyFrom(src)
	return nil
}

func (tl *TestLayer) Shape() *etensor.Shape {
	for _, v := range tl.tensors {
		return v.ShapeObj()
	}
	return nil
}

func testLinear(t *testing.T, activationFn ActivationFunc) {
	const tol = 1.0e-6

	dec := Linear{}
	dec.Init(2, 2, -1, activationFn)
	trgs := []float32{0, 1}
	outs := []float32{0, 0}
	var lastSSE float32
	for i := 0; i < 100; i++ {
		if i%2 == 0 {
			dec.Inputs[0] = 1
			dec.Inputs[1] = 0
			trgs[0] = 1
			trgs[1] = 0
		} else {
			dec.Inputs[0] = 0
			dec.Inputs[1] = 1
			trgs[0] = 0
			trgs[1] = 1
		}
		dec.Forward()
		dec.Output(&outs)
		if i > 2 {
			if i%2 == 0 {
				if outs[0] < outs[1] {
					t.Errorf("err: %d\t output: %g !> other: %g\n", i, outs[0], outs[1])
				}
			} else {
				if outs[1] < outs[0] {
					t.Errorf("err: %d\t output: %g !> other: %g\n", i, outs[1], outs[0])
				}
			}
		}
		sse, err := dec.Train(trgs)
		if err != nil {
			t.Error(err)
		}
		if i > 2 {
			if (sse - lastSSE) > tol {
				t.Errorf("error: %d\t sse now is *larger* than previoust: %g > %g\n", i, sse, lastSSE)
			}
		}
		lastSSE = sse
	}
}

func TestLinearIdentity(t *testing.T) {
	testLinear(t, IdentityFunc)
}

func TestLinearLogistic(t *testing.T) {
	testLinear(t, LogisticFunc)
}

func TestInputPool1D(t *testing.T) {
	dec := Linear{}
	shape := etensor.NewShape([]int{1, 5, 6, 6}, nil, nil)
	vals := make([]float32, shape.Len())
	for i := range vals {
		vals[i] = float32(i)
	}
	tensor := etensor.NewFloat32Shape(shape, vals)
	layer := TestLayer{tensors: map[string]etensor.Tensor{"var0": tensor}}
	dec.InitPool(2, &layer, 0, IdentityFunc)
	dec.Input("var0", 0)
	expected := tensor.SubSpace([]int{0, 0}).(*etensor.Float32).Values
	assert.Equal(t, expected, dec.Inputs)

	dec.InitPool(2, &layer, 1, IdentityFunc)
	dec.Input("var0", 0)
	expected = tensor.SubSpace([]int{0, 1}).(*etensor.Float32).Values
	assert.Equal(t, expected, dec.Inputs)
}

func TestInputPool2D(t *testing.T) {
	dec := Linear{}
	shape := etensor.NewShape([]int{2, 5, 6, 6}, nil, nil)
	vals := make([]float32, shape.Len())
	for i := range vals {
		vals[i] = float32(i)
	}
	tensor := etensor.NewFloat32Shape(shape, vals)
	layer := TestLayer{tensors: map[string]etensor.Tensor{"var0": tensor}}
	dec.InitPool(2, &layer, 0, IdentityFunc)
	dec.Input("var0", 0)
	expected := tensor.SubSpace([]int{0, 0}).(*etensor.Float32).Values
	assert.Equal(t, expected, dec.Inputs)

	dec.InitPool(2, &layer, 1, IdentityFunc)
	dec.Input("var0", 0)
	expected = tensor.SubSpace([]int{0, 1}).(*etensor.Float32).Values
	assert.Equal(t, expected, dec.Inputs)

	dec.InitPool(2, &layer, 5, IdentityFunc)
	dec.Input("var0", 0)
	expected = tensor.SubSpace([]int{1, 0}).(*etensor.Float32).Values
	assert.Equal(t, expected, dec.Inputs)

	dec.InitPool(2, &layer, 9, IdentityFunc)
	dec.Input("var0", 0)
	expected = tensor.SubSpace([]int{1, 4}).(*etensor.Float32).Values
	assert.Equal(t, expected, dec.Inputs)
}
