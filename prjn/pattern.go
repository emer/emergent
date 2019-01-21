// Copyright (c) 2019, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package prjn

import (
	"github.com/apache/arrow/go/arrow/tensor"
	"github.com/emer/emergent/etensor"
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
	Connect(recv, send *etensor.Shape, same bool) (recvn, sendn *tensor.Int32, cons *etensor.Bits)

	// HasWeights returns true if this projection can provide initial synaptic weights
	// for connected units, via the Weights method.
	HasWeights() bool

	// Weights provides initial synaptic weights for each of the connected units.
	// For efficiency, the weights are provided for connected units -- i.e., only for
	// the cons = true bits -- values are in receiver-outer, sender-inner order.
	// Typically, these weights reflect an overall topographic pattern and do NOT include
	// random perturbations, which can be added on top.
	Weights(recvn, sendn *tensor.Int32, cons *etensor.Bits) []float32
}
