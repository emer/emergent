// Copyright (c) 2019, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package prjn

import (
	"math"
	"math/rand"
	"sort"

	"github.com/emer/emergent/erand"
	"github.com/emer/etable/etensor"
	"github.com/goki/ki/ints"
)

// PoolUnifRnd implements random pattern of connectivity between pools within layers.
// Pools are the outer-most two dimensions of a 4D layer shape.
// If either layer does not have pools, PoolUnifRnd works as UnifRnd does.
// If probability of connection (PCon) is 1, PoolUnifRnd works as PoolOnetoOne does.
type PoolUnifRnd struct {
	PoolOneToOne
	UnifRnd
}

func NewPoolUnifRnd() *PoolUnifRnd {
	newur := &PoolUnifRnd{}
	newur.PCon = 0.5
	return newur
}

func (ur *PoolUnifRnd) Name() string {
	return "PoolUnifRnd"
}

func (ur *PoolUnifRnd) Connect(send, recv *etensor.Shape, same bool) (sendn, recvn *etensor.Int32, cons *etensor.Bits) {
	if send.NumDims() == 4 && recv.NumDims() == 4 {
		return ur.ConnectPoolsRnd(send, recv, same)
	}
	return ur.ConnectRnd(send, recv, same)
}

// ConnectPoolsRnd is when both recv and send have pools
func (ur *PoolUnifRnd) ConnectPoolsRnd(send, recv *etensor.Shape, same bool) (sendn, recvn *etensor.Int32, cons *etensor.Bits) {
	if ur.PCon >= 1 {
		return ur.ConnectPools(send, recv, same)
	}
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

	noself := same && !ur.SelfCon
	var nsend int
	if noself {
		nsend = int(math.Round(float64(ur.PCon) * float64(sNu-1)))
	} else {
		nsend = int(math.Round(float64(ur.PCon) * float64(sNu)))
	}

	if ur.RndSeed == 0 {
		ur.RndSeed = int64(rand.Uint64())
	}
	rand.Seed(ur.RndSeed)

	sordlen := sNu
	if noself {
		sordlen--
	}

	sorder := rand.Perm(sordlen)
	slist := make([]int, nsend)

	if ur.NPools > 0 {
		npl = ints.MinInt(ur.NPools, rNp)
	}
	for i := 0; i < npl; i++ {
		rpi := ur.RecvStart + i
		spi := ur.SendStart + i
		if rpi >= rNp || spi >= sNp {
			break
		}
		for rui := 0; rui < rNu; rui++ {
			ri := rpi*rNu + rui
			rnv[ri] = int32(nsend)
			if noself { // need to exclude ri
				ix := 0
				for j := 0; j < sNu; j++ {
					ji := spi*sNu + j
					if ji != ri {
						sorder[ix] = j
						ix++
					}
				}
				erand.PermuteInts(sorder)
			}
			copy(slist, sorder)
			sort.Ints(slist)
			for sui := 0; sui < nsend; sui++ {
				si := spi*sNu + slist[sui]
				off := ri*sNtot + si
				cons.Values.Set(off, true)
			}
			erand.PermuteInts(sorder)
		}
		for sui := 0; sui < sNu; sui++ {
			nr := 0
			si := spi*sNu + sui
			for rui := 0; rui < rNu; rui++ {
				ri := rpi*rNu + rui
				off := ri*sNtot + si
				if cons.Values.Index(off) {
					nr++
				}
			}
			snv[si] = int32(nr)
		}
	}
	return
}

// ConnectRnd is a copy of UnifRnd.Connect with initial if statement modified
func (ur *PoolUnifRnd) ConnectRnd(send, recv *etensor.Shape, same bool) (sendn, recvn *etensor.Int32, cons *etensor.Bits) {
	if ur.PCon >= 1 {
		switch {
		case send.NumDims() == 2 && recv.NumDims() == 4:
			return ur.ConnectRecvPool(send, recv, same)
		case send.NumDims() == 4 && recv.NumDims() == 2:
			return ur.ConnectSendPool(send, recv, same)
		case send.NumDims() == 2 && recv.NumDims() == 2:
			return ur.ConnectOneToOne(send, recv, same)
		}
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
	// nrSEM := nrMean / mat32.Sqrt(nrMean)
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

	sordlen := slen
	if noself {
		sordlen--
	}

	sorder := rand.Perm(sordlen)
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
			erand.PermuteInts(sorder)
		}
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
