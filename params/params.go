// Copyright (c) 2019, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package params

import (
	"fmt"
	"log"

	"github.com/goki/ki/kit"
)

// Params is a name-value map for parameter values that can be applied
// to any numeric type in any object.
// The name must be a dot-separated path to a specific parameter, e.g., Prjn.Learn.Lrate
// The first part of the path is the overall target object type, e.g., "Prjn" or "Layer",
// which is used for determining if the parameter applies to a given object type.
//
// All of the params in one map must apply to the same target type because
// only the first item in the map (which could be any due to order randomization)
// is used for checking the type of the target.  Also, they all fall within the same
// Sel selector scope which is used to determine what specific objects to apply the
// parameters to.
type Params map[string]string

// ParamByNameTry returns given parameter, by name.
// Returns error if not found.
func (pr *Params) ParamByNameTry(name string) (string, error) {
	vl, ok := (*pr)[name]
	if !ok {
		err := fmt.Errorf("params.Params: parameter named %v not found", name)
		log.Println(err)
		return "", err
	}
	return vl, nil
}

// ParamByName returns given parameter by name (just does the map access)
// Returns "" if not found -- use Try version for error
func (pr *Params) ParamByName(name string) string {
	return (*pr)[name]
}

// SetParamByName sets given parameter by name to given value.
// (just a wrapper around map set function)
func (pr *Params) SetParamByName(name, value string) {
	(*pr)[name] = value
}

var KiT_Params = kit.Types.AddType(&Params{}, ParamsProps)

///////////////////////////////////////////////////////////////////////

// params.Sel specifies a selector for the scope of application of a set of
// parameters, using standard css selector syntax (. prefix = class, # prefix = name,
// and no prefix = type)
type Sel struct {
	Sel    string `width:"30" desc:"selector for what to apply the parameters to, using standard css selector syntax: .Example applies to anything with a Class tag of 'Example', #Example applies to anything with a Name of 'Example', and Example with no prefix applies to anything of type 'Example'"`
	Desc   string `width:"60" desc:"description of these parameter values -- what effect do they have?  what range was explored?  it is valuable to record this information as you explore the params."`
	Params Params `view:"no-inline" desc:"parameter values to apply to whatever matches the selector"`
}

var KiT_Sel = kit.Types.AddType(&Sel{}, SelProps)

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
type Sheet []*Sel

// ElemLabel satisfies the gi.SliceLabeler interface to provide labels for slice elements
func (sh *Sheet) ElemLabel(idx int) string {
	return (*sh)[idx].Sel
}

var KiT_Sheet = kit.Types.AddType(&Sheet{}, SheetProps)

// SelByNameTry returns given selector within the Sheet, by Name.
// Returns nil and error if not found.
func (sh *Sheet) SelByNameTry(sel string) (*Sel, error) {
	sl := sh.SelByName(sel)
	if sl == nil {
		err := fmt.Errorf("params.Sheet: Sel named %v not found", sel)
		log.Println(err)
		return nil, err
	}
	return sl, nil
}

// SelByName returns given selector within the Sheet, by Name.
// Returns nil if not found -- use Try version for error
func (sh *Sheet) SelByName(sel string) *Sel {
	for _, sl := range *sh {
		if sl.Sel == sel {
			return sl
		}
	}
	return nil
}

///////////////////////////////////////////////////////////////////////

// Sheets is a map of named sheets -- used in the Set
type Sheets map[string]*Sheet

var KiT_Sheets = kit.Types.AddType(&Sheets{}, SheetsProps)

///////////////////////////////////////////////////////////////////////

// Set is a collection of Sheet's that constitute a coherent set of parameters --
// a particular specific configuration of parameters, which the user selects to use.
// A good strategy is to have a "Base" set that has all the best parameters so far,
// and then other sets can modify relative to that one.  It is up to the Sim code to
// apply parameter sets in whatever order is desired.
//
// Within a params.Set, multiple different params.Sheet's can be organized,
// with each CSS-style sheet achieving a relatively complete parameter styling
// of a given element of the overal model, e.g., "Network", "Sim", "Env".
// Or Network could be further broken down into "Learn" vs. "Act" etc,
// or according to different brain areas ("Hippo", "PFC", "BG", etc).
// Again, this is entirely at the discretion of the modeler and must be
// performed under explict program control, especially because order is so critical.
//
// Note that there is NO deterministic ordering of the Sheets due to the use of
// a Go map structure, which specifically randomizes order, so simply iterating over them
// and applying may produce unexpected results -- it is better to lookup by name.
type Set struct {
	Name   string `desc:"unique name of this set of parameters"`
	Desc   string `width:"60" desc:"description of this param set -- when should it be used?  how is it different from the other sets?"`
	Sheets Sheets `desc:"Sheet's grouped according to their target and / or function, e.g., "Network" for all the network params (or "Learn" vs. "Act" for more fine-grained), and "Sim" for overall simulation control parameters, "Env" for environment parameters, etc.  It is completely up to your program to lookup these names and apply them as appropriate"`
}

var KiT_Set = kit.Types.AddType(&Set{}, SetProps)

// SheetByNameTry tries to find given sheet by name, and returns error
// if not found (also logs the error)
func (ps *Set) SheetByNameTry(name string) (*Sheet, error) {
	psht, ok := ps.Sheets[name]
	if !ok {
		err := fmt.Errorf("params.Set: %v Sheet named %v not found", ps.Name, name)
		log.Println(err)
		return nil, err
	}
	return psht, nil
}

// SheetByName finds given sheet by name -- returns nil if not found.
// Use this when sure the sheet exists -- otherwise use Try version.
func (ps *Set) SheetByName(name string) *Sheet {
	return ps.Sheets[name]
}

// ValidateSheets ensures that the sheet names are among those listed -- returns
// error message for any that are not.  Helps catch typos and makes sure params are
// applied properly.    Automatically logs errors.
func (ps *Set) ValidateSheets(valids []string) error {
	var invalids []string
	for nm := range ps.Sheets {
		got := false
		for _, vl := range valids {
			if nm == vl {
				got = true
				break
			}
		}
		if !got {
			invalids = append(invalids, nm)
		}
	}
	if len(invalids) > 0 {
		err := fmt.Errorf("params.Set: %v Invalid sheet names: %v", ps.Name, invalids)
		log.Println(err)
		return err
	}
	return nil
}

///////////////////////////////////////////////////////////////////////

// Sets is a collection of Set's that can be chosen among
// depending on different desired configurations etc.  Thus, each Set
// represents a collection of different possible specific configurations,
// and different such configurations can be chosen by name to apply as desired.
type Sets []*Set

var KiT_Sets = kit.Types.AddType(&Sets{}, SetsProps)

// SetByNameTry tries to find given set by name, and returns error
// if not found (also logs the error)
func (ps *Sets) SetByNameTry(name string) (*Set, error) {
	for _, st := range *ps {
		if st.Name == name {
			return st, nil
		}
	}
	err := fmt.Errorf("params.Sets: Param Set named %v not found", name)
	log.Println(err)
	return nil, err
}

// SetByName returns given set by name -- for use when confident
// that it exists, as a nil will return if not found with no error
func (ps *Sets) SetByName(name string) *Set {
	st, _ := ps.SetByNameTry(name)
	return st
}

// ValidateSheets ensures that the sheet names are among those listed -- returns
// error message for any that are not.  Helps catch typos and makes sure params are
// applied properly.  Automatically logs errors.
func (ps *Sets) ValidateSheets(valids []string) error {
	var err error
	for _, st := range *ps {
		er := st.ValidateSheets(valids)
		if er != nil {
			err = er
		}
	}
	return err
}

// ElemLabel satisfies the gi.SliceLabeler interface to provide labels for slice elements
func (ps *Sets) ElemLabel(idx int) string {
	return (*ps)[idx].Name
}
