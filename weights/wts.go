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

// Layer is temp structure for holding decoded weights, one for each layer
type Layer struct {
	Layer    string
	MetaData map[string]string // used for optional layer-level params, metadata such as 	ActMAvg, ActPAvg
	Prjns    []Prjn            // receiving projections
}

// Prjn is temp structure for holding decoded weights, one for each projection
type Prjn struct {
	From     string
	MetaData map[string]string // used for optional prjn-level params, metadata such as GScale
	Rs       []Recv
}

// Recv is temp structure for holding decoded weights, one for each recv unit
type Recv struct {
	Ri int
	N  int
	Si []int
	Wt []float32
}
