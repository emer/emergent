// Copyright (c) 2021, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package empi

import (
	"errors"
	"fmt"
	"math/rand"

	"github.com/emer/emergent/v2/empi/mpi"
)

// RandCheck checks that the current random numbers generated across each
// MPI processor are identical. Most emergent simulations depend on this
// being true, so it is good to check periodically to ensure!
func RandCheck(comm *mpi.Comm) error {
	ws := comm.Size()
	rnd := rand.Int()
	src := []int{rnd}
	agg := make([]int, ws)
	err := comm.AllGatherInt(agg, src)
	if err != nil {
		return err
	}
	errs := ""
	for i := range agg {
		if agg[i] != rnd {
			errs += fmt.Sprintf("%d ", i)
		}
	}
	if errs != "" {
		err = errors.New("empi.RandCheck: random numbers differ in procs: " + errs)
		mpi.Printf("%s\n", err)
	}
	return err
}
