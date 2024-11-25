// Copyright (c) 2019, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package paths

import (
	"testing"

	"cogentcore.org/core/tensor"
	"github.com/stretchr/testify/assert"
)

func CheckAllN(ns *tensor.Int32, trg int, t *testing.T) {
	sz := ns.Len()
	for i := 0; i < sz; i++ {
		n := int(ns.Value1D(i))
		if n != trg {
			t.Errorf("con n at idx: %d is not correct: %d trg: %d\n", i, n, trg)
		}
	}
}

func TestFull(t *testing.T) {
	send := tensor.NewShape(2, 3)
	recv := tensor.NewShape(3, 4)

	sNtot := send.Len()
	rNtot := recv.Len()

	pj := NewFull()
	sendn, recvn, cons := pj.Connect(send, recv, false)
	// fmt.Printf("full recv: 3x4 send: 2x3\n%s\n", string(ConsStringFull(send, recv, cons)))

	ex := `1 1 1 1 1 1 
1 1 1 1 1 1 
1 1 1 1 1 1 
1 1 1 1 1 1 
1 1 1 1 1 1 
1 1 1 1 1 1 
1 1 1 1 1 1 
1 1 1 1 1 1 
1 1 1 1 1 1 
1 1 1 1 1 1 
1 1 1 1 1 1 
1 1 1 1 1 1 
`
	assert.Equal(t, ex, string(ConsStringFull(send, recv, cons)))

	CheckAllN(sendn, rNtot, t)
	CheckAllN(recvn, sNtot, t)
}

func TestFullSelf(t *testing.T) {
	send := tensor.NewShape(2, 3)

	sNtot := send.Len()

	pj := NewFull()
	pj.SelfCon = false
	sendn, recvn, cons := pj.Connect(send, send, true)
	// fmt.Printf("full self no-con 2x3\n%s\n", string(ConsStringFull(send, send, cons)))

	ex := `0 1 1 1 1 1 
1 0 1 1 1 1 
1 1 0 1 1 1 
1 1 1 0 1 1 
1 1 1 1 0 1 
1 1 1 1 1 0 
`

	assert.Equal(t, ex, string(ConsStringFull(send, send, cons)))

	CheckAllN(sendn, sNtot-1, t)
	CheckAllN(recvn, sNtot-1, t)
}

func TestOneToOne(t *testing.T) {
	send := tensor.NewShape(3, 2)
	recv := tensor.NewShape(3, 2)

	pj := NewOneToOne()
	sendn, recvn, cons := pj.Connect(send, recv, false)
	// fmt.Printf("1-to-1 recv: 3x4 send: 3x4\n%s\n", string(ConsStringFull(send, recv, cons)))

	ex := `1 0 0 0 0 0 
0 1 0 0 0 0 
0 0 1 0 0 0 
0 0 0 1 0 0 
0 0 0 0 1 0 
0 0 0 0 0 1 
`
	assert.Equal(t, ex, string(ConsStringFull(send, recv, cons)))

	CheckAllN(sendn, 1, t)
	CheckAllN(recvn, 1, t)
}

func TestPoolOneToOne(t *testing.T) {
	send := tensor.NewShape(2, 3, 1, 2)
	recv := tensor.NewShape(2, 3, 1, 2)

	sNu := send.DimSize(2) * send.DimSize(3)
	rNu := recv.DimSize(2) * recv.DimSize(3)

	pj := NewPoolOneToOne()
	// pj.NCons = 4
	// pj.RecvStart = 1
	// pj.SendStart = 1
	sendn, recvn, cons := pj.Connect(send, recv, false)
	// fmt.Printf("pool 1-to-1 both 2x3 1x2\n%s\n", string(ConsStringFull(send, recv, cons)))

	ex := `1 1 0 0 0 0 0 0 0 0 0 0 
1 1 0 0 0 0 0 0 0 0 0 0 
0 0 1 1 0 0 0 0 0 0 0 0 
0 0 1 1 0 0 0 0 0 0 0 0 
0 0 0 0 1 1 0 0 0 0 0 0 
0 0 0 0 1 1 0 0 0 0 0 0 
0 0 0 0 0 0 1 1 0 0 0 0 
0 0 0 0 0 0 1 1 0 0 0 0 
0 0 0 0 0 0 0 0 1 1 0 0 
0 0 0 0 0 0 0 0 1 1 0 0 
0 0 0 0 0 0 0 0 0 0 1 1 
0 0 0 0 0 0 0 0 0 0 1 1 
`

	assert.Equal(t, ex, string(ConsStringFull(send, recv, cons)))

	CheckAllN(sendn, rNu, t)
	CheckAllN(recvn, sNu, t)
}

func TestPoolOneToOneRecv(t *testing.T) {
	send := tensor.NewShape(2, 3)
	recv := tensor.NewShape(2, 3, 1, 2)

	rNu := recv.DimSize(2) * recv.DimSize(3)

	pj := NewPoolOneToOne()
	sendn, recvn, cons := pj.Connect(send, recv, false)
	// fmt.Printf("pool 1-to-1 recv 2x3 1x2, send 2x3\n%s\n", string(ConsStringFull(send, recv, cons)))

	ex := `1 0 0 0 0 0 
1 0 0 0 0 0 
0 1 0 0 0 0 
0 1 0 0 0 0 
0 0 1 0 0 0 
0 0 1 0 0 0 
0 0 0 1 0 0 
0 0 0 1 0 0 
0 0 0 0 1 0 
0 0 0 0 1 0 
0 0 0 0 0 1 
0 0 0 0 0 1 
`
	assert.Equal(t, ex, string(ConsStringFull(send, recv, cons)))

	CheckAllN(sendn, rNu, t)
	CheckAllN(recvn, 1, t)
}

func TestPoolOneToOneSend(t *testing.T) {
	send := tensor.NewShape(2, 3, 1, 2)
	recv := tensor.NewShape(2, 3)

	sNu := send.DimSize(2) * send.DimSize(3)

	pj := NewPoolOneToOne()
	sendn, recvn, cons := pj.Connect(send, recv, false)
	// fmt.Printf("pool 1-to-1 send 2x3 1x2, recv 2x3\n%s\n", string(ConsStringFull(send, recv, cons)))

	ex := `1 1 0 0 0 0 0 0 0 0 0 0 
0 0 1 1 0 0 0 0 0 0 0 0 
0 0 0 0 1 1 0 0 0 0 0 0 
0 0 0 0 0 0 1 1 0 0 0 0 
0 0 0 0 0 0 0 0 1 1 0 0 
0 0 0 0 0 0 0 0 0 0 1 1 
`

	assert.Equal(t, ex, string(ConsStringFull(send, recv, cons)))

	CheckAllN(sendn, 1, t)
	CheckAllN(recvn, sNu, t)
}

func TestPoolTile(t *testing.T) {
	send := tensor.NewShape(4, 4, 1, 2)
	recv := tensor.NewShape(2, 2, 1, 3)

	sNu := send.DimSize(2) * send.DimSize(3)
	rNu := recv.DimSize(2) * recv.DimSize(3)

	pj := NewPoolTile()
	pj.Size.Set(2, 2)
	pj.Skip.Set(2, 2)
	pj.Start.Set(0, 0)
	sendn, recvn, cons := pj.Connect(send, recv, false)
	// fmt.Printf("pool tile send 4x4 1x2, recv 2x2 1x3\n%s\n", string(ConsStringFull(send, recv, cons)))

	ex := `1 1 1 1 0 0 0 0 1 1 1 1 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 
1 1 1 1 0 0 0 0 1 1 1 1 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 
1 1 1 1 0 0 0 0 1 1 1 1 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 
0 0 0 0 1 1 1 1 0 0 0 0 1 1 1 1 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 
0 0 0 0 1 1 1 1 0 0 0 0 1 1 1 1 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 
0 0 0 0 1 1 1 1 0 0 0 0 1 1 1 1 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 
0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 1 1 1 1 0 0 0 0 1 1 1 1 0 0 0 0 
0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 1 1 1 1 0 0 0 0 1 1 1 1 0 0 0 0 
0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 1 1 1 1 0 0 0 0 1 1 1 1 0 0 0 0 
0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 1 1 1 1 0 0 0 0 1 1 1 1 
0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 1 1 1 1 0 0 0 0 1 1 1 1 
0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 1 1 1 1 0 0 0 0 1 1 1 1 
`

	assert.Equal(t, ex, string(ConsStringFull(send, recv, cons)))

	CheckAllN(sendn, rNu, t)
	CheckAllN(recvn, pj.Size.X*pj.Size.Y*sNu, t)

	// send = tensor.NewShape(4, 4, 3, 3)
	// recv = tensor.NewShape(2, 2, 2, 2)
	// wts := &tensor.Float32{}
	// pj.TopoWeights(send, recv, wts)
	// fmt.Printf("topo wts\n%v\n", wts)
}

func TestPoolTileRecip(t *testing.T) {
	send := tensor.NewShape(4, 4, 1, 2)
	recv := tensor.NewShape(2, 2, 1, 3)

	sNu := send.DimSize(2) * send.DimSize(3)
	rNu := recv.DimSize(2) * recv.DimSize(3)

	pj := NewPoolTile()
	pj.Size.Set(2, 2)
	pj.Skip.Set(2, 2)
	pj.Start.Set(0, 0)
	pj.Recip = true
	sendn, recvn, cons := pj.Connect(recv, send, false)
	// fmt.Printf("pool tile recip send 4x4 1x2, recv 2x2 1x3\n%s\n", string(ConsStringFull(recv, send, cons)))

	ex := `1 1 1 0 0 0 0 0 0 0 0 0 
1 1 1 0 0 0 0 0 0 0 0 0 
1 1 1 0 0 0 0 0 0 0 0 0 
1 1 1 0 0 0 0 0 0 0 0 0 
0 0 0 1 1 1 0 0 0 0 0 0 
0 0 0 1 1 1 0 0 0 0 0 0 
0 0 0 1 1 1 0 0 0 0 0 0 
0 0 0 1 1 1 0 0 0 0 0 0 
1 1 1 0 0 0 0 0 0 0 0 0 
1 1 1 0 0 0 0 0 0 0 0 0 
1 1 1 0 0 0 0 0 0 0 0 0 
1 1 1 0 0 0 0 0 0 0 0 0 
0 0 0 1 1 1 0 0 0 0 0 0 
0 0 0 1 1 1 0 0 0 0 0 0 
0 0 0 1 1 1 0 0 0 0 0 0 
0 0 0 1 1 1 0 0 0 0 0 0 
0 0 0 0 0 0 1 1 1 0 0 0 
0 0 0 0 0 0 1 1 1 0 0 0 
0 0 0 0 0 0 1 1 1 0 0 0 
0 0 0 0 0 0 1 1 1 0 0 0 
0 0 0 0 0 0 0 0 0 1 1 1 
0 0 0 0 0 0 0 0 0 1 1 1 
0 0 0 0 0 0 0 0 0 1 1 1 
0 0 0 0 0 0 0 0 0 1 1 1 
0 0 0 0 0 0 1 1 1 0 0 0 
0 0 0 0 0 0 1 1 1 0 0 0 
0 0 0 0 0 0 1 1 1 0 0 0 
0 0 0 0 0 0 1 1 1 0 0 0 
0 0 0 0 0 0 0 0 0 1 1 1 
0 0 0 0 0 0 0 0 0 1 1 1 
0 0 0 0 0 0 0 0 0 1 1 1 
0 0 0 0 0 0 0 0 0 1 1 1 
`

	assert.Equal(t, ex, string(ConsStringFull(recv, send, cons)))

	CheckAllN(sendn, pj.Size.X*pj.Size.Y*sNu, t)
	CheckAllN(recvn, rNu, t)

	// send = tensor.NewShape(4, 4, 3, 3)
	// recv = tensor.NewShape(2, 2, 2, 2)
	// wts := &tensor.Float32{}
	// pj.TopoWeights(send, recv, wts)
	// fmt.Printf("topo wts\n%v\n", wts)
}

func TestPoolTile2(t *testing.T) {
	send := tensor.NewShape(5, 4, 1, 2)
	recv := tensor.NewShape(5, 4, 2, 1)

	sNu := send.DimSize(2) * send.DimSize(3)
	rNu := recv.DimSize(2) * recv.DimSize(3)

	pj := NewPoolTile()
	pj.Size.Set(3, 3)
	pj.Skip.Set(1, 1)
	pj.Start.Set(-1, -1)
	sendn, recvn, cons := pj.Connect(recv, send, false)
	// fmt.Printf("pool tile 3x3skip1 send 5x4 1x2, recv 5x5 2x1\n%s\n", string(ConsStringFull(recv, send, cons)))

	ex := `1 1 1 1 0 0 1 1 1 1 1 1 0 0 1 1 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 1 1 1 1 0 0 1 1 
1 1 1 1 0 0 1 1 1 1 1 1 0 0 1 1 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 1 1 1 1 0 0 1 1 
1 1 1 1 1 1 0 0 1 1 1 1 1 1 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 1 1 1 1 1 1 0 0 
1 1 1 1 1 1 0 0 1 1 1 1 1 1 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 1 1 1 1 1 1 0 0 
0 0 1 1 1 1 1 1 0 0 1 1 1 1 1 1 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 1 1 1 1 1 1 
0 0 1 1 1 1 1 1 0 0 1 1 1 1 1 1 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 1 1 1 1 1 1 
1 1 0 0 1 1 1 1 1 1 0 0 1 1 1 1 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 1 1 0 0 1 1 1 1 
1 1 0 0 1 1 1 1 1 1 0 0 1 1 1 1 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 1 1 0 0 1 1 1 1 
1 1 1 1 0 0 1 1 1 1 1 1 0 0 1 1 1 1 1 1 0 0 1 1 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 
1 1 1 1 0 0 1 1 1 1 1 1 0 0 1 1 1 1 1 1 0 0 1 1 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 
1 1 1 1 1 1 0 0 1 1 1 1 1 1 0 0 1 1 1 1 1 1 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 
1 1 1 1 1 1 0 0 1 1 1 1 1 1 0 0 1 1 1 1 1 1 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 
0 0 1 1 1 1 1 1 0 0 1 1 1 1 1 1 0 0 1 1 1 1 1 1 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 
0 0 1 1 1 1 1 1 0 0 1 1 1 1 1 1 0 0 1 1 1 1 1 1 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 
1 1 0 0 1 1 1 1 1 1 0 0 1 1 1 1 1 1 0 0 1 1 1 1 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 
1 1 0 0 1 1 1 1 1 1 0 0 1 1 1 1 1 1 0 0 1 1 1 1 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 
0 0 0 0 0 0 0 0 1 1 1 1 0 0 1 1 1 1 1 1 0 0 1 1 1 1 1 1 0 0 1 1 0 0 0 0 0 0 0 0 
0 0 0 0 0 0 0 0 1 1 1 1 0 0 1 1 1 1 1 1 0 0 1 1 1 1 1 1 0 0 1 1 0 0 0 0 0 0 0 0 
0 0 0 0 0 0 0 0 1 1 1 1 1 1 0 0 1 1 1 1 1 1 0 0 1 1 1 1 1 1 0 0 0 0 0 0 0 0 0 0 
0 0 0 0 0 0 0 0 1 1 1 1 1 1 0 0 1 1 1 1 1 1 0 0 1 1 1 1 1 1 0 0 0 0 0 0 0 0 0 0 
0 0 0 0 0 0 0 0 0 0 1 1 1 1 1 1 0 0 1 1 1 1 1 1 0 0 1 1 1 1 1 1 0 0 0 0 0 0 0 0 
0 0 0 0 0 0 0 0 0 0 1 1 1 1 1 1 0 0 1 1 1 1 1 1 0 0 1 1 1 1 1 1 0 0 0 0 0 0 0 0 
0 0 0 0 0 0 0 0 1 1 0 0 1 1 1 1 1 1 0 0 1 1 1 1 1 1 0 0 1 1 1 1 0 0 0 0 0 0 0 0 
0 0 0 0 0 0 0 0 1 1 0 0 1 1 1 1 1 1 0 0 1 1 1 1 1 1 0 0 1 1 1 1 0 0 0 0 0 0 0 0 
0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 1 1 1 1 0 0 1 1 1 1 1 1 0 0 1 1 1 1 1 1 0 0 1 1 
0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 1 1 1 1 0 0 1 1 1 1 1 1 0 0 1 1 1 1 1 1 0 0 1 1 
0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 1 1 1 1 1 1 0 0 1 1 1 1 1 1 0 0 1 1 1 1 1 1 0 0 
0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 1 1 1 1 1 1 0 0 1 1 1 1 1 1 0 0 1 1 1 1 1 1 0 0 
0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 1 1 1 1 1 1 0 0 1 1 1 1 1 1 0 0 1 1 1 1 1 1 
0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 1 1 1 1 1 1 0 0 1 1 1 1 1 1 0 0 1 1 1 1 1 1 
0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 1 1 0 0 1 1 1 1 1 1 0 0 1 1 1 1 1 1 0 0 1 1 1 1 
0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 1 1 0 0 1 1 1 1 1 1 0 0 1 1 1 1 1 1 0 0 1 1 1 1 
1 1 1 1 0 0 1 1 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 1 1 1 1 0 0 1 1 1 1 1 1 0 0 1 1 
1 1 1 1 0 0 1 1 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 1 1 1 1 0 0 1 1 1 1 1 1 0 0 1 1 
1 1 1 1 1 1 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 1 1 1 1 1 1 0 0 1 1 1 1 1 1 0 0 
1 1 1 1 1 1 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 1 1 1 1 1 1 0 0 1 1 1 1 1 1 0 0 
0 0 1 1 1 1 1 1 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 1 1 1 1 1 1 0 0 1 1 1 1 1 1 
0 0 1 1 1 1 1 1 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 1 1 1 1 1 1 0 0 1 1 1 1 1 1 
1 1 0 0 1 1 1 1 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 1 1 0 0 1 1 1 1 1 1 0 0 1 1 1 1 
1 1 0 0 1 1 1 1 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 1 1 0 0 1 1 1 1 1 1 0 0 1 1 1 1 
`

	assert.Equal(t, ex, string(ConsStringFull(recv, send, cons)))

	CheckAllN(recvn, pj.Size.X*pj.Size.Y*sNu, t)
	CheckAllN(sendn, pj.Size.X*pj.Size.Y*rNu, t)
}

func TestPoolTileRecip2(t *testing.T) {
	send := tensor.NewShape(5, 4, 1, 2)
	recv := tensor.NewShape(5, 4, 2, 1)

	sNu := send.DimSize(2) * send.DimSize(3)
	rNu := recv.DimSize(2) * recv.DimSize(3)

	pj := NewPoolTile()
	pj.Size.Set(3, 3)
	pj.Skip.Set(1, 1)
	pj.Start.Set(-1, -1)
	pj.Recip = true
	sendn, recvn, cons := pj.Connect(recv, send, false)
	// fmt.Printf("pool tile recip send 5x4 1x2, recv 5x4 2x1\n%s\n", string(ConsStringFull(recv, send, cons)))

	ex := `1 1 1 1 0 0 1 1 1 1 1 1 0 0 1 1 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 1 1 1 1 0 0 1 1 
1 1 1 1 0 0 1 1 1 1 1 1 0 0 1 1 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 1 1 1 1 0 0 1 1 
1 1 1 1 1 1 0 0 1 1 1 1 1 1 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 1 1 1 1 1 1 0 0 
1 1 1 1 1 1 0 0 1 1 1 1 1 1 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 1 1 1 1 1 1 0 0 
0 0 1 1 1 1 1 1 0 0 1 1 1 1 1 1 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 1 1 1 1 1 1 
0 0 1 1 1 1 1 1 0 0 1 1 1 1 1 1 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 1 1 1 1 1 1 
1 1 0 0 1 1 1 1 1 1 0 0 1 1 1 1 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 1 1 0 0 1 1 1 1 
1 1 0 0 1 1 1 1 1 1 0 0 1 1 1 1 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 1 1 0 0 1 1 1 1 
1 1 1 1 0 0 1 1 1 1 1 1 0 0 1 1 1 1 1 1 0 0 1 1 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 
1 1 1 1 0 0 1 1 1 1 1 1 0 0 1 1 1 1 1 1 0 0 1 1 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 
1 1 1 1 1 1 0 0 1 1 1 1 1 1 0 0 1 1 1 1 1 1 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 
1 1 1 1 1 1 0 0 1 1 1 1 1 1 0 0 1 1 1 1 1 1 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 
0 0 1 1 1 1 1 1 0 0 1 1 1 1 1 1 0 0 1 1 1 1 1 1 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 
0 0 1 1 1 1 1 1 0 0 1 1 1 1 1 1 0 0 1 1 1 1 1 1 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 
1 1 0 0 1 1 1 1 1 1 0 0 1 1 1 1 1 1 0 0 1 1 1 1 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 
1 1 0 0 1 1 1 1 1 1 0 0 1 1 1 1 1 1 0 0 1 1 1 1 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 
0 0 0 0 0 0 0 0 1 1 1 1 0 0 1 1 1 1 1 1 0 0 1 1 1 1 1 1 0 0 1 1 0 0 0 0 0 0 0 0 
0 0 0 0 0 0 0 0 1 1 1 1 0 0 1 1 1 1 1 1 0 0 1 1 1 1 1 1 0 0 1 1 0 0 0 0 0 0 0 0 
0 0 0 0 0 0 0 0 1 1 1 1 1 1 0 0 1 1 1 1 1 1 0 0 1 1 1 1 1 1 0 0 0 0 0 0 0 0 0 0 
0 0 0 0 0 0 0 0 1 1 1 1 1 1 0 0 1 1 1 1 1 1 0 0 1 1 1 1 1 1 0 0 0 0 0 0 0 0 0 0 
0 0 0 0 0 0 0 0 0 0 1 1 1 1 1 1 0 0 1 1 1 1 1 1 0 0 1 1 1 1 1 1 0 0 0 0 0 0 0 0 
0 0 0 0 0 0 0 0 0 0 1 1 1 1 1 1 0 0 1 1 1 1 1 1 0 0 1 1 1 1 1 1 0 0 0 0 0 0 0 0 
0 0 0 0 0 0 0 0 1 1 0 0 1 1 1 1 1 1 0 0 1 1 1 1 1 1 0 0 1 1 1 1 0 0 0 0 0 0 0 0 
0 0 0 0 0 0 0 0 1 1 0 0 1 1 1 1 1 1 0 0 1 1 1 1 1 1 0 0 1 1 1 1 0 0 0 0 0 0 0 0 
0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 1 1 1 1 0 0 1 1 1 1 1 1 0 0 1 1 1 1 1 1 0 0 1 1 
0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 1 1 1 1 0 0 1 1 1 1 1 1 0 0 1 1 1 1 1 1 0 0 1 1 
0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 1 1 1 1 1 1 0 0 1 1 1 1 1 1 0 0 1 1 1 1 1 1 0 0 
0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 1 1 1 1 1 1 0 0 1 1 1 1 1 1 0 0 1 1 1 1 1 1 0 0 
0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 1 1 1 1 1 1 0 0 1 1 1 1 1 1 0 0 1 1 1 1 1 1 
0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 1 1 1 1 1 1 0 0 1 1 1 1 1 1 0 0 1 1 1 1 1 1 
0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 1 1 0 0 1 1 1 1 1 1 0 0 1 1 1 1 1 1 0 0 1 1 1 1 
0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 1 1 0 0 1 1 1 1 1 1 0 0 1 1 1 1 1 1 0 0 1 1 1 1 
1 1 1 1 0 0 1 1 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 1 1 1 1 0 0 1 1 1 1 1 1 0 0 1 1 
1 1 1 1 0 0 1 1 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 1 1 1 1 0 0 1 1 1 1 1 1 0 0 1 1 
1 1 1 1 1 1 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 1 1 1 1 1 1 0 0 1 1 1 1 1 1 0 0 
1 1 1 1 1 1 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 1 1 1 1 1 1 0 0 1 1 1 1 1 1 0 0 
0 0 1 1 1 1 1 1 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 1 1 1 1 1 1 0 0 1 1 1 1 1 1 
0 0 1 1 1 1 1 1 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 1 1 1 1 1 1 0 0 1 1 1 1 1 1 
1 1 0 0 1 1 1 1 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 1 1 0 0 1 1 1 1 1 1 0 0 1 1 1 1 
1 1 0 0 1 1 1 1 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 1 1 0 0 1 1 1 1 1 1 0 0 1 1 1 1 
`

	assert.Equal(t, ex, string(ConsStringFull(recv, send, cons)))

	CheckAllN(sendn, pj.Size.X*pj.Size.Y*sNu, t)
	CheckAllN(recvn, pj.Size.X*pj.Size.Y*rNu, t)
}

func TestUniformRand(t *testing.T) {
	send := tensor.NewShape(2, 3)
	recv := tensor.NewShape(3, 4)

	sNtot := send.Len()
	rNtot := recv.Len()

	pj := NewUniformRand()
	pj.RandSeed = 10
	pj.PCon = 0.5
	sendn, recvn, cons := pj.Connect(send, recv, false)
	// fmt.Printf("unif rnd recv: 3x4 send: 2x3\n%s\n", string(ConsStringFull(send, recv, cons)))
	_ = recvn

	ex := `1 1 1 0 0 0 
0 1 1 1 0 0 
1 0 0 1 1 0 
0 1 0 1 1 0 
0 1 0 1 0 1 
0 0 0 1 1 1 
0 0 1 1 1 0 
0 0 1 0 1 1 
1 1 0 0 1 0 
0 1 1 0 0 1 
0 0 0 1 1 1 
0 1 0 0 1 1 
`
	assert.Equal(t, ex, string(ConsStringFull(send, recv, cons)))

	nrMax := 0
	nrMin := rNtot
	nrMean := 0
	for si := 0; si < sNtot; si++ {
		nr := int(sendn.Values[si])
		nrMax = max(nr)
		nrMin = min(nrMin, nr)
		nrMean += nr
	}
	// fmt.Printf("sendn: %v\n", sendn.Values)
	// fmt.Printf("unif rnd rNtot: %d  pcon: %g  max: %d  min: %d  mean: %g\n", rNtot, pj.PCon, nrMax, nrMin, float32(nrMean)/float32(sNtot))

	assert.Equal(t, []int32{3, 7, 5, 7, 8, 6}, sendn.Values)
	assert.Equal(t, rNtot, 12)
	assert.Equal(t, nrMax, 6)
	assert.Equal(t, nrMin, 3)

	// now test recip
	rpj := NewUniformRand()
	rpj.RandSeed = 10
	rpj.PCon = 0.5
	rpj.Recip = true
	sendn, recvn, cons = rpj.Connect(send, recv, false)
	// fmt.Printf("unif rnd recip recv: 3x4 send: 2x3\n%s\n", string(ConsStringFull(send, recv, cons)))

	ex2 := `0 1 0 0 0 0 
1 1 1 0 0 0 
1 0 1 0 1 0 
1 0 0 0 0 1 
1 1 0 1 1 1 
0 0 0 1 0 0 
1 1 1 1 0 0 
0 0 0 0 0 0 
1 1 1 1 1 1 
0 0 0 1 1 1 
0 1 1 0 1 1 
0 0 1 1 1 1 
`
	assert.Equal(t, ex2, string(ConsStringFull(send, recv, cons)))
}

func TestUniformRandLg(t *testing.T) {
	send := tensor.NewShape(20, 30)
	recv := tensor.NewShape(30, 40)

	sNtot := send.Len()
	rNtot := recv.Len()

	pj := NewUniformRand()
	pj.PCon = 0.05
	pj.RandSeed = 10
	sendn, recvn, cons := pj.Connect(send, recv, false)

	_ = recvn
	_ = cons

	nrMax := 0
	nrMin := rNtot
	nrMean := 0
	for si := 0; si < sNtot; si++ {
		nr := int(sendn.Values[si])
		nrMax = max(nr)
		nrMin = min(nrMin, nr)
		nrMean += nr
	}
	// fmt.Printf("unif rnd large rNtot: %d  pcon: %g  max: %d  min: %d  mean: %g\n", rNtot, pj.PCon, nrMax, nrMin, float32(nrMean)/float32(sNtot))

	assert.Equal(t, rNtot, 1200)
	assert.Equal(t, nrMax, 50)
	assert.Equal(t, nrMin, 41)

}

func TestUniformRandSelf(t *testing.T) {
	send := tensor.NewShape(2, 3)
	recv := tensor.NewShape(2, 3)

	sNtot := send.Len()
	rNtot := recv.Len()

	pj := NewUniformRand()
	pj.RandSeed = 10
	pj.PCon = 0.5
	pj.SelfCon = false
	sendn, recvn, cons := pj.Connect(send, recv, true)
	// fmt.Printf("unif rnd self: 2x3\n%s\n", string(ConsStringFull(send, recv, cons)))
	_ = recvn

	ex := `0 1 0 1 1 0 
1 0 1 1 0 0 
1 1 0 1 0 0 
0 1 1 0 0 1 
1 1 0 1 0 0 
1 1 0 1 0 0 
`

	assert.Equal(t, ex, string(ConsStringFull(send, recv, cons)))

	nrMax := 0
	nrMin := rNtot
	nrMean := 0
	for si := 0; si < sNtot; si++ {
		nr := int(sendn.Values[si])
		nrMax = max(nr)
		nrMin = min(nrMin, nr)
		nrMean += nr
	}
	// fmt.Printf("sendn: %v\n", sendn.Values)
	// fmt.Printf("unif rnd rNtot: %d  pcon: %g  max: %d  min: %d  mean: %g\n", rNtot, pj.PCon, nrMax, nrMin, float32(nrMean)/float32(sNtot))

	assert.Equal(t, rNtot, 6)
	assert.Equal(t, nrMax, 1)
	assert.Equal(t, nrMin, 1)
}

func TestPoolUniformRand(t *testing.T) {
	send := tensor.NewShape(2, 3, 2, 3)
	recv := tensor.NewShape(2, 3, 3, 4)

	sNtot := send.Len()
	rNtot := recv.Len()

	pj := NewPoolUniformRand()
	pj.RandSeed = 10
	pj.PCon = 0.5
	sendn, recvn, cons := pj.Connect(send, recv, false)
	// fmt.Printf("unif rnd recv: 2x3x3x4 send: 2x3x2x3\n%s\n", string(ConsStringFull(send, recv, cons)))
	_ = recvn

	ex := `1 1 1 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 
0 1 1 1 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 
1 0 0 1 1 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 
0 1 0 1 1 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 
0 1 0 1 0 1 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 
0 0 0 1 1 1 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 
0 0 1 1 1 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 
0 0 1 0 1 1 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 
1 1 0 0 1 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 
0 1 1 0 0 1 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 
0 0 0 1 1 1 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 
0 1 0 0 1 1 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 
0 0 0 0 0 0 1 1 0 0 0 1 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 
0 0 0 0 0 0 0 1 1 0 0 1 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 
0 0 0 0 0 0 0 0 1 1 0 1 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 
0 0 0 0 0 0 0 1 0 1 1 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 
0 0 0 0 0 0 0 1 0 1 0 1 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 
0 0 0 0 0 0 1 1 1 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 
0 0 0 0 0 0 1 0 0 0 1 1 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 
0 0 0 0 0 0 1 1 0 1 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 
0 0 0 0 0 0 1 0 1 1 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 
0 0 0 0 0 0 1 1 1 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 
0 0 0 0 0 0 1 1 0 0 1 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 
0 0 0 0 0 0 0 0 0 1 1 1 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 
0 0 0 0 0 0 0 0 0 0 0 0 0 0 1 1 1 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 
0 0 0 0 0 0 0 0 0 0 0 0 1 0 1 0 0 1 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 
0 0 0 0 0 0 0 0 0 0 0 0 0 1 1 0 1 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 
0 0 0 0 0 0 0 0 0 0 0 0 0 0 1 1 0 1 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 
0 0 0 0 0 0 0 0 0 0 0 0 0 1 0 0 1 1 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 
0 0 0 0 0 0 0 0 0 0 0 0 0 1 0 0 1 1 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 
0 0 0 0 0 0 0 0 0 0 0 0 0 1 1 1 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 
0 0 0 0 0 0 0 0 0 0 0 0 0 0 1 1 0 1 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 
0 0 0 0 0 0 0 0 0 0 0 0 0 1 0 1 0 1 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 
0 0 0 0 0 0 0 0 0 0 0 0 0 0 1 0 1 1 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 
0 0 0 0 0 0 0 0 0 0 0 0 0 1 1 0 1 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 
0 0 0 0 0 0 0 0 0 0 0 0 0 1 0 1 0 1 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 
0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 1 0 1 1 0 0 0 0 0 0 0 0 0 0 0 0 
0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 1 0 0 1 1 0 0 0 0 0 0 0 0 0 0 0 0 
0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 1 1 0 0 1 0 0 0 0 0 0 0 0 0 0 0 0 
0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 1 1 0 0 1 0 0 0 0 0 0 0 0 0 0 0 0 0 
0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 1 0 1 1 0 0 0 0 0 0 0 0 0 0 0 0 
0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 1 1 0 0 1 0 0 0 0 0 0 0 0 0 0 0 0 
0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 1 1 0 0 1 0 0 0 0 0 0 0 0 0 0 0 0 
0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 1 1 1 0 0 0 0 0 0 0 0 0 0 0 0 0 0 
0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 1 1 0 0 1 0 0 0 0 0 0 0 0 0 0 0 0 0 
0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 1 0 0 0 1 1 0 0 0 0 0 0 0 0 0 0 0 0 
0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 1 1 0 1 0 0 0 0 0 0 0 0 0 0 0 0 
0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 1 1 1 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 
0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 1 0 0 1 1 0 0 0 0 0 0 
0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 1 1 1 0 0 0 0 0 0 0 0 0 
0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 1 1 1 0 0 0 0 0 0 0 0 
0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 1 1 0 0 1 0 0 0 0 0 0 
0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 1 0 1 1 0 0 0 0 0 0 
0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 1 1 1 0 0 0 0 0 0 0 
0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 1 1 0 0 0 1 0 0 0 0 0 0 
0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 1 1 1 0 0 0 0 0 0 
0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 1 0 0 0 1 1 0 0 0 0 0 0 
0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 1 0 1 0 0 1 0 0 0 0 0 0 
0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 1 0 1 1 0 0 0 0 0 0 0 0 
0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 1 0 1 1 0 0 0 0 0 0 0 
0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 1 1 1 0 
0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 1 1 0 1 0 0 
0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 1 0 1 1 0 0 
0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 1 0 0 1 1 0 
0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 1 1 1 0 
0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 1 0 1 0 1 
0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 1 0 1 1 0 0 
0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 1 0 1 0 1 
0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 1 0 1 1 0 0 
0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 1 0 0 1 1 0 
0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 1 0 1 1 0 0 
0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 1 1 0 1 0 0 
`

	assert.Equal(t, ex, string(ConsStringFull(send, recv, cons)))

	nrMax := 0
	nrMin := rNtot
	nrMean := 0
	for si := 0; si < sNtot; si++ {
		nr := int(sendn.Values[si])
		nrMax = max(nr)
		nrMin = min(nrMin, nr)
		nrMean += nr
	}
	// fmt.Printf("sendn: %v\n", sendn.Values)
	// fmt.Printf("unif rnd rNtot: %d  pcon: %g  max: %d  min: %d  mean: %g\n", rNtot, pj.PCon, nrMax, nrMin, float32(nrMean)/float32(sNtot))

	assert.Equal(t, rNtot, 72)
	assert.Equal(t, nrMax, 2)
	assert.Equal(t, nrMin, 1)

	// now test recip
	// rpj := NewUniformRand()
	// rpj.PCon = 0.5
	// rpj.Recip = true
	// sendn, recvn, cons = rpj.Connect(send, recv, false)
	// fmt.Printf("unif rnd recip recv: 3x4 send: 2x3\n%s\n", string(ConsStringFull(send, recv, cons)))

	// _ = recvn
}

func TestPoolUniformRandLg(t *testing.T) {
	send := tensor.NewShape(2, 3, 20, 30)
	recv := tensor.NewShape(2, 3, 30, 40)

	sNtot := send.Len()
	rNtot := recv.Len()

	pj := NewPoolUniformRand()
	pj.PCon = 0.05
	pj.RandSeed = 10
	sendn, recvn, cons := pj.Connect(send, recv, false)

	_ = recvn
	_ = cons

	nrMax := 0
	nrMin := rNtot
	nrMean := 0
	for si := 0; si < sNtot; si++ {
		nr := int(sendn.Values[si])
		nrMax = max(nr)
		nrMin = min(nrMin, nr)
		nrMean += nr
	}
	// fmt.Printf("unif rnd large rNtot: %d  pcon: %g  max: %d  min: %d  mean: %g\n", rNtot, pj.PCon, nrMax, nrMin, float32(nrMean)/float32(sNtot))

	assert.Equal(t, rNtot, 7200)
	assert.Equal(t, nrMax, 66)
	assert.Equal(t, nrMin, 33)
}

func TestPoolUniformRandSelf(t *testing.T) {
	send := tensor.NewShape(2, 3, 2, 3)
	recv := tensor.NewShape(2, 3, 2, 3)

	sNtot := send.Len()
	rNtot := recv.Len()

	pj := NewPoolUniformRand()
	pj.PCon = 0.5
	pj.RandSeed = 10
	pj.SelfCon = false
	sendn, recvn, cons := pj.Connect(send, recv, true)
	// fmt.Printf("unif rnd self: 2x3x2x3\n%s\n", string(ConsStringFull(send, recv, cons)))
	_, _ = recvn, cons

	nrMax := 0
	nrMin := rNtot
	nrMean := 0
	for si := 0; si < sNtot; si++ {
		nr := int(sendn.Values[si])
		nrMax = max(nr)
		nrMin = min(nrMin, nr)
		nrMean += nr
	}
	// fmt.Printf("sendn: %v\n", sendn.Values)
	// fmt.Printf("unif rnd rNtot: %d  pcon: %g  max: %d  min: %d  mean: %g\n", rNtot, pj.PCon, nrMax, nrMin, float32(nrMean)/float32(sNtot))

	assert.Equal(t, rNtot, 36)
	assert.Equal(t, nrMax, 2)
	assert.Equal(t, nrMin, 1)
}

func TestPoolSameUnit(t *testing.T) {
	send := tensor.NewShape(1, 2, 2, 3)
	recv := tensor.NewShape(1, 2, 2, 3)

	sNp := send.DimSize(0) * send.DimSize(1)
	rNp := recv.DimSize(0) * recv.DimSize(1)

	pj := NewPoolSameUnit()
	sendn, recvn, cons := pj.Connect(send, recv, false)
	// fmt.Printf("pool same unit both 2x3 1x2\n%s\n", string(ConsStringFull(send, recv, cons)))

	ex := `1 0 0 0 0 0 1 0 0 0 0 0 
0 1 0 0 0 0 0 1 0 0 0 0 
0 0 1 0 0 0 0 0 1 0 0 0 
0 0 0 1 0 0 0 0 0 1 0 0 
0 0 0 0 1 0 0 0 0 0 1 0 
0 0 0 0 0 1 0 0 0 0 0 1 
1 0 0 0 0 0 1 0 0 0 0 0 
0 1 0 0 0 0 0 1 0 0 0 0 
0 0 1 0 0 0 0 0 1 0 0 0 
0 0 0 1 0 0 0 0 0 1 0 0 
0 0 0 0 1 0 0 0 0 0 1 0 
0 0 0 0 0 1 0 0 0 0 0 1 
`

	assert.Equal(t, ex, string(ConsStringFull(send, recv, cons)))

	CheckAllN(sendn, rNp, t)
	CheckAllN(recvn, sNp, t)
}

func TestPoolSameUnitRecv(t *testing.T) {
	send := tensor.NewShape(2, 3)
	recv := tensor.NewShape(1, 2, 2, 3)

	rNp := recv.DimSize(0) * recv.DimSize(1)
	pj := NewPoolSameUnit()
	sendn, recvn, cons := pj.Connect(send, recv, false)
	// fmt.Printf("pool same unit recv 2x3 1x2, send 2x3\n%s\n", string(ConsStringFull(send, recv, cons)))

	ex := `1 0 0 0 0 0 
0 1 0 0 0 0 
0 0 1 0 0 0 
0 0 0 1 0 0 
0 0 0 0 1 0 
0 0 0 0 0 1 
1 0 0 0 0 0 
0 1 0 0 0 0 
0 0 1 0 0 0 
0 0 0 1 0 0 
0 0 0 0 1 0 
0 0 0 0 0 1 
`

	assert.Equal(t, ex, string(ConsStringFull(send, recv, cons)))

	CheckAllN(sendn, rNp, t)
	CheckAllN(recvn, 1, t)
}

func TestPoolSameUnitSend(t *testing.T) {
	send := tensor.NewShape(1, 2, 2, 3)
	recv := tensor.NewShape(2, 3)

	sNp := send.DimSize(0) * send.DimSize(1)

	pj := NewPoolSameUnit()
	sendn, recvn, cons := pj.Connect(send, recv, false)
	// fmt.Printf("pool same unit send 2x3 1x2, recv 2x3\n%s\n", string(ConsStringFull(send, recv, cons)))

	ex := `1 0 0 0 0 0 1 0 0 0 0 0 
0 1 0 0 0 0 0 1 0 0 0 0 
0 0 1 0 0 0 0 0 1 0 0 0 
0 0 0 1 0 0 0 0 0 1 0 0 
0 0 0 0 1 0 0 0 0 0 1 0 
0 0 0 0 0 1 0 0 0 0 0 1 
`

	assert.Equal(t, ex, string(ConsStringFull(send, recv, cons)))

	CheckAllN(sendn, 1, t)
	CheckAllN(recvn, sNp, t)
}

func TestRect(t *testing.T) {
	send := tensor.NewShape(2, 3)
	recv := tensor.NewShape(2, 3)

	pj := NewRect()
	pj.Size.Set(2, 1)
	pj.Scale.Set(1, 1)
	sendn, recvn, cons := pj.Connect(send, recv, false)
	// fmt.Printf("rect 2x1 recv: 2x3 send: 2x3\n%s\n", string(ConsStringFull(send, recv, cons)))
	_ = sendn

	ex := `1 1 0 0 0 0 
0 1 1 0 0 0 
1 0 1 0 0 0 
0 0 0 1 1 0 
0 0 0 0 1 1 
0 0 0 1 0 1 
`

	assert.Equal(t, ex, string(ConsStringFull(send, recv, cons)))

	CheckAllN(sendn, 2, t)
	CheckAllN(recvn, 2, t)
}

func TestPoolRect(t *testing.T) {
	send := tensor.NewShape(2, 3, 2, 2)
	recv := tensor.NewShape(2, 3, 2, 2)

	pj := NewPoolRect()
	pj.Size.Set(2, 1)
	pj.Scale.Set(1, 1)
	sendn, recvn, cons := pj.Connect(send, recv, false)
	// fmt.Printf("rect 2x1 recv: 2x3 send: 2x3\n%s\n", string(ConsStringFull(send, recv, cons)))
	_ = sendn

	ex := `1 1 1 1 1 1 1 1 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 
1 1 1 1 1 1 1 1 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 
1 1 1 1 1 1 1 1 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 
1 1 1 1 1 1 1 1 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 
0 0 0 0 1 1 1 1 1 1 1 1 0 0 0 0 0 0 0 0 0 0 0 0 
0 0 0 0 1 1 1 1 1 1 1 1 0 0 0 0 0 0 0 0 0 0 0 0 
0 0 0 0 1 1 1 1 1 1 1 1 0 0 0 0 0 0 0 0 0 0 0 0 
0 0 0 0 1 1 1 1 1 1 1 1 0 0 0 0 0 0 0 0 0 0 0 0 
1 1 1 1 0 0 0 0 1 1 1 1 0 0 0 0 0 0 0 0 0 0 0 0 
1 1 1 1 0 0 0 0 1 1 1 1 0 0 0 0 0 0 0 0 0 0 0 0 
1 1 1 1 0 0 0 0 1 1 1 1 0 0 0 0 0 0 0 0 0 0 0 0 
1 1 1 1 0 0 0 0 1 1 1 1 0 0 0 0 0 0 0 0 0 0 0 0 
0 0 0 0 0 0 0 0 0 0 0 0 1 1 1 1 1 1 1 1 0 0 0 0 
0 0 0 0 0 0 0 0 0 0 0 0 1 1 1 1 1 1 1 1 0 0 0 0 
0 0 0 0 0 0 0 0 0 0 0 0 1 1 1 1 1 1 1 1 0 0 0 0 
0 0 0 0 0 0 0 0 0 0 0 0 1 1 1 1 1 1 1 1 0 0 0 0 
0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 1 1 1 1 1 1 1 1 
0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 1 1 1 1 1 1 1 1 
0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 1 1 1 1 1 1 1 1 
0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 1 1 1 1 1 1 1 1 
0 0 0 0 0 0 0 0 0 0 0 0 1 1 1 1 0 0 0 0 1 1 1 1 
0 0 0 0 0 0 0 0 0 0 0 0 1 1 1 1 0 0 0 0 1 1 1 1 
0 0 0 0 0 0 0 0 0 0 0 0 1 1 1 1 0 0 0 0 1 1 1 1 
0 0 0 0 0 0 0 0 0 0 0 0 1 1 1 1 0 0 0 0 1 1 1 1 
`
	assert.Equal(t, ex, string(ConsStringFull(send, recv, cons)))

	CheckAllN(sendn, 2*4, t)
	CheckAllN(recvn, 2*4, t)
}
