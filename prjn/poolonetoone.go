// Copyright (c) 2019, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package prjn

import (
	"github.com/emer/dtable/etensor"
	"github.com/goki/ki/ints"
)

// PoolOneToOne implements a one-to-one pattern of connectivity between the sub-pools of units within two layers
// sub-pools are present for layers with a 4D shape, as the outer-most
// two dimensions of that 4D shape
// if either layer does not have pools, then if the number of individual
// units matches the number of pools in the other layer, those are connected one-to-one
// otherwise each pool connects to the entire set of other units.
// if neither is 4D, then it is equivalent to OneToOne
type PoolOneToOne struct {
	NCons     int `desc:"number of recv pools to connect (0 for entire number of pools in recv layer)"`
	SendStart int `desc:"starting pool index for sending connections"`
	RecvStart int `desc:"starting pool index for recv connections"`
}

func NewPoolOneToOne() *PoolOneToOne {
	return &PoolOneToOne{}
}

func (ot *PoolOneToOne) Name() string {
	return "PoolOneToOne"
}

func (ot *PoolOneToOne) Connect(send, recv *etensor.Shape, same bool) (sendn, recvn *etensor.Int32, cons *etensor.Bits) {
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
func (ot *PoolOneToOne) ConnectPools(send, recv *etensor.Shape, same bool) (sendn, recvn *etensor.Int32, cons *etensor.Bits) {
	sendn, recvn, cons = NewTensors(send, recv)
	nsend := send.Len()
	// nrecv := recv.Len()
	nsendPl := send.Dim(0) * send.Dim(1)
	nrecvPl := recv.Dim(0) * recv.Dim(1)
	nsendUn := send.Dim(2) * send.Dim(3)
	nrecvUn := recv.Dim(2) * recv.Dim(3)
	rnv := recvn.Values
	snv := sendn.Values
	ncon := nrecvPl
	if ot.NCons > 0 {
		ncon = ints.MinInt(ot.NCons, nrecvPl)
	}
	for i := 0; i < ncon; i++ {
		rpi := ot.RecvStart + i
		spi := ot.SendStart + i
		if rpi >= nrecvPl || spi >= nsendPl {
			break
		}
		for rui := 0; rui < nrecvUn; rui++ {
			ri := rpi*nrecvUn + rui
			for sui := 0; sui < nsendUn; sui++ {
				si := spi*nsendUn + sui
				off := ri*nsend + si
				cons.Values.Set(off, true)
				rnv[ri] = int32(nsendUn)
				snv[si] = int32(nrecvUn)
			}
		}
	}
	return
}

// ConnectRecvPool is when recv has pools but send doesn't
func (ot *PoolOneToOne) ConnectRecvPool(send, recv *etensor.Shape, same bool) (sendn, recvn *etensor.Int32, cons *etensor.Bits) {
	sendn, recvn, cons = NewTensors(send, recv)
	nsend := send.Len()
	nrecvPl := recv.Dim(0) * recv.Dim(1)
	nrecvUn := recv.Dim(2) * recv.Dim(3)
	rnv := recvn.Values
	snv := sendn.Values
	ncon := nrecvPl
	if ot.NCons > 0 {
		ncon = ints.MinInt(ot.NCons, nrecvPl)
	}

	if nsend == nrecvPl { // one-to-one
		for i := 0; i < ncon; i++ {
			rpi := ot.RecvStart + i
			si := ot.SendStart + i
			if rpi >= nrecvPl || si >= nsend {
				break
			}
			for rui := 0; rui < nrecvUn; rui++ {
				ri := rpi*nrecvUn + rui
				off := ri*nsend + si
				cons.Values.Set(off, true)
				rnv[ri] = int32(1)
				snv[si] = int32(nrecvUn)
			}
		}
	} else { // full
		for i := 0; i < ncon; i++ {
			rpi := ot.RecvStart + i
			if rpi >= nrecvPl {
				break
			}
			for rui := 0; rui < nrecvUn; rui++ {
				ri := rpi*nrecvUn + rui
				for si := 0; si < nsend; si++ {
					off := ri*nsend + si
					cons.Values.Set(off, true)
					rnv[ri] = int32(nsend)
					snv[si] = int32(ncon * nrecvUn)
				}
			}
		}
	}
	return
}

// ConnectSendPool is when send has pools but recv doesn't
func (ot *PoolOneToOne) ConnectSendPool(send, recv *etensor.Shape, same bool) (sendn, recvn *etensor.Int32, cons *etensor.Bits) {
	sendn, recvn, cons = NewTensors(send, recv)
	nsend := send.Len()
	nrecv := recv.Len()
	nsendPl := send.Dim(0) * send.Dim(1)
	nsendUn := send.Dim(2) * send.Dim(3)
	rnv := recvn.Values
	snv := sendn.Values
	ncon := nsendPl
	if ot.NCons > 0 {
		ncon = ints.MinInt(ot.NCons, nsendPl)
	}

	if nrecv == nsendPl { // one-to-one
		for i := 0; i < ncon; i++ {
			spi := ot.SendStart + i
			ri := ot.RecvStart + i
			if ri >= nrecv || spi >= nsendPl {
				break
			}
			for sui := 0; sui < nsendUn; sui++ {
				si := spi*nsendUn + sui
				off := ri*nsend + si
				cons.Values.Set(off, true)
				rnv[ri] = int32(nsendUn)
				snv[si] = int32(1)
			}
		}
	} else { // full
		for i := 0; i < ncon; i++ {
			spi := ot.SendStart + i
			if spi >= nsendPl {
				break
			}
			for ri := 0; ri < nrecv; ri++ {
				for sui := 0; sui < nsendUn; sui++ {
					si := spi*nsendUn + sui
					off := ri*nsend + si
					cons.Values.Set(off, true)
					rnv[ri] = int32(ncon * nsendUn)
					snv[si] = int32(nrecv)
				}
			}
		}
	}
	return
}

// copy of OneToOne.Connect
func (ot *PoolOneToOne) ConnectOneToOne(send, recv *etensor.Shape, same bool) (sendn, recvn *etensor.Int32, cons *etensor.Bits) {
	sendn, recvn, cons = NewTensors(send, recv)
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

func (ot *PoolOneToOne) HasWeights() bool {
	return false
}

func (ot *PoolOneToOne) Weights(sendn, recvn *etensor.Int32, cons *etensor.Bits) []float32 {
	return nil
}
