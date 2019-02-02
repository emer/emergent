// Copyright (c) 2019, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// ra25 runs a simple random-associator 5x5 = 25 four-layer leabra network
package main

import (
	"github.com/emer/emergent/basic/leabra"
	"github.com/emer/emergent/dtable"
	"github.com/emer/emergent/emer"
	"github.com/emer/emergent/etensor"
	"github.com/emer/emergent/prjn"
)

var Net *leabra.Network
var InPats *dtable.Table

var Pars = emer.ParamStyle{
	"Prjn": {
		"Prjn.Learn.Norm.On":     1,
		"Prjn.Learn.Momentum.On": 1,
	},
	".TopDown": {
		"Prjn.WtScale.Rel": 0.2, // this is generally quite important
	},
}

func ConfigNet(net *leabra.Network) {
	net.Name = "RA25"
	inLay := net.AddLayer("Input", []int{5, 5}, leabra.Input)
	hid1Lay := net.AddLayer("Hidden1", []int{5, 5}, leabra.Hidden)
	hid2Lay := net.AddLayer("Hidden2", []int{5, 5}, leabra.Hidden)
	outLay := net.AddLayer("Output", []int{5, 5}, leabra.Target)

	net.ConnectLayers(inLay, hid1Lay, prjn.NewFull())
	net.ConnectLayers(hid1Lay, hid2Lay, prjn.NewFull())
	net.ConnectLayers(hid2Lay, outLay, prjn.NewFull())

	outHid2 := net.ConnectLayers(outLay, hid2Lay, prjn.NewFull())
	outHid2.Class = "TopDown"
	hid2Hid1 := net.ConnectLayers(hid2Lay, hid1Lay, prjn.NewFull())
	hid2Hid1.Class = "TopDown"

	net.Defaults()
	net.StyleParams(Pars)
	net.Build()
	net.InitWts()
}

func ConfigInPats(dt *dtable.Table) {
	dt.SetFromSchema(dtable.Schema{
		{"name", etensor.STRING, nil, nil},
		{"pattern", etensor.FLOAT32, []int{5, 5}, []string{"Y", "X"}},
	}, 25)

	// for pi := 0; pi < 4; pi++ {
	// 	InPats.Set([]int{pi, pi, 0}, 1)
	// }
}

func main() {
}
