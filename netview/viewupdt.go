// Copyright (c) 2019, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package netview

import (
	"strings"

	"github.com/emer/emergent/etime"
)

// ViewUpdt manages time scales for updating the NetView
type ViewUpdt struct {
	View      *NetView    `view:"-" desc:"the network view"`
	Testing   bool        `view:"-" desc:"whether in testing mode -- can be set in advance to drive appropriate updating"`
	Text      string      `view:"-" desc:"text to display at the bottom of the view"`
	On        bool        `desc:"toggles update of display on"`
	SkipInvis bool        `desc:"if true, do not record network data when the NetView is invisible -- this speeds up running when not visible, but the NetView display will not show the current state when switching back to it"`
	Train     etime.Times `desc:"at what time scale to update the display during training?"`
	Test      etime.Times `desc:"at what time scale to update the display during testing?"`
}

// Config configures for given NetView and default train, test times
func (vu *ViewUpdt) Config(nv *NetView, train, test etime.Times) {
	vu.View = nv
	vu.On = true
	vu.Train = train
	vu.Test = test
}

// UpdtTime returns the relevant update time based on testing flag
func (vu *ViewUpdt) UpdtTime(testing bool) etime.Times {
	if testing {
		return vu.Test
	}
	return vu.Train
}

// Update does an update if view is On, visible and active,
// including recording new data and driving update of display
func (vu *ViewUpdt) Update() {
	if !vu.On || vu.View == nil {
		return
	}
	if !vu.View.IsVisible() && vu.SkipInvis {
		vu.View.RecordCounters(vu.Text)
		return
	}
	vu.View.Record(vu.Text, -1) // -1 = use a dummy counter
	// note: essential to use Go version of update when called from another goroutine
	if vu.View.IsVisible() {
		vu.View.GoUpdate()
	}
}

// UpdateWhenStopped does an update when the network updating was stopped
// either via stepping or hitting the stop button -- this has different
// logic for the raster view vs. regular.
func (vu *ViewUpdt) UpdateWhenStopped() {
	if !vu.On || vu.View == nil {
		return
	}
	if !vu.View.IsVisible() && vu.SkipInvis {
		vu.View.RecordCounters(vu.Text)
		return
	}
	if !vu.View.Params.Raster.On { // always record when not in raster mode
		vu.View.Record(vu.Text, -1) // -1 = use a dummy counter
	}
	// note: essential to use Go version of update when called from another goroutine
	if vu.View.IsVisible() {
		vu.View.GoUpdate()
	}
}

// UpdateTime triggers an update at given timescale.
func (vu *ViewUpdt) UpdateTime(time etime.Times) {
	if !vu.On || vu.View == nil {
		return
	}
	viewUpdt := vu.UpdtTime(vu.Testing)
	if viewUpdt == time {
		vu.Update()
	} else {
		if viewUpdt < etime.Trial && time == etime.Trial {
			if vu.View.Params.Raster.On { // no extra rec here
				vu.View.Data.RecordLastCtrs(vu.Text)
				if vu.View.IsVisible() {
					vu.View.GoUpdate()
				}
			} else {
				vu.Update()
			}
		}
	}
}

// IsCycleUpdating returns true if the view is updating at a cycle level,
// either from raster or literal cycle level.
func (vu *ViewUpdt) IsCycleUpdating() bool {
	if !vu.On || vu.View == nil || !(vu.View.IsVisible() || !vu.SkipInvis) {
		return false
	}
	viewUpdt := vu.UpdtTime(vu.Testing)
	if viewUpdt > etime.ThetaCycle {
		return false
	}
	if viewUpdt == etime.Cycle {
		return true
	}
	if vu.View.Params.Raster.On {
		return true
	}
	return false
}

// IsViewingSynapse returns true if netview is actively viewing synapses.
func (vu *ViewUpdt) IsViewingSynapse() bool {
	if !vu.On || vu.View == nil || !(vu.View.IsVisible() || !vu.SkipInvis) {
		return false
	}
	vvar := vu.View.Var
	if strings.HasPrefix(vvar, "r.") || strings.HasPrefix(vvar, "s.") {
		return true
	}
	return false
}

// UpdateCycle triggers an update at the Cycle (Millisecond) timescale,
// using given text to display at bottom of view
func (vu *ViewUpdt) UpdateCycle(cyc int) {
	if !vu.On || vu.View == nil {
		return
	}
	viewUpdt := vu.UpdtTime(vu.Testing)
	if viewUpdt > etime.ThetaCycle {
		return
	}
	if vu.View.Params.Raster.On {
		vu.UpdateCycleRaster(cyc)
		return
	}
	switch viewUpdt {
	case etime.Cycle:
		vu.Update()
	case etime.FastSpike:
		if cyc%10 == 0 {
			vu.Update()
		}
	case etime.GammaCycle:
		if cyc%25 == 0 {
			vu.Update()
		}
	case etime.BetaCycle:
		if cyc%50 == 0 {
			vu.Update()
		}
	case etime.AlphaCycle:
		if cyc%100 == 0 {
			vu.Update()
		}
	case etime.ThetaCycle:
		if cyc%200 == 0 {
			vu.Update()
		}
	}
}

// UpdateCycleRaster raster version of Cycle update
func (vu *ViewUpdt) UpdateCycleRaster(cyc int) {
	if !vu.View.IsVisible() && vu.SkipInvis {
		vu.View.RecordCounters(vu.Text)
		return
	}
	viewUpdt := vu.UpdtTime(vu.Testing)
	vu.View.Record(vu.Text, cyc)
	switch viewUpdt {
	case etime.Cycle:
		vu.View.GoUpdate()
	case etime.FastSpike:
		if cyc%10 == 0 {
			vu.View.GoUpdate()
		}
	case etime.GammaCycle:
		if cyc%25 == 0 {
			vu.View.GoUpdate()
		}
	case etime.BetaCycle:
		if cyc%50 == 0 {
			vu.View.GoUpdate()
		}
	case etime.AlphaCycle:
		if cyc%100 == 0 {
			vu.View.GoUpdate()
		}
	case etime.ThetaCycle:
		if cyc%200 == 0 {
			vu.View.GoUpdate()
		}
	}
}

// RecordSyns records synaptic data -- stored separate from unit data
// and only needs to be called when synaptic values are updated.
// Should be done when the DWt values have been computed, before
// updating Wts and zeroing.
// NetView displays this recorded data when Update is next called.
func (vu *ViewUpdt) RecordSyns() {
	if !vu.On || vu.View == nil {
		return
	}
	if !vu.View.IsVisible() {
		if vu.SkipInvis || !vu.IsViewingSynapse() {
			return
		}
	}
	vu.View.RecordSyns()
}
