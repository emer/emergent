// Copyright (c) 2019, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package prjn

import (
	"math/rand"
	"sort"

	"github.com/emer/emergent/erand"
	"github.com/emer/etable/etensor"
	"github.com/g3n/engine/math32"
)

// UnifRnd implements uniform random pattern of connectivity between two layers
// uses a permuted (shuffled) list for without-replacement randomness
// and maintains its own local random seed for fully replicable results
// (if seed is not set when run, then random number generator is used to create seed)
// should reset seed after calling to resume sequence appropriately
type UnifRnd struct {
	PCon    float32 `min:"0" max:"1" desc:"probability of connection (0-1)"`
	RndSeed int64   `view:"-" desc:"the current random seed"`
	SelfCon bool    `desc:"if true, and connecting layer to itself (self projection), then make a self-connection from unit to itself"`
}

func NewUnifRnd() *UnifRnd {
	return &UnifRnd{}
}

func (ur *UnifRnd) Name() string {
	return "UnifRnd"
}

func (ur *UnifRnd) Connect(send, recv *etensor.Shape, same bool) (sendn, recvn *etensor.Int32, cons *etensor.Bits) {
	if ur.PCon >= 1 {
		return ur.ConnectFull(send, recv, same)
	}
	sendn, recvn, cons = NewTensors(send, recv)
	slen := send.Len()
	rlen := recv.Len()

	var nsend int
	if same && ur.SelfCon {
		nsend = int(math32.Round(ur.PCon * float32(slen-1)))
	} else {
		nsend = int(math32.Round(ur.PCon * float32(slen)))
	}

	// NOTE: this is reasonably accurate: mean + 3 * SEM, but we can just use
	// empirical values more easily and safely.

	// recv number is even distribution across recvs plus some imbalance factor
	// nrMean := float32(rlen*nsend) / float32(slen)
	// // add 3 * SEM as corrective factor
	// nrSEM := nrMean / math32.Sqrt(nrMean)
	// nrecv := int(nrMean + 3*nrSEM)
	// if nrecv > rlen {
	// 	nrecv = rlen
	// }

	rnv := recvn.Values
	for i := range rnv {
		rnv[i] = int32(nsend)
	}

	if ur.RndSeed == 0 {
		ur.RndSeed = int64(rand.Uint64())
	}
	rand.Seed(ur.RndSeed)

	sorder := rand.Perm(slen)
	slist := make([]int, nsend)
	for ri := 0; ri < rlen; ri++ {
		copy(slist, sorder)
		sort.Ints(slist) // keep list sorted for more efficient memory traversal etc
		for si := 0; si < nsend; si++ {
			off := ri*slen + slist[si]
			cons.Values.Set(off, true)
		}
		erand.PermuteInts(sorder)
	}

	// 	set send n's empirically
	snv := sendn.Values
	for si := range snv {
		nr := 0
		for ri := 0; ri < rlen; ri++ {
			off := ri*slen + si
			if cons.Values.Index(off) {
				nr++
			}
		}
		snv[si] = int32(nr)
	}
	return
}

func (ur *UnifRnd) ConnectFull(send, recv *etensor.Shape, same bool) (sendn, recvn *etensor.Int32, cons *etensor.Bits) {
	sendn, recvn, cons = NewTensors(send, recv)
	cons.Values.SetAll(true)
	nsend := send.Len()
	nrecv := recv.Len()
	if same && !ur.SelfCon {
		nsend--
		nrecv--
		for i := 0; i < nsend; i++ { // nsend = nrecv
			off := i*nsend + i
			cons.Values.Set(off, false)
		}
	}
	rnv := recvn.Values
	for i := range rnv {
		rnv[i] = int32(nsend)
	}
	snv := sendn.Values
	for i := range snv {
		snv[i] = int32(nrecv)
	}
	return
}

func (ur *UnifRnd) HasWeights() bool {
	return false
}

func (ur *UnifRnd) Weights(sendn, recvn *etensor.Int32, cons *etensor.Bits) []float32 {
	return nil
}
