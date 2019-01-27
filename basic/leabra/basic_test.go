// Copyright (c) 2019, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package leabra

import (
	"testing"

	"github.com/emer/emergent/etensor"
	"github.com/emer/emergent/prjn"
)

var TestNet Network
var InPats *etensor.Float32

func TestMakeNet(t *testing.T) {
	in := TestNet.AddLayer("Input", []int{4, 1}, Input)
	hid := TestNet.AddLayer("Hidden", []int{4, 1}, Hidden)
	out := TestNet.AddLayer("Output", []int{4, 1}, Target)

	inHid := TestNet.ConnectLayers(hid, in, prjn.NewOneToOne())
	HidOut := TestNet.ConnectLayers(out, hid, prjn.NewOneToOne())
	OutHid := TestNet.ConnectLayers(hid, out, prjn.NewOneToOne())
	_ = inHid
	_ = HidOut
	_ = OutHid

	TestNet.Defaults()
	TestNet.Build()
}

func TestInPats(t *testing.T) {
	InPats = etensor.NewFloat32([]int{4, 4, 1}, nil, []string{"pat", "Y", "X"})
	for pi := 0; pi < 4; pi++ {
		InPats.Set([]int{pi, pi, 0}, 1)
	}
}
