// Copyright (c) 2019, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package weights

import (
	"encoding/json"
	"io"
	"log"
)

// Prec is precision for weight output in text formats
var Prec = 4

// Network is used for decoding weights in ReadJSON
type Network struct {
	Network  string
	MetaData map[string]string // used for optional network-level params, metadata
	Layers   []Layer
}

// Layer is used for decoding weights in ReadJSON
type Layer struct {
	Layer    string
	MetaData map[string]string // used for optional layer-level params, metadata such as 	ActMAvg, ActPAvg
	Prjns    []Prjn
}

// Prjn is used for decoding weights in ReadJSON
type Prjn struct {
	From     string
	MetaData map[string]string // used for optional prjn-level params, metadata such as GScale
	Rs       []Recv
}

// Recv is used for decoding weights in ReadJSON
type Recv struct {
	Ri int
	N  int
	Si []int
	Wt []float32
}

// NetReadJSON reads weights for entire network in a JSON format into Network structure
func NetReadJSON(r io.Reader) (*Network, error) {
	nw := &Network{}
	dec := json.NewDecoder(r)
	err := dec.Decode(nw) // this is way to do it on reader instead of bytes
	if err == io.EOF {
		return nil, nil
	}
	if err != nil {
		log.Println(err)
	}
	return nw, nil
}

// LayReadJSON reads weights for layer in a JSON format into Layer structure
func LayReadJSON(r io.Reader) (*Layer, error) {
	lw := &Layer{}
	dec := json.NewDecoder(r)
	err := dec.Decode(lw) // this is way to do it on reader instead of bytes
	if err == io.EOF {
		return nil, nil
	}
	if err != nil {
		log.Println(err)
	}
	return lw, nil
}

// PrjnReadJSON reads weights for prjn in a JSON format into Prjn structure
func PrjnReadJSON(r io.Reader) (*Prjn, error) {
	pw := &Prjn{}
	dec := json.NewDecoder(r)
	err := dec.Decode(pw) // this is way to do it on reader instead of bytes
	if err == io.EOF {
		return nil, nil
	}
	if err != nil {
		log.Println(err)
	}
	return pw, nil
}
