// Copyright (c) 2022, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package egui

import (
	"github.com/emer/emergent/v2/emer"
	"github.com/emer/emergent/v2/netview"
	"goki.dev/gi"
)

// UpdateNetView updates the gui visualization of the network.
func (gui *GUI) UpdateNetView() {
	if gui.ViewUpdt != nil {
		gui.ViewUpdt.Update()
	}
}

// UpdateNetViewWhenStopped updates the gui visualization of the network.
// when stopped either via stepping or user hitting stop button.
func (gui *GUI) UpdateNetViewWhenStopped() {
	if gui.ViewUpdt != nil {
		gui.ViewUpdt.UpdateWhenStopped()
	}
}

// InitNetData initializes the NetData object to record NetView data
// when the GUI is not active
func (gui *GUI) InitNetData(net emer.Network, nrecs int) {
	gui.NetData = &netview.NetData{}
	gui.NetData.Init(net, nrecs, true, 1) // true = NoSynData, 1 = MaxData
}

// NetDataRecord records current netview data
// if InitNetData has been called and NetData exists.
func (gui *GUI) NetDataRecord(netViewText string) {
	if gui.NetData == nil {
		return
	}
	gui.NetData.Record(netViewText, -1, 100)
}

// SaveNetData saves NetData NetView data (if !nil)
// to a file named by the network name
// plus _extra name plus ".netdata.gz"
func (gui *GUI) SaveNetData(extra string) {
	if gui.NetData == nil {
		return
	}
	ndfn := gui.NetData.Net.Name() + "_" + extra + ".netdata.gz"
	gui.NetData.SaveJSON(gi.Filename(ndfn))
}
