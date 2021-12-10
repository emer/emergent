// Copyright (c) 2021, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package decoder

import (
	"math/rand"
	"sort"
)

// TopVoteInt returns the choice with the most votes among a list of votes
// as integer-valued choices, and also returns the number of votes for that item.
// In the case of ties, it chooses one at random (otherwise it would have a bias
// toward the lowest numbered item).
func TopVoteInt(votes []int) (int, int) {
	sort.Ints(votes)
	prv := votes[0]
	cur := prv
	top := prv
	topn := 1
	curn := 1
	n := len(votes)
	var ties []int
	for i := 1; i < n; i++ {
		cur = votes[i]
		if cur != prv {
			if curn > topn {
				top = prv
				topn = curn
				ties = []int{top}
			} else if curn == topn {
				ties = append(ties, prv)
			}
			curn = 1
			prv = cur
		} else {
			curn++
		}
	}
	if curn > topn {
		top = cur
		topn = curn
		ties = []int{top}
	} else if curn == topn {
		ties = append(ties, cur)
	}
	if len(ties) > 1 {
		ti := rand.Intn(len(ties))
		top = ties[ti]
	}
	return top, topn
}

// TopVoteString returns the choice with the most votes among a list of votes
// as string-valued choices, and also returns the number of votes for that item.
// In the case of ties, it chooses one at random (otherwise it would have a bias
// toward the lowest numbered item).
func TopVoteString(votes []string) (string, int) {
	sort.Strings(votes)
	prv := votes[0]
	cur := prv
	top := prv
	topn := 1
	curn := 1
	n := len(votes)
	var ties []string
	for i := 1; i < n; i++ {
		cur = votes[i]
		if cur != prv {
			if curn > topn {
				top = prv
				topn = curn
				ties = []string{top}
			} else if curn == topn {
				ties = append(ties, prv)
			}
			curn = 1
			prv = cur
		} else {
			curn++
		}
	}
	if curn > topn {
		top = cur
		topn = curn
		ties = []string{top}
	} else if curn == topn {
		ties = append(ties, cur)
	}
	if len(ties) > 1 {
		ti := rand.Intn(len(ties))
		top = ties[ti]
	}
	return top, topn
}
