// Copyright (c) 2022, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package egui

import (
	"goki.dev/gi/v2/gi"
	"goki.dev/girl/states"
	"goki.dev/girl/styles"
	"goki.dev/goosi/events"
	"goki.dev/icons"
)

// ToolbarItem holds the configuration values for a toolbar item
type ToolbarItem struct {
	Label   string
	Icon    icons.Icon
	Tooltip string
	Active  ToolGhosting
	Func    func()
}

// AddToolbarItem adds a toolbar item but also checks when it be active in the UI
func (gui *GUI) AddToolbarItem(item ToolbarItem) {
	itm := gi.NewButton(gui.Toolbar).SetText(item.Label).SetIcon(item.Icon).
		SetTooltip(item.Tooltip).OnClick(func(e events.Event) {
		item.Func()
	})
	switch item.Active {
	case ActiveStopped:
		itm.Style(func(s *styles.Style) {
			s.State.SetFlag(gui.IsRunning, states.Disabled)
		})
	case ActiveRunning:
		itm.Style(func(s *styles.Style) {
			s.State.SetFlag(!gui.IsRunning, states.Disabled)
		})
	}
}
