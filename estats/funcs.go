// Copyright (c) 2022, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package estats

import (
	"cogentcore.org/core/base/errors"
	"github.com/emer/emergent/v2/emer"
	"github.com/emer/etensor/tensor"
	"github.com/emer/etensor/tensor/stats/metric"
	"github.com/emer/etensor/tensor/stats/stats"
	"github.com/emer/etensor/tensor/table"
)

// funcs contains misc stats functions

// SetLayerTensor sets tensor of Unit values on a layer for given variable
// to a F32Tensor with name = layNm
// di is a data parallel index di, for networks capable of processing input patterns in parallel.
func (st *Stats) SetLayerTensor(net emer.Network, layNm, unitVar string, di int) *tensor.Float32 {
	ly := errors.Log1(net.AsEmer().EmerLayerByName(layNm)).AsEmer()
	tsr := st.F32TensorDi(layNm, di)
	ly.UnitValuesTensor(tsr, unitVar, di)
	return tsr
}

// SetLayerSampleTensor sets tensor of representative Unit values on a layer
// for given variable to a F32Tensor with name = layNm
// di is a data parallel index di, for networks capable of processing input patterns in parallel.
func (st *Stats) SetLayerSampleTensor(net emer.Network, layNm, unitVar string, di int) *tensor.Float32 {
	ly := errors.Log1(net.AsEmer().EmerLayerByName(layNm)).AsEmer()
	tsr := st.F32TensorDi(layNm, di)
	ly.UnitValuesSampleTensor(tsr, unitVar, di)
	return tsr
}

// LayerVarsCorrel returns the correlation between two variables on a given layer
// di is a data parallel index di, for networks capable of processing input patterns in parallel.
func (st *Stats) LayerVarsCorrel(net emer.Network, layNm, unitVarA, unitVarB string, di int) float32 {
	ly := errors.Log1(net.AsEmer().EmerLayerByName(layNm)).AsEmer()
	tsrA := st.F32TensorDi(layNm, di) // standard re-used storage tensor
	ly.UnitValuesTensor(tsrA, unitVarA, di)
	tsrB := st.F32TensorDi(layNm+"_alt", di) // alternative storage tensor
	ly.UnitValuesTensor(tsrB, unitVarB, di)
	return metric.Correlation32(tsrA.Values, tsrB.Values)
}

// LayerVarsCorrelRep returns the correlation between two variables on a given layer
// Rep version uses representative units.
// di is a data parallel index di, for networks capable of processing input patterns in parallel.
func (st *Stats) LayerVarsCorrelRep(net emer.Network, layNm, unitVarA, unitVarB string, di int) float32 {
	ly := errors.Log1(net.AsEmer().EmerLayerByName(layNm)).AsEmer()
	tsrA := st.F32TensorDi(layNm, di) // standard re-used storage tensor
	ly.UnitValuesSampleTensor(tsrA, unitVarA, di)
	tsrB := st.F32TensorDi(layNm+"_alt", di) // alternative storage tensor
	ly.UnitValuesSampleTensor(tsrB, unitVarB, di)
	return metric.Correlation32(tsrA.Values, tsrB.Values)
}

// ClosestStat finds the closest pattern in given column of given table of possible patterns,
// compared to layer activation pattern using given variable.  Returns the row number,
// correlation value, and value of a column named namecol for that row if non-empty.
// Column must be tensor.Float32
// di is a data parallel index di, for networks capable of processing input patterns in parallel.
func (st *Stats) ClosestPat(net emer.Network, layNm, unitVar string, di int, pats *table.Table, colnm, namecol string) (int, float32, string) {
	tsr := st.SetLayerTensor(net, layNm, unitVar, di)
	col := errors.Log1(pats.ColumnByName(colnm))
	// note: requires Increasing metric so using Inv
	row, cor := metric.ClosestRow32(tsr, col.(*tensor.Float32), metric.InvCorrelation32)
	cor = 1 - cor // convert back to correl
	nm := ""
	if namecol != "" {
		nm = pats.StringValue(namecol, row)
	}
	return row, cor, nm
}

//////////////////////////////////////////////
//  PCA Stats

// PCAStrongThr is the threshold for counting PCA eigenvalues as "strong"
// Applies to SVD as well.
var PCAStrongThr = 0.01

// PCAStats computes PCA statistics on recorded hidden activation patterns
// on given log table (IndexView), and given list of layer names
// and variable name -- columns named "layer_var".
// Helpful for measuring the overall information (variance) in the representations
// to detect a common failure mode where a few patterns dominate over everything ("hogs").
// Records Float stats as:
// layer_PCA_NStrong: number of eigenvalues above the PCAStrongThr threshold
// layer_PCA_Top5: average strength of top 5 eigenvalues
// layer_PCA_Next5: average strength of next 5 eigenvalues
// layer_PCA_Rest: average strength of remaining eigenvalues (if more than 10 total eigens)
// Uses SVD to compute much more efficiently than official PCA.
func (st *Stats) PCAStats(ix *table.IndexView, varNm string, layers []string) {
	svd := &st.SVD
	svd.Cond = PCAStrongThr
	for _, lnm := range layers {
		svd.TableColumn(ix, lnm+"_"+varNm, metric.Covariance64)
		ln := len(svd.Values)
		var nstr float64 // nstr := float64(svd.Rank)  this didn't work..
		for i, v := range svd.Values {
			if v < PCAStrongThr {
				nstr = float64(i)
				break
			}
		}
		var top5, next5 float64
		for i := 0; i < 5; i++ {
			if ln >= 5 {
				top5 += svd.Values[i]
			}
			if ln >= 10 {
				next5 += svd.Values[i+5]
			}
		}
		st.SetFloat(lnm+"_PCA_NStrong", nstr)
		st.SetFloat(lnm+"_PCA_Top5", top5/5)
		st.SetFloat(lnm+"_PCA_Next5", next5/5)
		if ln > 10 {
			sum := stats.Sum64(svd.Values)
			ravg := (sum - (top5 + next5)) / float64(ln-10)
			st.SetFloat(lnm+"_PCA_Rest", ravg)
		} else {
			st.SetFloat(lnm+"_PCA_Rest", 0)
		}
	}
}
