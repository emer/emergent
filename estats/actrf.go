// Copyright (c) 2022, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package estats

import (
	"fmt"
	"strings"

	"github.com/emer/emergent/emer"
	"github.com/emer/etable/etensor"
)

// InitActRFs initializes a set of activation-based receptive field (ActRF)
// statistics, which record activation-weighted averaging of other tensor
// states, which can be activations in other layers, or external sensory
// inputs, or any kind of analytic pattern that helps to decode what
// the network is doing.
// The input is a list of colon-separated "Layer:Source" strings,
// where 'Layer' refers to a name of a layer in the given network,
// and 'Source' is either the name of another layer (checked first)
// or the name of a tensor stored in F32Tensors (if layer name not found).
// If Source is not a layer, it must be populated prior to these calls.
func (st *Stats) InitActRFs(net emer.Network, arfs []string, varnm string) error {
	var err error
	for _, anm := range arfs {
		sp := strings.Split(anm, ":")
		lnm := sp[0]
		_, err = net.LayerByNameTry(lnm)
		if err != nil {
			fmt.Printf("estats.InitActRFs: %s\n", err)
			continue
		}

		lvt := st.SetLayerRepTensor(net, lnm, varnm, 0)
		tnm := sp[1]
		var tvt *etensor.Float32
		_, err = net.LayerByNameTry(tnm)
		if err == nil {
			tvt = st.SetLayerRepTensor(net, tnm, varnm, 0)
		} else {
			ok := false
			tvt, ok = st.F32Tensors[tnm]
			if !ok {
				fmt.Printf("estats.InitActRFs: Source tensor named: %s not found\n", tnm)
				continue
			}
		}
		st.ActRFs.AddRF(anm, lvt, tvt)
		// af.NormRF.SetMetaData("min", "0")
	}
	return err
}

// UpdateActRFs updates activation-based receptive fields with
// a new sample of data from current network state, and updated
// state values which must be already updated in F32Tensors.
// Must have called InitActRFs first -- see it for documentation.
// Uses RFs configured then, grabbing network values from variable
// varnm, and given threshold (0.01 recommended)
// di is a data parallel index di, for networks capable of processing input patterns in parallel.
func (st *Stats) UpdateActRFs(net emer.Network, varnm string, thr float32, di int) {
	for _, rf := range st.ActRFs.RFs {
		anm := rf.Name
		sp := strings.Split(anm, ":")
		lnm := sp[0]
		_, err := net.LayerByNameTry(lnm)
		if err != nil {
			continue
		}
		lvt := st.SetLayerRepTensor(net, lnm, varnm, di)
		tnm := sp[1]
		var tvt *etensor.Float32
		_, err = net.LayerByNameTry(tnm)
		if err == nil {
			tvt = st.SetLayerRepTensor(net, tnm, varnm, di)
		} else { // random state
			tvt = st.F32Tensor(tnm)
		}
		st.ActRFs.Add(anm, lvt, tvt, thr)
	}
}

// ActRFsAvgNorm calls Avg() then Norm() on ActRFs -- this is the
// standard way to visualize the RFs
func (st *Stats) ActRFsAvgNorm() {
	st.ActRFs.AvgNorm()
}
