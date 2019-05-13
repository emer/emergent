// Copyright (c) 2019, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package emer

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"reflect"
	"strings"

	"github.com/goki/gi/gi"
	"github.com/goki/ki/ki"
	"github.com/goki/ki/kit"
)

// note: using interface{} map here probably doesn't make sense, like it did in GoGi
// it requires managing the possible types that can be created in the interface{}
// and that adds a lot of complexity.  Simpler to just have basic fixed float32
// values and aggregations thereof.

// Params is a name-value map for floating point parameter values that can be applied
// to network layers or prjns, which is where the parameter values live.
// The name must be a dot-separated path to a specific parameter, e.g., Prjn.Learn.Lrate
// The first part of the path is the overall target object type, e.g., "Prjn" or "Layer".
// All of the params in one map must apply to the same target type.
type Params map[string]float32

var KiT_Params = kit.Types.AddType(&Params{}, ParamsProps)

// ParamSel specifies a selector for the scope of application of a set of
// parameters, using standard css selector syntax (. prefix = class, # prefix = name,
// and no prefix = type)
type ParamSel struct {
	Sel    string `desc:"selector for what to apply the parameters to, using standard css selector syntax: .Example applies to anything with a Class tag of 'Example', #Example applies to anything with a Name of 'Example', and Example with no prefix applies to anything of type 'Example' (e.g., typically Prjn or Layer are the only relevant types)"`
	Params Params `desc:"parameter values to apply to whatever matches the selector"`
}

var KiT_ParamSel = kit.Types.AddType(&ParamSel{}, ParamSelProps)

// ParamStyle is a CSS-like collection of Params values, each of which represents a different
// set of specific parameter values applied according to the Sel selector.
// .Class #Name or Type where Type is either Prjn or Layer to apply
// to every instance of that type in the network.
//
// The order of elements in the ParamStyle list is critical, as they are applied
// in the order given by the list (slice), and thus later ParamSel's can override
// those applied earlier.  Thus, you generally want to have more general Type-level
// parameters listed first, and then subsequently more specific ones (.Class and #Name)
type ParamStyle []ParamSel

var KiT_ParamStyle = kit.Types.AddType(&ParamStyle{}, ParamStyleProps)

// ParamSet is a collection of ParamStyle's that constitute a coherent set of parameters --
// a particular specific configuration of parameters.
// Relative to the basic ParamStyle, the ParamSet allows for separately-named subsets
// of parameters, to organize and manage more complex collections of parameters.
// Typically the different subsets apply to different parts or aspects of the model.
// Note that there is NO deterministic ordering of these sets due to the use of
// a Go map structure, which specifically randomizes order.  Thus, it is important
// that each subset apply to a different part of the network, or else unpredictable
// overwriting of parameters can occur.  Alternatively, different subsets can be
// specifically applied in a programmatically-specified order, instead of using
// the generic method that applies all of them.
type ParamSet map[string]ParamStyle

var KiT_ParamSet = kit.Types.AddType(&ParamSet{}, ParamSetProps)

// ParamSets is a collection of ParamSet's that can be chosen among
// depending on different desired configurations etc.  Thus, each ParamSet
// represents a collection of different possible specific configurations,
// and different such configurations can be chosen by name to apply as desired.
type ParamSets map[string]ParamSet

var KiT_ParamSets = kit.Types.AddType(&ParamSets{}, ParamSetsProps)

///////////////////////////////////////////////////////////////////////
//  Params

// FindParam parses the path and recursively tries to find the parameter pointed to
// by the path (dot-delimited field names).  Returns false if not found (emits error messages).
func FindParam(val reflect.Value, path string) (reflect.Value, bool) {
	npv := kit.NonPtrValue(val)
	if npv.Kind() != reflect.Struct {
		log.Printf("Params.FindParam: object is not a struct: %v kind: %v -- params must be on structs, path: %v\n", npv.String(), npv.Kind(), path)
		return npv, false
	}
	paths := strings.Split(path, ".")
	fnm := paths[0]
	fld := npv.FieldByName(fnm)
	if !fld.IsValid() {
		log.Printf("Params.FindParam: could not find Field named: %v in struct: %v kind: %v, path: %v\n", fnm, npv.String(), npv.Kind(), path)
		return fld, false
	}
	if len(paths) == 1 {
		return fld.Addr(), true
	}
	return FindParam(fld.Addr(), strings.Join(paths[1:], ".")) // need addr
}

// SetParam sets parameter at given path on given object to given value
// converts the float32 val as appropriate for target type.
// returns true if successful (error messages logged on failure)
func SetParam(obj interface{}, path string, val float32) bool {
	fld, ok := FindParam(reflect.ValueOf(obj), path)
	if !ok {
		return false
	}
	npf := kit.NonPtrValue(fld)
	switch npf.Kind() {
	case reflect.Float64:
		fallthrough
	case reflect.Float32:
		npf.SetFloat(float64(val))
		return true
	case reflect.Bool:
		npf.SetBool((val != 0))
		return true
	}
	return false
}

// todo: Get versions too, to read current values

// Target returns the first part of the path, indicating what type of overall object the params
// apply to.  Uses the first item in the map -- everything in the map must have the same target.
func (pr *Params) Target() string {
	for pt := range *pr {
		return strings.Split(pt, ".")[0]
	}
	return ""
}

// Path returns the second part of the path after the target, indicating the path to the specific
// parameter being set
func (pr *Params) Path(path string) string {
	return strings.Join(strings.Split(path, ".")[1:], ".")
}

// Set applies all parameter values to given object.
// object must already be the appropriate target type based on the first element of the path
// (see Target method).  if setMsg is true, then it will print a confirmation that the parameter
// was set (it always prints an error message if it fails to set the parameter at given path)
func (pr *Params) Set(obj interface{}, setMsg bool) {
	olbl := ""
	olblr, haslbl := obj.(gi.Labeler)
	if haslbl {
		olbl = olblr.Label()
	}
	for pt, v := range *pr {
		path := pr.Path(pt)
		ok := SetParam(obj, path, v)
		if ok {
			if setMsg {
				log.Printf("%v Set param path: %v to value: %v\n", olbl, pt, v)
			}
		} else {
			log.Printf("%v Failed to set param path: %v to value: %v\n", olbl, pt, v)
		}
	}
}

///////////////////////////////////////////////////////////////////////
//  ParamStyle

// StyleMatch returns true if given selector matches the target object properties
// (name, class, type name).  Class can be space separated list of names.
func StyleMatch(sel string, name, cls, typ string) bool {
	if sel == "" {
		return false
	}
	if sel[0] == '.' { // class
		return ClassMatch(sel[1:], cls)
	}
	if sel[0] == '#' { // name
		return name == sel[1:]
	}
	return typ == sel // type
}

// ClassMatch returns true if given class names -- handles space-separated multiple class names
func ClassMatch(sel, cls string) bool {
	clss := strings.Split(cls, " ")
	for _, cl := range clss {
		if strings.TrimSpace(cl) == sel {
			return true
		}
	}
	return false
}

///////////////////////////////////////////////////////////////////////
//  I/O

// OpenJSON opens params from a JSON-formatted file.
func (pr *Params) OpenJSON(filename gi.FileName) error {
	*pr = make(Params) // reset
	b, err := ioutil.ReadFile(string(filename))
	if err != nil {
		gi.PromptDialog(nil, gi.DlgOpts{Title: "File Not Found", Prompt: err.Error()}, true, false, nil, nil)
		log.Println(err)
		return err
	}
	return json.Unmarshal(b, pr)
}

// SaveJSON saves params to a JSON-formatted file.
func (pr *Params) SaveJSON(filename gi.FileName) error {
	b, err := json.MarshalIndent(pr, "", "  ")
	if err != nil {
		log.Println(err) // unlikely
		return err
	}
	err = ioutil.WriteFile(string(filename), b, 0644)
	if err != nil {
		gi.PromptDialog(nil, gi.DlgOpts{Title: "Could not Save to File", Prompt: err.Error()}, true, false, nil, nil)
		log.Println(err)
	}
	return err
}

// OpenJSON opens params from a JSON-formatted file.
func (pr *ParamSel) OpenJSON(filename gi.FileName) error {
	b, err := ioutil.ReadFile(string(filename))
	if err != nil {
		gi.PromptDialog(nil, gi.DlgOpts{Title: "File Not Found", Prompt: err.Error()}, true, false, nil, nil)
		log.Println(err)
		return err
	}
	return json.Unmarshal(b, pr)
}

// SaveJSON saves params to a JSON-formatted file.
func (pr *ParamSel) SaveJSON(filename gi.FileName) error {
	b, err := json.MarshalIndent(pr, "", "  ")
	if err != nil {
		log.Println(err) // unlikely
		return err
	}
	err = ioutil.WriteFile(string(filename), b, 0644)
	if err != nil {
		gi.PromptDialog(nil, gi.DlgOpts{Title: "Could not Save to File", Prompt: err.Error()}, true, false, nil, nil)
		log.Println(err)
	}
	return err
}

// OpenJSON opens params from a JSON-formatted file.
func (pr *ParamStyle) OpenJSON(filename gi.FileName) error {
	*pr = make(ParamStyle, 0) // reset
	b, err := ioutil.ReadFile(string(filename))
	if err != nil {
		gi.PromptDialog(nil, gi.DlgOpts{Title: "File Not Found", Prompt: err.Error()}, true, false, nil, nil)
		log.Println(err)
		return err
	}
	return json.Unmarshal(b, pr)
}

// SaveJSON saves params to a JSON-formatted file.
func (pr *ParamStyle) SaveJSON(filename gi.FileName) error {
	b, err := json.MarshalIndent(pr, "", "  ")
	if err != nil {
		log.Println(err) // unlikely
		return err
	}
	err = ioutil.WriteFile(string(filename), b, 0644)
	if err != nil {
		gi.PromptDialog(nil, gi.DlgOpts{Title: "Could not Save to File", Prompt: err.Error()}, true, false, nil, nil)
		log.Println(err)
	}
	return err
}

// OpenJSON opens params from a JSON-formatted file.
func (pr *ParamSet) OpenJSON(filename gi.FileName) error {
	*pr = make(ParamSet) // reset
	b, err := ioutil.ReadFile(string(filename))
	if err != nil {
		gi.PromptDialog(nil, gi.DlgOpts{Title: "File Not Found", Prompt: err.Error()}, true, false, nil, nil)
		log.Println(err)
		return err
	}
	return json.Unmarshal(b, pr)
}

// SaveJSON saves params to a JSON-formatted file.
func (pr *ParamSet) SaveJSON(filename gi.FileName) error {
	b, err := json.MarshalIndent(pr, "", "  ")
	if err != nil {
		log.Println(err) // unlikely
		return err
	}
	err = ioutil.WriteFile(string(filename), b, 0644)
	if err != nil {
		gi.PromptDialog(nil, gi.DlgOpts{Title: "Could not Save to File", Prompt: err.Error()}, true, false, nil, nil)
		log.Println(err)
	}
	return err
}

// OpenJSON opens params from a JSON-formatted file.
func (pr *ParamSets) OpenJSON(filename gi.FileName) error {
	*pr = make(ParamSets) // reset
	b, err := ioutil.ReadFile(string(filename))
	if err != nil {
		gi.PromptDialog(nil, gi.DlgOpts{Title: "File Not Found", Prompt: err.Error()}, true, false, nil, nil)
		log.Println(err)
		return err
	}
	return json.Unmarshal(b, pr)
}

// SaveJSON saves params to a JSON-formatted file.
func (pr *ParamSets) SaveJSON(filename gi.FileName) error {
	b, err := json.MarshalIndent(pr, "", "  ")
	if err != nil {
		log.Println(err) // unlikely
		return err
	}
	err = ioutil.WriteFile(string(filename), b, 0644)
	if err != nil {
		gi.PromptDialog(nil, gi.DlgOpts{Title: "Could not Save to File", Prompt: err.Error()}, true, false, nil, nil)
		log.Println(err)
	}
	return err
}

var ParamsProps = ki.Props{
	"ToolBar": ki.PropSlice{
		{"SaveJSON", ki.Props{
			"label": "Save As...",
			"desc":  "save to JSON formatted file",
			"icon":  "file-save",
			"Args": ki.PropSlice{
				{"File Name", ki.Props{
					"ext": ".params",
				}},
			},
		}},
		{"OpenJSON", ki.Props{
			"label": "Open...",
			"desc":  "open from JSON formatted file",
			"icon":  "file-open",
			"Args": ki.PropSlice{
				{"File Name", ki.Props{
					"ext": ".params",
				}},
			},
		}},
	},
}

var ParamSelProps = ki.Props{
	"ToolBar": ki.PropSlice{
		{"SaveJSON", ki.Props{
			"label": "Save As...",
			"desc":  "save to JSON formatted file",
			"icon":  "file-save",
			"Args": ki.PropSlice{
				{"File Name", ki.Props{
					"ext": ".params",
				}},
			},
		}},
		{"OpenJSON", ki.Props{
			"label": "Open...",
			"desc":  "open from JSON formatted file",
			"icon":  "file-open",
			"Args": ki.PropSlice{
				{"File Name", ki.Props{
					"ext": ".params",
				}},
			},
		}},
	},
}

var ParamStyleProps = ki.Props{
	"ToolBar": ki.PropSlice{
		{"SaveJSON", ki.Props{
			"label": "Save As...",
			"desc":  "save to JSON formatted file",
			"icon":  "file-save",
			"Args": ki.PropSlice{
				{"File Name", ki.Props{
					"ext": ".params",
				}},
			},
		}},
		{"OpenJSON", ki.Props{
			"label": "Open...",
			"desc":  "open from JSON formatted file",
			"icon":  "file-open",
			"Args": ki.PropSlice{
				{"File Name", ki.Props{
					"ext": ".params",
				}},
			},
		}},
	},
}

var ParamSetProps = ki.Props{
	"ToolBar": ki.PropSlice{
		{"SaveJSON", ki.Props{
			"label": "Save As...",
			"desc":  "save to JSON formatted file",
			"icon":  "file-save",
			"Args": ki.PropSlice{
				{"File Name", ki.Props{
					"ext": ".params",
				}},
			},
		}},
		{"OpenJSON", ki.Props{
			"label": "Open...",
			"desc":  "open from JSON formatted file",
			"icon":  "file-open",
			"Args": ki.PropSlice{
				{"File Name", ki.Props{
					"ext": ".params",
				}},
			},
		}},
	},
}

var ParamSetsProps = ki.Props{
	"ToolBar": ki.PropSlice{
		{"SaveJSON", ki.Props{
			"label": "Save As...",
			"desc":  "save to JSON formatted file",
			"icon":  "file-save",
			"Args": ki.PropSlice{
				{"File Name", ki.Props{
					"ext": ".params",
				}},
			},
		}},
		{"OpenJSON", ki.Props{
			"label": "Open...",
			"desc":  "open from JSON formatted file",
			"icon":  "file-open",
			"Args": ki.PropSlice{
				{"File Name", ki.Props{
					"ext": ".params",
				}},
			},
		}},
	},
}
