// Copyright (c) 2022, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package egui

import (
	"github.com/emer/emergent/emer"
	"github.com/emer/emergent/etime"
	"github.com/emer/emergent/netview"
	"github.com/goki/gi/gi"
)

// UpdateNetView updates the gui visualization of the network.
// Set the NetViewText field prior to updating
func (gui *GUI) UpdateNetView() {
	if gui.NetView != nil && gui.NetView.IsVisible() {
		gui.NetView.Record(gui.NetViewText, -1)
		// note: essential to use Go version of update when called from another goroutine
		gui.NetView.GoUpdate() // note: using counters is significantly slower..
	}
}

// UpdateNetViewCycle updates the gui visualization of the network
// at given time scale relative to given cycle.  For all times
// longer than cycle intervals (ThetaCycle and above), it returns.
// Set the NetViewText field prior to updating
func (gui *GUI) UpdateNetViewCycle(time etime.Times, cyc int) {
	switch time {
	case etime.Cycle:
		gui.UpdateNetView()
	case etime.FastSpike:
		if cyc%10 == 0 {
			gui.UpdateNetView()
		}
	case etime.GammaCycle:
		if cyc%25 == 0 {
			gui.UpdateNetView()
		}
	case etime.AlphaCycle:
		if cyc%100 == 0 {
			gui.UpdateNetView()
		}
	}
}

// InitNetData initializes the NetData object to record NetView data
// when the GUI is not active (located in egui package because of
// the NetViewText that is also recorded)
func (gui *GUI) InitNetData(net emer.Network, nrecs int) {
	gui.NetData = &netview.NetData{}
	gui.NetData.Init(net, nrecs)
}

// NetDataRecord records current netview data
// if InitNetData has been called and NetData exists.
func (gui *GUI) NetDataRecord() {
	if gui.NetData == nil {
		return
	}
	gui.NetData.Record(gui.NetViewText, -1)
}

// SaveNetData saves NetData NetView data (if !nil)
// to a file named by the network name
// plus _extra name plus ".netdata.gz"
func (gui *GUI) SaveNetData(extra string) {
	if gui.NetData == nil {
		return
	}
	ndfn := gui.NetData.Net.Name() + "_" + extra + ".netdata.gz"
	gui.NetData.SaveJSON(gi.FileName(ndfn))
}
