// Copyright (c) 2023, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package params

import (
	"errors"
	"fmt"
	"log"
	"reflect"
	"strings"

	"goki.dev/laser"
)

// ApplyMap applies given map[string]any values, where the map keys
// are a Path and the value is the value to apply (any appropriate type).
// This is not for Network params, which should use MapToSheet -- see emer.Params wrapper.
// If setMsg is true, then a message is printed to confirm each parameter that is set.
// It always prints a message if a parameter fails to be set, and returns an error.
func ApplyMap(obj any, vals map[string]any, setMsg bool) error {
	objv := reflect.ValueOf(obj)
	npv := laser.NonPtrValue(objv)
	if npv.Kind() == reflect.Map {
		err := laser.CopyMapRobust(obj, vals)
		if err != nil {
			log.Println(err)
			return err
		}
		if setMsg {
			log.Printf("ApplyMap: set map object to %#v\n", obj)
		}
	}
	var errs []error
	for k, v := range vals {
		fld, err := FindParam(objv, k)
		if err != nil {
			errs = append(errs, err)
		}
		err = laser.SetRobust(fld.Interface(), v)
		if err != nil {
			err = fmt.Errorf("ApplyMap: was not able to apply value: %v to field: %s", v, k)
			log.Println(err)
			errs = append(errs, err)
		}
		if setMsg {
			log.Printf("ApplyMap: set field: %s = %#v\n", k, laser.NonPtrValue(fld).Interface())
		}
	}
	return errors.Join(errs...)
}

// MapToSheet returns a Sheet from given map[string]any values,
// so the Sheet can be applied as such -- e.g., for the network
// ApplyParams method.
// The map keys are Selector:Path and the value is the value to apply.
func MapToSheet(vals map[string]any) (*Sheet, error) {
	sh := NewSheet()
	var errs []error
	for k, v := range vals {
		fld := strings.Split(k, ":")
		if len(fld) != 2 {
			err := fmt.Errorf("ApplyMap: map key value must be colon-separated Selector:Path, not: %s", k)
			log.Println(err)
			errs = append(errs, err)
			continue
		}
		vstr := laser.ToString(v)

		sl := &Sel{Sel: fld[0], SetName: "ApplyMap"}
		sl.Params = make(Params)
		sl.Params[fld[1]] = vstr
		*sh = append(*sh, sl)
	}
	return sh, errors.Join(errs...)
}
