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
	"Base": {
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
	"DefaultInhib": {
		{Sel: "#Output", Desc: "go back to default",
			Params: Params{
				"Layer.Inhib.Layer.Gi": "1.8",
			}},
	},
	"NoMomentum": {
		{Sel: "Path", Desc: "no norm or momentum",
			Params: Params{
				"Path.Learn.Norm.On":     "false",
				"Path.Learn.Momentum.On": "false",
			}},
	},
	"WtBalOn": {
		{Sel: "Path", Desc: "weight bal on",
			Params: Params{
				"Path.Learn.WtBal.On": "true",
			}},
	},
}

var trgCode = `params.Sets{
	"Base": {
		{Sel: "Path", Desc: "norm and momentum on works better, but wt bal is not better for smaller nets",
			Params: params.Params{
				"Path.Learn.Norm.On":     "true",
				"Path.Learn.Momentum.On": "true",
				"Path.Learn.WtBal.On":    "false",
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
		{Sel: ".Back", Desc: "top-down back-pathways MUST have lower relative weight scale, otherwise network hallucinates",
			Params: params.Params{
				"Path.WtScale.Rel": "0.2",
			}},
	},
	"DefaultInhib": {
		{Sel: "#Output", Desc: "go back to default",
			Params: params.Params{
				"Layer.Inhib.Layer.Gi": "1.8",
			}},
	},
	"NoMomentum": {
		{Sel: "Path", Desc: "no norm or momentum",
			Params: params.Params{
				"Path.Learn.Norm.On":     "false",
				"Path.Learn.Momentum.On": "false",
			}},
	},
	"WtBalOn": {
		{Sel: "Path", Desc: "weight bal on",
			Params: params.Params{
				"Path.Learn.WtBal.On": "true",
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
	cval, err := paramSets.ParamValue("Base", "Path", "Path.Learn.WtBal.On")
	if err != nil {
		t.Error(err)
	}
	// fmt.Printf("current value: %s\n", cval)
	if cval != "false" {
		t.Errorf("value should have been false: %s\n", cval)
	}
	err = paramSets.SetString("Base", "Path", "Path.Learn.WtBal.On", "true")
	if err != nil {
		t.Error(err)
	}
	cval, err = paramSets.ParamValue("Base", "Path", "Path.Learn.WtBal.On")
	// fmt.Printf("new value: %s\n", cval)
	if cval != "true" {
		t.Errorf("value should have been true: %s\n", cval)
	}
	err = paramSets.SetFloat("Base", "Path", "Path.Learn.WtBal.On", 5.1)
	if err != nil {
		t.Error(err)
	}
	cval, err = paramSets.ParamValue("Base", "Path", "Path.Learn.WtBal.On")
	// fmt.Printf("new value: %s\n", cval)
	if cval != "5.1" {
		t.Errorf("value should have been 5.1: %s\n", cval)
	}
	cval, err = paramSets.ParamValue("Basre", "Path", "Path.Learn.WtBal.On")
	if err == nil {
		t.Errorf("Should have had an error")
	}
	// fmt.Printf("error: %s\n", err)
	cval, err = paramSets.ParamValue("Base", "Paths", "Path.Learn.WtBal.On")
	if err == nil {
		t.Errorf("Should have had an error")
	}
	// fmt.Printf("error: %s\n", err)
}

var trgHypers = `{
  "Hidden1": {
    "Name": "Hidden1",
    "Type": "Layer",
    "Class": "Hidden",
    "Object": {
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
    "Name": "Hidden2",
    "Type": "Layer",
    "Class": "Hidden",
    "Object": {
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
    "Name": "Input",
    "Type": "Layer",
    "Class": "Input",
    "Object": {
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
    "Name": "Output",
    "Type": "Layer",
    "Class": "Target",
    "Object": {
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
		FlexVal{Name: "Input", Type: "Layer", Class: "Input", Object: Hypers{}},
		FlexVal{Name: "Hidden1", Type: "Layer", Class: "Hidden", Object: Hypers{}},
		FlexVal{Name: "Hidden2", Type: "Layer", Class: "Hidden", Object: Hypers{}},
		FlexVal{Name: "Output", Type: "Layer", Class: "Target", Object: Hypers{}},
	})
	basenet := paramSets["Base"]
	hypers.ApplySheet(basenet, false)

	dfs := hypers.JSONString()
	// fmt.Printf("%s", dfs)
	if dfs != trgHypers {
		t.Errorf("Param hypers output incorrect at: %v!\n", diff.LineDiff(dfs, trgHypers))
		// t.Errorf("ParamStyle output incorrect!\n%v\n", dfs)
	}
}
