// Copyright (c) 2023, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package econfig

import (
	"bufio"
	"io"
	"io/fs"
	"io/ioutil"
	"log"
	"os"

	"github.com/BurntSushi/toml" // either one of these works fine
	"github.com/goki/ki/dirs"
)

// Open reads config from given config file,
// looking on IncludePaths for the file.
func Open(cfg any, file string) error {
	filename, err := dirs.FindFileOnPaths(IncludePaths, file)
	if err != nil {
		log.Println(err)
		return err
	}
	// _, err = toml.DecodeFile(fp, cfg)
	fp, err := os.Open(filename)
	defer fp.Close()
	if err != nil {
		log.Println(err)
		return err
	}
	return Read(cfg, bufio.NewReader(fp))
}

// OpenFS reads config from given config file,
// using the fs.FS filesystem -- e.g., for embed files.
func OpenFS(cfg any, fsys fs.FS, file string) error {
	fp, err := fsys.Open(file)
	defer fp.Close()
	if err != nil {
		log.Println(err)
		return err
	}
	return Read(cfg, bufio.NewReader(fp))
}

// Read reads config from given reader,
// looking on IncludePaths for the file.
func Read(cfg any, reader io.Reader) error {
	b, err := ioutil.ReadAll(reader)
	if err != nil {
		log.Println(err)
		return err
	}
	return ReadBytes(cfg, b)
}

// ReadBytes reads config from given bytes,
// looking on IncludePaths for the file.
func ReadBytes(cfg any, b []byte) error {
	err := toml.Unmarshal(b, cfg)
	if err != nil {
		log.Println(err)
		return err
	}
	return nil
}

// OpenWithIncludes reads config from given config file,
// looking on IncludePaths for the file,
// and opens any Includes specified in the given config file
// in the natural include order so includee overwrites included settings.
// Is equivalent to Open if there are no Includes.
// Returns an error if any of the include files cannot be found on IncludePath.
func OpenWithIncludes(cfg any, file string) error {
	err := Open(cfg, file)
	if err != nil {
		return err
	}
	incfg, ok := cfg.(Includer)
	if !ok {
		return err
	}
	incs, err := IncludeStack(incfg)
	ni := len(incs)
	if ni == 0 {
		return err
	}
	for i := ni - 1; i >= 0; i-- {
		inc := incs[i]
		err = Open(cfg, inc)
		if err != nil {
			log.Println(err)
		}
	}
	// reopen original
	Open(cfg, file)
	*incfg.IncludesPtr() = incs
	return err
}

/////////////////////////////////////////////////////////
//  Saving

// Save writes config to given config file.
func Save(cfg any, file string) error {
	// _, err = toml.DecodeFile(fp, cfg)
	fp, err := os.Create(file)
	defer fp.Close()
	if err != nil {
		log.Println(err)
		return err
	}
	bw := bufio.NewWriter(fp)
	err = Write(cfg, bw)
	if err != nil {
		log.Println(err)
		return err
	}
	err = bw.Flush()
	if err != nil {
		log.Println(err)
	}
	return err
}

// Write writes config to given writer.
func Write(cfg any, writer io.Writer) error {
	enc := toml.NewEncoder(writer)
	return enc.Encode(cfg)
}
