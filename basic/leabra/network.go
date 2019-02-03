// Copyright (c) 2019, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package leabra

import (
	"fmt"
	"io"
	"log"
	"os"
	"sync"

	"github.com/emer/emergent/emer"
	"github.com/emer/emergent/prjn"
	"github.com/emer/emergent/timer"
	"github.com/goki/gi/gi"
	"github.com/goki/ki/indent"
	"github.com/goki/ki/ints"
)

// LayFunChan is a channel that runs layer functions
type LayFunChan chan func(ly *Layer)

// leabra.NetworkStru holds the basic structural components of a network (layers)
type NetworkStru struct {
	Name   string `desc:"overall name of network -- helps discriminate if there are multiple"`
	Layers []emer.Layer
	LayMap map[string]emer.Layer `desc:"map of name to layers -- layer names must be unique"`

	NThreads int            `inactive:"+" desc:"number of parallel threads (go routines) to use -- this is computed directly from the Layers which you must explicitly allocate to different threads -- updated during Build of network"`
	ThrLay   [][]emer.Layer `inactive:"+" desc:"layers per thread -- outer group is threads and inner is layers operated on by that thread -- based on user-assigned threads, initialized during Build"`
	ThrChans []LayFunChan   `view:"-" desc:"layer function channels, per thread"`
	ThrTimes []timer.Time   `desc:"timers for each thread, so you can see how evenly the workload is being distributed"`
	wg       sync.WaitGroup
}

// emer.Network interface methods:
func (nt *NetworkStru) NetName() string               { return nt.Name }
func (nt *NetworkStru) Label() string                 { return nt.Name }
func (nt *NetworkStru) NLayers() int                  { return len(nt.Layers) }
func (nt *NetworkStru) LayerIndex(idx int) emer.Layer { return nt.Layers[idx] }

// LayerByName returns a layer by looking it up by name in the layer map (nil if not found).
// Will create the layer map if it is nil or a different size than layers slice,
// but otherwise needs to be updated manually.
func (nt *NetworkStru) LayerByName(name string) emer.Layer {
	if nt.LayMap == nil || len(nt.LayMap) != len(nt.Layers) {
		nt.MakeLayMap()
	}
	ly := nt.LayMap[name]
	return ly
}

// LayerByNameCheck returns a layer by looking it up by name -- emits a log error message
// if layer is not found
func (nt *NetworkStru) LayerByNameCheck(name string) (emer.Layer, error) {
	ly := nt.LayerByName(name)
	if ly == nil {
		err := fmt.Errorf("Layer named: %v not found in Network: %v\n", name, nt.Name)
		return ly, err
	}
	return ly, nil
}

// MakeLayMap updates layer map based on current layers
func (nt *NetworkStru) MakeLayMap() {
	nt.LayMap = make(map[string]emer.Layer, len(nt.Layers))
	for _, ly := range nt.Layers {
		nt.LayMap[ly.LayName()] = ly
	}
}

// BuildThreads constructs the layer thread allocation based on Thread setting in the layers
func (nt *NetworkStru) BuildThreads() {
	nthr := 0
	for _, ly := range nt.Layers {
		if ly.IsOff() {
			continue
		}
		nthr = ints.MaxInt(nthr, ly.LayThread())
	}
	nt.NThreads = nthr + 1
	nt.ThrLay = make([][]emer.Layer, nt.NThreads)
	nt.ThrChans = make([]LayFunChan, nt.NThreads)
	nt.ThrTimes = make([]timer.Time, nt.NThreads)
	for _, ly := range nt.Layers {
		if ly.IsOff() {
			continue
		}
		th := ly.LayThread()
		nt.ThrLay[th] = append(nt.ThrLay[th], ly)
	}
	for th := 0; th < nt.NThreads; th++ {
		if len(nt.ThrLay[th]) == 0 {
			log.Printf("Network BuildThreads: Network %v has no layers for thread: %v\n", nt.Name, th)
		}
		nt.ThrChans[th] = make(LayFunChan)
	}
}

// StdVertLayout arranges layers in a standard vertical (z axis stack) layout, by setting
// the Rel settings
func (nt *NetworkStru) StdVertLayout() {
	lstnm := ""
	for li, ly := range nt.Layers {
		if li == 0 {
			ly.SetLayRel(emer.Rel{Rel: emer.NoRel})
			lstnm = ly.LayName()
		} else {
			ly.SetLayRel(emer.Rel{Rel: emer.Above, Other: lstnm, XAlign: emer.Middle, YAlign: emer.Front})
		}
	}
}

// todo: compute positions

//////////////////////////////////////////////////////////////////////////////////////
//  Network

// leabra.Network has parameters for running a basic rate-coded Leabra network
type Network struct {
	NetworkStru
	WtBalInterval int `def:"10" desc:"how frequently to update the weight balance average weight factor -- relatively expensive"`
	WtBalCtr      int `inactive:"+" desc:"counter for how long it has been since last WtBal"`
}

// Defaults sets all the default parameters for all layers and projections
func (nt *Network) Defaults() {
	nt.WtBalInterval = 10
	nt.WtBalCtr = 0
	for li, ly := range nt.Layers {
		ly.Defaults()
		ly.(*Layer).Index = li
	}
}

// UpdateParams updates all the derived parameters if any have changed, for all layers
// and projections
func (nt *Network) UpdateParams() {
	for _, ly := range nt.Layers {
		ly.UpdateParams()
	}
}

// StyleParams applies a given styles to layers and receiving projections,
// depending on the style specification (.Class, #Name, Type) and target value of params
// If setMsg is true, then a message is printed to confirm each parameter that is set.
// it always prints a message if a parameter fails to be set.
func (nt *Network) StyleParams(psty emer.ParamStyle, setMsg bool) {
	for _, ly := range nt.Layers {
		ly.StyleParams(psty, setMsg)
	}
}

// Layer returns the leabra.Layer version of the layer
func (nt *Network) Layer(idx int) *Layer {
	return nt.Layers[idx].(*Layer)
}

// AddLayer adds a new layer with given name and shape to the network
func (nt *Network) AddLayer(name string, shape []int, typ LayerType) *Layer {
	ly := &Layer{}
	ly.Name = name
	ly.SetShape(shape)
	ly.Type = typ
	nt.Layers = append(nt.Layers, ly)
	nt.MakeLayMap()
	return ly
}

// ConnectLayerNames establishes a projection between two layers, referenced by name
// adding to the recv and send projection lists on each side of the connection.
// Returns error if not successful.
// Does not yet actually connect the units within the layers -- that requires Build.
func (nt *Network) ConnectLayersNames(send, recv string, pat prjn.Pattern) (rlay, slay emer.Layer, pj *Prjn, err error) {
	rlay, err = nt.LayerByNameCheck(recv)
	if err != nil {
		return
	}
	slay, err = nt.LayerByNameCheck(send)
	if err != nil {
		return
	}
	pj = nt.ConnectLayers(rlay.(*Layer), slay.(*Layer), pat)
	return
}

// ConnectLayers establishes a projection between two layers,
// adding to the recv and send projection lists on each side of the connection.
// Returns false if not successful. Does not yet actually connect the units within the layers -- that
// requires Build.
func (nt *Network) ConnectLayers(send, recv *Layer, pat prjn.Pattern) *Prjn {
	pj := &Prjn{}
	pj.Recv = recv
	pj.Send = send
	pj.Pat = pat
	recv.RecvPrjns.Add(pj)
	send.SendPrjns.Add(pj)
	return pj
}

// Build constructs the layer and projection state based on the layer shapes and patterns
// of interconnectivity
func (nt *Network) Build() {
	for li, ly := range nt.Layers {
		ly.(*Layer).Index = li
		if ly.IsOff() {
			continue
		}
		ly.(*Layer).Build()
	}
	nt.BuildThreads()
	nt.StartThreads()
}

// SaveWtsJSON saves network weights (and any other state that adapts with learning)
// to a JSON-formatted file
func (nt *Network) SaveWtsJSON(filename gi.FileName) error {
	fp, err := os.Create(string(filename))
	defer fp.Close()
	if err != nil {
		log.Println(err)
		return err
	}
	nt.WriteWtsJSON(fp)
	return nil
}

// OpenWtsJSON opens network weights (and any other state that adapts with learning)
// from a JSON-formatted file
func (nt *Network) OpenWtsJSON(filename gi.FileName) error {
	fp, err := os.Open(string(filename))
	defer fp.Close()
	if err != nil {
		log.Println(err)
		return err
	}
	return nt.ReadWtsJSON(fp)
}

// WriteWtsJSON writes the weights from this layer from the receiver-side perspective
// in a JSON text format.  We build in the indentation logic to make it much faster and
// more efficient.
func (nt *Network) WriteWtsJSON(w io.Writer) {
	depth := 0
	w.Write(indent.TabBytes(depth))
	w.Write([]byte("{\n"))
	depth++
	w.Write(indent.TabBytes(depth))
	w.Write([]byte(fmt.Sprintf("\"%v\": [\n", nt.Name)))
	depth++
	for _, ly := range nt.Layers {
		if ly.IsOff() {
			continue
		}
		ly.(*Layer).WriteWtsJSON(w, depth)
	}
	depth--
	w.Write(indent.TabBytes(depth))
	w.Write([]byte("]\n"))
	depth--
	w.Write(indent.TabBytes(depth))
	w.Write([]byte("}\n"))
}

// ReadWtsJSON reads the weights from this layer from the receiver-side perspective
// in a JSON text format.  We build in the indentation logic to make it much faster and
// more efficient.
func (nt *Network) ReadWtsJSON(r io.Reader) error {
	return nil
}

// below are all the computational algorithm methods, which generally just call layer
// methods..

// todo: use goroutines here!

//////////////////////////////////////////////////////////////////////////////////////
//  Init methods

// InitWts initializes synaptic weights and all other associated long-term state variables
// including running-average state values (e.g., layer running average activations etc)
func (nt *Network) InitWts() {
	nt.WtBalCtr = 0
	for _, ly := range nt.Layers {
		if ly.IsOff() {
			continue
		}
		ly.(*Layer).InitWts()
	}
	// separate pass to enforce symmetry
	for _, ly := range nt.Layers {
		if ly.IsOff() {
			continue
		}
		ly.(*Layer).InitWtSym()
	}
}

// InitActs fully initializes activation state -- not automatically called
func (nt *Network) InitActs() {
	for _, ly := range nt.Layers {
		if ly.IsOff() {
			continue
		}
		ly.(*Layer).InitActs()
	}
}

// InitExt initializes external input state -- call prior to applying external inputs to layers
func (nt *Network) InitExt() {
	for _, ly := range nt.Layers {
		if ly.IsOff() {
			continue
		}
		ly.(*Layer).InitExt()
	}
}

// TrialInit handles all initialization at start of new input pattern, including computing
// netinput scaling from running average activation etc.
func (nt *Network) TrialInit() {
	for _, ly := range nt.Layers {
		if ly.IsOff() {
			continue
		}
		ly.(*Layer).TrialInit()
	}
}

//////////////////////////////////////////////////////////////////////////////////////
//  Act methods

// Cycle runs one cycle of activation updating:
// * Sends Ge increments from sending to receiving layers
// * Average and Max Ge stats
// * Inhibition based on Ge stats and Act Stats (computed at end of Cycle)
// * Activation from Ge, Gi, and Gl
// * Average and Max Act stats
func (nt *Network) Cycle() {
	nt.SendGeDelta() // also does integ
	nt.AvgMaxGe()
	nt.InhibFmGeAct()
	nt.ActFmG()
	nt.AvgMaxAct()
}

// StartThreads starts up the threads for computation
func (nt *Network) StartThreads() {
	for th := 0; th < nt.NThreads; th++ {
		go nt.ThrWorker(th) // start the worker thread for this channel
	}
}

// todo: stop threads too!

// ThrWorker is the worker function run by the worker threads
func (nt *Network) ThrWorker(tt int) {
	for fun := range nt.ThrChans[tt] {
		// fmt.Printf("worker %v got fun\n", tt)
		thly := nt.ThrLay[tt]
		nt.ThrTimes[tt].Start()
		for _, ly := range thly {
			if ly.IsOff() {
				continue
			}
			fun(ly.(*Layer))
		}
		nt.ThrTimes[tt].Stop()
		nt.wg.Done()
	}
}

// ThrChanLayFun runs given layer computation function on thread worker
// using the ThrChans channel
func (nt *Network) ThrChanLayFun(tt int, fun func(ly *Layer)) {
	nt.wg.Add(1)
	nt.ThrChans[tt] <- func(ly *Layer) {
		fun(ly)
	}
}

// ThrLayFun calls function on layer, using threaded (go routine) computation if NThreads > 1
// and otherwise just iterates over layers in the current thread.
func (nt *Network) ThrLayFun(fun func(ly *Layer)) {
	if nt.NThreads <= 1 {
		for _, ly := range nt.Layers {
			if ly.IsOff() {
				continue
			}
			fun(ly.(*Layer))
		}
	} else {
		for th := 0; th < nt.NThreads; th++ {
			// fmt.Printf("calling worker %v\n", th)
			nt.ThrChanLayFun(th, fun)
		}
		nt.wg.Wait()
	}
}

// ThrTimerReport reports the amount of time spent in each thread
func (nt *Network) ThrTimerReport() {
	if nt.NThreads <= 1 {
		fmt.Printf("ThrTimerReport: not running multiple threads\n")
		return
	}
	fmt.Printf("ThrTimerReport: %v, NThreads: %v\n", nt.Name, nt.NThreads)
	fmt.Printf("\tThr\tTotal Secs\tPct\n")
	pcts := make([]float64, nt.NThreads)
	tot := 0.0
	for th := 0; th < nt.NThreads; th++ {
		pcts[th] = nt.ThrTimes[th].TotalSecs()
		tot += pcts[th]
	}
	for th := 0; th < nt.NThreads; th++ {
		fmt.Printf("\t%v \t%6g\t%6g\n", th, pcts[th], pcts[th]/tot)
	}
}

// ThrTimerReset resets the per-thread timers
func (nt *Network) ThrTimerReset() {
	for th := 0; th < nt.NThreads; th++ {
		nt.ThrTimes[th].Reset()
	}
}

// SendGeDelta sends change in activation since last sent, if above thresholds
// and integrates sent deltas into GeRaw and time-integrated Ge values
func (nt *Network) SendGeDelta() {
	nt.ThrLayFun(func(ly *Layer) { ly.SendGeDelta() })
	nt.ThrLayFun(func(ly *Layer) { ly.GeFmGeInc() })
}

// AvgMaxGe computes the average and max Ge stats, used in inhibition
func (nt *Network) AvgMaxGe() {
	nt.ThrLayFun(func(ly *Layer) { ly.AvgMaxGe() })
}

// InhibiFmGeAct computes inhibition Gi from Ge and Act stats within relevant Pools
func (nt *Network) InhibFmGeAct() {
	nt.ThrLayFun(func(ly *Layer) { ly.InhibFmGeAct() })
}

// ActFmG computes rate-code activation from Ge, Gi, Gl conductances
func (nt *Network) ActFmG() {
	nt.ThrLayFun(func(ly *Layer) { ly.ActFmG() })
}

// AvgMaxGe computes the average and max Ge stats, used in inhibition
func (nt *Network) AvgMaxAct() {
	nt.ThrLayFun(func(ly *Layer) { ly.AvgMaxAct() })
}

// QuarterFinal does updating after end of a quarter
func (nt *Network) QuarterFinal(ltime *Time) {
	nt.ThrLayFun(func(ly *Layer) { ly.QuarterFinal(ltime) })
}

//////////////////////////////////////////////////////////////////////////////////////
//  Learn methods

// DWt computes the weight change (learning) based on current running-average activation values
func (nt *Network) DWt() {
	nt.ThrLayFun(func(ly *Layer) { ly.DWt() })
}

// WtFmDWt updates the weights from delta-weight changes.
// Also calls WtBalFmWt every WtBalInterval times
func (nt *Network) WtFmDWt() {
	nt.ThrLayFun(func(ly *Layer) { ly.WtFmDWt() })
	nt.WtBalCtr++
	if nt.WtBalCtr >= nt.WtBalInterval {
		nt.WtBalCtr = 0
		nt.WtBalFmWt()
	}
}

// WtBalFmWt updates the weight balance factors based on average recv weights
func (nt *Network) WtBalFmWt() {
	nt.ThrLayFun(func(ly *Layer) { ly.WtBalFmWt() })
}
