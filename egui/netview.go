// Copyright (c) 2022, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package egui

// UpdateNetView updates the gui visualization of the network.
func (gui *GUI) UpdateNetView() {
	if gui.ViewUpdate != nil {
		gui.ViewUpdate.Update()
	}
}

// UpdateNetViewWhenStopped updates the gui visualization of the network.
// when stopped either via stepping or user hitting stop button.
func (gui *GUI) UpdateNetViewWhenStopped() {
	if gui.ViewUpdate != nil {
		gui.ViewUpdate.UpdateWhenStopped()
	}
}
