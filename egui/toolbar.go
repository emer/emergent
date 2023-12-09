// Copyright (c) 2022, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package egui

import (
	"goki.dev/gi/v2/gi"
	"goki.dev/ki/v2"
)

// ToolbarItem holds the configuration values for a toolbar item
type ToolbarItem struct {
	Label   string
	Icon    string
	Tooltip string
	Active  ToolGhosting
	Func    func()
}

// AddToolbarItem adds a toolbar item but also checks when it be active in the UI
func (gui *GUI) AddToolbarItem(item ToolbarItem) {
	switch item.Active {
	case ActiveStopped:
		gui.ToolBar.AddAction(gi.ActOpts{Label: item.Label, Icon: item.Icon, Tooltip: item.Tooltip, UpdateFunc: func(act *gi.Action) {
			act.SetActiveStateUpdt(!gui.IsRunning)
		}}, gui.Win.This(), func(recv, send ki.Ki, sig int64, data any) {
			item.Func()
		})
	case ActiveRunning:
		gui.ToolBar.AddAction(gi.ActOpts{Label: item.Label, Icon: item.Icon, Tooltip: item.Tooltip, UpdateFunc: func(act *gi.Action) {
			act.SetActiveStateUpdt(gui.IsRunning)
		}}, gui.Win.This(), func(recv, send ki.Ki, sig int64, data any) {
			item.Func()
		})
	case ActiveAlways:
		gui.ToolBar.AddAction(gi.ActOpts{Label: item.Label, Icon: item.Icon, Tooltip: item.Tooltip}, gui.Win.This(),
			func(recv, send ki.Ki, sig int64, data any) {
				item.Func()
			})
	}
}
