// Copyright (c) 2019, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package leabra

import (
	"log"

	"github.com/emer/emergent/emer"
	"github.com/emer/emergent/prjn"
)

// leabra.NetworkStru holds the basic structural components of a network (layers)
type NetworkStru struct {
	Name   string `desc:"overall name of network -- helps discriminate if there are multiple"`
	Layers []emer.Layer
	LayMap map[string]emer.Layer `desc:"map of name to layers -- layer names must be unique"`
}

// emer.Network interface methods:
func (nt *NetworkStru) NetName() string               { return nt.Name }
func (nt *NetworkStru) NLayers() int                  { return len(nt.Layers) }
func (nt *NetworkStru) LayerIndex(idx int) emer.Layer { return nt.Layers[idx] }

// LayerByName returns a layer by looking it up by name in the layer map
// will create the layer map if it is nil or a different size than layers slice
// but otherwise needs to be updated manually.
func (nt *NetworkStru) LayerByName(name string) (emer.Layer, bool) {
	if nt.LayMap == nil || len(nt.LayMap) != len(nt.Layers) {
		nt.MakeLayMap()
	}
	ly, has := nt.LayMap[name]
	return ly, has
}

// LayerByNameErrMsg returns a layer by looking it up by name -- emits a log error message
// if layer is not found
func (nt *NetworkStru) LayerByNameErrMsg(name string) (emer.Layer, bool) {
	ly, has := nt.LayerByName(name)
	if !has {
		log.Printf("Layer named: %v not found in Network: %v\n", name, nt.Name)
	}
	return ly, has
}

// MakeLayMap updates layer map based on current layers
func (nt *NetworkStru) MakeLayMap() {
	nt.LayMap = make(map[string]emer.Layer, len(nt.Layers))
	for _, ly := range nt.Layers {
		nt.LayMap[ly.LayName()] = ly
	}
}

//////////////////////////////////////////////////////////////////////////////////////
//  Network

// leabra.Network has parameters for running a basic rate-coded Leabra network
type Network struct {
	NetworkStru
}

// Layer returns the leabra.Layer version of the layer
func (nt *Network) Layer(idx int) *Layer {
	return nt.Layers[idx].(*Layer)
}

// AddLayer adds a new layer with given name and shape to the network
func (nt *Network) AddLayer(name string, shape []int) *Layer {
	ly := &Layer{}
	ly.Name = name
	ly.SetShape(shape)
	nt.Layers = append(nt.Layers, ly)
	nt.MakeLayMap()
	return ly
}

// ConnectLayers establishes a projection between two layers, adding to the recv and send
// projection lists on each side of the connection.  Returns false if not successful.
// Does not yet actually connect the units within the layers.
func (nt *Network) ConnectLayers(recv, send string, pat prjn.Pattern) (rlay, slay emer.Layer, pj *Prjn, ok bool) {
	ok = false
	rlay, has := nt.LayerByNameErrMsg(recv)
	if !has {
		return
	}
	slay, has = nt.LayerByNameErrMsg(send)
	if !has {
		return
	}
	pj = &Prjn{}
	pj.Recv = rlay
	pj.Send = slay
	pj.Pat = pat
	rlay.(*Layer).RecvPrjns.Add(pj)
	slay.(*Layer).SendPrjns.Add(pj)
	return
}

// Build constructs the layer and projection state based on the layer shapes and patterns
// of interconnectivity
func (nt *Network) Build() {
	for _, ly := range nt.Layers {
		if ly.IsOff() {
			continue
		}
		ly.(*Layer).Build()
	}
}

// below are all the computational algorithm methods, which generally just call layer
// methods..

// todo: use goroutines here!

//////////////////////////////////////////////////////////////////////////////////////
//  Init methods

// InitWts initializes synaptic weights and all other associated long-term state variables
// including running-average state values (e.g., layer running average activations etc)
func (nt *Network) InitWts() {
	for _, ly := range nt.Layers {
		if ly.IsOff() {
			continue
		}
		ly.(*Layer).InitWts()
	}
}

func (nt *Network) InitActs() {
	for _, ly := range nt.Layers {
		if ly.IsOff() {
			continue
		}
		ly.(*Layer).InitActs()
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

// SendGeDelta sends change in activation since last sent, if above thresholds
// and integrates sent deltas into GeRaw and time-integrated Ge values
func (nt *Network) SendGeDelta() {
	for _, ly := range nt.Layers {
		if ly.IsOff() {
			continue
		}
		ly.(*Layer).SendGeDelta()
	}
	for _, ly := range nt.Layers {
		if ly.IsOff() {
			continue
		}
		ly.(*Layer).GeFmGeInc()
	}
}

// AvgMaxGe computes the average and max Ge stats, used in inhibition
func (nt *Network) AvgMaxGe() {
	for _, ly := range nt.Layers {
		if ly.IsOff() {
			continue
		}
		ly.(*Layer).AvgMaxGe()
	}
}

// InhibiFmGeAct computes inhibition Gi from Ge and Act stats within relevant Pools
func (nt *Network) InhibFmGeAct() {
	for _, ly := range nt.Layers {
		if ly.IsOff() {
			continue
		}
		ly.(*Layer).InhibFmGeAct()
	}
}

// ActFmG computes rate-code activation from Ge, Gi, Gl conductances
func (nt *Network) ActFmG() {
	for _, ly := range nt.Layers {
		if ly.IsOff() {
			continue
		}
		ly.(*Layer).ActFmG()
	}
}

// AvgMaxGe computes the average and max Ge stats, used in inhibition
func (nt *Network) AvgMaxAct() {
	for _, ly := range nt.Layers {
		if ly.IsOff() {
			continue
		}
		ly.(*Layer).AvgMaxAct()
	}
}

// for reference: full cycle run from C++ leabra
/*
void LEABRA_NETWORK_STATE::Cycle_Run_Thr(int thr_no) {
  int tot_cyc = 1;
  if(times.cycle_qtr)
    tot_cyc = times.quarter;
  for(int cyc = 0; cyc < tot_cyc; cyc++) {
    Send_Netin_Thr(thr_no);
    ThreadSyncSpin(thr_no, 0);

    Compute_NetinInteg_Thr(thr_no);
    ThreadSyncSpin(thr_no, 1);

    StartTimer(NT_NETIN_STATS, thr_no);

    Compute_NetinStats_Thr(thr_no);
    if(deep.mod_net) {
      Compute_DeepModStats_Thr(thr_no);
    }
    ThreadSyncSpin(thr_no, 2);
    if(thr_no == 0) {
      Compute_NetinStats_Post();
      if(deep.mod_net) {
        Compute_DeepModStats_Post();
      }
    }
    ThreadSyncSpin(thr_no, 0);

    InitCycleNetinTmp_Thr(thr_no);

    EndTimer(NT_NETIN_STATS, thr_no);

    if(thr_no == 0) {
      Compute_Inhib();
    }
    ThreadSyncSpin(thr_no, 1);

    Compute_Act_Thr(thr_no);
    ThreadSyncSpin(thr_no, 2);

    if(thr_no == 0) {
      Compute_CycleStats_Pre(); // prior to act post!
    }
    ThreadSyncSpin(thr_no, 0);

    Compute_Act_Post_Thr(thr_no);
    ThreadSyncSpin(thr_no, 1);

    StartTimer(NT_CYCLE_STATS, thr_no);

    Compute_CycleStats_Thr(thr_no);
    ThreadSyncSpin(thr_no, 2);

    if(thr_no == 0) {
      Compute_CycleStats_Post();
    }
    ThreadSyncSpin(thr_no, 0);

    if(deep.on && deep.Quarter_DeepRawNow(quarter)) {
      int qtrcyc = cycle % times.quarter;
      if(qtrcyc % times.deep_cyc == 0) {
        Compute_DeepRaw_Thr(thr_no);
      }
    }
    ThreadSyncSpin(thr_no, 1);

    EndTimer(NT_CYCLE_STATS, thr_no);

    if(thr_no == 0) {
      Cycle_IncrCounters();
    }
  }
}

*/
