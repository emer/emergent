// Copyright (c) 2020, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package empi

import (
	"github.com/emer/emergent/v2/empi/mpi"
	"github.com/emer/etable/v2/etable"
)

// GatherTableRows does an MPI AllGather on given src table data, gathering into dest.
// dest will have np * src.Rows Rows, filled with each processor's data, in order.
// dest must be a clone of src: if not same number of cols, will be configured from src.
func GatherTableRows(dest, src *etable.Table, comm *mpi.Comm) {
	sr := src.Rows
	np := mpi.WorldSize()
	dr := np * sr
	if len(dest.Cols) != len(src.Cols) {
		dest.SetFromSchema(src.Schema(), dr)
	} else {
		dest.SetNumRows(dr)
	}
	for ci, st := range src.Cols {
		dt := dest.Cols[ci]
		GatherTensorRows(dt, st, comm)
	}
}

// ReduceTable does an MPI AllReduce on given src table data using given operation,
// gathering into dest.
// each processor must have the same table organization -- the tensor values are
// just aggregated directly across processors.
// dest will be a clone of src if not the same (cos & rows),
// does nothing for strings.
func ReduceTable(dest, src *etable.Table, comm *mpi.Comm, op mpi.Op) {
	sr := src.Rows
	if len(dest.Cols) != len(src.Cols) {
		dest.SetFromSchema(src.Schema(), sr)
	} else {
		dest.SetNumRows(sr)
	}
	for ci, st := range src.Cols {
		dt := dest.Cols[ci]
		ReduceTensor(dt, st, comm, op)
	}
}
