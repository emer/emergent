// Copyright (c) 2023, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package econfig

//go:generate core generate -add-types

import (
	"errors"
	"os"
	"reflect"

	"cogentcore.org/core/base/fsx"
	"cogentcore.org/core/base/mpi"
)

var (
	// DefaultEncoding is the default encoding format for config files.
	// currently toml is the only supported format, but others could be added
	// if needed.
	DefaultEncoding = "toml"

	// IncludePaths is a list of file paths to try for finding config files
	// specified in Include field or via the command line --config --cfg or -c args.
	// Set this prior to calling Config -- default is current directory '.' and 'configs'
	IncludePaths = []string{".", "configs"}

	//	NonFlagArgs are the command-line args that remain after all the flags have
	// been processed.  This is set after the call to Config.
	NonFlagArgs = []string{}

	// ConfigFile is the name of the config file actually loaded, specified by the
	// -config or -cfg command-line arg or the default file given in Config
	ConfigFile string

	// Help is variable target for -help or -h args
	Help bool
)

// Config is the overall config setting function, processing config files
// and command-line arguments, in the following order:
//   - Apply any `default:` field tag default values.
//   - Look for `--config`, `--cfg`, or `-c` arg, specifying a config file on the command line.
//   - Fall back on default config file name passed to `Config` function, if arg not found.
//   - Read any `Include[s]` files in config file in deepest-first (natural) order,
//     then the specified config file last.
//   - if multiple config files are listed, then the first one that exists is used
//   - Process command-line args based on Config field names, with `.` separator
//     for sub-fields.
//   - Boolean flags are set on with plain -flag; use No prefix to turn off
//     (or explicitly set values to true or false).
//
// Also processes -help or -h and prints usage and quits immediately.
func Config(cfg any, defaultFile ...string) ([]string, error) {
	ConfigFile = ""

	var errs []error
	err := SetFromDefaults(cfg)
	if err != nil {
		errs = append(errs, err)
	}

	allArgs := make(map[string]reflect.Value)
	CommandArgs(allArgs)

	args := os.Args[1:]
	_, err = ParseArgs(cfg, args, allArgs, false) // false = ignore non-matches

	if Help {
		mpi.Println(Usage(cfg))
		os.Exit(0)
	}

	var cfgFiles []string
	if ConfigFile != "" {
		files := fsx.FindFilesOnPaths(IncludePaths, ConfigFile)
		if len(files) > 0 {
			cfgFiles = append(cfgFiles, ConfigFile)
		}
	} else {
		for _, fn := range defaultFile {
			files := fsx.FindFilesOnPaths(IncludePaths, fn)
			if len(files) > 0 {
				cfgFiles = append(cfgFiles, fn)
			}
		}
	}
	for _, fn := range cfgFiles {
		err = OpenWithIncludes(cfg, fn)
		if err != nil {
			errs = append(errs, err)
		}
	}
	NonFlagArgs, err = SetFromArgs(cfg, args)
	if err != nil {
		errs = append(errs, err)
	}
	return args, errors.Join(errs...)
}
