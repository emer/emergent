// Copyright (c) 2019, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package actrf

import (
	"fmt"
	"log"

	"github.com/emer/etable/v2/etensor"
)

// RFs manages multiple named RF's -- each one must be initialized first
// but functions like Avg, Norm, and Reset can be called generically on all.
type RFs struct {

	// map of names to indexes of RFs
	NameMap map[string]int

	// the RFs
	RFs []*RF
}

// RFByName returns RF of given name, nil if not found
func (af *RFs) RFByName(name string) *RF {
	if af.NameMap == nil {
		return nil
	}
	idx, ok := af.NameMap[name]
	if ok {
		return af.RFs[idx]
	}
	return nil
}

// RFByNameTry returns RF of given name, nil and error msg if not found
func (af *RFs) RFByNameTry(name string) (*RF, error) {
	rf := af.RFByName(name)
	if rf == nil {
		return nil, fmt.Errorf("Name: %s not found in list of named RFs", name)
	}
	return rf, nil
}

// AddRF adds a new RF, calling Init on it using given act, src tensors
func (af *RFs) AddRF(name string, act, src etensor.Tensor) *RF {
	if af.NameMap == nil {
		af.NameMap = make(map[string]int)
	}
	sz := len(af.RFs)
	af.NameMap[name] = sz
	rf := &RF{}
	af.RFs = append(af.RFs, rf)
	rf.Init(name, act, src)
	return rf
}

// Add adds a new act sample to the accumulated data for given named rf
func (af *RFs) Add(name string, act, src etensor.Tensor, thr float32) error {
	rf, err := af.RFByNameTry(name)
	if err != nil {
		log.Println(err)
		return err
	}
	rf.Add(act, src, thr)
	return nil
}

// Reset resets Sum accumulations for all rfs
func (af *RFs) Reset() {
	for _, rf := range af.RFs {
		rf.Reset()
	}
}

// Avg computes RF as SumProd / SumTarg.  Does not Reset sums.
func (af *RFs) Avg() {
	for _, rf := range af.RFs {
		rf.Avg()
	}
}

// Norm computes unit norm of RF values -- must be called after Avg
func (af *RFs) Norm() {
	for _, rf := range af.RFs {
		rf.Norm()
	}
}

// AvgNorm computes RF as SumProd / SumTarg and then does Norm.
// This is what you typically want to call before viewing RFs.
// Does not Reset sums.
func (af *RFs) AvgNorm() {
	for _, rf := range af.RFs {
		rf.AvgNorm()
	}
}
