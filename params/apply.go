// Copyright (c) 2019, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package params

import (
	"fmt"
	"log"
	"reflect"
	"strconv"
	"strings"

	"github.com/goki/gi/gi"
	"github.com/goki/ki/kit"
)

// TargetType returns the first part of the path, indicating what type of
// object the params apply to.  Uses the first item in the map (which is random)
// everything in the map must have the same target.
func (pr *Params) TargetType() string {
	for pt := range *pr {
		return strings.Split(pt, ".")[0]
	}
	return ""
}

// Path returns the second part of the path after the target type,
// indicating the path to the specific parameter being set.
func (pr *Params) Path(path string) string {
	return strings.Join(strings.Split(path, ".")[1:], ".")
}

// Apply applies all parameter values to given object.
// Object must already be the appropriate target type based on
// the first element of the path (see TargetType method).
// If setMsg is true, then it will log a confirmation that the parameter
// was set (it always prints an error message if it fails to set the
// parameter at given path, and returns error if so).
func (pr *Params) Apply(obj interface{}, setMsg bool) error {
	objNm := ""
	if stylr, has := obj.(Styler); has {
		objNm = stylr.Name()
	} else if lblr, has := obj.(gi.Labeler); has {
		objNm = lblr.Label()
	}
	var rerr error
	for pt, v := range *pr {
		path := pr.Path(pt)
		err := SetParam(obj, path, v)
		if err == nil {
			if setMsg {
				log.Printf("%v Set param path: %v to value: %v\n", objNm, pt, v)
			}
		} else {
			rerr = err // could accumulate but..
		}
	}
	return rerr
}

///////////////////////////////////////////////////////////////////////
//  Sel

// Apply checks if Sel selector applies to this object according to (.Class, #Name, Type)
// using the params.Styler interface, and returns false if it does not.
// The TargetType of the Params is always tested against the obj's type name first.
// If it does apply, or is not a Styler, then the Params values are set.
// If setMsg is true, then a message is printed to confirm each parameter that is set.
// It always prints a message if a parameter fails to be set, and returns an error.
func (ps *Sel) Apply(obj interface{}, setMsg bool) (bool, error) {
	if !ps.TargetTypeMatch(obj) {
		return false, nil
	}
	if !ps.SelMatch(obj) {
		return false, nil
	}
	err := ps.Params.Apply(obj, setMsg)
	return true, err
}

// TargetTypeMatch return true if target type applies to object
func (ps *Sel) TargetTypeMatch(obj interface{}) bool {
	trg := ps.Params.TargetType()
	if stylr, has := obj.(Styler); has {
		tnm := stylr.TypeName()
		if tnm == trg {
			return true
		}
	}
	tnm := kit.NonPtrType(reflect.TypeOf(obj)).Name()
	return tnm == trg
}

// SelMatch returns true if Sel selector matches the target object properties
func (ps *Sel) SelMatch(obj interface{}) bool {
	stylr, has := obj.(Styler)
	if !has {
		return true // default match if no styler..
	}
	gotyp := kit.NonPtrType(reflect.TypeOf(obj)).Name()
	return SelMatch(ps.Sel, stylr.Name(), stylr.Class(), stylr.TypeName(), gotyp)
}

// SelMatch returns true if Sel selector matches the target object properties
func SelMatch(sel string, name, cls, styp, gotyp string) bool {
	if sel == "" {
		return false
	}
	if sel[0] == '.' { // class
		return ClassMatch(sel[1:], cls)
	}
	if sel[0] == '#' { // name
		return name == sel[1:]
	}
	return styp == sel || gotyp == sel // type
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
//  Sheet

// Apply applies entire sheet to given object, using param.Sel's in order
// see param.Sel.Apply() for details.
// returns true if any Sel's applied, and error if any errors.
// If setMsg is true, then a message is printed to confirm each parameter that is set.
// It always prints a message if a parameter fails to be set, and returns an error.
func (ps *Sheet) Apply(obj interface{}, setMsg bool) (bool, error) {
	applied := false
	var rerr error
	for _, sl := range *ps {
		app, err := sl.Apply(obj, setMsg)
		if app {
			applied = true
		}
		if err != nil {
			rerr = err
		}
	}
	return applied, rerr
}

///////////////////////////////////////////////////////////////////////
//  Core Find / Set / Get Param

// FindParam parses the path and recursively tries to find the parameter pointed to
// by the path (dot-delimited field names).
// Returns error if not found, and always also emits error messages --
// the target type should already have been identified and this should only
// be called when there is an expectation of the path working.
func FindParam(val reflect.Value, path string) (reflect.Value, error) {
	npv := kit.NonPtrValue(val)
	if npv.Kind() != reflect.Struct {
		err := fmt.Errorf("params.FindParam: object is not a struct: %v kind: %v -- params must be on structs, path: %v\n", npv.String(), npv.Kind(), path)
		log.Println(err)
		return npv, err
	}
	paths := strings.Split(path, ".")
	fnm := paths[0]
	fld := npv.FieldByName(fnm)
	if !fld.IsValid() {
		err := fmt.Errorf("params.FindParam: could not find Field named: %v in struct: %v kind: %v, path: %v\n", fnm, npv.String(), npv.Kind(), path)
		log.Println(err)
		return fld, err
	}
	if len(paths) == 1 {
		return fld.Addr(), nil
	}
	return FindParam(fld.Addr(), strings.Join(paths[1:], ".")) // need addr
}

// SetParam sets parameter at given path on given object to given value
// converts the string param val as appropriate for target type.
// returns error if path not found or cannot set (always logged).
func SetParam(obj interface{}, path string, val string) error {
	fld, err := FindParam(reflect.ValueOf(obj), path)
	if err != nil {
		return err
	}
	npf := kit.NonPtrValue(fld)
	switch npf.Kind() {
	case reflect.String:
		npf.SetString(val)
	case reflect.Float64, reflect.Float32:
		r, err := strconv.ParseFloat(val, 64)
		if err != nil {
			log.Println(err)
			return err
		}
		npf.SetFloat(r)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		r, err := strconv.ParseInt(val, 0, 64)
		if err != nil {
			enerr := kit.SetEnumValueFromString(fld, val)
			if enerr != nil {
				log.Println(err)
				return err
			}
		} else {
			npf.SetInt(r)
		}
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		r, err := strconv.ParseInt(val, 0, 64)
		if err != nil {
			log.Println(err)
			return err
		}
		npf.SetUint(uint64(r))
	case reflect.Bool:
		r, err := strconv.ParseBool(val)
		if err != nil {
			log.Println(err)
			return err
		}
		npf.SetBool(r)
	default:
		err := fmt.Errorf("params.SetParam: field is not of a numeric type -- only numeric types supported. value: %v, kind: %v, path: %v\n", npf.String(), npf.Kind(), path)
		log.Println(err)
		return err
	}
	return nil
}

// GetParam gets parameter value at given path on given object.
// converts target type to float64.
// returns error if path not found or target is not a numeric type (always logged).
func GetParam(obj interface{}, path string) (float64, error) {
	fld, err := FindParam(reflect.ValueOf(obj), path)
	if err != nil {
		return 0, err
	}
	npf := kit.NonPtrValue(fld)
	switch npf.Kind() {
	case reflect.Float64, reflect.Float32:
		return npf.Float(), nil
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return float64(npf.Int()), nil
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return float64(npf.Uint()), nil
	case reflect.Bool:
		if npf.Bool() {
			return 1, nil
		} else {
			return 0, nil
		}
	default:
		err := fmt.Errorf("params.GetParam: field is not of a numeric type -- only numeric types supported. value: %v, kind: %v, path: %v\n", npf.String(), npf.Kind(), path)
		log.Println(err)
		return 0, err
	}
}
