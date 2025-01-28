// Copyright (c) 2019, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package confusion

//go:generate core generate -add-types

import (
	"fmt"
	"math"

	"cogentcore.org/core/core"
	"github.com/emer/etensor/tensor"
	"github.com/emer/etensor/tensor/stats/simat"
)

// Matrix computes the confusion matrix, with rows representing
// the ground truth correct class, and columns representing the
// actual answer produced.  Correct answers are along the diagonal.
type Matrix struct { //git:add

	// normalized probability of confusion: Row = ground truth class, Col = actual response for that class.
	Prob tensor.Float64 `display:"no-inline"`

	// incremental sums
	Sum tensor.Float64 `display:"no-inline"`

	// counts per ground truth (rows)
	N tensor.Float64 `display:"no-inline"`

	// visualization using SimMat
	Vis simat.SimMat `display:"no-inline"`

	// true pos/neg, false pos/neg for each class, generated from the confusion matrix
	TFPN tensor.Float64 `display:"no-inline"`

	// precision, recall and F1 score by class
	ClassScores tensor.Float64 `display:"no-inline"`

	// micro F1, macro F1 and weighted F1 scores for entire matrix ignoring class
	MatrixScores tensor.Float64 `display:"no-inline"`
}

// Init initializes the Matrix for given number of classes,
// and resets the data to zero.
func (cm *Matrix) Init(n int) {
	cm.Prob.SetShape([]int{n, n}, "N", "N")
	cm.Sum.SetShape([]int{n, n}, "N", "N")
	cm.N.SetShape([]int{n}, "N")
	cm.TFPN.SetShape([]int{n, 4}, "TP", "FP", "FN", "TN")
	cm.ClassScores.SetShape([]int{n, 3}, "Precision", "Recall", "F1")
	cm.MatrixScores.SetShape([]int{3}, "Precision", "Recall", "F1")
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
	cm.Vis.Columns = lbls
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
	ncat := cm.Sum.DimSize(0)
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
				v := cm.Sum.FloatRowCell(r, c)
				cm.TFPN.SetFloatRowCell(class, 0, v)
			} else if r == class && c != class { // False Positive
				fn += cm.Sum.FloatRowCell(r, c)
				cm.TFPN.SetFloatRowCell(class, 1, fp)
			} else if r != class && c == class { // False Negative
				fp += cm.Sum.FloatRowCell(r, c)
				cm.TFPN.SetFloatRowCell(class, 2, fn)
			} else { //                             True Negative
				tn += cm.Sum.FloatRowCell(r, c)
				cm.TFPN.SetFloatRowCell(class, 3, tn)
			}
		}
	}
	cm.TFPN.SetFloatRowCell(class, 1, fp)
	cm.TFPN.SetFloatRowCell(class, 2, fn)
	cm.TFPN.SetFloatRowCell(class, 3, tn)
}

func (cm *Matrix) ScoreClass(class int) {
	tp := cm.TFPN.FloatRowCell(class, 0)
	fp := cm.TFPN.FloatRowCell(class, 1)
	fn := cm.TFPN.FloatRowCell(class, 2)

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
		tp += cm.TFPN.FloatRowCell(i, 0)
		fp += cm.TFPN.FloatRowCell(i, 1)
		fn += cm.TFPN.FloatRowCell(i, 2)
	}

	// micro F1 - ignores class
	f1 := 2 * tp / ((2 * tp) + fp + fn) // 2 x (Precision x Recall) / (Precision + Recall)
	cm.MatrixScores.SetFloat1D(0, f1)

	// macro F1 - unweighted average of class F1 scores
	// some classes might not have any instances so check NaN
	f1 = 0.0
	for i := 0; i < n; i++ {
		classf1 := cm.ClassScores.FloatRowCell(i, 2)
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
		classf1 := cm.ClassScores.FloatRowCell(i, 2) * cm.N.Float1D(i)
		if math.IsNaN(classf1) == false {
			f1 += classf1
		}
		totalN += cm.N.Float1D(i)
	}
	cm.MatrixScores.SetFloat1D(2, f1/totalN)
}

// SaveCSV saves Prob result to a CSV file, comma separated
func (cm *Matrix) SaveCSV(fname core.Filename) {
	tensor.SaveCSV(&cm.Prob, fname, ',')
}

// OpenCSV opens Prob result from a CSV file, comma separated
func (cm *Matrix) OpenCSV(fname core.Filename) {
	tensor.OpenCSV(&cm.Prob, fname, ',')
}

/*
var MatrixProps = tree.Props{
	"ToolBar": tree.PropSlice{
		{"SaveCSV", tree.Props{
			"label": "Save CSV...",
			"icon":  "file-save",
			"desc":  "Save CSV-formatted confusion probabilities (Probs)",
			"Args": tree.PropSlice{
				{"CSV File Name", tree.Props{
					"ext": ".csv",
				}},
			},
		}},
		{"OpenCSV", tree.Props{
			"label": "Open CSV...",
			"icon":  "file-open",
			"desc":  "Open CSV-formatted confusion probabilities (Probs)",
			"Args": tree.PropSlice{
				{"Weights File Name", tree.Props{
					"ext": ".csv",
				}},
			},
		}},
	},
}
*/
