// Copyright (c) 2019, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package dtable

import (
	"github.com/emer/emergent/etensor"
	"testing"
)

func TestAdd4DCol(t *testing.T) {
	dt := Table{}
	err := dt.AddCol(etensor.NewFloat32([]int{11, 1, 1, 16}, nil, nil), "Vals")
	if err != nil {
		t.Error(err)
	}

	col := dt.ColByName("Vals")
	if col.NumDims() != 4 {
		t.Errorf("Add4DCol: # of dims != 4\n")
	}

	if col.Dim(0) != 11 {
		t.Errorf("Add4DCol: dim 0 len != 11\n")
	}

	if col.Dim(1) != 1 {
		t.Errorf("Add4DCol: dim 0 len != 11\n")
	}

	if col.Dim(2) != 1 {
		t.Errorf("Add4DCol: dim 0 len != 11\n")
	}

	if col.Dim(3) != 16 {
		t.Errorf("Add4DCol: dim 0 len != 11\n")
	}
}
