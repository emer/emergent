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

var trgSearch = "[\n\t{\n\t\t\"Param\": \"Layer.Inhib.Layer.Gi\",\n\t\t\"Sel\": {\n\t\t\t\"Sel\": \"#Hidden\",\n\t\t\t\"Desc\": \"output definitely needs lower inhib -- true for smaller layers in general\",\n\t\t\t\"Params\": {\n\t\t\t\t\"Layer.Inhib.Layer.Gi\": \"1.4\"\n\t\t\t},\n\t\t\t\"Hypers\": {\n\t\t\t\t\"Layer.Inhib.Layer.Gi\": {\n\t\t\t\t\t\"Tweak\": \"incr\"\n\t\t\t\t}\n\t\t\t}\n\t\t},\n\t\t\"Search\": [\n\t\t\t{\n\t\t\t\t\"Name\": \"Hidden\",\n\t\t\t\t\"Type\": \"Layer\",\n\t\t\t\t\"Path\": \"Layer.Inhib.Layer.Gi\",\n\t\t\t\t\"Start\": 1.4,\n\t\t\t\t\"Values\": [\n\t\t\t\t\t1.3,\n\t\t\t\t\t1.5\n\t\t\t\t]\n\t\t\t}\n\t\t]\n\t},\n\t{\n\t\t\"Param\": \"Path.WtScale.Rel\",\n\t\t\"Sel\": {\n\t\t\t\"Sel\": \".Back\",\n\t\t\t\"Desc\": \"top-down back-pathways MUST have lower relative weight scale, otherwise network hallucinates\",\n\t\t\t\"Params\": {\n\t\t\t\t\"Path.WtScale.Rel\": \"0.2\"\n\t\t\t},\n\t\t\t\"Hypers\": {\n\t\t\t\t\"Path.WtScale.Rel\": {\n\t\t\t\t\t\"Tweak\": \"log\"\n\t\t\t\t}\n\t\t\t}\n\t\t},\n\t\t\"Search\": [\n\t\t\t{\n\t\t\t\t\"Name\": \"HiddenToInput\",\n\t\t\t\t\"Type\": \"Path\",\n\t\t\t\t\"Path\": \"Path.WtScale.Rel\",\n\t\t\t\t\"Start\": 0.2,\n\t\t\t\t\"Values\": [\n\t\t\t\t\t0.1,\n\t\t\t\t\t0.5\n\t\t\t\t]\n\t\t\t}\n\t\t]\n\t},\n\t{\n\t\t\"Param\": \"Layer.Inhib.Layer.Gi\",\n\t\t\"Sel\": {\n\t\t\t\"Sel\": \"Layer\",\n\t\t\t\"Desc\": \"using default 1.8 inhib for all of network -- can explore\",\n\t\t\t\"Params\": {\n\t\t\t\t\"Layer.Inhib.Layer.Gi\": \"1.8\"\n\t\t\t},\n\t\t\t\"Hypers\": {\n\t\t\t\t\"Layer.Inhib.Layer.Gi\": {\n\t\t\t\t\t\"Tweak\": \"[1.75, 1.85]\"\n\t\t\t\t}\n\t\t\t}\n\t\t},\n\t\t\"Search\": [\n\t\t\t{\n\t\t\t\t\"Name\": \"Input\",\n\t\t\t\t\"Type\": \"Layer\",\n\t\t\t\t\"Path\": \"Layer.Inhib.Layer.Gi\",\n\t\t\t\t\"Start\": 1.8,\n\t\t\t\t\"Values\": [\n\t\t\t\t\t1.75,\n\t\t\t\t\t1.85\n\t\t\t\t]\n\t\t\t}\n\t\t]\n\t},\n\t{\n\t\t\"Param\": \"Path.Learn.LRate\",\n\t\t\"Sel\": {\n\t\t\t\"Sel\": \"Path\",\n\t\t\t\"Desc\": \"norm and momentum on works better, but wt bal is not better for smaller nets\",\n\t\t\t\"Params\": {\n\t\t\t\t\"Path.Learn.LRate\": \"0.02\",\n\t\t\t\t\"Path.Learn.Momentum\": \"0.9\"\n\t\t\t},\n\t\t\t\"Hypers\": {\n\t\t\t\t\"Path.Learn.LRate\": {\n\t\t\t\t\t\"Tweak\": \"log\"\n\t\t\t\t},\n\t\t\t\t\"Path.Learn.Momentum\": {\n\t\t\t\t\t\"Tweak\": \"incr\"\n\t\t\t\t}\n\t\t\t}\n\t\t},\n\t\t\"Search\": [\n\t\t\t{\n\t\t\t\t\"Name\": \"HiddenToInput\",\n\t\t\t\t\"Type\": \"Path\",\n\t\t\t\t\"Path\": \"Path.Learn.LRate\",\n\t\t\t\t\"Start\": 0.02,\n\t\t\t\t\"Values\": [\n\t\t\t\t\t0.01,\n\t\t\t\t\t0.05\n\t\t\t\t]\n\t\t\t},\n\t\t\t{\n\t\t\t\t\"Name\": \"InputToHidden\",\n\t\t\t\t\"Type\": \"Path\",\n\t\t\t\t\"Path\": \"Path.Learn.LRate\",\n\t\t\t\t\"Start\": 0.02,\n\t\t\t\t\"Values\": [\n\t\t\t\t\t0.01,\n\t\t\t\t\t0.05\n\t\t\t\t]\n\t\t\t}\n\t\t]\n\t},\n\t{\n\t\t\"Param\": \"Path.Learn.Momentum\",\n\t\t\"Sel\": {\n\t\t\t\"Sel\": \"Path\",\n\t\t\t\"Desc\": \"norm and momentum on works better, but wt bal is not better for smaller nets\",\n\t\t\t\"Params\": {\n\t\t\t\t\"Path.Learn.LRate\": \"0.02\",\n\t\t\t\t\"Path.Learn.Momentum\": \"0.9\"\n\t\t\t},\n\t\t\t\"Hypers\": {\n\t\t\t\t\"Path.Learn.LRate\": {\n\t\t\t\t\t\"Tweak\": \"log\"\n\t\t\t\t},\n\t\t\t\t\"Path.Learn.Momentum\": {\n\t\t\t\t\t\"Tweak\": \"incr\"\n\t\t\t\t}\n\t\t\t}\n\t\t},\n\t\t\"Search\": [\n\t\t\t{\n\t\t\t\t\"Name\": \"HiddenToInput\",\n\t\t\t\t\"Type\": \"Path\",\n\t\t\t\t\"Path\": \"Path.Learn.Momentum\",\n\t\t\t\t\"Start\": 0.9,\n\t\t\t\t\"Values\": [\n\t\t\t\t\t0.8,\n\t\t\t\t\t1\n\t\t\t\t]\n\t\t\t},\n\t\t\t{\n\t\t\t\t\"Name\": \"InputToHidden\",\n\t\t\t\t\"Type\": \"Path\",\n\t\t\t\t\"Path\": \"Path.Learn.Momentum\",\n\t\t\t\t\"Start\": 0.9,\n\t\t\t\t\"Values\": [\n\t\t\t\t\t0.8,\n\t\t\t\t\t1\n\t\t\t\t]\n\t\t\t}\n\t\t]\n\t}\n]\n"

func TestTweakHypers(t *testing.T) {
	hypers := Flex{}
	hypers.Init([]FlexVal{
		FlexVal{Nm: "Input", Type: "Layer", Cls: "Input", Obj: Hypers{}},
		FlexVal{Nm: "Hidden", Type: "Layer", Cls: "Hidden", Obj: Hypers{}},
		FlexVal{Nm: "InputToHidden", Type: "Path", Cls: "Forward", Obj: Hypers{}},
		FlexVal{Nm: "HiddenToInput", Type: "Path", Cls: "Back", Obj: Hypers{}},
	})
	basenet := tweakSets["Base"]
	hypers.ApplySheet(basenet, false)

	// fmt.Println("hypers:", reflectx.StringJSON(hypers))

	srch := TweaksFromHypers(hypers)
	ss := reflectx.StringJSON(srch)
	// fmt.Println("\n\n##########\n", ss)
	assert.Equal(t, trgSearch, ss)
}
