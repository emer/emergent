// Copyright (c) 2023, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package econfig

import (
	"io/fs"
	"strings"

	"cogentcore.org/core/iox/tomlx"
	"github.com/emer/emergent/v2/empi/mpi"
)

// OpenWithIncludes reads config from given config file,
// looking on IncludePaths for the file,
// and opens any Includes specified in the given config file
// in the natural include order so includee overwrites included settings.
// Is equivalent to Open if there are no Includes.
// Returns an error if any of the include files cannot be found on IncludePath.
func OpenWithIncludes(cfg any, file string) error {
	err := tomlx.OpenFromPaths(cfg, file, IncludePaths)
	if err != nil {
		return err
	}
	incsObj, hasIncludes := cfg.(Includeser)
	incObj, hasInclude := cfg.(Includer)
	if !hasInclude && !hasIncludes {
		return nil // no further processing
	}
	var incs []string
	if hasIncludes {
		incs, err = IncludesStack(incsObj)
	} else {
		incs, err = IncludeStack(incObj)
	}
	ni := len(incs)
	if err != nil || ni == 0 {
		return err
	}
	for i := ni - 1; i >= 0; i-- {
		inc := incs[i]
		err = tomlx.OpenFromPaths(cfg, inc, IncludePaths)
		if err != nil {
			mpi.Println(err)
		}
	}
	// reopen original
	tomlx.OpenFromPaths(cfg, file, IncludePaths)
	if hasIncludes {
		*incsObj.IncludesPtr() = incs
	} else {
		*incObj.IncludePtr() = strings.Join(incs, ",")
	}
	return err
}

// OpenFS reads config from given TOML file,
// using the fs.FS filesystem -- e.g., for embed files.
func OpenFS(cfg any, fsys fs.FS, file string) error {
	return tomlx.OpenFS(cfg, fsys, file)
}

// Save writes TOML to given file.
func Save(cfg any, file string) error {
	return tomlx.Save(cfg, file)
}
