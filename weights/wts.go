// Copyright (c) 2019, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package weights

// Network is temp structure for holding decoded weights
type Network struct {
	Network  string
	MetaData map[string]string // used for optional network-level params, metadata
	Layers   []Layer
}

func (nt *Network) SetMetaData(key, val string) {
	if nt.MetaData == nil {
		nt.MetaData = make(map[string]string)
	}
	nt.MetaData[key] = val
}

// Layer is temp structure for holding decoded weights, one for each layer
type Layer struct {
	Layer    string
	MetaData map[string]string // used for optional layer-level params, metadata such as 	ActMAvg, ActPAvg
	Prjns    []Prjn            // receiving projections
}

func (ly *Layer) SetMetaData(key, val string) {
	if ly.MetaData == nil {
		ly.MetaData = make(map[string]string)
	}
	ly.MetaData[key] = val
}

// Prjn is temp structure for holding decoded weights, one for each projection
type Prjn struct {
	From     string
	MetaData map[string]string // used for optional prjn-level params, metadata such as GScale
	Rs       []Recv
}

func (pj *Prjn) SetMetaData(key, val string) {
	if pj.MetaData == nil {
		pj.MetaData = make(map[string]string)
	}
	pj.MetaData[key] = val
}

// Recv is temp structure for holding decoded weights, one for each recv unit
type Recv struct {
	Ri    int
	N     int
	Si    []int
	Wt    []float32
	Scale []float32
}
