// Copyright (c) 2019, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package emer

import (
	"log"
	"reflect"
	"strings"

	"github.com/goki/gi/gi"
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

// ParamStyle is a CSS-like collection of Params values, each of which represents a different
// set of specific parameter values.  The name is a CSS-style .Class #Name or Type
// where Type is either Prjn or Layer to apply to every instance in the network.
type ParamStyle map[string]Params

// ParamSet is a collection of ParamStyle's that constitute a coherent set of parameters --
// a particular specific configuration of parameters.
type ParamSet map[string]ParamStyle

// ParamSets is a collection of ParamSet's that can be chosen among depending on different desired
// configurations etc -- a collection of different possible specific configurations
type ParamSets map[string]ParamSet

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
// (see Target method)
func (pr *Params) Set(obj interface{}) {
	olbl := ""
	olblr, haslbl := obj.(gi.Labeler)
	if haslbl {
		olbl = olblr.Label()
	}
	for pt, v := range *pr {
		path := pr.Path(pt)
		ok := SetParam(obj, path, v)
		if ok {
			log.Printf("%v Set param path: %v to value: %v\n", olbl, pt, v)
		} else {
			log.Printf("%v Failed to set param path: %v to value: %v\n", olbl, pt, v)
		}
	}
}

///////////////////////////////////////////////////////////////////////
//  ParamStyle

// StyleMatch returns true if given style specifier matches the target object properties
// (name, class, type name).  Class can be space separated list of names.
func StyleMatch(sty string, name, class, typ string) bool {
	if sty == "" {
		return false
	}
	if sty[0] == '.' { // class
		return ClassMatch(sty[1:], class)
	}
	if sty[0] == '#' { // name
		return name == sty[1:]
	}
	return typ == sty // type
}

// ClassMatch returns true if given class names -- handles space-separated multiple class names
func ClassMatch(class, sty string) bool {
	cls := strings.Split(class, " ")
	for _, cl := range cls {
		if strings.TrimSpace(cl) == sty {
			return true
		}
	}
	return false
}
