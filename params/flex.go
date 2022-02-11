// Copyright (c) 2019, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package params

import (
	"encoding/json"
	"io"
	"io/ioutil"
	"log"

	"github.com/goki/gi/gi"
)

// FlexVal is a specific flexible value for the Flex parameter map
// that implements the StylerObj interface for CSS-style selection logic
type FlexVal struct {
	Nm   string      `desc:"name of this specific object -- matches #Name selections"`
	Type string      `desc:"type name of this object -- matches plain TypeName selections"`
	Cls  string      `desc:"space-separated list of class name(s) -- match the .Class selections"`
	Obj  interface{} `desc:"actual object with data that is set by the parameters"`
}

func (fv *FlexVal) TypeName() string {
	return fv.Type
}

func (fv *FlexVal) Class() string {
	return fv.Cls
}

func (fv *FlexVal) Name() string {
	return fv.Nm
}

func (fv *FlexVal) Object() interface{} {
	return fv.Obj
}

// Flex supports arbitrary named parameter values that can be set
// by a Set of parameters, as a map of interface{} objects.
// First initialize the map with set of names and a type to create
// blank values, then apply the Set to it.
type Flex map[string]*FlexVal

// Init initializes the Flex map with given set of flex values.
func (fl *Flex) Init(vals []FlexVal) {
	*fl = make(Flex, len(vals))
	for _, vl := range vals {
		(*fl)[vl.Nm] = &vl
	}
}

// WriteJSON saves hypers to a JSON-formatted file.
func (fl *Flex) WriteJSON(w io.Writer) error {
	b, err := json.MarshalIndent(fl, "", "  ")
	if err != nil {
		log.Println(err) // unlikely
		return err
	}
	w.Write(b)
	return err
}

// SaveJSON saves hypers to a JSON-formatted file.
func (fl *Flex) SaveJSON(filename gi.FileName) error {
	b, err := json.MarshalIndent(fl, "", "  ")
	if err != nil {
		log.Println(err) // unlikely
		return err
	}
	err = ioutil.WriteFile(string(filename), b, 0644)
	if err != nil {
		log.Println(err)
	}
	return err
}
