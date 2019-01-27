// Copyright (c) 2019, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package leabra

import (
	"testing"

	"github.com/emer/emergent/prjn"
)

var TestNet Network

func TestMakeNet(t *testing.T) {
	in := TestNet.AddLayer("Input", []int{5, 5}, Input)
	hid := TestNet.AddLayer("Hidden", []int{5, 5}, Hidden)
	out := TestNet.AddLayer("Output", []int{5, 5}, Target)

	inHid := TestNet.ConnectLayers(hid, in, prjn.NewFull())
	HidOut := TestNet.ConnectLayers(out, hid, prjn.NewFull())
	OutHid := TestNet.ConnectLayers(hid, out, prjn.NewFull())
	_ = inHid
	_ = HidOut
	_ = OutHid

	TestNet.Defaults()
}
