// Copyright (c) 2019, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package prjn

import (
	"github.com/emer/etable/v2/etensor"
)

// OneToOne implements point-to-point one-to-one pattern of connectivity between two layers
type OneToOne struct {

	// number of recv connections to make (0 for entire size of recv layer)
	NCons int

	// starting unit index for sending connections
	SendStart int

	// starting unit index for recv connections
	RecvStart int
}

func NewOneToOne() *OneToOne {
	return &OneToOne{}
}

func (ot *OneToOne) Name() string {
	return "OneToOne"
}

func (ot *OneToOne) Connect(send, recv *etensor.Shape, same bool) (sendn, recvn *etensor.Int32, cons *etensor.Bits) {
	sendn, recvn, cons = NewTensors(send, recv)
	nsend := send.Len()
	nrecv := recv.Len()
	rnv := recvn.Values
	snv := sendn.Values
	ncon := nrecv
	if ot.NCons > 0 {
		ncon = min(ot.NCons, nrecv)
	}
	for i := 0; i < ncon; i++ {
		ri := ot.RecvStart + i
		si := ot.SendStart + i
		if ri >= nrecv || si >= nsend {
			break
		}
		off := ri*nsend + si
		cons.Values.Set(off, true)
		rnv[ri] = 1
		snv[si] = 1
	}
	return
}
