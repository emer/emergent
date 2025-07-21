// Copyright (c) 2019, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package params

/*
var tweakSets = Sets{
	"Base": = {
		{Sel: "Path", Doc: "norm and momentum on works better, but wt bal is not better for smaller nets",
			Params: Params{
				pt.Learn.LRate =    "0.02",
				pt.Learn.Momentum = "0.9",
			},
			Hypers: Hypers{
				pt.Learn.LRate =    {"Tweak = "log"},
				pt.Learn.Momentum = {"Tweak = "incr"},
			}},
		{Sel: "Layer", Doc: "using default 1.8 inhib for all of network -- can explore",
			Params: Params{
				ly.Inhib.Layer.Gi = "1.8",
			},
			Hypers: Hypers{
				ly.Inhib.Layer.Gi = {"Tweak = "[1.75, 1.85]"},
			}},
		{Sel: "#Hidden", Doc: "output definitely needs lower inhib -- true for smaller layers in general",
			Params: Params{
				ly.Inhib.Layer.Gi = "1.4",
			},
			Hypers: Hypers{
				ly.Inhib.Layer.Gi = {"Tweak = "incr"},
			}},
		{Sel: ".Back", Doc: "top-down back-pathways MUST have lower relative weight scale, otherwise network hallucinates",
			Params: Params{
				pt.WtScale.Rel = "0.2",
			},
			Hypers: Hypers{
				pt.WtScale.Rel = {"Tweak = "log"},
			}},
	},
}

*/

/*
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
*/

/*
var trgSearch = `[
	{
		"Param": ly.Inhib.Layer.Gi",
		"Sel": {
			"Sel": "#Hidden",
			"Doc": "output definitely needs lower inhib -- true for smaller layers in general",
			"Params": {
				ly.Inhib.Layer.Gi": "1.4"
			},
			"Hypers": {
				ly.Inhib.Layer.Gi": {
					"Tweak": "incr"
				}
			}
		},
		"Search": [
			{
				"Name": "Hidden",
				"Type": "Layer",
				"Path": ly.Inhib.Layer.Gi",
				"Start": 1.4,
				"Values": [
					1.3,
					1.5
				]
			}
		]
	},
	{
		"Param": pt.WtScale.Rel",
		"Sel": {
			"Sel": ".Back",
			"Doc": "top-down back-pathways MUST have lower relative weight scale, otherwise network hallucinates",
			"Params": {
				pt.WtScale.Rel": "0.2"
			},
			"Hypers": {
				pt.WtScale.Rel": {
					"Tweak": "log"
				}
			}
		},
		"Search": [
			{
				"Name": "HiddenToInput",
				"Type": "Path",
				"Path": pt.WtScale.Rel",
				"Start": 0.2,
				"Values": [
					0.1,
					0.5
				]
			}
		]
	},
	{
		"Param": ly.Inhib.Layer.Gi",
		"Sel": {
			"Sel": "Layer",
			"Doc": "using default 1.8 inhib for all of network -- can explore",
			"Params": {
				ly.Inhib.Layer.Gi": "1.8"
			},
			"Hypers": {
				ly.Inhib.Layer.Gi": {
					"Tweak": "[1.75, 1.85]"
				}
			}
		},
		"Search": [
			{
				"Name": "Input",
				"Type": "Layer",
				"Path": ly.Inhib.Layer.Gi",
				"Start": 1.8,
				"Values": [
					1.75,
					1.85
				]
			}
		]
	},
	{
		"Param": pt.Learn.LRate",
		"Sel": {
			"Sel": "Path",
			"Doc": "norm and momentum on works better, but wt bal is not better for smaller nets",
			"Params": {
				pt.Learn.LRate": "0.02",
				pt.Learn.Momentum": "0.9"
			},
			"Hypers": {
				pt.Learn.LRate": {
					"Tweak": "log"
				},
				pt.Learn.Momentum": {
					"Tweak": "incr"
				}
			}
		},
		"Search": [
			{
				"Name": "HiddenToInput",
				"Type": "Path",
				"Path": pt.Learn.LRate",
				"Start": 0.02,
				"Values": [
					0.01,
					0.05
				]
			},
			{
				"Name": "InputToHidden",
				"Type": "Path",
				"Path": pt.Learn.LRate",
				"Start": 0.02,
				"Values": [
					0.01,
					0.05
				]
			}
		]
	},
	{
		"Param": pt.Learn.Momentum",
		"Sel": {
			"Sel": "Path",
			"Doc": "norm and momentum on works better, but wt bal is not better for smaller nets",
			"Params": {
				pt.Learn.LRate": "0.02",
				pt.Learn.Momentum": "0.9"
			},
			"Hypers": {
				pt.Learn.LRate": {
					"Tweak": "log"
				},
				pt.Learn.Momentum": {
					"Tweak": "incr"
				}
			}
		},
		"Search": [
			{
				"Name": "HiddenToInput",
				"Type": "Path",
				"Path": pt.Learn.Momentum",
				"Start": 0.9,
				"Values": [
					0.8,
					1
				]
			},
			{
				"Name": "InputToHidden",
				"Type": "Path",
				"Path": pt.Learn.Momentum",
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

*/
