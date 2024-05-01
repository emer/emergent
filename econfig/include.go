// Copyright (c) 2023, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// note: FindFileOnPaths adapted from viper package https://github.com/spf13/viper
// Copyright (c) 2014 Steve Francia

package econfig

import (
	"errors"
	"reflect"

	"cogentcore.org/core/base/iox/tomlx"
	"cogentcore.org/core/base/reflectx"
)

// Includeser enables processing of Includes []string field with files to include in Config objects.
type Includeser interface {
	// IncludesPtr returns a pointer to the Includes []string field containing file(s) to include
	// before processing the current config file.
	IncludesPtr() *[]string
}

// Includer enables processing of Include string field with file to include in Config objects.
type Includer interface {
	// IncludePtr returns a pointer to the Include string field containing single file to include
	// before processing the current config file.
	IncludePtr() *string
}

// IncludesStack returns the stack of include files in the natural
// order in which they are encountered (nil if none).
// Files should then be read in reverse order of the slice.
// Returns an error if any of the include files cannot be found on IncludePath.
// Does not alter cfg.
func IncludesStack(cfg Includeser) ([]string, error) {
	clone := reflect.New(reflectx.NonPointerType(reflect.TypeOf(cfg))).Interface().(Includeser)
	*clone.IncludesPtr() = *cfg.IncludesPtr()
	return includesStackImpl(clone, nil)
}

// includeStackImpl implements IncludeStack, operating on cloned cfg
// todo: could use a more efficient method to just extract the include field..
func includesStackImpl(clone Includeser, includes []string) ([]string, error) {
	incs := *clone.IncludesPtr()
	ni := len(incs)
	if ni == 0 {
		return includes, nil
	}
	for i := ni - 1; i >= 0; i-- {
		includes = append(includes, incs[i]) // reverse order so later overwrite earlier
	}
	var errs []error
	for _, inc := range incs {
		*clone.IncludesPtr() = nil
		err := tomlx.OpenFromPaths(clone, inc, IncludePaths)
		if err == nil {
			includes, err = includesStackImpl(clone, includes)
			if err != nil {
				errs = append(errs, err)
			}
		} else {
			errs = append(errs, err)
		}
	}
	return includes, errors.Join(errs...)
}

// IncludeStack returns the stack of include files in the natural
// order in which they are encountered (nil if none).
// Files should then be read in reverse order of the slice.
// Returns an error if any of the include files cannot be found on IncludePath.
// Does not alter cfg.
func IncludeStack(cfg Includer) ([]string, error) {
	clone := reflect.New(reflectx.NonPointerType(reflect.TypeOf(cfg))).Interface().(Includer)
	*clone.IncludePtr() = *cfg.IncludePtr()
	return includeStackImpl(clone, nil)
}

// includeStackImpl implements IncludeStack, operating on cloned cfg
// todo: could use a more efficient method to just extract the include field..
func includeStackImpl(clone Includer, includes []string) ([]string, error) {
	inc := *clone.IncludePtr()
	if inc == "" {
		return includes, nil
	}
	includes = append(includes, inc)
	var errs []error
	*clone.IncludePtr() = ""
	err := tomlx.OpenFromPaths(clone, inc, IncludePaths)
	if err == nil {
		includes, err = includeStackImpl(clone, includes)
		if err != nil {
			errs = append(errs, err)
		}
	} else {
		errs = append(errs, err)
	}
	return includes, errors.Join(errs...)
}
