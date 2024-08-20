// Copyright (c) 2019, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package params

//go:generate core generate -add-types

import (
	"fmt"

	"cogentcore.org/core/base/errors"
)

// Params is a name-value map for parameter values that can be applied
// to any numeric type in any object.
// The name must be a dot-separated path to a specific parameter, e.g., Path.Learn.Lrate
// The first part of the path is the overall target object type, e.g., "Path" or "Layer",
// which is used for determining if the parameter applies to a given object type.
//
// All of the params in one map must apply to the same target type because
// only the first item in the map (which could be any due to order randomization)
// is used for checking the type of the target.  Also, they all fall within the same
// Sel selector scope which is used to determine what specific objects to apply the
// parameters to.
type Params map[string]string //types:add

// ParamByName returns given parameter, by name.
// Returns and logs error if not found.
func (pr *Params) ParamByName(name string) (string, error) {
	vl, ok := (*pr)[name]
	if !ok {
		return "", errors.Log(fmt.Errorf("params.Params: parameter named %v not found", name))
	}
	return vl, nil
}

// SetByName sets given parameter by name to given value.
// (just a wrapper around map set function)
func (pr *Params) SetByName(name, value string) {
	(*pr)[name] = value
}

///////////////////////////////////////////////////////////////////////

// params.Sel specifies a selector for the scope of application of a set of
// parameters, using standard css selector syntax (. prefix = class, # prefix = name,
// and no prefix = type)
type Sel struct { //types:add

	// selector for what to apply the parameters to, using standard css selector syntax: .Example applies to anything with a Class tag of 'Example', #Example applies to anything with a Name of 'Example', and Example with no prefix applies to anything of type 'Example'
	Sel string `width:"30"`

	// description of these parameter values -- what effect do they have?  what range was explored?  it is valuable to record this information as you explore the params.
	Desc string `width:"60"`

	// parameter values to apply to whatever matches the selector
	Params Params `display:"no-inline"`

	// Put your hyperparams here
	Hypers Hypers

	// number of times this selector matched a target during the last Apply process -- a warning is issued for any that remain at 0 -- see Sheet SelMatchReset and SelNoMatchWarn methods
	NMatch int `table:"-" toml:"-" json:"-" xml:"-" edit:"-"`

	// name of current Set being applied
	SetName string `table:"-" toml:"-" json:"-" xml:"-" edit:"-"`
}

// SetFloat sets the value of given parameter
func (sl *Sel) SetFloat(param string, val float64) {
	sl.Params.SetByName(param, fmt.Sprintf("%g", val))
}

// SetString sets the value of given parameter
func (sl *Sel) SetString(param string, val string) {
	sl.Params.SetByName(param, val)
}

// ParamVal returns the value of given parameter
func (sl *Sel) ParamValue(param string) (string, error) {
	return sl.Params.ParamByName(param)
}

///////////////////////////////////////////////////////////////////////

// Sheet is a CSS-like style-sheet of params.Sel values, each of which represents
// a different set of specific parameter values applied according to the Sel selector:
// .Class #Name or Type.
//
// The order of elements in the Sheet list is critical, as they are applied
// in the order given by the list (slice), and thus later Sel's can override
// those applied earlier.  Thus, you generally want to have more general Type-level
// parameters listed first, and then subsequently more specific ones (.Class and #Name)
//
// This is the highest level of params that has an Apply method -- above this level
// application must be done under explicit program control.
type Sheet []*Sel //types:add

// NewSheet returns a new Sheet
func NewSheet() *Sheet {
	sh := make(Sheet, 0)
	return &sh
}

// ElemLabel satisfies the core.SliceLabeler interface to provide labels for slice elements
func (sh *Sheet) ElemLabel(idx int) string {
	return (*sh)[idx].Sel
}

// SelByName returns given selector within the Sheet, by Name.
// Returns and logs error if not found.
func (sh *Sheet) SelByName(sel string) (*Sel, error) {
	for _, sl := range *sh {
		if sl.Sel == sel {
			return sl, nil
		}
	}
	return nil, errors.Log(fmt.Errorf("params.Sheet: Sel named %v not found", sel))
}

// SetFloat sets the value of given parameter, in selection sel
func (sh *Sheet) SetFloat(sel, param string, val float64) error {
	sp, err := sh.SelByName(sel)
	if err != nil {
		return err
	}
	sp.SetFloat(param, val)
	return nil
}

// SetString sets the value of given parameter, in selection sel
func (sh *Sheet) SetString(sel, param string, val string) error {
	sp, err := sh.SelByName(sel)
	if err != nil {
		return err
	}
	sp.SetString(param, val)
	return nil
}

// ParamVal returns the value of given parameter, in selection sel
func (sh *Sheet) ParamValue(sel, param string) (string, error) {
	sp, err := sh.SelByName(sel)
	if err != nil {
		return "", err
	}
	return sp.ParamValue(param)
}

///////////////////////////////////////////////////////////////////////

// Sets is a collection of Sheets that can be chosen among
// depending on different desired configurations etc.  Thus, each Set
// represents a collection of different possible specific configurations,
// and different such configurations can be chosen by name to apply as desired.
type Sets map[string]*Sheet //git:add

// SheetByName tries to find given set by name.
// Returns and logs error if not found.
func (ps *Sets) SheetByName(name string) (*Sheet, error) {
	st, ok := (*ps)[name]
	if ok {
		return st, nil
	}
	return nil, errors.Log(fmt.Errorf("params.Sets: Param Sheet named %s not found", name))
}

// SetFloat sets the value of given parameter, in selection sel,
// in sheet and set.
func (ps *Sets) SetFloat(sheet, sel, param string, val float64) error {
	sp, err := ps.SheetByName(sheet)
	if err != nil {
		return err
	}
	return sp.SetFloat(sel, param, val)
}

// SetString sets the value of given parameter, in selection sel,
// in sheet and set.  Returns error if anything is not found.
func (ps *Sets) SetString(sheet, sel, param string, val string) error {
	sp, err := ps.SheetByName(sheet)
	if err != nil {
		return err
	}
	return sp.SetString(sel, param, val)
}

// ParamVal returns the value of given parameter, in selection sel,
// in sheet and set.  Returns error if anything is not found.
func (ps *Sets) ParamValue(sheet, sel, param string) (string, error) {
	sp, err := ps.SheetByName(sheet)
	if err != nil {
		return "", err
	}
	return sp.ParamValue(sel, param)
}
