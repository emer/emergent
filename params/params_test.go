// Copyright (c) 2019, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package params

import (
	"bytes"
	"testing"

	"github.com/andreyvit/diff"
)

var paramSets = Sets{
	"Base": {Desc: "these are the best params", Sheets: Sheets{
		"Network": &Sheet{
			{Sel: "Path", Desc: "norm and momentum on works better, but wt bal is not better for smaller nets",
				Params: Params{
					"Path.Learn.Norm.On":     "true",
					"Path.Learn.Momentum.On": "true",
					"Path.Learn.WtBal.On":    "false",
				}},
			{Sel: "Layer", Desc: "using default 1.8 inhib for all of network -- can explore",
				Params: Params{
					"Layer.Inhib.Layer.Gi": "1.8",
				},
				Hypers: Hypers{
					"Layer.Inhib.Layer.Gi": {"Min": "0.5", "StdDev": "0.1"},
				},
			},
			{Sel: "#Output", Desc: "output definitely needs lower inhib -- true for smaller layers in general",
				Params: Params{
					"Layer.Inhib.Layer.Gi": "1.4",
				}},
			{Sel: ".Back", Desc: "top-down back-pathways MUST have lower relative weight scale, otherwise network hallucinates",
				Params: Params{
					"Path.WtScale.Rel": "0.2",
				}},
		},
		"Sim": &Sheet{ // sim params apply to sim object
			{Sel: "Sim", Desc: "best params always finish in this time",
				Params: Params{
					"Sim.MaxEpcs": "50",
				}},
		},
	}},
	"DefaultInhib": {Desc: "output uses default inhib instead of lower", Sheets: Sheets{
		"Network": &Sheet{
			{Sel: "#Output", Desc: "go back to default",
				Params: Params{
					"Layer.Inhib.Layer.Gi": "1.8",
				}},
		},
		"Sim": &Sheet{ // sim params apply to sim object
			{Sel: "Sim", Desc: "takes longer -- generally doesn't finish..",
				Params: Params{
					"Sim.MaxEpcs": "100",
				}, Hypers: Hypers{
					"Sim.MaxEps": {"Val": "90", "Min": "40", "Max": "2000"},
				}},
		},
	}},
	"NoMomentum": {Desc: "no momentum or normalization", Sheets: Sheets{
		"Network": &Sheet{
			{Sel: "Path", Desc: "no norm or momentum",
				Params: Params{
					"Path.Learn.Norm.On":     "false",
					"Path.Learn.Momentum.On": "false",
				}},
		},
	}},
	"WtBalOn": {Desc: "try with weight bal on", Sheets: Sheets{
		"Network": &Sheet{
			{Sel: "Path", Desc: "weight bal on",
				Params: Params{
					"Path.Learn.WtBal.On": "true",
				}},
		},
	}},
}

var trgCode = `params.Sets{
	{Desc: "these are the best params", Sheets: params.Sheets{
		"Network": &params.Sheet{
			{Sel: "Path", Desc: "norm and momentum on works better, but wt bal is not better for smaller nets",
				Params: params.Params{
					"Path.Learn.Momentum.On": "true",
					"Path.Learn.Norm.On": "true",
					"Path.Learn.WtBal.On": "false",
				}},
			{Sel: "Layer", Desc: "using default 1.8 inhib for all of network -- can explore",
				Params: params.Params{
					"Layer.Inhib.Layer.Gi": "1.8",
				}params.Hypers{
					"Layer.Inhib.Layer.Gi": map["Min":"0.5" "StdDev":"0.1"],
				}},
			{Sel: "#Output", Desc: "output definitely needs lower inhib -- true for smaller layers in general",
				Params: params.Params{
					"Layer.Inhib.Layer.Gi": "1.4",
				}},
			{Sel: ".Back", Desc: "top-down back-pathways MUST have lower relative weight scale, otherwise network hallucinates",
				Params: params.Params{
					"Path.WtScale.Rel": "0.2",
				}},
		},
		"Sim": &params.Sheet{
			{Sel: "Sim", Desc: "best params always finish in this time",
				Params: params.Params{
					"Sim.MaxEpcs": "50",
				}},
		},
	}},
	{Desc: "output uses default inhib instead of lower", Sheets: params.Sheets{
		"Network": &params.Sheet{
			{Sel: "#Output", Desc: "go back to default",
				Params: params.Params{
					"Layer.Inhib.Layer.Gi": "1.8",
				}},
		},
		"Sim": &params.Sheet{
			{Sel: "Sim", Desc: "takes longer -- generally doesn't finish..",
				Params: params.Params{
					"Sim.MaxEpcs": "100",
				}params.Hypers{
					"Sim.MaxEps": map["Max":"2000" "Min":"40" "Val":"90"],
				}},
		},
	}},
	{Desc: "no momentum or normalization", Sheets: params.Sheets{
		"Network": &params.Sheet{
			{Sel: "Path", Desc: "no norm or momentum",
				Params: params.Params{
					"Path.Learn.Momentum.On": "false",
					"Path.Learn.Norm.On": "false",
				}},
		},
	}},
	{Desc: "try with weight bal on", Sheets: params.Sheets{
		"Network": &params.Sheet{
			{Sel: "Path", Desc: "weight bal on",
				Params: params.Params{
					"Path.Learn.WtBal.On": "true",
				}},
		},
	}},
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
	cval, err := paramSets.ParamValue("Base", "Network", "Path", "Path.Learn.WtBal.On")
	if err != nil {
		t.Error(err)
	}
	// fmt.Printf("current value: %s\n", cval)
	if cval != "false" {
		t.Errorf("value should have been false: %s\n", cval)
	}
	err = paramSets.SetString("Base", "Network", "Path", "Path.Learn.WtBal.On", "true")
	if err != nil {
		t.Error(err)
	}
	cval, err = paramSets.ParamValue("Base", "Network", "Path", "Path.Learn.WtBal.On")
	// fmt.Printf("new value: %s\n", cval)
	if cval != "true" {
		t.Errorf("value should have been true: %s\n", cval)
	}
	err = paramSets.SetFloat("Base", "Network", "Path", "Path.Learn.WtBal.On", 5.1)
	if err != nil {
		t.Error(err)
	}
	cval, err = paramSets.ParamValue("Base", "Network", "Path", "Path.Learn.WtBal.On")
	// fmt.Printf("new value: %s\n", cval)
	if cval != "5.1" {
		t.Errorf("value should have been 5.1: %s\n", cval)
	}
	cval, err = paramSets.ParamValue("Basre", "Network2", "Path", "Path.Learn.WtBal.On")
	if err == nil {
		t.Errorf("Should have had an error")
	}
	// fmt.Printf("error: %s\n", err)
	cval, err = paramSets.ParamValue("Base", "Network2", "Path", "Path.Learn.WtBal.On")
	if err == nil {
		t.Errorf("Should have had an error")
	}
	cval, err = paramSets.ParamValue("Base", "Network", "Paths", "Path.Learn.WtBal.On")
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
    },
    "History": [
      {
        "Sel": "Layer",
        "Desc": "using default 1.8 inhib for all of network -- can explore",
        "Params": {
          "Layer.Inhib.Layer.Gi": "1.8"
        },
        "Hypers": {
          "Layer.Inhib.Layer.Gi": {
            "Min": "0.5",
            "StdDev": "0.1"
          }
        }
      }
    ]
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
    },
    "History": [
      {
        "Sel": "Layer",
        "Desc": "using default 1.8 inhib for all of network -- can explore",
        "Params": {
          "Layer.Inhib.Layer.Gi": "1.8"
        },
        "Hypers": {
          "Layer.Inhib.Layer.Gi": {
            "Min": "0.5",
            "StdDev": "0.1"
          }
        }
      }
    ]
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
    },
    "History": [
      {
        "Sel": "Layer",
        "Desc": "using default 1.8 inhib for all of network -- can explore",
        "Params": {
          "Layer.Inhib.Layer.Gi": "1.8"
        },
        "Hypers": {
          "Layer.Inhib.Layer.Gi": {
            "Min": "0.5",
            "StdDev": "0.1"
          }
        }
      }
    ]
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
    },
    "History": [
      {
        "Sel": "Layer",
        "Desc": "using default 1.8 inhib for all of network -- can explore",
        "Params": {
          "Layer.Inhib.Layer.Gi": "1.8"
        },
        "Hypers": {
          "Layer.Inhib.Layer.Gi": {
            "Min": "0.5",
            "StdDev": "0.1"
          }
        }
      },
      {
        "Sel": "#Output",
        "Desc": "output definitely needs lower inhib -- true for smaller layers in general",
        "Params": {
          "Layer.Inhib.Layer.Gi": "1.4"
        },
        "Hypers": null
      }
    ]
  }
}`

func TestFlexHypers(t *testing.T) {
	hypers := Flex{}
	hypers.Init([]FlexVal{
		FlexVal{Nm: "Input", Type: "Layer", Cls: "Input", Obj: Hypers{}},
		FlexVal{Nm: "Hidden1", Type: "Layer", Cls: "Hidden", Obj: Hypers{}},
		FlexVal{Nm: "Hidden2", Type: "Layer", Cls: "Hidden", Obj: Hypers{}},
		FlexVal{Nm: "Output", Type: "Layer", Cls: "Target", Obj: Hypers{}},
	})
	basenet := paramSets.SetByName("Base").Sheets["Network"]
	hypers.ApplySheet(basenet, false)

	dfs := hypers.JSONString()
	// fmt.Printf("%s", dfs)
	if dfs != trgHypers {
		t.Errorf("Param hypers output incorrect at: %v!\n", diff.LineDiff(dfs, trgHypers))
		// t.Errorf("ParamStyle output incorrect!\n%v\n", dfs)
	}
}
