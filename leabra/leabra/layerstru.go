// Copyright (c) 2019, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package leabra

import (
	"github.com/emer/emergent/emer"
	"github.com/emer/emergent/etensor"
)

// leabra.LayerStru manages the structural elements of the layer, which are common
// to any Layer type
type LayerStru struct {
	LeabraLay LeabraLayer    `copy:"-" json:"-" xml:"-" view:"-" desc:"we need a pointer to ourselves as an LeabraLayer (which subsumes emer.Layer), which can always be used to extract the true underlying type of object when layer is embedded in other structs -- function receivers do not have this ability so this is necessary."`
	Name      string         `desc:"Name of the layer -- this must be unique within the network, which has a map for quick lookup and layers are typically accessed directly by name"`
	Class     string         `desc:"Class is for applying parameter styles, can be space separated multple tags"`
	Off       bool           `desc:"inactivate this layer -- allows for easy experimentation"`
	Shape     etensor.Shape  `desc:"shape of the layer -- can be 2D for basic layers and 4D for layers with sub-groups (hypercolumns) -- order is outer-to-inner (row major) so Y then X for 2D and for 4D: Y-X unit pools then Y-X units within pools"`
	Type      emer.LayerType `desc:"type of layer -- Hidden, Input, Target, Compare"`
	Thread    int            `desc:"the thread number (go routine) to use in updating this layer. The user is responsible for allocating layers to threads, trying to maintain an even distribution across layers and establishing good break-points."`
	Rel       emer.Rel       `desc:"Spatial relationship to other layer, determines positioning"`
	Pos       emer.Vec3i     `desc:"position of lower-left-hand corner of layer in 3D space, computed from Rel"`
	Index     int            `desc:"a 0..n-1 index of the position of the layer within list of layers in the network. For Leabra networks, it only has significance in determining who gets which weights for enforcing initial weight symmetry -- higher layers get weights from lower layers."`
	RecvPrjns emer.PrjnList  `desc:"list of receiving projections into this layer from other layers"`
	SendPrjns emer.PrjnList  `desc:"list of sending projections from this layer to other layers"`
}

// emer.Layer interface methods

// InitName MUST be called to initialize the layer's pointer to itself as an emer.Layer
// which enables the proper interface methods to be called.  Also sets the name.
func (ls *LayerStru) InitName(lay emer.Layer, name string) {
	ls.LeabraLay = lay.(LeabraLayer)
	ls.Name = name
}

func (ls *LayerStru) AsLeabra() LeabraLayer        { return ls.LeabraLay }
func (ls *LayerStru) LayName() string              { return ls.Name }
func (ls *LayerStru) Label() string                { return ls.Name }
func (ls *LayerStru) LayClass() string             { return ls.Class }
func (ls *LayerStru) SetClass(cls string)          { ls.Class = cls }
func (ls *LayerStru) IsOff() bool                  { return ls.Off }
func (ls *LayerStru) LayShape() *etensor.Shape     { return &ls.Shape }
func (ls *LayerStru) LayThread() int               { return ls.Thread }
func (ls *LayerStru) SetThread(thr int)            { ls.Thread = thr }
func (ls *LayerStru) LayRel() emer.Rel             { return ls.Rel }
func (ls *LayerStru) SetLayRel(rel emer.Rel)       { ls.Rel = rel }
func (ls *LayerStru) LayPos() emer.Vec3i           { return ls.Pos }
func (ls *LayerStru) LayIndex() int                { return ls.Index }
func (ls *LayerStru) SetIndex(idx int)             { ls.Index = idx }
func (ls *LayerStru) RecvPrjnList() *emer.PrjnList { return &ls.RecvPrjns }
func (ls *LayerStru) NRecvPrjns() int              { return len(ls.RecvPrjns) }
func (ls *LayerStru) RecvPrjn(idx int) emer.Prjn   { return ls.RecvPrjns[idx] }
func (ls *LayerStru) SendPrjnList() *emer.PrjnList { return &ls.SendPrjns }
func (ls *LayerStru) NSendPrjns() int              { return len(ls.SendPrjns) }
func (ls *LayerStru) SendPrjn(idx int) emer.Prjn   { return ls.SendPrjns[idx] }

// SetShape sets the layer shape and also uses default dim names
func (ls *LayerStru) SetShape(shape []int) {
	var dnms []string
	if len(shape) == 2 {
		dnms = []string{"X", "Y"}
	} else if len(shape) == 4 {
		dnms = []string{"GX", "GY", "X", "Y"} // group X,Y
	}
	ls.Shape.SetShape(shape, nil, dnms) // row major default
}

// NPools returns the number of unit sub-pools according to the shape parameters.
// Currently supported for a 4D shape, where the unit pools are the first 2 Y,X dims
// and then the units within the pools are the 2nd 2 Y,X dims
func (ls *LayerStru) NPools() int {
	if ls.Shape.NumDims() != 4 {
		return 0
	}
	sh := ls.Shape.Shape()
	return int(sh[0] * sh[1])
}

// RecipToSendPrjn finds the reciprocal projection relative to the given sending projection
// found within the SendPrjns of this layer.  This is then a recv prjn within this layer:
//  S=A -> R=B recip: R=A <- S=B -- ly = A -- we are the sender of srj and recv of rpj.
// returns false if not found.
func (ls *LayerStru) RecipToSendPrjn(spj emer.Prjn) (emer.Prjn, bool) {
	for _, rpj := range ls.RecvPrjns {
		if rpj.SendLay() == spj.RecvLay() {
			return rpj, true
		}
	}
	return nil, false
}

// Config configures the basic properties of the layer
func (ls *LayerStru) Config(shape []int, typ emer.LayerType) {
	ls.SetShape(shape)
	ls.Type = typ
}

// StyleParam applies a given style to either this layer or the receiving projections in this layer
// depending on the style specification (.Class, #Name, Type) and target value of params.
// returns true if applied successfully.
// If setMsg is true, then a message is printed to confirm each parameter that is set.
// it always prints a message if a parameter fails to be set.
func (ls *LayerStru) StyleParam(sty string, pars emer.Params, setMsg bool) bool {
	if emer.StyleMatch(sty, ls.Name, ls.Class, "Layer") {
		if ls.LeabraLay.SetParams(pars, setMsg) { // note: going through LeabraLay interface is key
			return true // done -- otherwise, might be for prjns
		}
	}
	set := false
	for _, pj := range ls.RecvPrjns {
		did := pj.StyleParam(sty, pars, setMsg)
		if did {
			set = true
		}
	}
	return set
}

// StyleParams applies a given styles to either this layer or the receiving projections in this layer
// depending on the style specification (.Class, #Name, Type) and target value of params
// If setMsg is true, then a message is printed to confirm each parameter that is set.
// it always prints a message if a parameter fails to be set.
func (ls *LayerStru) StyleParams(psty emer.ParamStyle, setMsg bool) {
	for sty, pars := range psty {
		ls.StyleParam(sty, pars, setMsg)
	}
}
