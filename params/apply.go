// Copyright (c) 2019, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package params

import (
	"errors"
	"fmt"
	"log"
	"log/slog"
	"reflect"
	"strings"

	"cogentcore.org/core/base/labels"
	"cogentcore.org/core/base/reflectx"
)

// PathAfterType returns the portion of a path string after the initial
// type, e.g., Layer.Acts.Kir.Gbar -> Acts.Kir.Gbar
func PathAfterType(path string) string {
	return strings.Join(strings.Split(path, ".")[1:], ".")
}

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
	return PathAfterType(path)
}

// Apply applies all parameter values to given object.
// Object must already be the appropriate target type based on
// the first element of the path (see TargetType method).
// If setMsg is true, then it will log a confirmation that the parameter
// was set (it always prints an error message if it fails to set the
// parameter at given path, and returns error if so).
func (pr *Params) Apply(obj any, setMsg bool) error {
	objNm := ""
	if styler, has := obj.(Styler); has {
		objNm = styler.StyleName()
		if styob, has := obj.(StylerObject); has {
			obj = styob.StyleObject()
		}
	} else if lblr, has := obj.(labels.Labeler); has {
		objNm = lblr.Label()
	}
	var errs []error
	for pt, v := range *pr {
		path := pr.Path(pt)
		if hv, ok := obj.(Hypers); ok {
			if cv, has := hv[pt]; has { // full path
				cv["Val"] = v
			} else {
				hv[pt] = HyperValues{"Val": v}
			}
			continue
		}
		err := SetParam(obj, path, v)
		if err == nil {
			if setMsg {
				log.Printf("%v Set param path: %v to value: %v\n", objNm, pt, v)
			}
		} else {
			errs = append(errs, err)
		}
	}
	return errors.Join(errs...)
}

///////////////////////////////////////////////////////////////////////
//  Hypers

// TargetType returns the first part of the path, indicating what type of
// object the params apply to.  Uses the first item in the map (which is random)
// everything in the map must have the same target.
func (pr *Hypers) TargetType() string {
	for pt := range *pr {
		return strings.Split(pt, ".")[0]
	}
	return ""
}

// Path returns the second part of the path after the target type,
// indicating the path to the specific parameter being set.
func (pr *Hypers) Path(path string) string {
	return strings.Join(strings.Split(path, ".")[1:], ".")
}

// Apply applies all parameter values to given object.
// Object must already be the appropriate target type based on
// the first element of the path (see TargetType method).
// If setMsg is true, then it will log a confirmation that the parameter
// was set (it always prints an error message if it fails to set the
// parameter at given path, and returns error if so).
func (pr *Hypers) Apply(obj any, setMsg bool) error {
	objNm := ""
	if styler, has := obj.(Styler); has {
		objNm = styler.StyleName()
		if styob, has := obj.(StylerObject); has {
			obj = styob.StyleObject()
		}
	} else if lblr, has := obj.(labels.Labeler); has {
		objNm = lblr.Label()
	}
	if hv, ok := obj.(Hypers); ok {
		hv.CopyFrom(*pr)
		return nil
	}
	var errs []error
	for pt, v := range *pr {
		path := pr.Path(pt)
		val, ok := v["Val"]
		if !ok {
			continue
		}
		err := SetParam(obj, path, val)
		if err == nil {
			if setMsg {
				log.Printf("%v Set hypers path: %v to value: %v\n", objNm, pt, v)
			}
		} else {
			errs = append(errs, err)
		}
	}
	return errors.Join(errs...)
}

///////////////////////////////////////////////////////////////////////
//  Sel

// Apply checks if Sel selector applies to this object according to (.Class, #Name, Type)
// using the params.Styler interface, and returns false if it does not.
// The TargetType of the Params is always tested against the obj's type name first.
// If it does apply, or is not a Styler, then the Params values are set.
// If setMsg is true, then a message is printed to confirm each parameter that is set.
// It always prints a message if a parameter fails to be set, and returns an error.
func (ps *Sel) Apply(obj any, setMsg bool) (bool, error) {
	if !ps.TargetTypeMatch(obj) {
		return false, nil
	}
	if !ps.SelMatch(obj) {
		return false, nil
	}
	errp := ps.Params.Apply(obj, setMsg)
	errh := ps.Hypers.Apply(obj, setMsg)
	if errp != nil {
		return true, errp
	}
	return true, errh
}

// TargetTypeMatch return true if target type applies to object
func (ps *Sel) TargetTypeMatch(obj any) bool {
	trg := ps.Params.TargetType()
	if styler, has := obj.(Styler); has {
		tnm := styler.StyleType()
		if tnm == trg {
			return true
		}
	}
	trgh := ps.Hypers.TargetType()
	if styler, has := obj.(Styler); has {
		tnm := styler.StyleType()
		if tnm == trgh {
			return true
		}
	}
	tnm := reflectx.NonPointerType(reflect.TypeOf(obj)).Name()
	return tnm == trg || tnm == trgh
}

// SelMatch returns true if Sel selector matches the target object properties
func (ps *Sel) SelMatch(obj any) bool {
	styler, has := obj.(Styler)
	if !has {
		return true // default match if no styler..
	}
	if styob, has := obj.(StylerObject); has {
		obj = styob.StyleObject()
	}
	gotyp := reflectx.NonPointerType(reflect.TypeOf(obj)).Name()
	return SelMatch(ps.Sel, styler.StyleName(), styler.StyleClass(), styler.StyleType(), gotyp)
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

// ClassMatch returns true if given class names.
// handles space-separated multiple class names
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
func (ps *Sheet) Apply(obj any, setMsg bool) (bool, error) {
	applied := false
	var errs []error
	for _, sl := range *ps {
		app, err := sl.Apply(obj, setMsg)
		if app {
			applied = true
			sl.NMatch++
			if hist, ok := obj.(History); ok {
				hist.ParamsApplied(sl)
			}
		}
		if err != nil {
			errs = append(errs, err)
		}
	}
	return applied, errors.Join(errs...)
}

// SelMatchReset resets the Sel.NMatch counter used to find cases where no Sel
// matched any target objects.  Call at start of application process, which
// may be at an outer-loop of Apply calls (e.g., for a Network, Apply is called
// for each Layer and Path), so this must be called separately.
// See SelNoMatchWarn for warning call at end.
func (ps *Sheet) SelMatchReset(setName string) {
	for _, sl := range *ps {
		sl.NMatch = 0
		sl.SetName = setName
	}
}

// SelNoMatchWarn issues warning messages for any Sel selectors that had no
// matches during the last Apply process -- see SelMatchReset.
// The setName and objName provide info about the Set and obj being applied.
// Returns an error message with the non-matching sets if any, else nil.
func (ps *Sheet) SelNoMatchWarn(setName, objName string) error {
	msg := ""
	for _, sl := range *ps {
		if sl.NMatch == 0 {
			msg += "\tSel: " + sl.Sel + "\n"
		}
	}
	if msg != "" {
		msg = fmt.Sprintf("param.Sheet from Set: %s for object: %s had the following non-matching Selectors:\n%s", setName, objName, msg)
		log.Println(msg) // todo: slog?
		return errors.New(msg)
	}
	return nil
}

///////////////////////////////////////////////////////////////////////
//  Core Find / Set / Get Param

// FindParam parses the path and recursively tries to find the parameter pointed to
// by the path (dot-delimited field names).
// Returns error if not found, and always also emits error messages --
// the target type should already have been identified and this should only
// be called when there is an expectation of the path working.
func FindParam(val reflect.Value, path string) (reflect.Value, error) {
	npv := reflectx.NonPointerValue(val)
	if npv.Kind() != reflect.Struct {
		if !npv.IsValid() {
			err := fmt.Errorf("params.FindParam: object is nil -- must Build *before* applying params!  path: %v\n", path)
			slog.Error(err.Error())
			return npv, err
		}
		err := fmt.Errorf("params.FindParam: object is not a struct: %v kind: %v -- params must be on structs, path: %v\n", npv.String(), npv.Kind(), path)
		slog.Error(err.Error())
		return npv, err
	}
	paths := strings.Split(path, ".")
	fnm := paths[0]
	fld := npv.FieldByName(fnm)
	if !fld.IsValid() {
		err := fmt.Errorf("params.FindParam: could not find Field named: %v in struct: %v kind: %v, path: %v\n", fnm, npv.String(), npv.Kind(), path)
		slog.Error(err.Error())
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
func SetParam(obj any, path string, val string) error {
	npv := reflectx.NonPointerValue(reflect.ValueOf(obj))
	if npv.Kind() == reflect.Map { // only for string maps
		npv.SetMapIndex(reflect.ValueOf(path), reflect.ValueOf(val))
		return nil
	}

	fld, err := FindParam(reflect.ValueOf(obj), path)
	if err != nil {
		return err
	}
	err = reflectx.SetRobust(fld.Interface(), val)
	if err != nil {
		slog.Error("params.SetParam: field could not be set", "path", path, "value", val, "err", err)
		return err
	}
	return nil
}

// GetParam gets parameter value at given path on given object.
// converts target type to float64.
// returns error if path not found or target is not a numeric type (always logged).
func GetParam(obj any, path string) (float64, error) {
	fld, err := FindParam(reflect.ValueOf(obj), path)
	if err != nil {
		return 0, err
	}
	npf := reflectx.NonPointerValue(fld)
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
		slog.Error(err.Error())
		return 0, err
	}
}
