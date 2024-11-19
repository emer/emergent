// Copyright (c) 2019, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package params

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

type test struct {
	Name     string
	Class    string
	Norm     bool
	Momentum bool
	WtBal    bool
	WtScale  float32
}

func (t *test) StyleName() string  { return t.Name }
func (t *test) StyleClass() string { return t.Class }

var paramSets = Sheets[*test]{
	"Base": {
		{Sel: "", Doc: "norm and momentum on works better, but wt bal is not better for smaller nets",
			Set: func(t *test) {
				t.Norm = true
				t.Momentum = true
				t.WtBal = false
			}},
		{Sel: ".Back", Doc: "top-down back-pathways MUST have lower relative weight scale, otherwise network hallucinates",
			Set: func(t *test) {
				t.WtScale = 0.2
			}},
		{Sel: "#ToOutput", Doc: "to output must be stronger",
			Set: func(t *test) {
				t.WtScale = 2.0
			}},
	},
	"NoMomentum": {
		{Sel: "", Doc: "no norm or momentum",
			Set: func(t *test) {
				t.Norm = false
				t.Momentum = false
			}},
	},
	"WtBalOn": {
		{Sel: "", Doc: "weight bal on",
			Set: func(t *test) {
				t.WtBal = true
			}},
	},
}

func TestSet(t *testing.T) {
	tf := &test{}
	tf.Name = "Forward"
	tb := &test{}
	tb.Class = "Back"
	to := &test{}
	to.Name = "ToOutput"

	paramSets["Base"].Apply(tf)
	assert.Equal(t, true, tf.Norm)

	paramSets["Base"].Apply(tb)
	assert.Equal(t, float32(0.2), tb.WtScale)

	paramSets["Base"].Apply(to)
	assert.Equal(t, float32(2.0), to.WtScale)

	paramSets["NoMomentum"].Apply(tf)
	assert.Equal(t, false, tf.Norm)
}
