// Copyright (c) 2019, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package dtable

import (
	"testing"

	"github.com/emer/emergent/etensor"
)

func TestAdd4DCol(t *testing.T) {
	dt := Table{}
	err := dt.AddCol(etensor.NewFloat32([]int{1, 11, 1, 16}, nil, nil), "Vals")
	if err != nil {
		t.Error(err)
	}

	col := dt.ColByName("Vals")
	if col.NumDims() != 4 {
		t.Errorf("Add4DCol: # of dims != 4\n")
	}

	if col.Dim(0) != 1 {
		t.Errorf("Add4DCol: dim 0 len != 1, was: %v\n", col.Dim(0))
	}

	if col.Dim(1) != 11 {
		t.Errorf("Add4DCol: dim 0 len != 11, was: %v\n", col.Dim(1))
	}

	if col.Dim(2) != 1 {
		t.Errorf("Add4DCol: dim 0 len != 1, was: %v\n", col.Dim(2))
	}

	if col.Dim(3) != 16 {
		t.Errorf("Add4DCol: dim 0 len != 16, was: %v\n", col.Dim(3))
	}
}
