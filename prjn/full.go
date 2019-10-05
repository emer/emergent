// Copyright (c) 2019, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package prjn

import (
	"github.com/emer/etable/etensor"
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
	sendn, recvn, cons = NewTensors(send, recv)
	cons.Values.SetAll(true)
	nsend := send.Len()
	nrecv := recv.Len()
	if same && !fp.SelfCon {
		for i := 0; i < nsend; i++ { // nsend = nrecv
			off := i*nsend + i
			cons.Values.Set(off, false)
		}
		nsend--
		nrecv--
	}
	rnv := recvn.Values
	for i := range rnv {
		rnv[i] = int32(nsend)
	}
	snv := sendn.Values
	for i := range snv {
		snv[i] = int32(nrecv)
	}
	return
}
