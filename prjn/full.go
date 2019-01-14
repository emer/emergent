// Copyright (c) 2019, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package prjn

import (
	"github.com/apache/arrow/go/arrow/tensor"
	"github.com/emer/emergent/etensor"
)

// Full implements full all-to-all pattern of connectivity between two layers
type Full struct {
	SelfCon bool `desc:"if true, and connecting layer to itself (self projection), then make a self-connection from unit to itself"`
}

func (fp *Full) Connect(recv, send *etensor.Shape, same bool) (recvn, sendn *tensor.Int32, cons *etensor.Bits) {
	// todo: exclude self!
	recvn = tensor.NewInt32(nil, recv.Shape(), recv.Strides(), recv.DimNames())
	sendn = tensor.NewInt32(nil, send.Shape(), send.Strides(), send.DimNames())
	csh := etensor.AddShapes(recv, send)
	cons = etensor.NewBits(csh.Shape(), csh.Strides(), csh.DimNames())
	cons.Values.SetAll(true)
	nsend := send.Len()
	nrecv := recv.Len()
	rnv := recvn.Int32Values()
	for i := 0; i < nrecv; i++ {
		rnv[i] = int32(nsend)
	}
	snv := sendn.Int32Values()
	for i := 0; i < nsend; i++ {
		snv[i] = int32(nrecv)
	}
	return
}

func (fp *Full) HasWeights() bool {
	return false
}

func (fp *Full) Weights(recvn, sendn *tensor.Int32, cons *etensor.Bits) []float32 {
	return nil
}
