// Copyright (c) 2023, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// note: parsing code adapted from pflag package https://github.com/spf13/pflag
// Copyright (c) 2012 Alex Ogier. All rights reserved.
// Copyright (c) 2012 The Go Authors. All rights reserved.

package econfig

import (
	"fmt"
	"log"
	"os"
	"reflect"
	"strings"

	"github.com/goki/ki/kit"
	"github.com/iancoleman/strcase"
)

// SetFromArgs sets Config values from command-line args,
// based on the field names in the Config struct.
// Returns any args that did not start with a `-` flag indicator.
// For more robust error processing, it is assumed that all flagged args (-)
// must refer to fields in the config, so any that fail to match trigger
// an error.  Errors can also result from parsing.
// Errors are automatically logged because these are user-facing.
func SetFromArgs(cfg any) (leftovers []string, err error) {
	leftovers, err = parseArgs(cfg, os.Args[1:])
	if err != nil {
		fmt.Println(Usage(cfg))
	}
	return
}

// parseArgs does the actual arg parsing
func parseArgs(cfg any, args []string) ([]string, error) {
	longArgs, shortArgs := FieldArgNames(cfg)
	var leftovers []string
	var err error
	for len(args) > 0 {
		s := args[0]
		args = args[1:]
		if len(s) == 0 || s[0] != '-' || len(s) == 1 {
			leftovers = append(leftovers, s)
			continue
		}

		if s[1] == '-' {
			if len(s) == 2 { // "--" terminates the flags
				// f.argsLenAtDash = len(f.args)
				leftovers = append(leftovers, args...)
				break
			}
			args, err = parseLongArg(s, args, longArgs)
		} else {
			args, err = parseShortArg(s, args, shortArgs)
		}
		if err != nil {
			return leftovers, err
		}
	}
	return leftovers, nil
}

func parseLongArg(s string, args []string, longArgs map[string]reflect.Value) (a []string, err error) {
	a = args
	name := s[2:]
	if len(name) == 0 || name[0] == '-' || name[0] == '=' {
		err = fmt.Errorf("SetFromArgs: bad flag syntax: %s", s)
		log.Println(err)
		return
	}

	split := strings.SplitN(name, "=", 2)
	name = split[0]
	fval, exists := longArgs[name]
	if !exists {
		err = fmt.Errorf("SetFromArgs: flag name not recognized: %s", name)
		log.Println(err)
		return
	}

	var value string
	if len(split) == 2 {
		// '--flag=arg'
		value = split[1]
	} else if len(a) > 0 {
		// '--flag arg'
		value = a[0]
		a = a[1:]
	} else {
		// '--flag' (arg was required)
		err = fmt.Errorf("SetFromArgs: flag needs an argument: %s", s)
		log.Println(err)
		return
	}

	err = setArgValue(name, fval, value)
	return
}

func setArgValue(name string, fval reflect.Value, value string) error {
	ok := kit.SetRobust(fval.Interface(), value) // overkill but whatever
	if !ok {
		err := fmt.Errorf("SetFromArgs: not able to set field from arg: %s val: %s", name, value)
		log.Println(err)
		return err
	}
	return nil
}

func parseSingleShortArg(shorthands string, args []string, shortArgs map[string]reflect.Value) (outShorts string, outArgs []string, err error) {
	outArgs = args
	// 	if strings.HasPrefix(shorthands, "test.") {
	// 		return
	// 	}
	outShorts = shorthands[1:]
	c := string(shorthands[0])

	fval, exists := shortArgs[c]

	if !exists {
		err = fmt.Errorf("SetFromArgs: unknown shorthand flag: %q in -%s", c, shorthands)
		log.Println(err)
		return
	}

	// todo: make sure that next field doesn't start with -

	var value string
	if len(shorthands) > 2 && shorthands[1] == '=' {
		// '-f=arg'
		value = shorthands[2:]
		outShorts = ""
	} else if len(args) > 0 {
		if len(args[0]) > 1 && string(args[0][0]) != "-" {
			value = args[0]
			outArgs = args[1:]
		} else {
			value = "true"
		}
	} else {
		value = "true"
	}
	err = setArgValue(c, fval, value)
	return
}

func parseShortArg(s string, args []string, shortArgs map[string]reflect.Value) (a []string, err error) {
	a = args
	shorthands := s[1:]

	// "shorthands" can be a series of shorthand letters of flags (e.g. "-vvv").
	for len(shorthands) > 0 {
		shorthands, a, err = parseSingleShortArg(shorthands, args, shortArgs)
		if err != nil {
			return
		}
	}
	return
}

// FieldArgNames returns map of all the different ways the field names
// can be specified as arg flags, mapping to the reflect.Value
func FieldArgNames(obj any) (longArgs, shortArgs map[string]reflect.Value) {
	longArgs = make(map[string]reflect.Value)
	shortArgs = make(map[string]reflect.Value)
	FieldArgNamesStruct(obj, "", longArgs, shortArgs)
	return
}

// FieldArgNamesStruct returns map of all the different ways the field names
// can be specified as arg flags, mapping to the reflect.Value
func FieldArgNamesStruct(obj any, path string, longArgs, shortArgs map[string]reflect.Value) {
	typ := kit.NonPtrType(reflect.TypeOf(obj))
	val := kit.NonPtrValue(reflect.ValueOf(obj))
	for i := 0; i < typ.NumField(); i++ {
		f := typ.Field(i)
		fv := val.Field(i)
		if kit.NonPtrType(f.Type).Kind() == reflect.Struct {
			nwPath := f.Name
			if path != "" {
				nwPath = path + "." + nwPath
			}
			FieldArgNamesStruct(kit.PtrValue(fv).Interface(), nwPath, longArgs, shortArgs)
			continue
		}
		pval := kit.PtrValue(fv)
		nm := f.Name
		if path != "" {
			nm = path + "." + nm
		}
		longArgs[nm] = pval
		longArgs[strings.ToLower(nm)] = pval
		longArgs[strcase.ToKebab(nm)] = pval
		longArgs[strcase.ToSnake(nm)] = pval
		longArgs[strcase.ToScreamingSnake(nm)] = pval
		sh, ok := f.Tag.Lookup("short")
		if ok && sh != "" {
			if _, has := shortArgs[sh]; has {
				log.Println("Short arg named:", sh, "already defined")
			}
			shortArgs[sh] = pval
		}
	}
}
