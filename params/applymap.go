// Copyright (c) 2023, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package params

import (
	"fmt"
	"log"
	"strings"
)

// ApplyMap applies given map[string]any values, where the map keys
// are Selector:Path and the value is the value to apply.
// returns true if any Sel's applied, and error if any errors.
// If setMsg is true, then a message is printed to confirm each parameter that is set.
// It always prints a message if a parameter fails to be set, and returns an error.
func ApplyMap(obj any, vals map[string]any, setMsg bool) (bool, error) {
	applied := false
	var rerr error
	for k, v := range vals {
		fld := strings.Split(k, ":")
		if len(fld) != 2 {
			rerr = fmt.Errorf("ApplyMap: map key value must be colon-separated Selector:Path, not: %s", k)
			continue
		}
		vstr, ok := v.(string)
		if !ok {
			rerr = fmt.Errorf("ApplyMap: map value must be  a string type")
			continue
		}

		sl := &Sel{Sel: fld[0], SetName: "ApplyMap"}
		sl.Params = make(Params)
		sl.Params[fld[1]] = vstr
		fmt.Printf("applying: sel: %s  params: %#v\n", sl.Sel, sl.Params)
		app, err := sl.Apply(obj, setMsg)
		if err != nil {
			log.Println(err)
			rerr = err
		}
		if app {
			applied = true
			sl.NMatch++
			if hist, ok := obj.(History); ok {
				hist.ParamsApplied(sl)
			}
		}
	}
	return applied, rerr
}

// MapToSheet returns a Sheet from given map[string]any values,
// so the Sheet can be applied as such -- e.g., for the network
// ApplyParams method.
// The map keys are Selector:Path and the value is the value to apply.
func MapToSheet(vals map[string]any) (*Sheet, error) {
	var rerr error
	sh := NewSheet()
	for k, v := range vals {
		fld := strings.Split(k, ":")
		if len(fld) != 2 {
			rerr = fmt.Errorf("ApplyMap: map key value must be colon-separated Selector:Path, not: %s", k)
			continue
		}
		vstr, ok := v.(string)
		if !ok {
			rerr = fmt.Errorf("ApplyMap: map value must be  a string type")
			continue
		}

		sl := &Sel{Sel: fld[0], SetName: "ApplyMap"}
		sl.Params = make(Params)
		sl.Params[fld[1]] = vstr
		*sh = append(*sh, sl)
	}
	return sh, rerr
}
