// Copyright (c) 2019, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package prjn

import (
	"fmt"
	"testing"

	"github.com/emer/etable/etensor"
	"github.com/goki/ki/ints"
)

func CheckAllN(ns *etensor.Int32, trg int, t *testing.T) {
	sz := ns.Len()
	for i := 0; i < sz; i++ {
		n := int(ns.Value1D(i))
		if n != trg {
			t.Errorf("con n at idx: %d is not correct: %d trg: %d\n", i, n, trg)
		}
	}
}

func TestFull(t *testing.T) {
	send := etensor.NewShape([]int{2, 3}, nil, nil)
	recv := etensor.NewShape([]int{3, 4}, nil, nil)

	sNtot := send.Len()
	rNtot := recv.Len()

	pj := NewFull()
	sendn, recvn, cons := pj.Connect(send, recv, false)
	fmt.Printf("full recv: 3x4 send: 2x3\n%s\n", string(ConsStringFull(send, recv, cons)))

	CheckAllN(sendn, rNtot, t)
	CheckAllN(recvn, sNtot, t)
}

func TestFullSelf(t *testing.T) {
	send := etensor.NewShape([]int{2, 3}, nil, nil)

	sNtot := send.Len()

	pj := NewFull()
	pj.SelfCon = false
	sendn, recvn, cons := pj.Connect(send, send, true)
	fmt.Printf("full self no-con 2x3\n%s\n", string(ConsStringFull(send, send, cons)))

	CheckAllN(sendn, sNtot-1, t)
	CheckAllN(recvn, sNtot-1, t)
}

func TestOneToOne(t *testing.T) {
	send := etensor.NewShape([]int{3, 4}, nil, nil)
	recv := etensor.NewShape([]int{3, 4}, nil, nil)

	pj := NewOneToOne()
	sendn, recvn, cons := pj.Connect(send, recv, false)
	fmt.Printf("1-to-1 recv: 3x4 send: 3x4\n%s\n", string(ConsStringFull(send, recv, cons)))

	CheckAllN(sendn, 1, t)
	CheckAllN(recvn, 1, t)
}

func TestPoolOneToOne(t *testing.T) {
	send := etensor.NewShape([]int{2, 3, 1, 2}, nil, nil)
	recv := etensor.NewShape([]int{2, 3, 1, 2}, nil, nil)

	sNu := send.Dim(2) * send.Dim(3)
	rNu := recv.Dim(2) * recv.Dim(3)

	pj := NewPoolOneToOne()
	// pj.NCons = 4
	// pj.RecvStart = 1
	// pj.SendStart = 1
	sendn, recvn, cons := pj.Connect(send, recv, false)
	fmt.Printf("pool 1-to-1 both 2x3 1x2\n%s\n", string(ConsStringFull(send, recv, cons)))

	CheckAllN(sendn, rNu, t)
	CheckAllN(recvn, sNu, t)
}

func TestPoolOneToOneRecv(t *testing.T) {
	send := etensor.NewShape([]int{2, 3}, nil, nil)
	recv := etensor.NewShape([]int{2, 3, 1, 2}, nil, nil)

	rNu := recv.Dim(2) * recv.Dim(3)

	pj := NewPoolOneToOne()
	sendn, recvn, cons := pj.Connect(send, recv, false)
	fmt.Printf("pool 1-to-1 recv 2x3 1x2, send 2x3\n%s\n", string(ConsStringFull(send, recv, cons)))

	CheckAllN(sendn, rNu, t)
	CheckAllN(recvn, 1, t)
}

func TestPoolOneToOneSend(t *testing.T) {
	send := etensor.NewShape([]int{2, 3, 1, 2}, nil, nil)
	recv := etensor.NewShape([]int{2, 3}, nil, nil)

	sNu := send.Dim(2) * send.Dim(3)

	pj := NewPoolOneToOne()
	sendn, recvn, cons := pj.Connect(send, recv, false)
	fmt.Printf("pool 1-to-1 send 2x3 1x2, recv 2x3\n%s\n", string(ConsStringFull(send, recv, cons)))

	CheckAllN(sendn, 1, t)
	CheckAllN(recvn, sNu, t)
}

func TestPoolTile(t *testing.T) {
	send := etensor.NewShape([]int{4, 4, 1, 2}, nil, nil)
	recv := etensor.NewShape([]int{2, 2, 1, 3}, nil, nil)

	sNu := send.Dim(2) * send.Dim(3)
	rNu := recv.Dim(2) * recv.Dim(3)

	pj := NewPoolTile()
	pj.Size.Set(2, 2)
	pj.Skip.Set(2, 2)
	pj.Start.Set(0, 0)
	sendn, recvn, cons := pj.Connect(send, recv, false)
	fmt.Printf("pool tile send 4x4 1x2, recv 2x2 1x3\n%s\n", string(ConsStringFull(send, recv, cons)))

	CheckAllN(sendn, rNu, t)
	CheckAllN(recvn, 4*sNu, t)

	// send = etensor.NewShape([]int{4, 4, 3, 3}, nil, nil)
	// recv = etensor.NewShape([]int{2, 2, 2, 2}, nil, nil)
	// wts := &etensor.Float32{}
	// pj.TopoWts(send, recv, wts)
	// fmt.Printf("topo wts\n%v\n", wts)
}

func TestUnifRnd(t *testing.T) {
	send := etensor.NewShape([]int{2, 3}, nil, nil)
	recv := etensor.NewShape([]int{3, 4}, nil, nil)

	sNtot := send.Len()
	rNtot := recv.Len()

	pj := NewUnifRnd()
	pj.PCon = 0.5
	sendn, recvn, cons := pj.Connect(send, recv, false)
	fmt.Printf("unif rnd recv: 3x4 send: 2x3\n%s\n", string(ConsStringFull(send, recv, cons)))

	_ = recvn

	nrMax := 0
	nrMin := rNtot
	nrMean := 0
	for si := 0; si < sNtot; si++ {
		nr := int(sendn.Values[si])
		nrMax = ints.MaxInt(nrMax, nr)
		nrMin = ints.MinInt(nrMin, nr)
		nrMean += nr
	}
	fmt.Printf("sendn: %v\n", sendn.Values)
	fmt.Printf("unif rnd rNtot: %d  pcon: %g  max: %d  min: %d  mean: %g\n", rNtot, pj.PCon, nrMax, nrMin, float32(nrMean)/float32(sNtot))
}

func TestUnifRndLg(t *testing.T) {
	send := etensor.NewShape([]int{20, 30}, nil, nil)
	recv := etensor.NewShape([]int{30, 40}, nil, nil)

	sNtot := send.Len()
	rNtot := recv.Len()

	pj := NewUnifRnd()
	pj.PCon = 0.05
	sendn, recvn, cons := pj.Connect(send, recv, false)

	_ = recvn
	_ = cons

	nrMax := 0
	nrMin := rNtot
	nrMean := 0
	for si := 0; si < sNtot; si++ {
		nr := int(sendn.Values[si])
		nrMax = ints.MaxInt(nrMax, nr)
		nrMin = ints.MinInt(nrMin, nr)
		nrMean += nr
	}
	fmt.Printf("unif rnd large rNtot: %d  pcon: %g  max: %d  min: %d  mean: %g\n", rNtot, pj.PCon, nrMax, nrMin, float32(nrMean)/float32(sNtot))
}
