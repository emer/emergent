// Copyright (c) 2019, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package emer

import (
	"fmt"
	"log"
	"log/slog"
	"strings"

	"github.com/emer/emergent/v2/params"
)

// Params handles standard parameters for a Network and other objects.
// Assumes a Set named "Base" has the base-level parameters, which are
// always applied first, followed optionally by additional Set(s)
// that can have different parameters to try.
type Params struct {

	// full collection of param sets to use
	Params params.Sets `display:"no-inline"`

	// optional additional set(s) of parameters to apply after Base -- can use multiple names separated by spaces (don't put spaces in Set names!)
	ExtraSets string

	// optional additional tag to add to file names, logs to identify params / run config
	Tag string

	// map of objects to apply parameters to -- the key is the name of the Sheet for each object, e.g.,
	Objects map[string]any `display:"-" Network", "Sim" are typically used"`

	// list of hyper parameters compiled from the network parameters, using the layers and pathways from the network, so that the same styling logic as for regular parameters can be used
	NetHypers params.Flex `display:"-"`

	// print out messages for each parameter that is set
	SetMsg bool
}

// AddNetwork adds network to those configured by params -- replaces any existing
// network that was set previously.
func (pr *Params) AddNetwork(net Network) {
	pr.AddObject("Network", net)
}

// AddSim adds Sim object to those configured by params -- replaces any existing.
func (pr *Params) AddSim(sim any) {
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

// AddObject adds given object with given sheet name that applies to this object.
// It is based on a map keyed on the name, so any existing object is replaced
// (safe to call repeatedly).
func (pr *Params) AddObject(name string, object any) {
	if pr.Objects == nil {
		pr.Objects = make(map[string]any)
	}
	pr.Objects[name] = object
}

// Name returns name of current set of parameters, including Tag.
// if ExtraSets is empty then it returns "Base", otherwise returns ExtraSets
func (pr *Params) Name() string {
	rn := ""
	if pr.Tag != "" {
		rn += pr.Tag + "_"
	}
	if pr.ExtraSets == "" {
		rn += "Base"
	} else {
		rn += pr.ExtraSets
	}
	return rn
}

// RunName returns standard name simulation run based on params Name()
// and starting run number if > 0 (large models are often run separately)
func (pr *Params) RunName(startRun int) string {
	rn := pr.Name()
	if startRun > 0 {
		rn += fmt.Sprintf("_%03d", startRun)
	}
	return rn
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
	for _, obj := range pr.Objects {
		if hist, ok := obj.(params.History); ok {
			hist.ParamsHistoryReset()
		}
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
func (pr *Params) SetAllSet(setName string) error {
	pset, err := pr.Params.SetByNameTry(setName)
	if err != nil {
		return err
	}
	for nm, obj := range pr.Objects {
		sh, ok := pset.Sheets[nm]
		if !ok {
			continue
		}
		sh.SelMatchReset(setName)
		if nm == "Network" {
			net := obj.(Network)
			pr.SetNetworkSheet(net, sh, setName)
		} else if nm == "NetSize" {
			ns := obj.(*NetSize)
			ns.ApplySheet(sh, pr.SetMsg)
		} else {
			sh.Apply(obj, pr.SetMsg)
		}
		err = sh.SelNoMatchWarn(setName, nm)
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

// SetNetworkMap applies params from given map of values
// The map keys are Selector:Path and the value is the value to apply, as a string.
func (pr *Params) SetNetworkMap(net Network, vals map[string]any) error {
	sh, err := params.MapToSheet(vals)
	if err != nil {
		log.Println(err)
		return err
	}
	pr.SetNetworkSheet(net, sh, "ApplyMap")
	return nil
}

// SetNetworkSheet applies params from given sheet
func (pr *Params) SetNetworkSheet(net Network, sh *params.Sheet, setName string) {
	net.ApplyParams(sh, pr.SetMsg)
	hypers := NetworkHyperParams(net, sh)
	if setName == "Base" {
		pr.NetHypers = hypers
	} else {
		pr.NetHypers.CopyFrom(hypers)
	}
}

// SetObjectSet sets parameters for given Set name to given object
func (pr *Params) SetObjectSet(objName, setName string) error {
	pset, err := pr.Params.SetByNameTry(setName)
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
	sh.SelMatchReset(setName)
	if objName == "Network" {
		net := obj.(Network)
		pr.SetNetworkSheet(net, sh, setName)
	} else if objName == "NetSize" {
		ns := obj.(*NetSize)
		ns.ApplySheet(sh, pr.SetMsg)
	} else {
		sh.Apply(obj, pr.SetMsg)
	}
	err = sh.SelNoMatchWarn(setName, objName)
	return err
}

// NetworkHyperParams returns the compiled hyper parameters from given Sheet
// for each layer and pathway in the network -- applies the standard css
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
	// separate pathways
	for li := 0; li < nl; li++ {
		ly := net.Layer(li)
		np := ly.NRecvPaths()
		for pi := 0; pi < np; pi++ {
			pj := ly.RecvPath(pi)
			nm := pj.Name()
			// typ := pj.Type().String()
			hypers[nm] = &params.FlexVal{Nm: nm, Type: "Path", Cls: pj.Class(), Obj: params.Hypers{}}
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

// SetFloatParam sets given float32 param value to layer or pathway
// (typ = Layer or Path) of given name, at given path (which can start
// with the typ name).
// Returns an error (and logs it automatically) for any failure.
func SetFloatParam(net Network, name, typ, path string, val float32) error {
	rpath := params.PathAfterType(path)
	prs := fmt.Sprintf("%g", val)
	switch typ {
	case "Layer":
		ly, err := net.LayerByNameTry(name)
		if err != nil {
			slog.Error(err.Error())
			return err
		}
		err = ly.SetParam(rpath, prs)
		if err != nil {
			slog.Error(err.Error())
			return err
		}
	case "Path":
		pj, err := net.PathByNameTry(name)
		if err != nil {
			slog.Error(err.Error())
			return err
		}
		err = pj.SetParam(rpath, prs)
		if err != nil {
			slog.Error(err.Error())
			return err
		}
	}
	return nil
}
