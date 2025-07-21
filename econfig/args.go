// Copyright (c) 2023, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// note: parsing code adapted from pflag package https://github.com/spf13/pflag
// Copyright (c) 2012 Alex Ogier. All rights reserved.
// Copyright (c) 2012 The Go Authors. All rights reserved.

package econfig

import (
	"fmt"
	"reflect"
	"strings"

	"cogentcore.org/core/base/iox/tomlx"
	"cogentcore.org/core/base/mpi"
	"cogentcore.org/core/base/reflectx"
	"cogentcore.org/core/base/strcase"
)

// SetFromArgs sets Config values from command-line args,
// based on the field names in the Config struct.
// Returns any args that did not start with a `-` flag indicator.
// For more robust error processing, it is assumed that all flagged args (-)
// must refer to fields in the config, so any that fail to match trigger
// an error.  Errors can also result from parsing.
// Errors are automatically logged because these are user-facing.
func SetFromArgs(cfg any, args []string) (nonFlags []string, err error) {
	allArgs := make(map[string]reflect.Value)
	CommandArgs(allArgs) // need these to not trigger not-found errors
	FieldArgNames(cfg, allArgs)
	nonFlags, err = ParseArgs(cfg, args, allArgs, true)
	if err != nil {
		mpi.Println(Usage(cfg))
	}
	return
}

// ParseArgs parses given args using map of all available args
// setting the value accordingly, and returning any leftover args.
// setting errNotFound = true causes args that are not in allArgs to
// trigger an error.  Otherwise, it just skips those.
func ParseArgs(cfg any, args []string, allArgs map[string]reflect.Value, errNotFound bool) ([]string, error) {
	var nonFlags []string
	var err error
	for len(args) > 0 {
		s := args[0]
		args = args[1:]
		if len(s) == 0 || s[0] != '-' || len(s) == 1 {
			nonFlags = append(nonFlags, s)
			continue
		}

		if s[1] == '-' && len(s) == 2 { // "--" terminates the flags
			// f.argsLenAtDash = len(f.args)
			nonFlags = append(nonFlags, args...)
			break
		}
		args, err = ParseArg(s, args, allArgs, errNotFound)
		if err != nil {
			return nonFlags, err
		}
	}
	return nonFlags, nil
}

func ParseArg(s string, args []string, allArgs map[string]reflect.Value, errNotFound bool) (a []string, err error) {
	a = args
	name := s[1:]
	if name[0] == '-' {
		name = name[1:]
	}
	if len(name) == 0 || name[0] == '-' || name[0] == '=' {
		err = fmt.Errorf("econfig.ParseArgs: bad flag syntax: %s", s)
		mpi.Println(err)
		return
	}

	if strings.HasPrefix(name, "test.") { // go test passes args..
		return
	}

	split := strings.SplitN(name, "=", 2)
	name = split[0]
	fval, exists := allArgs[name]
	if !exists {
		if errNotFound {
			err = fmt.Errorf("econfig.ParseArgs: flag name not recognized: %s", name)
			mpi.Println(err)
		}
		return
	}

	isbool := reflectx.NonPointerValue(fval).Kind() == reflect.Bool

	var value string
	switch {
	case len(split) == 2:
		// '--flag=arg'
		value = split[1]
	case isbool:
		// '--flag' bare
		lcnm := strings.ToLower(name)
		negate := false
		if len(lcnm) > 3 {
			if lcnm[:3] == "no_" || lcnm[:3] == "no-" {
				negate = true
			} else if lcnm[:2] == "no" {
				if _, has := allArgs[lcnm[2:]]; has { // e.g., nogui and gui is on list
					negate = true
				}
			}
		}
		if negate {
			value = "false"
		} else {
			value = "true"
		}
	case len(a) > 0:
		// '--flag arg'
		value = a[0]
		a = a[1:]
	default:
		// '--flag' (arg was required)
		err = fmt.Errorf("econfig.ParseArgs: flag needs an argument: %s", s)
		mpi.Println(err)
		return
	}

	err = SetArgValue(name, fval, value)
	return
}

// SetArgValue sets given arg name to given value, into settable reflect.Value
func SetArgValue(name string, fval reflect.Value, value string) error {
	nptyp := reflectx.NonPointerType(fval.Type())
	vk := nptyp.Kind()
	switch {
	// todo: enum
	// case vk >= reflect.Int && vk <= reflect.Uint64 && kit.Enums.TypeRegistered(nptyp):
	// 	return kit.Enums.SetAnyEnumValueFromString(fval, value)
	case vk == reflect.Map:
		mval := make(map[string]any)
		err := tomlx.ReadBytes(&mval, []byte("tmp="+value)) // use toml decoder
		if err != nil {
			mpi.Println(err)
			return err
		}
		err = reflectx.CopyMapRobust(fval.Interface(), mval["tmp"])
		if err != nil {
			mpi.Println(err)
			err = fmt.Errorf("econfig.ParseArgs: not able to set map field from arg: %s val: %s", name, value)
			mpi.Println(err)
			return err
		}
	case vk == reflect.Slice:
		mval := make(map[string]any)
		err := tomlx.ReadBytes(&mval, []byte("tmp="+value)) // use toml decoder
		if err != nil {
			mpi.Println(err)
			return err
		}
		err = reflectx.CopySliceRobust(fval.Interface(), mval["tmp"])
		if err != nil {
			mpi.Println(err)
			err = fmt.Errorf("econfig.ParseArgs: not able to set slice field from arg: %s val: %s", name, value)
			mpi.Println(err)
			return err
		}
	default:
		err := reflectx.SetRobust(fval.Interface(), value) // overkill but whatever
		if err != nil {
			err := fmt.Errorf("econfig.ParseArgs: not able to set field from arg: %s val: %s", name, value)
			mpi.Println(err)
			return err
		}
	}
	return nil
}

// FieldArgNames adds to given args map all the different ways the field names
// can be specified as arg flags, mapping to the reflect.Value
func FieldArgNames(obj any, allArgs map[string]reflect.Value) {
	fieldArgNamesStruct(obj, "", false, allArgs)
}

func addAllCases(nm, path string, pval reflect.Value, allArgs map[string]reflect.Value) {
	if nm == "Includes" {
		return // skip
	}
	if path != "" {
		nm = path + "." + nm
	}
	allArgs[nm] = pval
	allArgs[strings.ToLower(nm)] = pval
	allArgs[strcase.ToKebab(nm)] = pval
	allArgs[strcase.ToSnake(nm)] = pval
	allArgs[strcase.ToSNAKE(nm)] = pval
}

// fieldArgNamesStruct returns map of all the different ways the field names
// can be specified as arg flags, mapping to the reflect.Value
func fieldArgNamesStruct(obj any, path string, nest bool, allArgs map[string]reflect.Value) {
	ov := reflect.ValueOf(obj)
	if reflectx.IsNil(ov) {
		return
	}
	if ov.Kind() == reflect.Pointer && ov.IsNil() {
		return
	}
	val := reflectx.NonPointerValue(ov)
	typ := val.Type()
	for i := 0; i < typ.NumField(); i++ {
		f := typ.Field(i)
		fv := val.Field(i)
		if reflectx.NonPointerType(f.Type).Kind() == reflect.Struct {
			nwPath := f.Name
			if path != "" {
				nwPath = path + "." + nwPath
			}
			nwNest := nest
			if !nwNest {
				neststr, ok := f.Tag.Lookup("nest")
				if ok && (neststr == "+" || neststr == "true") {
					nwNest = true
				}
			}
			fieldArgNamesStruct(reflectx.PointerValue(fv).Interface(), nwPath, nwNest, allArgs)
			continue
		}
		pval := reflectx.PointerValue(fv)
		addAllCases(f.Name, path, pval, allArgs)
		if f.Type.Kind() == reflect.Bool {
			addAllCases("No"+f.Name, path, pval, allArgs)
		}
		// now process adding non-nested version of field
		if path == "" || nest {
			continue
		}
		neststr, ok := f.Tag.Lookup("nest")
		if ok && (neststr == "+" || neststr == "true") {
			continue
		}
		if _, has := allArgs[f.Name]; has {
			mpi.Printf("econfig Field: %s.%s cannot be added as a non-nested %s arg because it has already been registered -- add 'nest:'+'' field tag to the one you want to keep only as a nested arg with path, to eliminate this message\n", path, f.Name, f.Name)
			continue
		}
		addAllCases(f.Name, "", pval, allArgs)
		if f.Type.Kind() == reflect.Bool {
			addAllCases("No"+f.Name, "", pval, allArgs)
		}
	}
}

// CommandArgs adds non-field args that control the config process:
// -config -cfg -help -h
func CommandArgs(allArgs map[string]reflect.Value) {
	allArgs["config"] = reflect.ValueOf(&ConfigFile)
	allArgs["cfg"] = reflect.ValueOf(&ConfigFile)
	allArgs["help"] = reflect.ValueOf(&Help)
	allArgs["h"] = reflect.ValueOf(&Help)
}
