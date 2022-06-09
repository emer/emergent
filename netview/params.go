// Copyright (c) 2019, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package netview

import (
	"log"
	"reflect"
	"strconv"

	"github.com/emer/etable/minmax"
	"github.com/goki/gi/giv"
)

// RasterParams holds parameters controlling the raster plot view
type RasterParams struct {
	On         bool    `desc:"if true, show a raster plot over time, otherwise units"`
	XAxis      bool    `desc:"if true, the raster counter (time) is plotted across the X axis -- otherwise the Z depth axis"`
	Max        int     `desc:"maximum count for the counter defining the raster plot"`
	UnitSize   float32 `min:"0.1" max:"1" step:"0.1" def:"1" desc:"size of a single unit, where 1 = full width and no space.. 1 default"`
	UnitHeight float32 `min:"0.1" max:"1" step:"0.1" def:"0.2" desc:"height multiplier for units, where 1 = full height.. 0.2 default"`
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
type Params struct {
	Raster     RasterParams     `view:"inline" desc:"raster plot parameters"`
	NoSynData  bool             `desc:"do not record synapse level data -- turn this on for very large networks where recording the entire synaptic state would be prohibitive"`
	PrjnType   string           `desc:"if non-empty, this is the type projection to show when there are multiple projections from the same layer -- e.g., Inhib, Lateral, Forward, etc"`
	MaxRecs    int              `min:"1" desc:"maximum number of records to store to enable rewinding through prior states"`
	UnitSize   float32          `min:"0.1" max:"1" step:"0.1" def:"0.9" desc:"size of a single unit, where 1 = full width and no space.. .9 default"`
	LayNmSize  float32          `min:"0.01" max:".1" step:"0.01" def:"0.05" desc:"size of the layer name labels -- entire network view is unit sized"`
	ColorMap   giv.ColorMapName `desc:"name of color map to use"`
	ZeroAlpha  float32          `min:"0" max:"1" step:"0.1" def:"0.5" desc:"opacity (0-1) of zero values -- greater magnitude values become increasingly opaque on either side of this minimum"`
	NetView    *NetView         `copy:"-" json:"-" xml:"-" view:"-" desc:"our netview, for update method"`
	NFastSteps int              `desc:"the number of records to jump for fast forward/backward"`
}

func (nv *Params) Defaults() {
	nv.Raster.Defaults()
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
		nv.ColorMap = giv.ColorMapName("ColdHot")
	}
	if nv.NFastSteps == 0 {
		nv.NFastSteps = 10
	}
}

// Update satisfies the gi.Updater interface and will trigger display update on edits
func (nv *Params) Update() {
	if nv.NetView != nil {
		nv.NetView.Config()
		nv.NetView.Update()
	}
}

// VarParams holds parameters for display of each variable
type VarParams struct {
	Var     string         `desc:"name of the variable"`
	ZeroCtr bool           `desc:"keep Min - Max centered around 0, and use negative heights for units -- else use full min-max range for height (no negative heights)"`
	Range   minmax.Range32 `view:"inline" desc:"range to display"`
	MinMax  minmax.F32     `view:"inline" desc:"if not using fixed range, this is the actual range of data"`
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
