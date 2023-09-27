// Copyright (c) 2021, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package decoder

import (
	"bufio"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"os"
	"path/filepath"
	"sort"

	"github.com/emer/emergent/emer"
	"github.com/emer/empi/mpi"
	"github.com/emer/etable/etensor"
	"github.com/goki/mat32"
)

// SoftMax is a softmax decoder, which is the best choice for a 1-hot classification
// using the widely-used SoftMax function: https://en.wikipedia.org/wiki/Softmax_function
type SoftMax struct {

	// [def: 0.1] learning rate
	Lrate float32 `def:"0.1" desc:"learning rate"`

	// layers to decode
	Layers []emer.Layer `desc:"layers to decode"`

	// number of different categories to decode
	NCats int `desc:"number of different categories to decode"`

	// unit values
	Units []SoftMaxUnit `desc:"unit values"`

	// sorted list of indexes into Units, in descending order from strongest to weakest -- i.e., Sorted[0] has the most likely categorization, and its activity is Units[Sorted[0]].Act
	Sorted []int `desc:"sorted list of indexes into Units, in descending order from strongest to weakest -- i.e., Sorted[0] has the most likely categorization, and its activity is Units[Sorted[0]].Act"`

	// number of inputs -- total sizes of layer inputs
	NInputs int `desc:"number of inputs -- total sizes of layer inputs"`

	// input values, copied from layers
	Inputs []float32 `desc:"input values, copied from layers"`

	// current target index of correct category
	Target int `desc:"current target index of correct category"`

	// [view: -] for holding layer values
	ValsTsrs map[string]*etensor.Float32 `view:"-" desc:"for holding layer values"`

	// synaptic weights: outer loop is units, inner loop is inputs
	Weights etensor.Float32 `desc:"synaptic weights: outer loop is units, inner loop is inputs"`

	// [view: -] mpi communicator -- MPI users must set this to their comm -- do direct assignment
	Comm *mpi.Comm `view:"-" desc:"mpi communicator -- MPI users must set this to their comm -- do direct assignment"`

	// delta weight changes: only for MPI mode -- outer loop is units, inner loop is inputs
	MPIDWts etensor.Float32 `desc:"delta weight changes: only for MPI mode -- outer loop is units, inner loop is inputs"`
}

// SoftMaxUnit has variables for softmax decoder unit
type SoftMaxUnit struct {

	// final activation = e^Ge / sum e^Ge
	Act float32 `desc:"final activation = e^Ge / sum e^Ge"`

	// net input = sum x * w
	Net float32 `desc:"net input = sum x * w"`

	// exp(Net)
	Exp float32 `desc:"exp(Net)"`
}

// InitLayer initializes detector with number of categories and layers
func (sm *SoftMax) InitLayer(ncats int, layers []emer.Layer) {
	sm.Layers = layers
	nin := 0
	for _, ly := range sm.Layers {
		nin += ly.Shape().Len()
	}
	sm.Init(ncats, nin)
}

// Init initializes detector with number of categories and number of inputs
func (sm *SoftMax) Init(ncats, ninputs int) {
	sm.NInputs = ninputs
	sm.Lrate = 0.1 // seems pretty good
	sm.NCats = ncats
	sm.Units = make([]SoftMaxUnit, ncats)
	sm.Sorted = make([]int, ncats)
	sm.Inputs = make([]float32, sm.NInputs)
	sm.Weights.SetShape([]int{sm.NCats, sm.NInputs}, nil, []string{"Cats", "Inputs"})
	for i := range sm.Weights.Values {
		sm.Weights.Values[i] = .1
	}
}

// Decode decodes the given variable name from layers (forward pass)
// See Sorted list of indexes for the decoding output -- i.e., Sorted[0]
// is the most likely -- that is returned here as a convenience.
// di is a data parallel index di, for networks capable
// of processing input patterns in parallel.
func (sm *SoftMax) Decode(varNm string, di int) int {
	sm.Input(varNm, di)
	sm.Forward()
	sm.Sort()
	return sm.Sorted[0]
}

// Train trains the decoder with given target correct answer (0..NCats-1)
func (sm *SoftMax) Train(targ int) {
	sm.Target = targ
	sm.Back()
}

// TrainMPI trains the decoder with given target correct answer (0..NCats-1)
// MPI version uses mpi to synchronize weight changes across parallel nodes.
func (sm *SoftMax) TrainMPI(targ int) {
	sm.Target = targ
	sm.BackMPI()
}

// ValsTsr gets value tensor of given name, creating if not yet made
func (sm *SoftMax) ValsTsr(name string) *etensor.Float32 {
	if sm.ValsTsrs == nil {
		sm.ValsTsrs = make(map[string]*etensor.Float32)
	}
	tsr, ok := sm.ValsTsrs[name]
	if !ok {
		tsr = &etensor.Float32{}
		sm.ValsTsrs[name] = tsr
	}
	return tsr
}

// Input grabs the input from given variable in layers
// di is a data parallel index di, for networks capable
// of processing input patterns in parallel.
func (sm *SoftMax) Input(varNm string, di int) {
	off := 0
	for _, ly := range sm.Layers {
		tsr := sm.ValsTsr(ly.Name())
		ly.UnitValsTensor(tsr, varNm, di)
		for j, v := range tsr.Values {
			sm.Inputs[off+j] = v
		}
		off += ly.Shape().Len()
	}
}

// Forward compute the forward pass from input
func (sm *SoftMax) Forward() {
	max := float32(-math.MaxFloat32)
	for ui := range sm.Units {
		u := &sm.Units[ui]
		net := float32(0)
		off := ui * sm.NInputs
		for j, in := range sm.Inputs {
			net += sm.Weights.Values[off+j] * in
		}
		u.Net = net
		if net > max {
			max = net
		}
	}
	sum := float32(0)
	for ui := range sm.Units {
		u := &sm.Units[ui]
		u.Net -= max
		u.Exp = mat32.FastExp(u.Net)
		sum += u.Exp
	}
	for ui := range sm.Units {
		u := &sm.Units[ui]
		u.Act = u.Exp / sum
	}
}

// Sort updates Sorted indexes of the current Unit category activations sorted
// from highest to lowest.  i.e., the 0-index value has the strongest
// decoded output category, 1 the next-strongest, etc.
func (sm *SoftMax) Sort() {
	for i := range sm.Sorted {
		sm.Sorted[i] = i
	}
	sort.Slice(sm.Sorted, func(i, j int) bool {
		return sm.Units[sm.Sorted[i]].Act > sm.Units[sm.Sorted[j]].Act
	})
}

// Back compute the backward error propagation pass
func (sm *SoftMax) Back() {
	lr := sm.Lrate
	for ui := range sm.Units {
		u := &sm.Units[ui]
		var del float32
		if ui == sm.Target {
			del = lr * (1 - u.Act)
		} else {
			del = -lr * u.Act
		}
		off := ui * sm.NInputs
		for j, in := range sm.Inputs {
			sm.Weights.Values[off+j] += del * in
		}
	}
}

// BackMPI compute the backward error propagation pass
// MPI version shares weight changes across nodes
func (sm *SoftMax) BackMPI() {
	if sm.MPIDWts.Len() != sm.Weights.Len() {
		sm.MPIDWts.CopyShapeFrom(&sm.Weights)
	}
	lr := sm.Lrate
	for ui := range sm.Units {
		u := &sm.Units[ui]
		var del float32
		if ui == sm.Target {
			del = lr * (1 - u.Act)
		} else {
			del = -lr * u.Act
		}
		off := ui * sm.NInputs
		for j, in := range sm.Inputs {
			sm.MPIDWts.Values[off+j] = del * in
		}
	}

	sm.Comm.AllReduceF32(mpi.OpSum, sm.MPIDWts.Values, nil)

	for i, dw := range sm.MPIDWts.Values {
		sm.Weights.Values[i] += dw
	}
}

type softMaxForSerialization struct {
	Weights []float32 `json:"weights"`
}

// Save saves the decoder weights to given file path.
// If path ends in .gz, it will be gzipped.
func (sm *SoftMax) Save(path string) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()
	ext := filepath.Ext(path)
	var writer io.Writer
	if ext == ".gz" {
		gw := gzip.NewWriter(file)
		defer gw.Close()
		writer = gw
	} else {
		bw := bufio.NewWriter(file)
		defer bw.Flush()
		writer = bw
	}
	encoder := json.NewEncoder(writer)
	return encoder.Encode(softMaxForSerialization{Weights: sm.Weights.Values})
}

// Load loads the decoder weights from given file path.
// If the shape of the decoder does not match the shape of the saved weights,
// an error will be returned.
func (sm *SoftMax) Load(path string) error {
	ext := filepath.Ext(path)
	var reader io.Reader
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()
	if ext == ".gz" {
		gr, err := gzip.NewReader(file)
		if err != nil {
			return err
		}
		defer gr.Close()
		reader = gr
	} else {
		reader = bufio.NewReader(file)
	}
	decoder := json.NewDecoder(reader)
	var s softMaxForSerialization
	if err := decoder.Decode(&s); err != nil {
		return err
	}
	if len(sm.Weights.Values) != len(s.Weights) {
		return fmt.Errorf("loaded weights length %d does not match expected length %d", len(s.Weights), len(sm.Weights.Values))
	}
	sm.Weights.Values = s.Weights
	return nil
}
