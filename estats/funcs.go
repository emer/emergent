// Copyright (c) 2022, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package estats

/*

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
	return float32(metric.Correlation(tsrA, tsrB).Float1D(0))
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
	return float32(metric.Correlation(tsrA, tsrB).Float1D(0))
}

// ClosestStat finds the closest pattern in given column of given table of possible patterns,
// compared to layer activation pattern using given variable.  Returns the row number,
// correlation value, and value of a column named namecol for that row if non-empty.
// Column must be tensor.Float32
// di is a data parallel index di, for networks capable of processing input patterns in parallel.
func (st *Stats) ClosestPat(net emer.Network, layNm, unitVar string, di int, pats *table.Table, colnm, namecol string) (int, float32, string) {
	tsr := st.SetLayerTensor(net, layNm, unitVar, di)
	col := pats.Column(colnm)
	// note: requires Increasing metric so using Inv
	rc := metric.ClosestRow(metric.InvCorrelation, tsr, col)
	row := rc.Int1D(0)
	cor := 1 - float32(rc.Float1D(1)) // convert back to correl
	nm := ""
	if namecol != "" {
		nm = pats.Column(namecol).String1D(row)
	}
	return row, cor, nm
}
*/
