// Copyright (c) 2019, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package params

import (
	"testing"

	"cogentcore.org/core/base/reflectx"
	"github.com/stretchr/testify/assert"
)

var tweakSets = Sets{
	"Base": {
		{Sel: "Path", Desc: "norm and momentum on works better, but wt bal is not better for smaller nets",
			Params: Params{
				"Path.Learn.LRate":    "0.02",
				"Path.Learn.Momentum": "0.9",
			},
			Hypers: Hypers{
				"Path.Learn.LRate":    {"Tweak": "log"},
				"Path.Learn.Momentum": {"Tweak": "incr"},
			}},
		{Sel: "Layer", Desc: "using default 1.8 inhib for all of network -- can explore",
			Params: Params{
				"Layer.Inhib.Layer.Gi": "1.8",
			},
			Hypers: Hypers{
				"Layer.Inhib.Layer.Gi": {"Tweak": "[1.75, 1.85]"},
			}},
		{Sel: "#Hidden", Desc: "output definitely needs lower inhib -- true for smaller layers in general",
			Params: Params{
				"Layer.Inhib.Layer.Gi": "1.4",
			},
			Hypers: Hypers{
				"Layer.Inhib.Layer.Gi": {"Tweak": "incr"},
			}},
		{Sel: ".Back", Desc: "top-down back-pathways MUST have lower relative weight scale, otherwise network hallucinates",
			Params: Params{
				"Path.WtScale.Rel": "0.2",
			},
			Hypers: Hypers{
				"Path.WtScale.Rel": {"Tweak": "log"},
			}},
	},
}

func TestTweak(t *testing.T) {
	logvals := []float32{.1, .2, .5, 1, 1.5, 12, .015}
	logtargs := []float32{.05, .2, .1, .5, .2, 1, .5, 2, 1.2, 2, 11, 15, .012, .02}
	for i, v := range logvals {
		ps := Tweak(v, true, false)
		for j, p := range ps {
			tp := logtargs[i*2+j]
			if p != tp {
				t.Errorf("log mismatch for v=%g: got %g != target %g\n", v, p, tp)
			}
		}
	}
	incrvals := []float32{.1, .3, 1.5, 25, .008}
	incrtargs := []float32{.09, .11, .2, .4, 1.4, 1.6, 24, 26, .007, .009}
	for i, v := range incrvals {
		ps := Tweak(v, false, true)
		for j, p := range ps {
			tp := incrtargs[i*2+j]
			if p != tp {
				t.Errorf("incr mismatch for v=%g: got %g != target %g\n", v, p, tp)
			}
		}
	}
}

var trgSearch = `[
	{
		"Param": "Layer.Inhib.Layer.Gi",
		"Sel": {
			"Sel": "#Hidden",
			"Desc": "output definitely needs lower inhib -- true for smaller layers in general",
			"Params": {
				"Layer.Inhib.Layer.Gi": "1.4"
			},
			"Hypers": {
				"Layer.Inhib.Layer.Gi": {
					"Tweak": "incr"
				}
			}
		},
		"Search": [
			{
				"Name": "Hidden",
				"Type": "Layer",
				"Path": "Layer.Inhib.Layer.Gi",
				"Start": 1.4,
				"Values": [
					1.3,
					1.5
				]
			}
		]
	},
	{
		"Param": "Path.WtScale.Rel",
		"Sel": {
			"Sel": ".Back",
			"Desc": "top-down back-pathways MUST have lower relative weight scale, otherwise network hallucinates",
			"Params": {
				"Path.WtScale.Rel": "0.2"
			},
			"Hypers": {
				"Path.WtScale.Rel": {
					"Tweak": "log"
				}
			}
		},
		"Search": [
			{
				"Name": "HiddenToInput",
				"Type": "Path",
				"Path": "Path.WtScale.Rel",
				"Start": 0.2,
				"Values": [
					0.1,
					0.5
				]
			}
		]
	},
	{
		"Param": "Layer.Inhib.Layer.Gi",
		"Sel": {
			"Sel": "Layer",
			"Desc": "using default 1.8 inhib for all of network -- can explore",
			"Params": {
				"Layer.Inhib.Layer.Gi": "1.8"
			},
			"Hypers": {
				"Layer.Inhib.Layer.Gi": {
					"Tweak": "[1.75, 1.85]"
				}
			}
		},
		"Search": [
			{
				"Name": "Input",
				"Type": "Layer",
				"Path": "Layer.Inhib.Layer.Gi",
				"Start": 1.8,
				"Values": [
					1.75,
					1.85
				]
			}
		]
	},
	{
		"Param": "Path.Learn.LRate",
		"Sel": {
			"Sel": "Path",
			"Desc": "norm and momentum on works better, but wt bal is not better for smaller nets",
			"Params": {
				"Path.Learn.LRate": "0.02",
				"Path.Learn.Momentum": "0.9"
			},
			"Hypers": {
				"Path.Learn.LRate": {
					"Tweak": "log"
				},
				"Path.Learn.Momentum": {
					"Tweak": "incr"
				}
			}
		},
		"Search": [
			{
				"Name": "HiddenToInput",
				"Type": "Path",
				"Path": "Path.Learn.LRate",
				"Start": 0.02,
				"Values": [
					0.01,
					0.05
				]
			},
			{
				"Name": "InputToHidden",
				"Type": "Path",
				"Path": "Path.Learn.LRate",
				"Start": 0.02,
				"Values": [
					0.01,
					0.05
				]
			}
		]
	},
	{
		"Param": "Path.Learn.Momentum",
		"Sel": {
			"Sel": "Path",
			"Desc": "norm and momentum on works better, but wt bal is not better for smaller nets",
			"Params": {
				"Path.Learn.LRate": "0.02",
				"Path.Learn.Momentum": "0.9"
			},
			"Hypers": {
				"Path.Learn.LRate": {
					"Tweak": "log"
				},
				"Path.Learn.Momentum": {
					"Tweak": "incr"
				}
			}
		},
		"Search": [
			{
				"Name": "HiddenToInput",
				"Type": "Path",
				"Path": "Path.Learn.Momentum",
				"Start": 0.9,
				"Values": [
					0.8,
					1
				]
			},
			{
				"Name": "InputToHidden",
				"Type": "Path",
				"Path": "Path.Learn.Momentum",
				"Start": 0.9,
				"Values": [
					0.8,
					1
				]
			}
		]
	}
]
`

func TestTweakHypers(t *testing.T) {
	hypers := Flex{}
	hypers.Init([]FlexVal{
		FlexVal{Name: "Input", Type: "Layer", Class: "Input", Object: Hypers{}},
		FlexVal{Name: "Hidden", Type: "Layer", Class: "Hidden", Object: Hypers{}},
		FlexVal{Name: "InputToHidden", Type: "Path", Class: "Forward", Object: Hypers{}},
		FlexVal{Name: "HiddenToInput", Type: "Path", Class: "Back", Object: Hypers{}},
	})
	basenet := tweakSets["Base"]
	hypers.ApplySheet(basenet, false)

	// fmt.Println("hypers:", reflectx.StringJSON(hypers))

	srch := TweaksFromHypers(hypers)
	ss := reflectx.StringJSON(srch)
	// fmt.Println("\n\n##########\n", ss)
	assert.Equal(t, trgSearch, ss)
}
