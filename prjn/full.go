// Copyright (c) 2019, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package prjn

import (
	"github.com/emer/dtable/etensor"
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

func (fp *Full) Connect(send, recv *etensor.Shape, same bool) (sendn, recvn *etensor.Int32, cons *etensor.Bits) {
	// todo: exclude self!
	sendn, recvn, cons = NewTensors(send, recv)
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

func (fp *Full) Weights(sendn, recvn *etensor.Int32, cons *etensor.Bits) []float32 {
	return nil
}
