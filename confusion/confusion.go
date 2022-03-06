// Copyright (c) 2019, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package confusion

import (
	"fmt"

	"github.com/emer/etable/etensor"
	"github.com/emer/etable/simat"
	"github.com/goki/gi/gi"
	"github.com/goki/ki/ki"
	"github.com/goki/ki/kit"
)

// Matrix computes the confusion matrix, with rows representing
// the ground truth correct class, and columns representing the
// actual answer produced.  Correct answers are along the diagonal.
type Matrix struct {
	Prob etensor.Float64 `view:"no-inline" desc:"normalized probability of confusion: Row = ground truth class, Col = actual response for that class."`
	Sum  etensor.Float64 `view:"no-inline" desc:"incremental sums"`
	N    etensor.Float64 `view:"no-inline" desc:"counts per ground truth (rows)"`
	Vis  simat.SimMat    `view:"no-inline" desc:"visualization using SimMat"`
}

var KiT_Matrix = kit.Types.AddType(&Matrix{}, MatrixProps)

// Init initializes the Matrix for given number of classes,
// and resets the data to zero.
func (cm *Matrix) Init(n int) {
	cm.Prob.SetShape([]int{n, n}, nil, []string{"N", "N"})
	cm.Sum.SetShape([]int{n, n}, nil, []string{"N", "N"})
	cm.N.SetShape([]int{n}, nil, []string{"N"})
	cm.Vis.Mat = &cm.Prob
	cm.Reset()
}

// Reset resets the data to zero
func (cm *Matrix) Reset() {
	cm.Prob.SetZeros()
	cm.Sum.SetZeros()
	cm.N.SetZeros()
}

// SetLabels sets the class labels, for visualization in Vis
func (cm *Matrix) SetLabels(lbls []string) {
	cm.Vis.Rows = lbls
	cm.Vis.Cols = lbls
}

// InitFromLabels does initialization based on given labels.
// Calls Init on len(lbls) and SetLabels.
// Default fontSize = 12 if 0 or -1 passed
func (cm *Matrix) InitFromLabels(lbls []string, fontSize int) {
	cm.Init(len(lbls))
	cm.SetLabels(lbls)
	if fontSize <= 0 {
		fontSize = 12
	}
	cm.Prob.SetMetaData("font-size", fmt.Sprintf("%d", fontSize))
}

// Incr increments the data for given class ground truth
// and response.
func (cm *Matrix) Incr(class, resp int) {
	ix := []int{class, resp}
	sum := cm.Sum.Value(ix)
	sum++
	cm.Sum.Set(ix, sum)
	n := cm.N.Value1D(class)
	n++
	cm.N.Set1D(class, n)
}

// Probs computes the probabilities based on accumulated data
func (cm *Matrix) Probs() {
	n := cm.N.Len()
	for cl := 0; cl < n; cl++ {
		cn := cm.N.Value1D(cl)
		if cn == 0 {
			continue
		}
		for ri := 0; ri < n; ri++ {
			ix := []int{cl, ri}
			sum := cm.Sum.Value(ix)
			cm.Prob.Set(ix, sum/cn)
		}
	}
}

// SaveCSV saves Prob result to a CSV file, comma separated
func (cm *Matrix) SaveCSV(filename gi.FileName) {
	etensor.SaveCSV(&cm.Prob, filename, ',')
}

// OpenCSV opens Prob result from a CSV file, comma separated
func (cm *Matrix) OpenCSV(filename gi.FileName) {
	etensor.OpenCSV(&cm.Prob, filename, ',')
}

var MatrixProps = ki.Props{
	"ToolBar": ki.PropSlice{
		{"SaveCSV", ki.Props{
			"label": "Save CSV...",
			"icon":  "file-save",
			"desc":  "Save CSV-formatted confusion probabilities (Probs)",
			"Args": ki.PropSlice{
				{"CSV File Name", ki.Props{
					"ext": ".csv",
				}},
			},
		}},
		{"OpenCSV", ki.Props{
			"label": "Open CSV...",
			"icon":  "file-open",
			"desc":  "Open CSV-formatted confusion probabilities (Probs)",
			"Args": ki.PropSlice{
				{"Weights File Name", ki.Props{
					"ext": ".csv",
				}},
			},
		}},
	},
}
