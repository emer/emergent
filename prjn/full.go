// Copyright (c) 2019, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package prjn

import (
	"github.com/emer/emergent/etensor"
)

// Full implements full all-to-all pattern of connectivity between two layers
type Full struct {
	SelfCon bool `desc:"if true, and connecting layer to itself (self projection), then make a self-connection from unit to itself"`
}

func NewFull() *Full {
	return &Full{}
}

func (fp *Full) Name() string {
	return "Full"
}

func (fp *Full) Connect(recv, send *etensor.Shape, same bool) (recvn, sendn *etensor.Int32, cons *etensor.Bits) {
	// todo: exclude self!
	recvn, sendn, cons = NewTensors(recv, send)
	cons.Values.SetAll(true)
	nsend := send.Len()
	nrecv := recv.Len()
	rnv := recvn.Values
	for i := 0; i < nrecv; i++ {
		rnv[i] = int32(nsend)
	}
	snv := sendn.Values
	for i := 0; i < nsend; i++ {
		snv[i] = int32(nrecv)
	}
	return
}

func (fp *Full) HasWeights() bool {
	return false
}

func (fp *Full) Weights(recvn, sendn *etensor.Int32, cons *etensor.Bits) []float32 {
	return nil
}
