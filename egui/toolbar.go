// Copyright (c) 2022, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package egui

import (
	"cogentcore.org/core/core"
	"cogentcore.org/core/events"
	"cogentcore.org/core/icons"
	"cogentcore.org/core/styles"
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
func (gui *GUI) AddToolbarItem(p *core.Plan, item ToolbarItem) {
	core.AddAt(p, item.Label, func(w *core.Button) {
		w.SetText(item.Label).SetIcon(item.Icon).
			SetTooltip(item.Tooltip).OnClick(func(e events.Event) {
			item.Func()
		})
		switch item.Active {
		case ActiveStopped:
			w.FirstStyler(func(s *styles.Style) { s.SetEnabled(!gui.IsRunning) })
		case ActiveRunning:
			w.FirstStyler(func(s *styles.Style) { s.SetEnabled(gui.IsRunning) })
		}
	})
}
