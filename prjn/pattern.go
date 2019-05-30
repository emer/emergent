// Copyright (c) 2019, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

/*
Package prjn is a separate package for defining patterns of connectivity between layers
(i.e., the ProjectionSpecs from C++ emergent).  This is done using a fully independent
structure that *only* knows about the shapes of the two layers, and it returns a fully general
bitmap representation of the pattern of connectivity between them.

The algorithm-specific leabra.Prjn code then uses these patterns to do all the nitty-gritty
of connecting up neurons.

This makes the projection code *much* simpler compared to the ProjectionSpec in C++ emergent,
which was involved in both creating the pattern and also all the complexity of setting up the
actual connections themselves.  This should be the *last* time any of those projection patterns
need to be written (having re-written this code too many times in the C++ version as the details
of memory allocations changed).
*/
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

	// HasWeights returns true if this projection can provide initial synaptic weights
	// for connected units, via the Weights method.
	HasWeights() bool

	// Weights provides initial synaptic weights for each of the connected units.
	// For efficiency, the weights are provided for connected units -- i.e., only for
	// the cons = true bits -- values are in receiver-outer, sender-inner order.
	// Typically, these weights reflect an overall topographic pattern and do NOT include
	// random perturbations, which can be added on top.
	Weights(sendn, recvn *etensor.Int32, cons *etensor.Bits) []float32
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
