// Copyright (c) 2019, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package netview

import (
	"github.com/goki/gi/gi"
	"github.com/goki/gi/gi3d"
	"github.com/goki/gi/giv"
	"github.com/goki/gi/oswin"
	"github.com/goki/gi/oswin/mouse"
	"github.com/goki/ki/ki"
	"github.com/goki/ki/kit"
)

// LayName is the Layer name as a Text2D within the NetView
type LayName struct {
	gi3d.Text2D
	NetView *NetView `copy:"-" json:"-" xml:"-" view:"-" desc:"our netview"`
}

var KiT_LayName = kit.Types.AddType(&LayName{}, nil)

func (ln *LayName) ConnectEvents3D(sc *gi3d.Scene) {
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
}
