// Copyright (c) 2019, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package actrf

import (
	"cogentcore.org/lab/base/mpi"
	"cogentcore.org/lab/tensor/tensormpi"
)

// MPISum aggregates RF Sum data across all processors in given mpi communicator.
// It adds to SumProd and SumSrc. Call this prior to calling NormAvg().
func (af *RF) MPISum(comm *mpi.Comm) {
	if mpi.WorldSize() == 1 {
		return
	}
	tensormpi.ReduceTensor(&af.MPITmp, &af.SumProd, comm, mpi.OpSum)
	af.SumProd.CopyFrom(&af.MPITmp)
	tensormpi.ReduceTensor(&af.MPITmp, &af.SumSrc, comm, mpi.OpSum)
	af.SumSrc.CopyFrom(&af.MPITmp)
}

// MPISum aggregates RF Sum data across all processors in given mpi communicator.
// It adds to SumProd and SumSrc. Call this prior to calling NormAvg().
func (af *RFs) MPISum(comm *mpi.Comm) {
	for _, rf := range af.RFs {
		rf.MPISum(comm)
	}
}
