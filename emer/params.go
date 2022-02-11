// Copyright (c) 2019, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package emer

import (
	"fmt"
	"strings"

	"github.com/emer/emergent/params"
)

// Params handles standard parameters for a Network and other objects.
// Assumes a Set named "Base" has the base-level parameters, which are
// always applied first, followed optionally by additional Set(s)
// that can have different parameters to try.
type Params struct {
	Params    params.Sets            `view:"no-inline" desc:"full collection of param sets to use"`
	ExtraSets string                 `desc:"optional additional set(s) of parameters to apply after Base -- can use multiple names separated by spaces (don't put spaces in Set names!)"`
	Objects   map[string]interface{} `view:"-" desc:"map of objects to apply parameters to -- the key is the name of the Sheet for each object, e.g., "Network", "Sim" are typically used"`
	NetHypers params.Flex            `view:"-" desc:"list of hyper parameters compiled from the network parameters, using the layers and projections from the network, so that the same styling logic as for regular parameters can be used"`
	SetMsg    bool                   `desc:"print out messages for each parameter that is set"`
}

// AddNetwork adds network to those configured by params
func (pr *Params) AddNetwork(net Network) {
	pr.AddObject("Network", net)
}

// AddSim adds Sim object to those configured by params
func (pr *Params) AddSim(sim interface{}) {
	pr.AddObject("Sim", sim)
}

// AddNetSize adds a new Network Schema object to those configured by params.
// The network schema can be retrieved using NetSize() method, and also the
// direct LayX, ..Y,  PoolX, ..Y methods can be used to directly access values.
func (pr *Params) AddNetSize() *NetSize {
	ns := &NetSize{}
	pr.AddObject("NetSize", ns)
	return ns
}

// NetSize returns the NetSize network size configuration object
// nil if it was not added
func (pr *Params) NetSize() *NetSize {
	ns, has := pr.Objects["NetSize"]
	if !has {
		return nil
	}
	return ns.(*NetSize)
}

// AddLayers adds layer(s) of given class to the NetSize for sizing params.
// Most efficient to add each class separately en-mass.
func (pr *Params) AddLayers(names []string, class string) {
	nsi, has := pr.Objects["NetSize"]
	var ns *NetSize
	if !has {
		ns = pr.AddNetSize()
	} else {
		ns = nsi.(*NetSize)
	}
	ns.AddLayers(names, class)
}

// LayX returns the X value = horizontal size of 2D layer or number of pools
// (outer dimension) for 4D layer, for given layer from NetSize, if it set there.
// Otherwise returns the provided default value
func (pr *Params) LayX(name string, def int) int {
	ns := pr.NetSize()
	if ns == nil {
		return def
	}
	return ns.LayX(name, def)
}

// LayY returns the Y value = vertical size of 2D layer or number of pools
// (outer dimension) for 4D layer, for given layer from NetSize, if it set there.
// Otherwise returns the provided default value
func (pr *Params) LayY(name string, def int) int {
	ns := pr.NetSize()
	if ns == nil {
		return def
	}
	return ns.LayY(name, def)
}

// PoolX returns the Pool X value (4D inner dim) = size of pool in units
// for given layer from NetSize if it set there.
// Otherwise returns the provided default value
func (pr *Params) PoolX(name string, def int) int {
	ns := pr.NetSize()
	if ns == nil {
		return def
	}
	return ns.PoolX(name, def)
}

// PoolY returns the Pool X value (4D inner dim) = size of pool in units
// for given layer from NetSize if it set there.
// Otherwise returns the provided default value
func (pr *Params) PoolY(name string, def int) int {
	ns := pr.NetSize()
	if ns == nil {
		return def
	}
	return ns.PoolY(name, def)
}

// AddObject adds given object with given sheet name to set of those to set params on
func (pr *Params) AddObject(name string, object interface{}) {
	if pr.Objects == nil {
		pr.Objects = make(map[string]interface{})
	}
	pr.Objects[name] = object
}

// Name returns name of current set of parameters -- if ExtraSets is empty
// then it returns "Base", otherwise returns ExtraSets
func (pr *Params) Name() string {
	if pr.ExtraSets == "" {
		return "Base"
	}
	return pr.ExtraSets
}

// Validate checks that there are sheets with the names for the
// Objects that have been added.
func (pr *Params) Validate() error {
	names := []string{}
	for nm := range pr.Objects {
		names = append(names, nm)
	}
	return pr.Params.ValidateSheets(names)
}

// SetAll sets all parameters, using "Base" Set then any ExtraSets,
// for all the Objects that have been added.  Does a Validate call first.
func (pr *Params) SetAll() error {
	err := pr.Validate()
	if err != nil {
		return err
	}
	err = pr.SetAllSet("Base")
	if pr.ExtraSets != "" && pr.ExtraSets != "Base" {
		sps := strings.Fields(pr.ExtraSets)
		for _, ps := range sps {
			err = pr.SetAllSet(ps)
		}
	}
	return err
}

// SetAllSet sets parameters for given Set name to all Objects
func (pr *Params) SetAllSet(setNm string) error {
	pset, err := pr.Params.SetByNameTry(setNm)
	if err != nil {
		return err
	}
	for nm, obj := range pr.Objects {
		sh, ok := pset.Sheets[nm]
		if !ok {
			continue
		}
		if nm == "Network" {
			net := obj.(Network)
			net.ApplyParams(sh, pr.SetMsg)
			hypers := NetworkHyperParams(net, sh)
			if setNm == "Base" {
				pr.NetHypers = hypers
			} else {
				pr.NetHypers.CopyFrom(hypers)
			}
		} else if nm == "NetSize" {
			ns := obj.(*NetSize)
			ns.ApplySheet(sh, pr.SetMsg)
		} else {
			sh.Apply(obj, pr.SetMsg)
		}
	}
	return err
}

// SetObject sets parameters, using "Base" Set then any ExtraSets,
// for the given object name (e.g., "Network" or "Sim" etc).
// Does not do Validate or collect hyper parameters.
func (pr *Params) SetObject(objName string) error {
	err := pr.SetObjectSet(objName, "Base")
	if pr.ExtraSets != "" && pr.ExtraSets != "Base" {
		sps := strings.Fields(pr.ExtraSets)
		for _, ps := range sps {
			err = pr.SetObjectSet(objName, ps)
		}
	}
	return err
}

// SetObjectSet sets parameters for given Set name to given object
func (pr *Params) SetObjectSet(objName, setNm string) error {
	pset, err := pr.Params.SetByNameTry(setNm)
	if err != nil {
		return err
	}
	sh, ok := pset.Sheets[objName]
	if !ok {
		err = fmt.Errorf("Params.SetObjectSet: sheet named: %s not found", objName)
		return err
	}
	obj, ok := pr.Objects[objName]
	if !ok {
		err = fmt.Errorf("Params.SetObjectSet: Object named: %s not found", objName)
		return err
	}
	if objName == "Network" {
		net := obj.(Network)
		net.ApplyParams(sh, pr.SetMsg)
	} else if objName == "NetSize" {
		ns := obj.(*NetSize)
		ns.ApplySheet(sh, pr.SetMsg)
	} else {
		sh.Apply(obj, pr.SetMsg)
	}
	return err
}

// NetworkHyperParams returns the compiled hyper parameters from given Sheet
// for each layer and projection in the network -- applies the standard css
// styling logic for the hyper parameters.
func NetworkHyperParams(net Network, sheet *params.Sheet) params.Flex {
	hypers := params.Flex{}
	nl := net.NLayers()
	for li := 0; li < nl; li++ {
		ly := net.Layer(li)
		nm := ly.Name()
		// typ := ly.Type().String()
		hypers[nm] = &params.FlexVal{Nm: nm, Type: "Layer", Cls: ly.Class(), Obj: params.Hypers{}}
	}
	// separate projections
	for li := 0; li < nl; li++ {
		ly := net.Layer(li)
		np := ly.NRecvPrjns()
		for pi := 0; pi < np; pi++ {
			pj := ly.RecvPrjn(pi)
			nm := pj.Name()
			// typ := pj.Type().String()
			hypers[nm] = &params.FlexVal{Nm: nm, Type: "Prjn", Cls: pj.Class(), Obj: params.Hypers{}}
		}
	}
	for nm, vl := range hypers {
		sheet.Apply(vl, false)
		hv := vl.Obj.(params.Hypers)
		hv.DeleteValOnly()
		if len(hv) == 0 {
			delete(hypers, nm)
		}
	}
	return hypers
}
