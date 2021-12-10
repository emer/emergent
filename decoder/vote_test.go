// Copyright (c) 2021, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package decoder

import (
	"fmt"
	"math/rand"
	"testing"
	"time"
)

func TestVoteInt(t *testing.T) {
	rand.Seed(time.Now().UTC().UnixNano())
	votes := []int{1, 2, 1, 5, 3, 2, 3, 2, 4, 2}
	tv, tn := TopVoteInt(votes)
	// fmt.Printf("top vote: %d got %d votes\n", tv, tn)
	if tv != 2 || tn != 4 {
		t.Errorf("tv %d != 2, tn %d != 4\n", tv, tn)
	}
	votes = []int{1, 2, 1, 5, 3, 1, 3, 1, 4, 2}
	tv, tn = TopVoteInt(votes)
	// fmt.Printf("top vote: %d got %d votes\n", tv, tn)
	if tv != 1 || tn != 4 {
		t.Errorf("tv %d != 1, tn %d != 4\n", tv, tn)
	}
	votes = []int{1, 2, 1, 5, 3, 5, 3, 5, 4, 2}
	tv, tn = TopVoteInt(votes)
	// fmt.Printf("top vote: %d got %d votes\n", tv, tn)
	if tv != 5 || tn != 3 {
		t.Errorf("tv %d != 5, tn %d != 3\n", tv, tn)
	}
	// tie -- run multiple times to see
	votes = []int{1, 2, 1, 5, 3, 5, 3, 5, 4, 2, 2}
	tv, tn = TopVoteInt(votes)
	fmt.Printf("top vote: %d got %d votes\n", tv, tn)
	if tv != 5 || tn != 3 {
		t.Errorf("tv %d != 5, tn %d != 3\n", tv, tn)
	}
}
