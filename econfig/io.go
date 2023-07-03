// Copyright (c) 2023, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package econfig

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml" // either one of these works fine
	// "github.com/pelletier/go-toml/v2"
)

// Open reads config from given config file,
// looking on IncludePaths for the file.
func Open(cfg any, file string) error {
	filename, err := FindFileOnPaths(IncludePaths, file)
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

// FindFileOnPaths attempts to locate given file on given list of paths,
// returning the full Abs path to file if found, else error
func FindFileOnPaths(paths []string, file string) (string, error) {
	for _, path := range paths {
		filePath := filepath.Join(path, file)
		ok, _ := FileExists(filePath)
		if ok {
			return filePath, nil
		}
	}
	return "", fmt.Errorf("FindFileOnPaths: unable to find file: %s on paths: %v\n", file, paths)
}

func FileExists(filePath string) (bool, error) {
	fileInfo, err := os.Stat(filePath)
	if err == nil {
		return !fileInfo.IsDir(), nil
	}
	if errors.Is(err, fs.ErrNotExist) {
		return false, nil
	}
	return false, err
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
