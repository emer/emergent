// Copyright (c) 2019, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package prjn

import (
	"github.com/emer/etable/etensor"
	"github.com/goki/ki/ints"
)

// PoolOneToOne implements one-to-one connectivity between pools within layers.
// Pools are the outer-most two dimensions of a 4D layer shape.
// If either layer does not have pools, then if the number of individual
// units matches the number of pools in the other layer, those are connected one-to-one
// otherwise each pool connects to the entire set of other units.
// If neither is 4D, then it is equivalent to OneToOne.
type PoolOneToOne struct {
	NPools    int `desc:"number of recv pools to connect (0 for entire number of pools in recv layer)"`
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
	sNtot := send.Len()
	// rNtot := recv.Len()
	sNp := send.Dim(0) * send.Dim(1)
	rNp := recv.Dim(0) * recv.Dim(1)
	sNu := send.Dim(2) * send.Dim(3)
	rNu := recv.Dim(2) * recv.Dim(3)
	rnv := recvn.Values
	snv := sendn.Values
	npl := rNp
	if ot.NPools > 0 {
		npl = ints.MinInt(ot.NPools, rNp)
	}
	for i := 0; i < npl; i++ {
		rpi := ot.RecvStart + i
		spi := ot.SendStart + i
		if rpi >= rNp || spi >= sNp {
			break
		}
		for rui := 0; rui < rNu; rui++ {
			ri := rpi*rNu + rui
			for sui := 0; sui < sNu; sui++ {
				si := spi*sNu + sui
				off := ri*sNtot + si
				cons.Values.Set(off, true)
				rnv[ri] = int32(sNu)
				snv[si] = int32(rNu)
			}
		}
	}
	return
}

// ConnectRecvPool is when recv has pools but send doesn't
func (ot *PoolOneToOne) ConnectRecvPool(send, recv *etensor.Shape, same bool) (sendn, recvn *etensor.Int32, cons *etensor.Bits) {
	sendn, recvn, cons = NewTensors(send, recv)
	sNtot := send.Len()
	rNp := recv.Dim(0) * recv.Dim(1)
	rNu := recv.Dim(2) * recv.Dim(3)
	rnv := recvn.Values
	snv := sendn.Values
	npl := rNp
	if ot.NPools > 0 {
		npl = ints.MinInt(ot.NPools, rNp)
	}

	if sNtot == rNp { // one-to-one
		for i := 0; i < npl; i++ {
			rpi := ot.RecvStart + i
			si := ot.SendStart + i
			if rpi >= rNp || si >= sNtot {
				break
			}
			for rui := 0; rui < rNu; rui++ {
				ri := rpi*rNu + rui
				off := ri*sNtot + si
				cons.Values.Set(off, true)
				rnv[ri] = int32(1)
				snv[si] = int32(rNu)
			}
		}
	} else { // full
		for i := 0; i < npl; i++ {
			rpi := ot.RecvStart + i
			if rpi >= rNp {
				break
			}
			for rui := 0; rui < rNu; rui++ {
				ri := rpi*rNu + rui
				for si := 0; si < sNtot; si++ {
					off := ri*sNtot + si
					cons.Values.Set(off, true)
					rnv[ri] = int32(sNtot)
					snv[si] = int32(npl * rNu)
				}
			}
		}
	}
	return
}

// ConnectSendPool is when send has pools but recv doesn't
func (ot *PoolOneToOne) ConnectSendPool(send, recv *etensor.Shape, same bool) (sendn, recvn *etensor.Int32, cons *etensor.Bits) {
	sendn, recvn, cons = NewTensors(send, recv)
	sNtot := send.Len()
	rNtot := recv.Len()
	sNp := send.Dim(0) * send.Dim(1)
	sNu := send.Dim(2) * send.Dim(3)
	rnv := recvn.Values
	snv := sendn.Values
	npl := sNp
	if ot.NPools > 0 {
		npl = ints.MinInt(ot.NPools, sNp)
	}

	if rNtot == sNp { // one-to-one
		for i := 0; i < npl; i++ {
			spi := ot.SendStart + i
			ri := ot.RecvStart + i
			if ri >= rNtot || spi >= sNp {
				break
			}
			for sui := 0; sui < sNu; sui++ {
				si := spi*sNu + sui
				off := ri*sNtot + si
				cons.Values.Set(off, true)
				rnv[ri] = int32(sNu)
				snv[si] = int32(1)
			}
		}
	} else { // full
		for i := 0; i < npl; i++ {
			spi := ot.SendStart + i
			if spi >= sNp {
				break
			}
			for ri := 0; ri < rNtot; ri++ {
				for sui := 0; sui < sNu; sui++ {
					si := spi*sNu + sui
					off := ri*sNtot + si
					cons.Values.Set(off, true)
					rnv[ri] = int32(npl * sNu)
					snv[si] = int32(rNtot)
				}
			}
		}
	}
	return
}

// copy of OneToOne.Connect
func (ot *PoolOneToOne) ConnectOneToOne(send, recv *etensor.Shape, same bool) (sendn, recvn *etensor.Int32, cons *etensor.Bits) {
	sendn, recvn, cons = NewTensors(send, recv)
	sNtot := send.Len()
	rNtot := recv.Len()
	rnv := recvn.Values
	snv := sendn.Values
	npl := rNtot
	if ot.NPools > 0 {
		npl = ints.MinInt(ot.NPools, rNtot)
	}
	for i := 0; i < npl; i++ {
		ri := ot.RecvStart + i
		si := ot.SendStart + i
		if ri >= rNtot || si >= sNtot {
			break
		}
		off := ri*sNtot + si
		cons.Values.Set(off, true)
		rnv[ri] = 1
		snv[si] = 1
	}
	return
}
