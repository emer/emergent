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

	"cogentcore.org/core/core"
)

// FlexVal is a specific flexible value for the Flex parameter map
// that implements the StylerObject interface for CSS-style selection logic.
// The field names are abbreviated because full names are used in StylerObject.
type FlexVal struct {
	// name of this specific object, matches #Name selections
	Name string

	// type name of this object, matches plain TypeName selections
	Type string

	// space-separated list of class name(s), match the .Class selections
	Class string

	// actual object with data that is set by the parameters
	Object any

	// History of params applied
	History HistoryImpl `table:"-"`
}

func (fv *FlexVal) StyleType() string {
	return fv.Type
}

func (fv *FlexVal) StyleClass() string {
	return fv.Class
}

func (fv *FlexVal) StyleName() string {
	return fv.Name
}

func (fv *FlexVal) StyleObject() any {
	return fv.Object
}

func (fv *FlexVal) CopyFrom(cp *FlexVal) {
	fv.Name = cp.Name // these should be the same, but copy anyway
	fv.Type = cp.Type
	fv.Class = cp.Class
	if hyp, ok := fv.Object.(Hypers); ok { // this is the main use-case
		if cph, ok := cp.Object.(Hypers); ok {
			hyp.CopyFrom(cph)
		}
	}
}

// ParamsHistoryReset resets parameter application history
func (fv *FlexVal) ParamsHistoryReset() {
	fv.History.ParamsHistoryReset()
}

// ParamsApplied is just to satisfy History interface so reset can be applied
func (fv *FlexVal) ParamsApplied(sel *Sel) {
	fv.History.ParamsApplied(sel)
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

func (fl *Flex) StyleType() string { // note: assuming all same type for this purpose
	for _, fv := range *fl {
		return fv.StyleType()
	}
	return "Flex"
}

func (fl *Flex) StyleClass() string {
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
		(*fl)[vl.Name] = &inst
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
func (fl *Flex) SaveJSON(filename core.Filename) error {
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
