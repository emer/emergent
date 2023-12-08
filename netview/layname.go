// Copyright (c) 2019, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package netview

import (
	"goki.dev/xyz"
)

// LayName is the Layer name as a Text2D within the NetView
type LayName struct {
	xyz.Text2D

	// [view: -] our netview
	NetView *NetView `copy:"-" json:"-" xml:"-" view:"-" desc:"our netview"`
}

func (ln *LayName) HandleMouseEvents(sc *xyz.Scene) {
	/*
		ln.ConnectEvent(sc.Win, oswin.MouseEvent, gi.RegPri, func(recv, send ki.Ki, sig int64, d interface{}) {
			me := d.(*mouse.Event)
			if me.Action != mouse.Release || !sc.IsVisible() { // LayObj steals some Press so use Release
				return
			}
			nv := ln.NetView
			lay := nv.Net.LayerByName(ln.Text)
			if lay != nil {
				giv.StructViewDialog(nv.ViewportSafe(), lay, giv.DlgOpts{Title: "Layer: " + lay.Name()}, nil, nil)
			}
			me.SetProcessed()
		})
	*/
}
