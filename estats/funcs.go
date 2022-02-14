// Copyright (c) 2022, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package estats

import (
	"github.com/emer/emergent/emer"
	"github.com/emer/etable/etable"
	"github.com/emer/etable/etensor"
	"github.com/emer/etable/metric"
	"github.com/emer/etable/norm"
)

// funcs contains misc stats functions

// SetLayerTensor sets tensor of Unit values on a layer for given variable
// to a F32Tensor with name = layNm
func (st *Stats) SetLayerTensor(net emer.Network, layNm, unitVar string) *etensor.Float32 {
	ly := net.LayerByName(layNm)
	tsr := st.F32Tensor(layNm)
	ly.UnitValsTensor(tsr, unitVar)
	return tsr
}

// ClosestStat finds the closest pattern in given column of given table of possible patterns,
// compared to layer activation pattern using given variable.  Returns the row number,
// correlation value, and value of a column named namecol for that row if non-empty.
// Column must be etensor.Float32
func (st *Stats) ClosestPat(net emer.Network, layNm, unitVar string, pats *etable.Table, colnm, namecol string) (int, float32, string) {
	tsr := st.SetLayerTensor(net, layNm, unitVar)
	col := pats.ColByName(colnm)
	// note: requires Increasing metric so using Inv
	row, cor := metric.ClosestRow32(tsr, col.(*etensor.Float32), metric.InvCorrelation32)
	cor = 1 - cor // convert back to correl
	nm := ""
	if namecol != "" {
		nm = pats.CellString(namecol, row)
	}
	return row, cor, nm
}

//////////////////////////////////////////////
//  PCA Stats

// PCAStrongThr is the threshold for counting PCA eigenvalues as "strong"
var PCAStrongThr = 0.01

// PCAStats computes PCA statistics on recorded hidden activation patterns
// on given log table (IdxView), and given list of layer names
// and variable name -- columns named "layer_var".
// Helpful for measuring the overall information (variance) in the representations
// to detect a common failure mode where a few patterns dominate over everything ("hogs").
// Records Float stats as:
// layer_PCA_NStrong: number of eigenvalues above the PCAStrongThr threshold
// layer_PCA_Top5: average strength of top 5 eigenvalues
// layer_PCA_Next5: average strength of next 5 eigenvalues
// layer_PCA_Rest: average strength of remaining eigenvalues (if more than 10 total eigens)
func (st *Stats) PCAStats(ix *etable.IdxView, varNm string, layers []string) {
	pca := &st.PCA
	for _, lnm := range layers {
		pca.TableCol(ix, lnm+"_"+varNm, metric.Covariance64)
		var nstr float64
		ln := len(pca.Values)
		for i, v := range pca.Values {
			if v >= PCAStrongThr {
				nstr = float64(ln - i)
				break
			}
		}
		var top5, next5 float64
		for i := 0; i < 5; i++ {
			if ln >= 5 {
				top5 += pca.Values[ln-1-i]
			}
			if ln >= 10 {
				next5 += pca.Values[ln-6-i]
			}
		}
		st.SetFloat(lnm+"_PCA_NStrong", nstr)
		st.SetFloat(lnm+"_PCA_Top5", top5/5)
		st.SetFloat(lnm+"_PCA_Next5", next5/5)
		if ln > 10 {
			sum := norm.Sum64(pca.Values)
			ravg := (sum - (top5 + next5)) / float64(ln-10)
			st.SetFloat(lnm+"_PCA_Rest", ravg)
		} else {
			st.SetFloat(lnm+"_PCA_Rest", 0)
		}
	}
}
