// Copyright (c) 2019, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package prjn

import (
	"github.com/emer/emergent/etensor"
	"github.com/goki/ki/ints"
)

// OneToOne implements point-to-point one-to-one pattern of connectivity between two layers
type OneToOne struct {
	NCons     int `desc:"number of recv connections to make (0 for entire size of recv layer)"`
	RecvStart int `desc:"starting unit index for recv connections"`
	SendStart int `desc:"starting unit index for sending connections"`
}

func NewOneToOne() *OneToOne {
	return &OneToOne{}
}

func (ot *OneToOne) Name() string {
	return "OneToOne"
}

func (ot *OneToOne) Connect(recv, send *etensor.Shape, same bool) (recvn, sendn *etensor.Int32, cons *etensor.Bits) {
	recvn, sendn, cons = NewTensors(recv, send)
	nsend := send.Len()
	nrecv := recv.Len()
	rnv := recvn.Values
	snv := sendn.Values
	ncon := nrecv
	if ot.NCons > 0 {
		ncon = ints.MinInt(ot.NCons, nrecv)
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

func (ot *OneToOne) HasWeights() bool {
	return false
}

func (ot *OneToOne) Weights(recvn, sendn *etensor.Int32, cons *etensor.Bits) []float32 {
	return nil
}
