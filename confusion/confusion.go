// Copyright (c) 2019, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package confusion

import (
	"fmt"
	"math"

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

	// [view: no-inline] normalized probability of confusion: Row = ground truth class, Col = actual response for that class.
	Prob etensor.Float64 `view:"no-inline" desc:"normalized probability of confusion: Row = ground truth class, Col = actual response for that class."`

	// [view: no-inline] incremental sums
	Sum etensor.Float64 `view:"no-inline" desc:"incremental sums"`

	// [view: no-inline] counts per ground truth (rows)
	N etensor.Float64 `view:"no-inline" desc:"counts per ground truth (rows)"`

	// [view: no-inline] visualization using SimMat
	Vis simat.SimMat `view:"no-inline" desc:"visualization using SimMat"`

	// [view: no-inline] true pos/neg, false pos/neg for each class, generated from the confusion matrix
	TFPN etensor.Float64 `view:"no-inline" desc:"true pos/neg, false pos/neg for each class, generated from the confusion matrix"`

	// [view: no-inline] precision, recall and F1 score by class
	ClassScores etensor.Float64 `view:"no-inline" desc:"precision, recall and F1 score by class"`

	// [view: no-inline] micro F1, macro F1 and weighted F1 scores for entire matrix ignoring class
	MatrixScores etensor.Float64 `view:"no-inline" desc:"micro F1, macro F1 and weighted F1 scores for entire matrix ignoring class"`
}

var KiT_Matrix = kit.Types.AddType(&Matrix{}, MatrixProps)

// Init initializes the Matrix for given number of classes,
// and resets the data to zero.
func (cm *Matrix) Init(n int) {
	cm.Prob.SetShape([]int{n, n}, nil, []string{"N", "N"})
	cm.Sum.SetShape([]int{n, n}, nil, []string{"N", "N"})
	cm.N.SetShape([]int{n}, nil, []string{"N"})
	cm.TFPN.SetShape([]int{n, 4}, nil, []string{"TP", "FP", "FN", "TN"})
	cm.ClassScores.SetShape([]int{n, 3}, nil, []string{"Precision", "Recall", "F1"})
	cm.MatrixScores.SetShape([]int{3}, nil, []string{"Precision", "Recall", "F1"})
	cm.Vis.Mat = &cm.Prob
	cm.Reset()
}

// Reset resets the data to zero
func (cm *Matrix) Reset() {
	cm.Prob.SetZeros()
	cm.Sum.SetZeros()
	cm.N.SetZeros()
	cm.TFPN.SetZeros()
	cm.ClassScores.SetZeros()
	cm.MatrixScores.SetZeros()
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

// Incr increments the data for given class ground truth and response.
func (cm *Matrix) Incr(class, resp int) {
	if class < 0 || resp < 0 {
		return
	}
	ncat := cm.Sum.Dim(0)
	if class >= ncat || resp >= ncat {
		return
	}
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

func (cm *Matrix) SumTFPN(class int) {
	fn := 0.0 // false negative
	fp := 0.0 // false positive
	tn := 0.0 // true negative

	n := cm.N.Len()
	for c := 0; c < n; c++ {
		for r := 0; r < n; r++ {
			if r == class && c == class { //        True Positive
				v := cm.Sum.FloatValRowCell(r, c)
				cm.TFPN.SetFloatRowCell(class, 0, v)
			} else if r == class && c != class { // False Positive
				fn += cm.Sum.FloatValRowCell(r, c)
				cm.TFPN.SetFloatRowCell(class, 1, fp)
			} else if r != class && c == class { // False Negative
				fp += cm.Sum.FloatValRowCell(r, c)
				cm.TFPN.SetFloatRowCell(class, 2, fn)
			} else { //                             True Negative
				tn += cm.Sum.FloatValRowCell(r, c)
				cm.TFPN.SetFloatRowCell(class, 3, tn)
			}
		}
	}
	cm.TFPN.SetFloatRowCell(class, 1, fp)
	cm.TFPN.SetFloatRowCell(class, 2, fn)
	cm.TFPN.SetFloatRowCell(class, 3, tn)
}

func (cm *Matrix) ScoreClass(class int) {
	tp := cm.TFPN.FloatValRowCell(class, 0)
	fp := cm.TFPN.FloatValRowCell(class, 1)
	fn := cm.TFPN.FloatValRowCell(class, 2)

	precision := tp / (tp + fp)
	cm.ClassScores.SetFloatRowCell(class, 0, precision)
	recall := tp / (tp + fn) // also called true positive rate and has other names
	cm.ClassScores.SetFloatRowCell(class, 1, recall)
	f1 := 2 * tp / ((2 * tp) + fp + fn) // 2 x (Precision x Recall) / (Precision + Recall)
	cm.ClassScores.SetFloatRowCell(class, 2, f1)
}

func (cm *Matrix) ScoreMatrix() {
	tp := 0.0
	fp := 0.0
	fn := 0.0

	n := cm.N.Len()
	for i := 0; i < n; i++ {
		tp += cm.TFPN.FloatValRowCell(i, 0)
		fp += cm.TFPN.FloatValRowCell(i, 1)
		fn += cm.TFPN.FloatValRowCell(i, 2)
	}

	// micro F1 - ignores class
	f1 := 2 * tp / ((2 * tp) + fp + fn) // 2 x (Precision x Recall) / (Precision + Recall)
	cm.MatrixScores.SetFloat1D(0, f1)

	// macro F1 - unweighted average of class F1 scores
	// some classes might not have any instances so check NaN
	f1 = 0.0
	for i := 0; i < n; i++ {
		classf1 := cm.ClassScores.FloatValRowCell(i, 2)
		if math.IsNaN(classf1) == false {
			f1 += classf1
		}
	}
	cm.MatrixScores.SetFloat1D(1, f1/float64(n))

	// weighted F1 - weighted average of class F1 scores
	// some classes might not have any instances so check NaN
	f1 = 0.0
	totalN := 0.0
	for i := 0; i < n; i++ {
		classf1 := cm.ClassScores.FloatValRowCell(i, 2) * cm.N.FloatVal1D(i)
		if math.IsNaN(classf1) == false {
			f1 += classf1
		}
		totalN += cm.N.FloatVal1D(i)
	}
	cm.MatrixScores.SetFloat1D(2, f1/totalN)
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
