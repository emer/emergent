// Copyright (c) 2022, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package estats

import (
	"cogentcore.org/core/base/errors"
	"cogentcore.org/core/tensor"
	"github.com/emer/emergent/v2/emer"
)

// ConfigRasters configures spike rasters for given maximum number of cycles
// and layer names.
func (st *Stats) ConfigRasters(net emer.Network, maxCyc int, layers []string) {
	st.Rasters = layers
	for _, lnm := range st.Rasters {
		ly := errors.Log1(net.AsEmer().EmerLayerByName(lnm)).AsEmer()
		sr := st.F32Tensor("Raster_" + lnm)
		nu := len(ly.SampleIndexes)
		if nu == 0 {
			nu = ly.Shape.Len()
		}
		sr.SetShapeSizes(nu, maxCyc)
	}
}

// SetRasterCol sets column of given raster from data
func (st *Stats) SetRasterCol(sr, tsr *tensor.Float32, col int) {
	for ni, v := range tsr.Values {
		sr.Set(v, ni, col)
	}
}

// RasterRec records data from layers configured with ConfigRasters
// using variable name, for given cycle number (X axis index)
// di is a data parallel index di, for networks capable of processing input patterns in parallel.
func (st *Stats) RasterRec(net emer.Network, cyc int, varNm string, di int) {
	for _, lnm := range st.Rasters {
		tsr := st.SetLayerSampleTensor(net, lnm, varNm, di)
		sr := st.F32Tensor("Raster_" + lnm)
		if sr.DimSize(1) <= cyc {
			continue
		}
		st.SetRasterCol(sr, tsr, cyc)
	}
}
