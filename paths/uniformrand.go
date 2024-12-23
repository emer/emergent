// Copyright (c) 2019, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package paths

import (
	"math"
	"math/rand"
	"sort"

	"cogentcore.org/lab/base/randx"
	"cogentcore.org/lab/tensor"
)

// UniformRand implements uniform random pattern of connectivity between two layers
// using a permuted (shuffled) list for without-replacement randomness,
// and maintains its own local random number source and seed
// which are initialized if Rand == nil -- usually best to keep this
// specific to each instance of a pathway so it is fully reproducible
// and doesn't interfere with other random number streams.
type UniformRand struct {

	// probability of connection (0-1)
	PCon float32 `min:"0" max:"1"`

	// if true, and connecting layer to itself (self pathway), then make a self-connection from unit to itself
	SelfCon bool

	// reciprocal connectivity: if true, switch the sending and receiving layers to create a symmetric top-down pathway -- ESSENTIAL to use same RandSeed between two paths to ensure symmetry
	Recip bool

	// random number source -- is created with its own separate source if nil
	Rand randx.Rand `display:"-"`

	// the current random seed -- will be initialized to a new random number from the global random stream when Rand is created.
	RandSeed int64 `display:"-"`
}

func NewUniformRand() *UniformRand {
	return &UniformRand{PCon: 0.5}
}

func (ur *UniformRand) Name() string {
	return "UniformRand"
}

func (ur *UniformRand) InitRand() {
	if ur.Rand != nil {
		ur.Rand.Seed(ur.RandSeed)
		return
	}
	if ur.RandSeed == 0 {
		ur.RandSeed = int64(rand.Uint64())
	}
	ur.Rand = randx.NewSysRand(ur.RandSeed)
}

func (ur *UniformRand) Connect(send, recv *tensor.Shape, same bool) (sendn, recvn *tensor.Int32, cons *tensor.Bool) {
	if ur.PCon >= 1 {
		return ur.ConnectFull(send, recv, same)
	}
	if ur.Recip {
		return ur.ConnectRecip(send, recv, same)
	}
	sendn, recvn, cons = NewTensors(send, recv)
	slen := send.Len()
	rlen := recv.Len()

	noself := same && !ur.SelfCon
	var nsend int
	if noself {
		nsend = int(math.Round(float64(ur.PCon) * float64(slen-1)))
	} else {
		nsend = int(math.Round(float64(ur.PCon) * float64(slen)))
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

	ur.InitRand()

	sordlen := slen
	if noself {
		sordlen--
	}

	sorder := ur.Rand.Perm(sordlen)
	slist := make([]int, nsend)
	for ri := 0; ri < rlen; ri++ {
		if noself { // need to exclude ri
			ix := 0
			for j := 0; j < slen; j++ {
				if j != ri {
					sorder[ix] = j
					ix++
				}
			}
			randx.PermuteInts(sorder, ur.Rand)
		}
		copy(slist, sorder)
		sort.Ints(slist) // keep list sorted for more efficient memory traversal etc
		for si := 0; si < nsend; si++ {
			off := ri*slen + slist[si]
			cons.Values.Set(true, off)
		}
		randx.PermuteInts(sorder, ur.Rand)
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

// ConnectRecip does reciprocal connectvity
func (ur *UniformRand) ConnectRecip(send, recv *tensor.Shape, same bool) (sendn, recvn *tensor.Int32, cons *tensor.Bool) {
	sendn, recvn, cons = NewTensors(send, recv)
	slen := recv.Len() // swapped
	rlen := send.Len()

	slenR := send.Len() // NOT swapped

	noself := same && !ur.SelfCon
	var nsend int
	if noself {
		nsend = int(math.Round(float64(ur.PCon) * float64(slen-1)))
	} else {
		nsend = int(math.Round(float64(ur.PCon) * float64(slen)))
	}

	rnv := sendn.Values // swapped
	for i := range rnv {
		rnv[i] = int32(nsend)
	}

	ur.InitRand()

	sordlen := slen
	if noself {
		sordlen--
	}

	sorder := ur.Rand.Perm(sordlen)
	slist := make([]int, nsend)
	for ri := 0; ri < rlen; ri++ {
		if noself { // need to exclude ri
			ix := 0
			for j := 0; j < slen; j++ {
				if j != ri {
					sorder[ix] = j
					ix++
				}
			}
			randx.PermuteInts(sorder, ur.Rand)
		}
		copy(slist, sorder)
		sort.Ints(slist) // keep list sorted for more efficient memory traversal etc
		for si := 0; si < nsend; si++ {
			off := slist[si]*slenR + ri
			cons.Values.Set(true, off)
		}
		randx.PermuteInts(sorder, ur.Rand)
	}

	// set send n's empirically
	snv := recvn.Values // swapped
	for si := range snv {
		nr := 0
		for ri := 0; ri < rlen; ri++ { // actually si
			off := si*slenR + ri
			if cons.Values.Index(off) {
				nr++
			}
		}
		snv[si] = int32(nr)
	}
	return
}

func (ur *UniformRand) ConnectFull(send, recv *tensor.Shape, same bool) (sendn, recvn *tensor.Int32, cons *tensor.Bool) {
	sendn, recvn, cons = NewTensors(send, recv)
	cons.Values.SetAll(true)
	nsend := send.Len()
	nrecv := recv.Len()
	if same && !ur.SelfCon {
		for i := 0; i < nsend; i++ { // nsend = nrecv
			off := i*nsend + i
			cons.Values.Set(false, off)
		}
		nsend--
		nrecv--
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
