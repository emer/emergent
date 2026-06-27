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

// RasterSettings holds parameters controlling the raster plot view
type RasterSettings struct { //types:add

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

func (rs *RasterSettings) Defaults() {
	if rs.Max == 0 {
		rs.Max = 200
	}
	if rs.UnitSize == 0 {
		rs.UnitSize = 1
	}
	if rs.UnitHeight == 0 {
		rs.UnitHeight = .2
	}
}

// Settings holds parameters controlling how the view is rendered
type Settings struct { //types:add

	// whether to display the pathways between layers as arrows
	Paths bool

	// PathType has name(s) to display (space separated), for path arrows,
	// and when there are multiple pathways from the same layer.
	// Uses the parameter Class names in addition to type,
	// and case insensitive "contains" logic for each name.
	PathType string

	// width of the path arrows, in normalized units
	PathWidth float32 `min:"0.0001" max:".05" step:"0.001" default:"0.002"`

	// raster plot parameters
	Raster RasterSettings `display:"inline"`

	// do not record synapse level data -- turn this on for very large networks where recording the entire synaptic state would be prohibitive
	NoSynData bool

	// maximum number of records to store to enable rewinding through prior states
	MaxRecs int `min:"1"`

	// number of variable columns
	NVarCols int

	// size of a single unit, where 1 = full width and no space.. .9 default
	UnitSize float32 `min:"0.1" max:"1" step:"0.1" default:"0.9"`

	// size of the layer name labels -- entire network view is unit sized
	LayerNameSize float32 `min:"0.01" max:".1" step:"0.01" default:"0.05"`

	// name of color map to use
	ColorMap core.ColorMapName

	// opacity (0-1) of zero values -- greater magnitude values become increasingly opaque on either side of this minimum
	ZeroAlpha float32 `min:"0" max:"1" step:"0.1" default:"0.5"`

	// the number of records to jump for fast forward/backward
	NFastSteps int
}

func (ns *Settings) Defaults() {
	ns.Raster.Defaults()
	if ns.NVarCols == 0 {
		ns.NVarCols = NVarCols
		ns.Paths = true
		ns.PathWidth = 0.002
	}
	if ns.MaxRecs == 0 {
		ns.MaxRecs = 210 // 200 cycles + 8 phase updates max + 2 extra..
	}
	if ns.UnitSize == 0 {
		ns.UnitSize = .9
	}
	if ns.LayerNameSize == 0 {
		ns.LayerNameSize = .05
	}
	if ns.ZeroAlpha == 0 {
		ns.ZeroAlpha = 0.5
	}
	if ns.ColorMap == "" {
		ns.ColorMap = core.ColorMapName("ColdHot")
	}
	if ns.NFastSteps == 0 {
		ns.NFastSteps = 10
	}
}

// VarSettings holds parameters for display of each variable
type VarSettings struct {

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
func (vs *VarSettings) Defaults() {
	if vs.Range.Max == 0 && vs.Range.Min == 0 {
		vs.ZeroCtr = true
		vs.Range.SetMin(-1)
		vs.Range.SetMax(1)
	}
}

// SetProps parses Go struct-tag style properties for variable and sets values accordingly
// for customized defaults
func (vs *VarSettings) SetProps(pstr string) {
	rstr := reflect.StructTag(pstr)
	if tv, ok := rstr.Lookup("range"); ok {
		rg, err := strconv.ParseFloat(tv, 32)
		if err != nil {
			log.Printf("NetView.VarSettings.SetProps for Var: %v 'range:' err: %v on val: %v\n", vs.Var, err, tv)
		} else {
			vs.Range.Max = float32(rg)
			vs.Range.Min = -float32(rg)
			vs.ZeroCtr = true
		}
	}
	if tv, ok := rstr.Lookup("min"); ok {
		rg, err := strconv.ParseFloat(tv, 32)
		if err != nil {
			log.Printf("NetView.VarSettings.SetProps for Var: %v 'min:' err: %v on val: %v\n", vs.Var, err, tv)
		} else {
			vs.Range.Min = float32(rg)
			vs.ZeroCtr = false
		}
	}
	if tv, ok := rstr.Lookup("max"); ok {
		rg, err := strconv.ParseFloat(tv, 32)
		if err != nil {
			log.Printf("NetView.VarSettings.SetProps for Var: %v 'max:' err: %v on val: %v\n", vs.Var, err, tv)
		} else {
			vs.Range.Max = float32(rg)
			vs.ZeroCtr = false
		}
	}
	if tv, ok := rstr.Lookup("auto-scale"); ok {
		if tv == "+" {
			vs.Range.FixMin = false
			vs.Range.FixMax = false
		} else {
			vs.Range.FixMin = true
			vs.Range.FixMax = true
		}
	}
	if tv, ok := rstr.Lookup("zeroctr"); ok {
		if tv == "+" {
			vs.ZeroCtr = true
		} else {
			vs.ZeroCtr = false
		}
	}
}
