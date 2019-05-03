// Copyright (c) 2019, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package prjn

import (
	"fmt"
	"testing"

	"github.com/emer/etable/etensor"
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

	nsend := send.Len()
	nrecv := recv.Len()

	pj := NewFull()
	sendn, recvn, cons := pj.Connect(send, recv, false)
	fmt.Printf("full recv: 3x4 send: 2x3\n%s\n", string(ConsStringFull(send, recv, cons)))

	CheckAllN(sendn, nrecv, t)
	CheckAllN(recvn, nsend, t)

	// todo: test self
}

func TestOneToOne(t *testing.T) {
	send := etensor.NewShape([]int{3, 4}, nil, nil)
	recv := etensor.NewShape([]int{3, 4}, nil, nil)

	// nsend := send.Len()
	// nrecv := recv.Len()

	pj := NewOneToOne()
	sendn, recvn, cons := pj.Connect(send, recv, false)
	fmt.Printf("1-to-1 recv: 3x4 send: 3x4\n%s\n", string(ConsStringFull(send, recv, cons)))

	CheckAllN(sendn, 1, t)
	CheckAllN(recvn, 1, t)
}

func TestPoolOneToOne(t *testing.T) {
	send := etensor.NewShape([]int{2, 3, 1, 2}, nil, nil)
	recv := etensor.NewShape([]int{2, 3, 1, 2}, nil, nil)

	// nsend := send.Len()
	// nrecv := recv.Len()

	nsendUn := send.Dim(2) * send.Dim(3)
	nrecvUn := recv.Dim(2) * recv.Dim(3)

	pj := NewPoolOneToOne()
	// pj.NCons = 4
	// pj.RecvStart = 1
	// pj.SendStart = 1
	sendn, recvn, cons := pj.Connect(send, recv, false)
	fmt.Printf("pool 1-to-1 both 2x3 1x2\n%s\n", string(ConsStringFull(send, recv, cons)))

	CheckAllN(sendn, nrecvUn, t)
	CheckAllN(recvn, nsendUn, t)
}

func TestPoolOneToOneRecv(t *testing.T) {
	send := etensor.NewShape([]int{2, 3}, nil, nil)
	recv := etensor.NewShape([]int{2, 3, 1, 2}, nil, nil)

	// nsend := send.Len()
	// nrecv := recv.Len()

	nrecvUn := recv.Dim(2) * recv.Dim(3)

	pj := NewPoolOneToOne()
	sendn, recvn, cons := pj.Connect(send, recv, false)
	fmt.Printf("pool 1-to-1 recv 2x3 1x2, send 2x3\n%s\n", string(ConsStringFull(send, recv, cons)))

	CheckAllN(sendn, nrecvUn, t)
	CheckAllN(recvn, 1, t)
}

func TestPoolOneToOneSend(t *testing.T) {
	send := etensor.NewShape([]int{2, 3, 1, 2}, nil, nil)
	recv := etensor.NewShape([]int{2, 3}, nil, nil)

	// nsend := send.Len()
	// nrecv := recv.Len()

	nsendUn := send.Dim(2) * send.Dim(3)

	pj := NewPoolOneToOne()
	sendn, recvn, cons := pj.Connect(send, recv, false)
	fmt.Printf("pool 1-to-1 send 2x3 1x2, recv 2x3\n%s\n", string(ConsStringFull(send, recv, cons)))

	CheckAllN(sendn, 1, t)
	CheckAllN(recvn, nsendUn, t)
}
