// Copyright (c) 2019, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package params

import (
	"testing"

	"cogentcore.org/core/laser"
	"github.com/andreyvit/diff"
)

var tweakSets = Sets{
	"Base": {Desc: "these are the best params", Sheets: Sheets{
		"Network": &Sheet{
			{Sel: "Prjn", Desc: "norm and momentum on works better, but wt bal is not better for smaller nets",
				Params: Params{
					"Prjn.Learn.LRate":    "0.02",
					"Prjn.Learn.Momentum": "0.9",
				},
				Hypers: Hypers{
					"Prjn.Learn.LRate":    {"Tweak": "log"},
					"Prjn.Learn.Momentum": {"Tweak": "incr"},
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
			{Sel: ".Back", Desc: "top-down back-projections MUST have lower relative weight scale, otherwise network hallucinates",
				Params: Params{
					"Prjn.WtScale.Rel": "0.2",
				},
				Hypers: Hypers{
					"Prjn.WtScale.Rel": {"Tweak": "log"},
				}},
		},
	}},
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
    "Name": "Hidden",
    "Type": "Layer",
    "Path": "Layer.Inhib.Layer.Gi",
    "Start": 1.4,
    "Values": [
      1.3,
      1.5
    ]
  },
  {
    "Name": "HiddenToInput",
    "Type": "Prjn",
    "Path": "Prjn.Learn.LRate",
    "Start": 0.02,
    "Values": [
      0.01,
      0.05
    ]
  },
  {
    "Name": "HiddenToInput",
    "Type": "Prjn",
    "Path": "Prjn.Learn.Momentum",
    "Start": 0.9,
    "Values": [
      0.8,
      1
    ]
  },
  {
    "Name": "HiddenToInput",
    "Type": "Prjn",
    "Path": "Prjn.WtScale.Rel",
    "Start": 0.2,
    "Values": [
      0.1,
      0.5
    ]
  },
  {
    "Name": "Input",
    "Type": "Layer",
    "Path": "Layer.Inhib.Layer.Gi",
    "Start": 1.8,
    "Values": [
      1.75,
      1.85
    ]
  },
  {
    "Name": "InputToHidden",
    "Type": "Prjn",
    "Path": "Prjn.Learn.LRate",
    "Start": 0.02,
    "Values": [
      0.01,
      0.05
    ]
  },
  {
    "Name": "InputToHidden",
    "Type": "Prjn",
    "Path": "Prjn.Learn.Momentum",
    "Start": 0.9,
    "Values": [
      0.8,
      1
    ]
  }
]`

func TestTweakHypers(t *testing.T) {
	hypers := Flex{}
	hypers.Init([]FlexVal{
		FlexVal{Nm: "Input", Type: "Layer", Cls: "Input", Obj: Hypers{}},
		FlexVal{Nm: "Hidden", Type: "Layer", Cls: "Hidden", Obj: Hypers{}},
		FlexVal{Nm: "InputToHidden", Type: "Prjn", Cls: "Forward", Obj: Hypers{}},
		FlexVal{Nm: "HiddenToInput", Type: "Prjn", Cls: "Back", Obj: Hypers{}},
	})
	basenet := tweakSets.SetByName("Base").Sheets["Network"]
	hypers.ApplySheet(basenet, false)

	srch := TweakFromHypers(hypers)
	ss := laser.StringJSON(srch)
	// fmt.Printf("%s", ss)
	if ss != trgSearch {
		t.Errorf("Test Tweak Search output incorrect at: %v!\n", diff.LineDiff(ss, trgSearch))
	}
}
