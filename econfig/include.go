// Copyright (c) 2023, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// note: FindFileOnPaths adapted from viper package https://github.com/spf13/viper
// Copyright (c) 2014 Steve Francia

package econfig

import (
	"errors"
	"fmt"
	"log"
	"reflect"
	"strings"

	"github.com/goki/ki/kit"
)

// SetFromIncludes sets config from included files.
// Returns an error if any of the include files cannot be found on IncludePath.
func SetFromIncludes(cfg any) error {
	incs, err := IncludeStack(cfg)
	ni := len(incs)
	if ni == 0 {
		return err
	}
	for i := ni - 1; i >= 0; i-- {
		inc := incs[i]
		err = Open(cfg, inc)
	}
	return err
}

// IncludeStack returns the stack of include files in the natural
// order in which they are encountered (nil if none).
// Files should then be read in reverse order of the slice.
// Returns an error if any of the include files cannot be found on IncludePath.
// Does not alter cfg.
func IncludeStack(cfg any) ([]string, error) {
	clone := reflect.New(reflect.TypeOf(cfg).Elem())
	return includeStackImpl(clone, nil)
}

// includeStackImpl implements IncludeStack, operating on cloned cfg
// todo: could use a more efficient method to just extract the include field..
func includeStackImpl(clone any, includes []string) ([]string, error) {
	incs := GetIncludes(clone)
	if len(incs) == 0 {
		return includes, nil
	}
	includes = append(includes, incs...)
	var errs []error
	for _, inc := range incs {
		ResetIncludeField(clone) // key to reset prior to loading so only getting new
		err := Open(clone, inc)
		if err == nil {
			includes, err = includeStackImpl(clone, includes)
			if err != nil {
				errs = append(errs, err)
			}
		} else {
			errs = append(errs, err)
		}
	}
	return includes, AllErrors(errs)
}

// GetIncludes returns the []string list of element(s) in the Include or Includes field
// (works with string or []string types).  Returns nil if none found.
// (case insensitive to field name)
func GetIncludes(cfg any) []string {
	fv, ok := FindIncludeField(cfg)
	if !ok {
		return nil
	}
	if fv.Kind() == reflect.String {
		return []string{fv.String()}
	}
	if fv.Kind() == reflect.Slice {
		if iss, ok := fv.Interface().([]string); ok {
			return iss
		}
	}
	return nil
}

// FindIncludeField returns the reflect.Value for a field named
// Include or Includes -- must be a string or []string.
// Returns false if not found.
func FindIncludeField(cfg any) (reflect.Value, bool) {
	typ := kit.NonPtrType(reflect.TypeOf(cfg))
	val := kit.NonPtrValue(reflect.ValueOf(cfg))
	for i := 0; i < typ.NumField(); i++ {
		f := typ.Field(i)
		fnm := strings.ToLower(f.Name)
		if !(fnm == "include" || fnm == "includes") {
			continue
		}
		if !(f.Type.Kind() == reflect.String || f.Type.Kind() == reflect.Slice) {
			err := fmt.Errorf("econfig.IncludeField: field named %s must be either a string or []string", f.Name)
			log.Println(err)
			return val, false
		}
		return val.Field(i), true
	}
	return val, false
}

// ResetIncludeField sets the Include or Includes field to zero / nil value.
// This is called prior to loading imports.
func ResetIncludeField(cfg any) {
	fv, ok := FindIncludeField(cfg)
	if !ok { // shouldn't happen
		return
	}
	fv.SetZero()
}

// AllErrors returns an err as a concatenation of errors (nil if none)
func AllErrors(errs []error) error {
	if len(errs) == 0 {
		return nil
	}
	ers := make([]string, len(errs))
	for i, err := range errs {
		ers[i] = err.Error()
	}
	return errors.New(strings.Join(ers, "\n"))
}
