// Copyright (c) 2023, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package econfig

var (
	// DefaultEncoding is the default encoding format for config files.
	// currently toml is the only supported format, but others could be added
	// if needed.
	DefaultEncoding = "toml"

	// IncludePaths is a list of file paths to try for finding config files
	// specified in Include field or via the command line --config --cfg or -c args.
	// Set this prior to calling Config -- default is just current directory '.'
	IncludePaths = []string{"."}
)

// Config is the overall config setting function, processing config files
// and command-line arguments, in the following order:
//   - Apply any `def:` field tag default values.
//   - Look for `--config`, `--cfg`, or `-c` arg, specifying a config file on the command line.
//   - Fall back on default config file name passed to `Config` function, if arg not found.
//   - Read any `Include[s]` files in config file in deepest-first (natural) order, then the specified config file last.
//   - Process command-line args based on Config field names, with `.` separator for sub-fields (see field tags for shorthand and aliases)
func Config(cfg any, defaultFile string) ([]string, error) {
	var errs []error
	err := SetFromDefaults(cfg)
	if err != nil {
		errs = append(errs, err)
	}
	// todo: register args, process config, use default instead
	// SetFromConfigArg(cfg, defaultFile)
	err = OpenWithIncludes(cfg, defaultFile)
	if err != nil {
		errs = append(errs, err)
	}
	args, err := SetFromArgs(cfg)
	if err != nil {
		errs = append(errs, err)
	}
	return args, AllErrors(errs)
}
