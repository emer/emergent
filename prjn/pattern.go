// Copyright (c) 2019, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package prjn

import (
	"github.com/emer/etable/etensor"
)

// Pattern defines a pattern of connectivity between two layers.
// The pattern is stored efficiently using a bitslice tensor of binary values indicating
// presence or absence of connection between two items.
// A receiver-based organization is generally assumed but connectivity can go either way.
type Pattern interface {
	// Name returns the name of the pattern -- i.e., the "type" name of the actual pattern generatop
	Name() string

	// Connect connects layers with the given shapes, returning the pattern of connectivity
	// as a bits tensor with shape = recv + send shapes, using row-major ordering with outer-most
	// indexes first (i.e., for each recv unit, there is a full inner-level of sender bits).
	// The number of connections for each recv and each send unit are also returned in
	// recvn and send tensors, each the shape of send and recv respectively.
	// The same flag should be set to true if the send and recv layers are the same (i.e., a self-connection)
	// often there are some different options for such connections.
	Connect(send, recv *etensor.Shape, same bool) (sendn, recvn *etensor.Int32, cons *etensor.Bits)
}

// NewTensors returns the tensors used for Connect method, based on layer sizes
func NewTensors(send, recv *etensor.Shape) (sendn, recvn *etensor.Int32, cons *etensor.Bits) {
	sendn = etensor.NewInt32Shape(send, nil)
	recvn = etensor.NewInt32Shape(recv, nil)
	csh := etensor.AddShapes(recv, send)
	cons = etensor.NewBitsShape(csh)
	return
}

// ConsStringFull returns a []byte string showing the pattern of connectivity.
// if perRecv is true then it displays the sending connections
// per each recv unit -- otherwise it shows the entire matrix
// as a 2D matrix
func ConsStringFull(send, recv *etensor.Shape, cons *etensor.Bits) []byte {
	nsend := send.Len()
	nrecv := recv.Len()

	one := []byte("1 ")
	zero := []byte("0 ")

	sz := nrecv * (nsend*2 + 1)
	b := make([]byte, 0, sz)

	for ri := 0; ri < nrecv; ri++ {
		for si := 0; si < nsend; si++ {
			off := ri*nsend + si
			cn := cons.Value1D(off)
			if cn {
				b = append(b, one...)
			} else {
				b = append(b, zero...)
			}
		}
		b = append(b, byte('\n'))
	}
	return b
}

// ConsStringPerRecv returns a []byte string showing the pattern of connectivity
// organized by receiving unit, showing the sending connections per each
func ConsStringPerRecv(send, recv *etensor.Shape, cons *etensor.Bits) []byte {
	return nil
}
