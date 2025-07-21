// Copyright (c) 2019, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package netview

import (
	"strings"

	"github.com/emer/emergent/v2/etime"
)

// ViewUpdate manages time scales for updating the NetView
type ViewUpdate struct {

	// View is the network view.
	View *NetView `display:"-"`

	// whether in testing mode -- can be set in advance to drive appropriate updating
	Testing bool `display:"-"`

	// text to display at the bottom of the view
	Text string `display:"-"`

	// toggles update of display on
	On bool

	// SkipInvis means do not record network data when the NetView is invisible.
	// This speeds up running when not visible, but the NetView display will
	// not show the current state when switching back to it.
	SkipInvis bool

	// at what time scale to update the display during training?
	Train etime.Times

	// at what time scale to update the display during testing?
	Test etime.Times
}

// Config configures for given NetView and default train, test times
func (vu *ViewUpdate) Config(nv *NetView, train, test etime.Times) {
	vu.View = nv
	vu.On = true
	vu.Train = train
	vu.Test = test
	vu.SkipInvis = true // more often running than debugging probably
}

// GetUpdateTime returns the relevant update time based on testing flag
func (vu *ViewUpdate) GetUpdateTime(testing bool) etime.Times {
	if testing {
		return vu.Test
	}
	return vu.Train
}

// GoUpdate does an update if view is On, visible and active,
// including recording new data and driving update of display.
// This version is only for calling from a separate goroutine,
// not the main event loop (see also Update).
func (vu *ViewUpdate) GoUpdate() {
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
		vu.View.GoUpdateView()
	}
}

// Update does an update if view is On, visible and active,
// including recording new data and driving update of display.
// This version is only for calling from the main event loop
// (see also GoUpdate).
func (vu *ViewUpdate) Update() {
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
		vu.View.UpdateView()
	}
}

// UpdateWhenStopped does an update when the network updating was stopped
// either via stepping or hitting the stop button.
// This has different logic for the raster view vs. regular.
// This is only for calling from a separate goroutine,
// not the main event loop.
func (vu *ViewUpdate) UpdateWhenStopped() {
	if !vu.On || vu.View == nil {
		return
	}
	if !vu.View.IsVisible() && vu.SkipInvis {
		vu.View.RecordCounters(vu.Text)
		return
	}
	if !vu.View.Options.Raster.On { // always record when not in raster mode
		vu.View.Record(vu.Text, -1) // -1 = use a dummy counter
	}
	// todo: updating is not available here -- needed?
	// if vu.View.Scene.Is(core.ScUpdating) {
	// 	return
	// }
	vu.View.GoUpdateView()
}

// UpdateTime triggers an update at given timescale.
func (vu *ViewUpdate) UpdateTime(time etime.Times) {
	if !vu.On || vu.View == nil {
		return
	}
	viewUpdate := vu.GetUpdateTime(vu.Testing)
	if viewUpdate == time {
		vu.GoUpdate()
	} else {
		if viewUpdate < etime.Trial && time == etime.Trial {
			if vu.View.Options.Raster.On { // no extra rec here
				vu.View.Data.RecordLastCtrs(vu.Text)
				if vu.View.IsVisible() {
					vu.View.GoUpdateView()
				}
			} else {
				vu.GoUpdate()
			}
		}
	}
}

// IsCycleUpdating returns true if the view is updating at a cycle level,
// either from raster or literal cycle level.
func (vu *ViewUpdate) IsCycleUpdating() bool {
	if !vu.On || vu.View == nil || !(vu.View.IsVisible() || !vu.SkipInvis) {
		return false
	}
	viewUpdate := vu.GetUpdateTime(vu.Testing)
	if viewUpdate > etime.ThetaCycle {
		return false
	}
	if viewUpdate == etime.Cycle {
		return true
	}
	if vu.View.Options.Raster.On {
		return true
	}
	return false
}

// IsViewingSynapse returns true if netview is actively viewing synapses.
func (vu *ViewUpdate) IsViewingSynapse() bool {
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
func (vu *ViewUpdate) UpdateCycle(cyc int) {
	if !vu.On || vu.View == nil {
		return
	}
	viewUpdate := vu.GetUpdateTime(vu.Testing)
	if viewUpdate > etime.ThetaCycle {
		return
	}
	if vu.View.Options.Raster.On {
		vu.UpdateCycleRaster(cyc)
		return
	}
	switch viewUpdate {
	case etime.Cycle:
		vu.GoUpdate()
	case etime.FastSpike:
		if cyc%10 == 0 {
			vu.GoUpdate()
		}
	case etime.GammaCycle:
		if cyc%25 == 0 {
			vu.GoUpdate()
		}
	case etime.BetaCycle:
		if cyc%50 == 0 {
			vu.GoUpdate()
		}
	case etime.AlphaCycle:
		if cyc%100 == 0 {
			vu.GoUpdate()
		}
	case etime.ThetaCycle:
		if cyc%200 == 0 {
			vu.GoUpdate()
		}
	}
}

// UpdateCycleRaster raster version of Cycle update
func (vu *ViewUpdate) UpdateCycleRaster(cyc int) {
	if !vu.View.IsVisible() && vu.SkipInvis {
		vu.View.RecordCounters(vu.Text)
		return
	}
	viewUpdate := vu.GetUpdateTime(vu.Testing)
	vu.View.Record(vu.Text, cyc)
	switch viewUpdate {
	case etime.Cycle:
		vu.View.GoUpdateView()
	case etime.FastSpike:
		if cyc%10 == 0 {
			vu.View.GoUpdateView()
		}
	case etime.GammaCycle:
		if cyc%25 == 0 {
			vu.View.GoUpdateView()
		}
	case etime.BetaCycle:
		if cyc%50 == 0 {
			vu.View.GoUpdateView()
		}
	case etime.AlphaCycle:
		if cyc%100 == 0 {
			vu.View.GoUpdateView()
		}
	case etime.ThetaCycle:
		if cyc%200 == 0 {
			vu.View.GoUpdateView()
		}
	}
}

// RecordSyns records synaptic data -- stored separate from unit data
// and only needs to be called when synaptic values are updated.
// Should be done when the DWt values have been computed, before
// updating Wts and zeroing.
// NetView displays this recorded data when Update is next called.
func (vu *ViewUpdate) RecordSyns() {
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
