// Copyright (c) 2019, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package prjn

import (
	"github.com/emer/etable/etensor"
)

// PoolSameUnit connects a given unit to the unit at the same index
// across all the pools in a layer.
// Pools are the outer-most two dimensions of a 4D layer shape.
// This is most sensible when pools have same numbers of units in send and recv.
// This is typically used for lateral topography-inducing connectivity
// and can also serve to reduce a pooled layer down to a single pool.
// The logic works if either layer does not have pools.
// If neither is 4D, then it is equivalent to OneToOne.
type PoolSameUnit struct {
	SelfCon bool `desc:"if true, and connecting layer to itself (self projection), then make a self-connection from unit to itself"`
}

func NewPoolSameUnit() *PoolSameUnit {
	return &PoolSameUnit{}
}

func (ot *PoolSameUnit) Name() string {
	return "PoolSameUnit"
}

func (ot *PoolSameUnit) Connect(send, recv *etensor.Shape, same bool) (sendn, recvn *etensor.Int32, cons *etensor.Bits) {
	switch {
	case send.NumDims() == 4 && recv.NumDims() == 4:
		return ot.ConnectPools(send, recv, same)
	case send.NumDims() == 2 && recv.NumDims() == 4:
		return ot.ConnectRecvPool(send, recv, same)
	case send.NumDims() == 4 && recv.NumDims() == 2:
		return ot.ConnectSendPool(send, recv, same)
	case send.NumDims() == 2 && recv.NumDims() == 2:
		return ot.ConnectOneToOne(send, recv, same)
	}
	return
}

// ConnectPools is when both recv and send have pools
func (ot *PoolSameUnit) ConnectPools(send, recv *etensor.Shape, same bool) (sendn, recvn *etensor.Int32, cons *etensor.Bits) {
	sendn, recvn, cons = NewTensors(send, recv)
	sNtot := send.Len()
	sNp := send.Dim(0) * send.Dim(1)
	rNp := recv.Dim(0) * recv.Dim(1)
	sNu := send.Dim(2) * send.Dim(3)
	rNu := recv.Dim(2) * recv.Dim(3)
	rnv := recvn.Values
	snv := sendn.Values
	for rpi := 0; rpi < rNp; rpi++ {
		for rui := 0; rui < rNu; rui++ {
			if rui >= sNu {
				break
			}
			ri := rpi*rNu + rui
			for spi := 0; spi < sNp; spi++ {
				if same && !ot.SelfCon && spi == rpi {
					continue
				}
				si := spi*sNu + rui
				off := ri*sNtot + si
				cons.Values.Set(off, true)
				rnv[ri]++
				snv[si]++
			}
		}
	}
	return
}

// ConnectRecvPool is when recv has pools but send doesn't
func (ot *PoolSameUnit) ConnectRecvPool(send, recv *etensor.Shape, same bool) (sendn, recvn *etensor.Int32, cons *etensor.Bits) {
	sendn, recvn, cons = NewTensors(send, recv)
	sNtot := send.Len()
	rNp := recv.Dim(0) * recv.Dim(1)
	sNu := send.Dim(0) * send.Dim(1)
	rNu := recv.Dim(2) * recv.Dim(3)
	rnv := recvn.Values
	snv := sendn.Values
	for rpi := 0; rpi < rNp; rpi++ {
		for rui := 0; rui < rNu; rui++ {
			if rui >= sNu {
				break
			}
			ri := rpi*rNu + rui
			si := rui
			off := ri*sNtot + si
			cons.Values.Set(off, true)
			rnv[ri]++
			snv[si]++
		}
	}
	return
}

// ConnectSendPool is when send has pools but recv doesn't
func (ot *PoolSameUnit) ConnectSendPool(send, recv *etensor.Shape, same bool) (sendn, recvn *etensor.Int32, cons *etensor.Bits) {
	sendn, recvn, cons = NewTensors(send, recv)
	sNtot := send.Len()
	sNp := send.Dim(0) * send.Dim(1)
	sNu := send.Dim(2) * send.Dim(3)
	rNu := recv.Dim(0) * recv.Dim(1)
	rnv := recvn.Values
	snv := sendn.Values
	for rui := 0; rui < rNu; rui++ {
		if rui >= sNu {
			break
		}
		ri := rui
		for spi := 0; spi < sNp; spi++ {
			si := spi*sNu + rui
			off := ri*sNtot + si
			cons.Values.Set(off, true)
			rnv[ri]++
			snv[si]++
		}
	}
	return
}

// copy of OneToOne.Connect
func (ot *PoolSameUnit) ConnectOneToOne(send, recv *etensor.Shape, same bool) (sendn, recvn *etensor.Int32, cons *etensor.Bits) {
	sendn, recvn, cons = NewTensors(send, recv)
	sNtot := send.Len()
	sNu := send.Dim(0) * send.Dim(1)
	rNu := recv.Dim(0) * recv.Dim(1)
	rnv := recvn.Values
	snv := sendn.Values
	for rui := 0; rui < rNu; rui++ {
		if rui >= sNu {
			break
		}
		ri := rui
		si := rui
		off := ri*sNtot + si
		cons.Values.Set(off, true)
		rnv[ri] = 1
		snv[si] = 1
	}
	return
}
