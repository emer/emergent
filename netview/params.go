// Copyright (c) 2019, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package netview

import (
	"log"
	"reflect"
	"strconv"

	"cogentcore.org/core/core"
	"cogentcore.org/core/math32/minmax"
)

// NVarCols is the default number of variable columns in the NetView
var NVarCols = 2

// RasterParams holds parameters controlling the raster plot view
type RasterParams struct { //types:add

	// if true, show a raster plot over time, otherwise units
	On bool

	// if true, the raster counter (time) is plotted across the X axis -- otherwise the Z depth axis
	XAxis bool

	// maximum count for the counter defining the raster plot
	Max int

	// size of a single unit, where 1 = full width and no space.. 1 default
	UnitSize float32 `min:"0.1" max:"1" step:"0.1" default:"1"`

	// height multiplier for units, where 1 = full height.. 0.2 default
	UnitHeight float32 `min:"0.1" max:"1" step:"0.1" default:"0.2"`
}

func (nv *RasterParams) Defaults() {
	if nv.Max == 0 {
		nv.Max = 200
	}
	if nv.UnitSize == 0 {
		nv.UnitSize = 1
	}
	if nv.UnitHeight == 0 {
		nv.UnitHeight = .2
	}
}

// Params holds parameters controlling how the view is rendered
type Params struct { //types:add

	// whether to display the pathways between layers as arrows
	Paths bool

	// width of the path arrows, in normalized units
	PathWidth float32 `default:"0.002"`

	// raster plot parameters
	Raster RasterParams `display:"inline"`

	// do not record synapse level data -- turn this on for very large networks where recording the entire synaptic state would be prohibitive
	NoSynData bool

	// if non-empty, this is the type pathway to show when there are multiple pathways from the same layer -- e.g., Inhib, Lateral, Forward, etc
	PathType string

	// maximum number of records to store to enable rewinding through prior states
	MaxRecs int `min:"1"`

	// number of variable columns
	NVarCols int

	// size of a single unit, where 1 = full width and no space.. .9 default
	UnitSize float32 `min:"0.1" max:"1" step:"0.1" default:"0.9"`

	// size of the layer name labels -- entire network view is unit sized
	LayNmSize float32 `min:"0.01" max:".1" step:"0.01" default:"0.05"`

	// name of color map to use
	ColorMap core.ColorMapName

	// opacity (0-1) of zero values -- greater magnitude values become increasingly opaque on either side of this minimum
	ZeroAlpha float32 `min:"0" max:"1" step:"0.1" default:"0.5"`

	// our netview, for update method
	NetView *NetView `copier:"-" json:"-" xml:"-" display:"-"`

	// the number of records to jump for fast forward/backward
	NFastSteps int
}

func (nv *Params) Defaults() {
	nv.Raster.Defaults()
	if nv.NVarCols == 0 {
		nv.NVarCols = NVarCols
		nv.Paths = true
		nv.PathWidth = 0.002
	}
	if nv.MaxRecs == 0 {
		nv.MaxRecs = 210 // 200 cycles + 8 phase updates max + 2 extra..
	}
	if nv.UnitSize == 0 {
		nv.UnitSize = .9
	}
	if nv.LayNmSize == 0 {
		nv.LayNmSize = .05
	}
	if nv.ZeroAlpha == 0 {
		nv.ZeroAlpha = 0.5
	}
	if nv.ColorMap == "" {
		nv.ColorMap = core.ColorMapName("ColdHot")
	}
	if nv.NFastSteps == 0 {
		nv.NFastSteps = 10
	}
}

// Update satisfies the core.Updater interface and will trigger display update on edits
func (nv *Params) Update() {
	if nv.NetView != nil {
		nv.NetView.Update()
	}
}

// VarParams holds parameters for display of each variable
type VarParams struct { //types:add

	// name of the variable
	Var string

	// keep Min - Max centered around 0, and use negative heights for units -- else use full min-max range for height (no negative heights)
	ZeroCtr bool

	// range to display
	Range minmax.Range32 `display:"inline"`

	// if not using fixed range, this is the actual range of data
	MinMax minmax.F32 `display:"inline"`
}

// Defaults sets default values if otherwise not set
func (vp *VarParams) Defaults() {
	if vp.Range.Max == 0 && vp.Range.Min == 0 {
		vp.ZeroCtr = true
		vp.Range.SetMin(-1)
		vp.Range.SetMax(1)
	}
}

// SetProps parses Go struct-tag style properties for variable and sets values accordingly
// for customized defaults
func (vp *VarParams) SetProps(pstr string) {
	rstr := reflect.StructTag(pstr)
	if tv, ok := rstr.Lookup("range"); ok {
		rg, err := strconv.ParseFloat(tv, 32)
		if err != nil {
			log.Printf("NetView.VarParams.SetProps for Var: %v 'range:' err: %v on val: %v\n", vp.Var, err, tv)
		} else {
			vp.Range.Max = float32(rg)
			vp.Range.Min = -float32(rg)
			vp.ZeroCtr = true
		}
	}
	if tv, ok := rstr.Lookup("min"); ok {
		rg, err := strconv.ParseFloat(tv, 32)
		if err != nil {
			log.Printf("NetView.VarParams.SetProps for Var: %v 'min:' err: %v on val: %v\n", vp.Var, err, tv)
		} else {
			vp.Range.Min = float32(rg)
			vp.ZeroCtr = false
		}
	}
	if tv, ok := rstr.Lookup("max"); ok {
		rg, err := strconv.ParseFloat(tv, 32)
		if err != nil {
			log.Printf("NetView.VarParams.SetProps for Var: %v 'max:' err: %v on val: %v\n", vp.Var, err, tv)
		} else {
			vp.Range.Max = float32(rg)
			vp.ZeroCtr = false
		}
	}
	if tv, ok := rstr.Lookup("auto-scale"); ok {
		if tv == "+" {
			vp.Range.FixMin = false
			vp.Range.FixMax = false
		} else {
			vp.Range.FixMin = true
			vp.Range.FixMax = true
		}
	}
	if tv, ok := rstr.Lookup("zeroctr"); ok {
		if tv == "+" {
			vp.ZeroCtr = true
		} else {
			vp.ZeroCtr = false
		}
	}
}
