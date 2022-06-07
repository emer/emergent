// Copyright (c) 2019, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package netview

import "github.com/emer/emergent/etime"

// ViewUpdt manages time scales for updating the NetView
type ViewUpdt struct {
	View    *NetView    `view:"-" desc:"the network view"`
	Testing bool        `view:"-" desc:"whether in testing mode -- can be set in advance to drive appropriate updating"`
	Text    string      `view:"-" desc:"text to display at the bottom of the view"`
	RastCtr int         `view:"-" desc:"raster counter used for raster mode of plotting"`
	On      bool        `desc:"toggles update of display on"`
	Train   etime.Times `desc:"at what time scale to update the display during training?"`
	Test    etime.Times `desc:"at what time scale to update the display during testing?"`
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
func (vu *ViewUpdt) Update() {
	if !vu.On || vu.View == nil || !vu.View.IsVisible() {
		return
	}
	vu.View.Record(vu.Text, vu.RastCtr)
	// note: essential to use Go version of update when called from another goroutine
	vu.View.GoUpdate()
}

// UpdateTime triggers an update at given timescale.
func (vu *ViewUpdt) UpdateTime(time etime.Times) {
	viewUpdt := vu.UpdtTime(vu.Testing)
	if viewUpdt == time || (viewUpdt < etime.Trial && time == etime.Trial) {
		vu.Update()
	}
}

// UpdateCycle triggers an update at the Cycle (Millisecond) timescale,
// using given text to display at bottom of view
func (vu *ViewUpdt) UpdateCycle(cyc int) {
	viewUpdt := vu.UpdtTime(vu.Testing)
	if viewUpdt > etime.ThetaCycle {
		return
	}
	if vu.View.Params.Raster {
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
	viewUpdt := vu.UpdtTime(vu.Testing)
	vu.RastCtr = cyc
	vu.View.Record(vu.Text, vu.RastCtr)
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
