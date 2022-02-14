// Copyright (c) 2022, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package estats

import (
	"github.com/emer/emergent/emer"
	"github.com/emer/etable/etensor"
)

// ConfigRasters configures spike rasters
func (st *Stats) ConfigRasters(net emer.Network, layers []string) {
	ncy := 200 // max cycles
	// spike rast
	for _, lnm := range layers {
		ly := net.LayerByName(lnm)
		sr := st.F32Tensor("Raster_" + lnm)
		sr.SetShape([]int{ly.Shape().Len(), ncy}, nil, []string{"Nrn", "Cyc"})
	}
}

// SetRasterCol sets column of given raster from data
func (st *Stats) SetRasterCol(sr, tsr *etensor.Float32, col int) {
	for ni, v := range tsr.Values {
		sr.Set([]int{ni, col}, v)
	}
}

// RasterRec records data from given layers, variable name to raster
// for given cycle number (X axis index)
func (st *Stats) RasterRec(net emer.Network, cyc int, varNm string, layers []string) {
	for _, lnm := range layers {
		tsr := st.SetLayerTensor(net, lnm, varNm)
		sr := st.F32Tensor("Raster_" + lnm)
		st.SetRasterCol(sr, tsr, cyc)
	}
}
