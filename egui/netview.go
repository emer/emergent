// Copyright (c) 2022, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package egui

import (
	"github.com/emer/emergent/emer"
	"github.com/emer/emergent/netview"
	"github.com/goki/gi/gi"
)

// UpdateNetView updates the gui visualization of the network.
// Set the NetViewText field prior to updating
func (gui *GUI) UpdateNetView() {
	if gui.ViewUpdt != nil {
		gui.ViewUpdt.Update()
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
	gui.NetData.SaveJSON(gi.FileName(ndfn))
}
