// Copyright (c) 2022, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package estats

import (
	"fmt"

	"github.com/emer/emergent/actrf"
	"github.com/emer/emergent/confusion"
	"github.com/emer/emergent/timer"
	"github.com/emer/etable/etensor"
	"github.com/emer/etable/pca"
	"github.com/emer/etable/simat"
)

// Stats provides maps for storing statistics as named scalar and tensor values.
// These stats are available in the elog.Context for use during logging.
type Stats struct {
	Floats     map[string]float64
	Strings    map[string]string
	Ints       map[string]int
	F32Tensors map[string]*etensor.Float32 `desc:"float32 tensors used for grabbing values from layers"`
	F64Tensors map[string]*etensor.Float64 `desc:"float64 tensors as needed for other computations"`
	IntTensors map[string]*etensor.Int     `desc:"int tensors as needed for other computations"`
	Confusion  confusion.Matrix            `view:"no-inline" desc:"confusion matrix"`
	SimMats    map[string]*simat.SimMat    `desc:"similarity matrix for comparing pattern similarities"`
	PCA        pca.PCA                     `desc:"one PCA object can be reused for all PCA computations"`
	SVD        pca.SVD                     `desc:"one SVD object can be reused for all SVD computations"`
	ActRFs     actrf.RFs                   `view:"no-inline" desc:"activation-based receptive fields"`
	Rasters    []string                    `desc:"list of layer names configured for recording raster plots"`
	Timers     map[string]*timer.Time      `desc:"named timers available for timing how long different computations take (wall-clock time)"`
}

// Init must be called before use to create all the maps
func (st *Stats) Init() {
	st.Floats = make(map[string]float64)
	st.Strings = make(map[string]string)
	st.Ints = make(map[string]int)
	st.F32Tensors = make(map[string]*etensor.Float32)
	st.F64Tensors = make(map[string]*etensor.Float64)
	st.IntTensors = make(map[string]*etensor.Int)
	st.SimMats = make(map[string]*simat.SimMat)
	st.Timers = make(map[string]*timer.Time)
	st.PCA.Init()
	st.SVD.Init()
	st.SVD.Cond = PCAStrongThr
}

// Print returns a formatted Name: Value string of stat values,
// suitable for displaying at the bottom of the NetView or otherwise printing.
// Looks for names of stats in order of fields in Stats object (Floats, Strings, Ints)
func (st *Stats) Print(stats []string) string {
	var str string
	for _, nm := range stats {
		if str != "" {
			str += "\t"
		}
		str += fmt.Sprintf("%s: \t", nm)
		if val, has := st.Floats[nm]; has {
			str += fmt.Sprintf("%.4g", val)
			continue
		}
		if val, has := st.Strings[nm]; has {
			str += fmt.Sprintf("%s", val)
			continue
		}
		if val, has := st.Ints[nm]; has {
			str += fmt.Sprintf("%d", val)
			continue
		}
		str += "<not found!>"
	}
	return str
}

// PrintVals returns values of given stats with given formats,
// and delimiter
func (st *Stats) PrintVals(stats, fmts []string, delim string) string {
	var str string
	for i, nm := range stats {
		fm := fmts[i]
		if str != "" {
			str += delim
		}
		if val, has := st.Floats[nm]; has {
			str += fmt.Sprintf(fm, val)
			continue
		}
		if val, has := st.Strings[nm]; has {
			str += fmt.Sprintf(fm, val)
			continue
		}
		if val, has := st.Ints[nm]; has {
			str += fmt.Sprintf(fm, val)
			continue
		}
		str += "0"
	}
	return str
}

func (st *Stats) SetFloat(name string, value float64) {
	st.Floats[name] = value
}

func (st *Stats) SetString(name string, value string) {
	st.Strings[name] = value
}

func (st *Stats) SetInt(name string, value int) {
	st.Ints[name] = value
}

func (st *Stats) Float(name string) float64 {
	val, has := st.Floats[name]
	if has {
		return val
	}
	fmt.Printf("Value named: %s not found in Stats\n", name)
	return 0
}

func (st *Stats) String(name string) string {
	val, has := st.Strings[name]
	if has {
		return val
	}
	fmt.Printf("Value named: %s not found in Stats\n", name)
	return ""
}

func (st *Stats) Int(name string) int {
	val, has := st.Ints[name]
	if has {
		return val
	}
	fmt.Printf("Value named: %s not found in Stats\n", name)
	return 0
}

// F32Tensor returns a float32 tensor of given name, creating if not yet made
func (st *Stats) F32Tensor(name string) *etensor.Float32 {
	tsr, has := st.F32Tensors[name]
	if !has {
		tsr = &etensor.Float32{}
		st.F32Tensors[name] = tsr
	}
	return tsr
}

// F64Tensor returns a float64 tensor of given name, creating if not yet made
func (st *Stats) F64Tensor(name string) *etensor.Float64 {
	tsr, has := st.F64Tensors[name]
	if !has {
		tsr = &etensor.Float64{}
		st.F64Tensors[name] = tsr
	}
	return tsr
}

// IntTensor returns a int tensor of given name, creating if not yet made
func (st *Stats) IntTensor(name string) *etensor.Int {
	tsr, has := st.IntTensors[name]
	if !has {
		tsr = &etensor.Int{}
		st.IntTensors[name] = tsr
	}
	return tsr
}

// SetF32Tensor sets a float32 tensor of given name.
// Just does: st.F32Tensors[name] = tsr
func (st *Stats) SetF32Tensor(name string, tsr *etensor.Float32) {
	st.F32Tensors[name] = tsr
}

// SetF64Tensor sets a float64 tensor of given name.
// Just does: st.F64Tensors[name] = tsr
func (st *Stats) SetF64Tensor(name string, tsr *etensor.Float64) {
	st.F64Tensors[name] = tsr
}

// SetIntTensor sets a int tensor of given name.
// Just does: st.IntTensors[name] = tsr
func (st *Stats) SetIntTensor(name string, tsr *etensor.Int) {
	st.IntTensors[name] = tsr
}

// SimMat returns a SimMat similarity matrix of given name, creating if not yet made
func (st *Stats) SimMat(name string) *simat.SimMat {
	sm, has := st.SimMats[name]
	if !has {
		sm = &simat.SimMat{}
		st.SimMats[name] = sm
	}
	return sm
}

// Timer returns timer of given name, creating if not yet made
func (st *Stats) Timer(name string) *timer.Time {
	tmr, has := st.Timers[name]
	if !has {
		tmr = &timer.Time{}
		st.Timers[name] = tmr
	}
	return tmr
}

// StartTimer starts given named timer
func (st *Stats) StartTimer(name string) *timer.Time {
	tmr := st.Timer(name)
	tmr.Start()
	return tmr
}

// StopTimer stops given timer
func (st *Stats) StopTimer(name string) *timer.Time {
	tmr := st.Timer(name)
	tmr.Stop()
	return tmr
}

// ResetTimer resets given named timer
func (st *Stats) ResetTimer(name string) *timer.Time {
	tmr := st.Timer(name)
	tmr.Reset()
	return tmr
}

// ResetStartTimer resets then starts given named timer
func (st *Stats) ResetStartTimer(name string) *timer.Time {
	tmr := st.Timer(name)
	tmr.ResetStart()
	return tmr
}
