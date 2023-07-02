// Copyright (c) 2023, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// note: parsing code adapted from pflag package https://github.com/spf13/pflag
// Copyright (c) 2012 Alex Ogier. All rights reserved.
// Copyright (c) 2012 The Go Authors. All rights reserved.

package econfig

// SetFromArgs sets Config values from command-line args,
// based on the field names in the Config struct.
// Returns any args that did not start with a `-` flag indicator.
// For more robust error processing, it is assumed that all flagged args (-)
// must refer to fields in the config, so any that fail to match trigger
// an error.  Errors can also result from parsing.
// Errors are automatically logged because these are user-facing.
func SetFromArgs(cfg any) ([]string, error) {
	return nil, nil
}

// parseArgs does the actual arg parsing
func parseArgs(cfg any, args []string) ([]string, error) {
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
			// args, err = f.parseLongArg(s, args, fn)
		} else {
			// args, err = f.parseShortArg(s, args, fn)
		}
		if err != nil {
			return leftovers, err
		}
	}
	return leftovers, nil
}
