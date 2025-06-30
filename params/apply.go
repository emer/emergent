// Copyright (c) 2019, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package params

import (
	"errors"
	"fmt"
	"log"
	"slices"
	"strings"
)

// Apply checks if Sel selector applies to this object according to (.Class, #Name, Type)
// using the Styler interface, and returns false if it does not. If it does apply,
// then the Set function is called on the object.
func (ps *Sel[T]) Apply(obj T) bool {
	if !ps.SelMatch(obj) {
		return false
	}
	ps.Set(obj)
	return true
}

// SelMatch returns true if Sel selector matches the target object properties.
func (ps *Sel[T]) SelMatch(obj T) bool {
	if ps.Sel == "" {
		return true
	}
	if ps.Sel[0] == '.' { // class
		return ClassMatch(ps.Sel[1:], obj.StyleClass())
	}
	if ps.Sel[0] == '#' { // name
		return obj.StyleName() == ps.Sel[1:]
	}
	return true // type always matches
}

// ClassMatch returns true if given class names match.
// Handles space-separated multiple class names.
func ClassMatch(sel, cls string) bool {
	return slices.Contains(strings.Fields(cls), sel)
}

////////  Sheet

// Apply applies entire sheet to given object, using Sel's in order.
// returns true if any Sel's applied, and error if any errors.
func (ps *Sheet[T]) Apply(obj T) bool {
	applied := false
	for _, sl := range *ps {
		app := sl.Apply(obj)
		if app {
			applied = true
			sl.NMatch++
		}
	}
	return applied
}

// SelMatchReset resets the Sel.NMatch counter used to find cases where no Sel
// matched any target objects. Call at start of application process, which
// may be at an outer-loop of Apply calls (e.g., for a Network, Apply is called
// for each Layer and Path), so this must be called separately.
// See SelNoMatchWarn for warning call at end.
func (ps *Sheet[T]) SelMatchReset() {
	for _, sl := range *ps {
		sl.NMatch = 0
	}
}

// SelNoMatchWarn issues warning messages for any Sel selectors that had no
// matches during the last Apply process -- see SelMatchReset.
// The setName and objName provide info about the Set and obj being applied.
// Returns an error message with the non-matching sets if any, else nil.
func (ps *Sheet[T]) SelNoMatchWarn(setName, objName string) error {
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
