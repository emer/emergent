// Copyright (c) 2019, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package params

//go:generate core generate -add-types

import (
	"fmt"

	"cogentcore.org/core/base/errors"
)

// Sel specifies a selector for the scope of application of a set of
// parameters, using standard css selector syntax (. prefix = class, # prefix = name,
// and no prefix = type). Type always matches, and generally should come first as an
// initial set of defaults.
type Sel[T Styler] struct {

	// Sel is the selector for what to apply the parameters to,
	// using standard css selector syntax:
	//	- .Example applies to anything with a Class tag of 'Example'
	//	- #Example applies to anything with a Name of 'Example'
	//	- Example with no prefix or blank selector always applies as the presumed Type.
	Sel string `width:"30"`

	// Doc is documentation of these parameter values: what effect
	// do they have? what range was explored? It is valuable to record
	// this information as you explore the params.
	Doc string `width:"60"`

	// Set function applies parameter values to the given object of the target type.
	Set func(v T) `display:"-"`

	// NMatch is the number of times this selector matched a target
	// during the last Apply process. A warning is issued for any
	// that remain at 0: See Sheet SelMatchReset and SelNoMatchWarn methods.
	NMatch int `table:"-" toml:"-" json:"-" xml:"-" edit:"-"`
}

////////

// Sheet is a CSS-like style-sheet of params.Sel values, each of which represents
// a different set of specific parameter values applied according to the Sel selector:
// .Class #Name or Type.
//
// The order of elements in the Sheet list is critical, as they are applied
// in the order given by the list (slice), and thus later Sel's can override
// those applied earlier. Generally put more general Type-level parameters first,
// and then subsequently more specific ones (.Class and #Name).
type Sheet[T Styler] []*Sel[T]

// NewSheet returns a new Sheet for given type.
func NewSheet[T Styler]() *Sheet[T] {
	sh := make(Sheet[T], 0)
	return &sh
}

// ElemLabel satisfies the core.SliceLabeler interface to provide labels for slice elements.
func (sh *Sheet[T]) ElemLabel(idx int) string {
	return (*sh)[idx].Sel
}

// SelByName returns given selector within the Sheet, by Name.
// Returns and logs error if not found.
func (sh *Sheet[T]) SelByName(sel string) (*Sel[T], error) {
	for _, sl := range *sh {
		if sl.Sel == sel {
			return sl, nil
		}
	}
	return nil, errors.Log(fmt.Errorf("params.Sheet: Sel named %v not found", sel))
}

////////

// Sheets are named collections of Sheet elements that can be chosen among
// depending on different desired configurations.
// Conventionally, there is always a Base configuration with basic-level
// defaults, and then any number of more specific sets to apply after that.
type Sheets[T Styler] map[string]*Sheet[T]

// SheetByName tries to find given set by name.
// Returns and logs error if not found.
func (ps *Sheets[T]) SheetByName(name string) (*Sheet[T], error) {
	st, ok := (*ps)[name]
	if ok {
		return st, nil
	}
	return nil, errors.Log(fmt.Errorf("params.Sheets: Param Sheet named %q not found", name))
}
