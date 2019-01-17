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

// ConnectLayers establishes a projection between two layers, adding to the recv and send
// projection lists on each side of the connection.  Returns false if not successful.
// Does not yet actually connect the units within the layers.
func (nt *Network) ConnectLayers(recv, send string, pat prjn.Pat) (rlay, slay emer.Layer, prjn *Prjn, ok bool) {
	ok = false
	rlay, has := nt.LayerByNameErrMsg(recv)
	if !has {
		return
	}
	slay, has = nt.LayerByNameErrMsg(send)
	if !has {
		return
	}
	prjn = &Prjn{}
	prjn.Recv = rlay
	prjn.Send = slay
	prjn.Pat = pat
	// rlay.RecvPrjns.Add(prjn)
	// slay.SendPrjns.Add(prjn)
	return
}
