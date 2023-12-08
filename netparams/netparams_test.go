// Copyright (c) 2019, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package netparams

import (
	"bytes"
	"testing"

	"github.com/andreyvit/diff"
	"github.com/emer/emergent/v2/params"
	// "github.com/andreyvit/diff"
)

var paramSets = Sets{
	"Base": {
		{Sel: "Prjn", Desc: "norm and momentum on works better, but wt bal is not better for smaller nets",
			Params: params.Params{
				"Prjn.Learn.Norm.On":     "true",
				"Prjn.Learn.Momentum.On": "true",
				"Prjn.Learn.WtBal.On":    "false",
			}},
		{Sel: "Layer", Desc: "using default 1.8 inhib for all of network -- can explore",
			Params: params.Params{
				"Layer.Inhib.Layer.Gi": "1.8",
			},
			Hypers: params.Hypers{
				"Layer.Inhib.Layer.Gi": {"Min": "0.5", "StdDev": "0.1"},
			},
		},
		{Sel: "#Output", Desc: "output definitely needs lower inhib -- true for smaller layers in general",
			Params: params.Params{
				"Layer.Inhib.Layer.Gi": "1.4",
			}},
		{Sel: ".Back", Desc: "top-down back-projections MUST have lower relative weight scale, otherwise network hallucinates",
			Params: params.Params{
				"Prjn.WtScale.Rel": "0.2",
			}},
	},
	"DefaultInhib": {
		{Sel: "#Output", Desc: "go back to default",
			Params: params.Params{
				"Layer.Inhib.Layer.Gi": "1.8",
			}},
	},
	"NoMomentum": {
		{Sel: "Prjn", Desc: "no norm or momentum",
			Params: params.Params{
				"Prjn.Learn.Norm.On":     "false",
				"Prjn.Learn.Momentum.On": "false",
			}},
	},
	"WtBalOn": {
		{Sel: "Prjn", Desc: "weight bal on",
			Params: params.Params{
				"Prjn.Learn.WtBal.On": "true",
			}},
	},
}

var trgCode = `netparams.Sets{
	"Base": {
		{Sel: "Prjn", Desc: "norm and momentum on works better, but wt bal is not better for smaller nets",
			Params: params.Params{
				"Prjn.Learn.Norm.On":     "true",
				"Prjn.Learn.Momentum.On": "true",
				"Prjn.Learn.WtBal.On":    "false",
			}},
		{Sel: "Layer", Desc: "using default 1.8 inhib for all of network -- can explore",
			Params: params.Params{
				"Layer.Inhib.Layer.Gi": "1.8",
			},
			Hypers: params.Hypers{
				"Layer.Inhib.Layer.Gi": {"Min": "0.5", "StdDev": "0.1"},
			},
		},
		{Sel: "#Output", Desc: "output definitely needs lower inhib -- true for smaller layers in general",
			Params: params.Params{
				"Layer.Inhib.Layer.Gi": "1.4",
			}},
		{Sel: ".Back", Desc: "top-down back-projections MUST have lower relative weight scale, otherwise network hallucinates",
			Params: params.Params{
				"Prjn.WtScale.Rel": "0.2",
			}},
	},
	"DefaultInhib": {
		{Sel: "#Output", Desc: "go back to default",
			Params: params.Params{
				"Layer.Inhib.Layer.Gi": "1.8",
			}},
	},
	"NoMomentum": {
		{Sel: "Prjn", Desc: "no norm or momentum",
			Params: params.Params{
				"Prjn.Learn.Norm.On":     "false",
				"Prjn.Learn.Momentum.On": "false",
			}},
	},
	"WtBalOn": {
		{Sel: "Prjn", Desc: "weight bal on",
			Params: params.Params{
				"Prjn.Learn.WtBal.On": "true",
			}},
	},
}
`

func TestParamSetsWriteGo(t *testing.T) {
	t.Skip("todo: need to sort the map for this to work now")
	var buf bytes.Buffer
	paramSets.WriteGoCode(&buf, 0)
	dfb := buf.Bytes()
	dfs := string(dfb)
	// fmt.Printf("%v", dfs)
	if dfs != trgCode {
		t.Errorf("ParamStyle output incorrect at: %v!\n", diff.LineDiff(dfs, trgCode))
		// t.Errorf("ParamStyle output incorrect!\n%v\n", dfs)
	}
}

func TestParamSetsSet(t *testing.T) {
	cval, err := paramSets.ParamVal("Base", "Prjn", "Prjn.Learn.WtBal.On")
	if err != nil {
		t.Error(err)
	}
	// fmt.Printf("current value: %s\n", cval)
	if cval != "false" {
		t.Errorf("value should have been false: %s\n", cval)
	}
	err = paramSets.SetString("Base", "Prjn", "Prjn.Learn.WtBal.On", "true")
	if err != nil {
		t.Error(err)
	}
	cval, err = paramSets.ParamVal("Base", "Prjn", "Prjn.Learn.WtBal.On")
	// fmt.Printf("new value: %s\n", cval)
	if cval != "true" {
		t.Errorf("value should have been true: %s\n", cval)
	}
	err = paramSets.SetFloat("Base", "Prjn", "Prjn.Learn.WtBal.On", 5.1)
	if err != nil {
		t.Error(err)
	}
	cval, err = paramSets.ParamVal("Base", "Prjn", "Prjn.Learn.WtBal.On")
	// fmt.Printf("new value: %s\n", cval)
	if cval != "5.1" {
		t.Errorf("value should have been 5.1: %s\n", cval)
	}
	cval, err = paramSets.ParamVal("Basre", "Prjn", "Prjn.Learn.WtBal.On")
	if err == nil {
		t.Errorf("Should have had an error")
	}
	// fmt.Printf("error: %s\n", err)
	cval, err = paramSets.ParamVal("Base", "Prjns", "Prjn.Learn.WtBal.On")
	if err == nil {
		t.Errorf("Should have had an error")
	}
	// fmt.Printf("error: %s\n", err)
}

var trgHypers = `{
  "Hidden1": {
    "Nm": "Hidden1",
    "Type": "Layer",
    "Cls": "Hidden",
    "Obj": {
      "Layer.Inhib.Layer.Gi": {
        "Min": "0.5",
        "StdDev": "0.1",
        "Val": "1.8"
      }
    }
  },
  "Hidden2": {
    "Nm": "Hidden2",
    "Type": "Layer",
    "Cls": "Hidden",
    "Obj": {
      "Layer.Inhib.Layer.Gi": {
        "Min": "0.5",
        "StdDev": "0.1",
        "Val": "1.8"
      }
    }
  },
  "Input": {
    "Nm": "Input",
    "Type": "Layer",
    "Cls": "Input",
    "Obj": {
      "Layer.Inhib.Layer.Gi": {
        "Min": "0.5",
        "StdDev": "0.1",
        "Val": "1.8"
      }
    }
  },
  "Output": {
    "Nm": "Output",
    "Type": "Layer",
    "Cls": "Target",
    "Obj": {
      "Layer.Inhib.Layer.Gi": {
        "Min": "0.5",
        "StdDev": "0.1",
        "Val": "1.4"
      }
    }
  }
}`
