// Copyright (c) 2019, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package params

import (
	"bytes"
	"encoding/json"
	"io"
	"io/ioutil"
	"log"

	"goki.dev/gi"
)

// FlexVal is a specific flexible value for the Flex parameter map
// that implements the StylerObj interface for CSS-style selection logic
type FlexVal struct {

	// name of this specific object -- matches #Name selections
	Nm string

	// type name of this object -- matches plain TypeName selections
	Type string

	// space-separated list of class name(s) -- match the .Class selections
	Cls string

	// actual object with data that is set by the parameters
	Obj any
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

func (fv *FlexVal) Object() any {
	return fv.Obj
}

func (fv *FlexVal) CopyFrom(cp *FlexVal) {
	fv.Nm = cp.Nm // these should be the same, but copy anyway
	fv.Type = cp.Type
	fv.Cls = cp.Cls
	if hyp, ok := fv.Obj.(Hypers); ok { // this is the main use-case
		if cph, ok := cp.Obj.(Hypers); ok {
			hyp.CopyFrom(cph)
		}
	}
}

// Flex supports arbitrary named parameter values that can be set
// by a Set of parameters, as a map of any objects.
// First initialize the map with set of names and a type to create
// blank values, then apply the Set to it.
type Flex map[string]*FlexVal

// Make makes the map if it is nil (otherwise does nothing)
func (fl *Flex) Make() {
	if *fl != nil {
		return
	}
	*fl = make(Flex)
}

func (fl *Flex) TypeName() string { // note: assuming all same type for this purpose
	for _, fv := range *fl {
		return fv.TypeName()
	}
	return "Flex"
}

func (fl *Flex) Class() string {
	return ""
}

func (fl *Flex) Name() string {
	return ""
}

// Init initializes the Flex map with given set of flex values.
func (fl *Flex) Init(vals []FlexVal) {
	*fl = make(Flex, len(vals))
	for _, vl := range vals {
		inst := vl
		(*fl)[vl.Nm] = &inst
	}
}

// ApplySheet applies given sheet of parameters to each element in Flex
func (fl *Flex) ApplySheet(sheet *Sheet, setMsg bool) {
	for _, vl := range *fl {
		sheet.Apply(vl, setMsg)
	}
}

// CopyFrom copies hyper vals from source
func (fl *Flex) CopyFrom(cp Flex) {
	fl.Make()
	for nm, fv := range cp {
		if sfv, has := (*fl)[nm]; has {
			sfv.CopyFrom(fv)
		} else {
			sfv := &FlexVal{}
			sfv.CopyFrom(fv)
			(*fl)[nm] = sfv
		}
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

// JSONString returns a string representation of Flex params
func (fl *Flex) JSONString() string {
	var buf bytes.Buffer
	fl.WriteJSON(&buf)
	return string(buf.Bytes())
}

// SaveJSON saves hypers to a JSON-formatted file.
func (fl *Flex) SaveJSON(filename gi.Filename) error {
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
