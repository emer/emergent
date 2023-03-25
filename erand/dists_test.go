// Copyright (c) 2023, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package erand

import (
	"math"
	"testing"

	"github.com/emer/etable/agg"
	"github.com/emer/etable/etable"
	"github.com/emer/etable/etensor"
)

func TestGaussianGen(t *testing.T) {
	nsamp := int(1e6)
	sch := etable.Schema{
		{"Val", etensor.FLOAT32, nil, nil},
	}
	dt := &etable.Table{}
	dt.SetFromSchema(sch, nsamp)

	mean := 0.5
	sig := 0.25
	tol := 1e-2

	for i := 0; i < nsamp; i++ {
		vl := GaussianGen(mean, sig, -1)
		dt.SetCellFloat("Val", i, vl)
	}
	ix := etable.NewIdxView(dt)
	desc := agg.DescAll(ix)

	meanRow := desc.RowsByString("Agg", "Mean", etable.Equals, etable.UseCase)[0]
	stdRow := desc.RowsByString("Agg", "Std", etable.Equals, etable.UseCase)[0]
	// minRow := desc.RowsByString("Agg", "Min", etable.Equals, etable.UseCase)[0]
	// maxRow := desc.RowsByString("Agg", "Max", etable.Equals, etable.UseCase)[0]

	actMean := desc.CellFloat("Val", meanRow)
	actStd := desc.CellFloat("Val", stdRow)

	if math.Abs(actMean-mean) > tol {
		t.Errorf("Gaussian: mean %g\t out of tolerance vs target: %g\n", actMean, mean)
	}
	if math.Abs(actStd-sig) > tol {
		t.Errorf("Gaussian: stdev %g\t out of tolerance vs target: %g\n", actStd, sig)
	}
	// b := bytes.NewBuffer(nil)
	// desc.WriteCSV(b, etable.Tab, etable.Headers)
	// fmt.Printf("%s\n", string(b.Bytes()))
}

func TestBinomialGen(t *testing.T) {
	nsamp := int(1e6)
	sch := etable.Schema{
		{"Val", etensor.FLOAT32, nil, nil},
	}
	dt := &etable.Table{}
	dt.SetFromSchema(sch, nsamp)

	n := 1.0
	p := 0.5
	tol := 1e-2

	for i := 0; i < nsamp; i++ {
		vl := BinomialGen(n, p, -1)
		dt.SetCellFloat("Val", i, vl)
	}
	ix := etable.NewIdxView(dt)
	desc := agg.DescAll(ix)

	meanRow := desc.RowsByString("Agg", "Mean", etable.Equals, etable.UseCase)[0]
	stdRow := desc.RowsByString("Agg", "Std", etable.Equals, etable.UseCase)[0]
	minRow := desc.RowsByString("Agg", "Min", etable.Equals, etable.UseCase)[0]
	maxRow := desc.RowsByString("Agg", "Max", etable.Equals, etable.UseCase)[0]

	actMean := desc.CellFloat("Val", meanRow)
	actStd := desc.CellFloat("Val", stdRow)
	actMin := desc.CellFloat("Val", minRow)
	actMax := desc.CellFloat("Val", maxRow)

	mean := n * p
	if math.Abs(actMean-mean) > tol {
		t.Errorf("Binomial: mean %g\t out of tolerance vs target: %g\n", actMean, mean)
	}
	sig := math.Sqrt(n * p * (1.0 - p))
	if math.Abs(actStd-sig) > tol {
		t.Errorf("Binomial: stdev %g\t out of tolerance vs target: %g\n", actStd, sig)
	}
	if actMin < 0 {
		t.Errorf("Binomial: min %g\t should not be < 0\n", actMin)
	}
	if actMax < 0 {
		t.Errorf("Binomial: max %g\t should not be > 1\n", actMax)
	}
	// b := bytes.NewBuffer(nil)
	// desc.WriteCSV(b, etable.Tab, etable.Headers)
	// fmt.Printf("%s\n", string(b.Bytes()))
}

func TestPoissonGen(t *testing.T) {
	nsamp := int(1e6)
	sch := etable.Schema{
		{"Val", etensor.FLOAT32, nil, nil},
	}
	dt := &etable.Table{}
	dt.SetFromSchema(sch, nsamp)

	lambda := 10.0
	tol := 1e-2

	for i := 0; i < nsamp; i++ {
		vl := PoissonGen(lambda, -1)
		dt.SetCellFloat("Val", i, vl)
	}
	ix := etable.NewIdxView(dt)
	desc := agg.DescAll(ix)

	meanRow := desc.RowsByString("Agg", "Mean", etable.Equals, etable.UseCase)[0]
	stdRow := desc.RowsByString("Agg", "Std", etable.Equals, etable.UseCase)[0]
	minRow := desc.RowsByString("Agg", "Min", etable.Equals, etable.UseCase)[0]
	// maxRow := desc.RowsByString("Agg", "Max", etable.Equals, etable.UseCase)[0]

	actMean := desc.CellFloat("Val", meanRow)
	actStd := desc.CellFloat("Val", stdRow)
	actMin := desc.CellFloat("Val", minRow)
	// actMax := desc.CellFloat("Val", maxRow)

	mean := lambda
	if math.Abs(actMean-mean) > tol {
		t.Errorf("Poisson: mean %g\t out of tolerance vs target: %g\n", actMean, mean)
	}
	sig := math.Sqrt(lambda)
	if math.Abs(actStd-sig) > tol {
		t.Errorf("Poisson: stdev %g\t out of tolerance vs target: %g\n", actStd, sig)
	}
	if actMin < 0 {
		t.Errorf("Poisson: min %g\t should not be < 0\n", actMin)
	}
	// if actMax < 0 {
	// 	t.Errorf("Poisson: max %g\t should not be > 1\n", actMax)
	// }
	// b := bytes.NewBuffer(nil)
	// desc.WriteCSV(b, etable.Tab, etable.Headers)
	// fmt.Printf("%s\n", string(b.Bytes()))
}

func TestGammaGen(t *testing.T) {
	nsamp := int(1e6)
	sch := etable.Schema{
		{"Val", etensor.FLOAT32, nil, nil},
	}
	dt := &etable.Table{}
	dt.SetFromSchema(sch, nsamp)

	alpha := 0.5
	beta := 0.8
	tol := 1e-2

	for i := 0; i < nsamp; i++ {
		vl := GammaGen(alpha, beta, -1)
		dt.SetCellFloat("Val", i, vl)
	}
	ix := etable.NewIdxView(dt)
	desc := agg.DescAll(ix)

	meanRow := desc.RowsByString("Agg", "Mean", etable.Equals, etable.UseCase)[0]
	stdRow := desc.RowsByString("Agg", "Std", etable.Equals, etable.UseCase)[0]

	actMean := desc.CellFloat("Val", meanRow)
	actStd := desc.CellFloat("Val", stdRow)

	mean := alpha / beta
	if math.Abs(actMean-mean) > tol {
		t.Errorf("Gamma: mean %g\t out of tolerance vs target: %g\n", actMean, mean)
	}
	sig := math.Sqrt(alpha / beta / beta)
	if math.Abs(actStd-sig) > tol {
		t.Errorf("Gamma: stdev %g\t out of tolerance vs target: %g\n", actStd, sig)
	}
	// b := bytes.NewBuffer(nil)
	// desc.WriteCSV(b, etable.Tab, etable.Headers)
	// fmt.Printf("%s\n", string(b.Bytes()))
}

func TestBetaGen(t *testing.T) {
	nsamp := int(1e6)
	sch := etable.Schema{
		{"Val", etensor.FLOAT32, nil, nil},
	}
	dt := &etable.Table{}
	dt.SetFromSchema(sch, nsamp)

	alpha := 0.5
	beta := 0.8
	tol := 1e-2

	for i := 0; i < nsamp; i++ {
		vl := BetaGen(alpha, beta, -1)
		dt.SetCellFloat("Val", i, vl)
	}
	ix := etable.NewIdxView(dt)
	desc := agg.DescAll(ix)

	meanRow := desc.RowsByString("Agg", "Mean", etable.Equals, etable.UseCase)[0]
	stdRow := desc.RowsByString("Agg", "Std", etable.Equals, etable.UseCase)[0]

	actMean := desc.CellFloat("Val", meanRow)
	actStd := desc.CellFloat("Val", stdRow)

	mean := alpha / (alpha + beta)
	if math.Abs(actMean-mean) > tol {
		t.Errorf("Beta: mean %g\t out of tolerance vs target: %g\n", actMean, mean)
	}
	vr := alpha * beta / ((alpha + beta) * (alpha + beta) * (alpha + beta + 1))
	sig := math.Sqrt(vr)
	if math.Abs(actStd-sig) > tol {
		t.Errorf("Beta: stdev %g\t out of tolerance vs target: %g\n", actStd, sig)
	}
	// b := bytes.NewBuffer(nil)
	// desc.WriteCSV(b, etable.Tab, etable.Headers)
	// fmt.Printf("%s\n", string(b.Bytes()))
}
