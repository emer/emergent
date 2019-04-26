// Copyright (c) 2019, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package patgen

import (
	"math/rand"

	"github.com/emer/dtable/etensor"
	"github.com/emer/emergent/erand"
)

// PermutedBinary sets the given tensor to contain nOn onVal values and the
// remainder are offVal values, using a permuted order of tensor elements (i.e.,
// randomly shuffled or permuted).
func PermutedBinary(tsr etensor.Tensor, nOn int, onVal, offVal float64) {
	ln := tsr.Len()
	if ln == 0 {
		return
	}
	pord := rand.Perm(ln)
	for i := 0; i < ln; i++ {
		if i < nOn {
			tsr.SetFloat1D(pord[i], onVal)
		} else {
			tsr.SetFloat1D(pord[i], offVal)
		}
	}
}

// PermutedBinaryRows treats the tensor as a column of rows as in a dtable.Table
// and sets each row to contain nOn onVal values and the remainder are offVal values,
// using a permuted order of tensor elements (i.e., randomly shuffled or permuted).
func PermutedBinaryRows(tsr etensor.Tensor, nOn int, onVal, offVal float64) {
	rows, cells := tsr.RowCellSize()
	if rows == 0 || cells == 0 {
		return
	}
	pord := rand.Perm(cells)
	for rw := 0; rw < rows; rw++ {
		stidx := rw * cells
		for i := 0; i < cells; i++ {
			if i < nOn {
				tsr.SetFloat1D(stidx+pord[i], onVal)
			} else {
				tsr.SetFloat1D(stidx+pord[i], offVal)
			}
		}
		erand.PermuteInts(pord)
	}
}

/*
bool taDataGen::PermutedBinary_MinDist(DataTable* data, const String& col_nm, int n_on,
                                       float dist, taMath::DistMetric metric,
                                       bool norm, float tol, int thr_no)
{
  if(!data) return false;
  if(col_nm.empty()) {
    bool rval = true;
    for(int pn = 0;pn<data->data.size;pn++) {
      DataCol* da = data->data.FastEl(pn);
      if(da->is_matrix && da->valType() == VT_FLOAT) {
        if(!PermutedBinary_MinDist(data, da->name, n_on, dist, metric, norm, tol, thr_no))
          rval = false;
      }
    }
    return rval;
  }
  DataCol* da = GetFloatMatrixDataCol(data, col_nm);
  if(!da) return false;
  bool larger_further = taMath::dist_larger_further(metric);
  int bogus_count = 0;
  data->DataUpdate(true);
  for(int i =0;i<da->rows();i++) {
    float_Matrix* mat = (float_Matrix*)da->GetValAsMatrix(i);
    taBase::Ref(mat);
    int cnt = 100 + (10 * (i + 1));   // 100 plus 10 more for every new stim
    bool ok = false;
    float min_d;
    do {
      PermutedBinaryMat(mat, n_on, 1.0f, 0.0f, thr_no);
      min_d = LastMinDist(da, i, metric, norm, tol);
      cnt--;
      if(larger_further)
        ok = (min_d >= dist);
      else
        ok = (min_d <= dist);
    } while(!ok && (cnt > 0));
    taBase::unRefDone(mat);

    if(cnt == 0) {
      taMisc::Warning("*** PermutedBinary_MinDist row:", String(i), "dist of:", (String)min_d,
                     "under dist limit:", (String)dist);
      bogus_count++;
    }
    if(bogus_count > 5) {
      taMisc::Warning("PermutedBinary_MinDist Giving up after 5 stimuli under the limit, set limits lower");
      data->DataUpdate(false);
      return false;
    }
  }
  data->DataUpdate(false);
  return true;
}

*/
