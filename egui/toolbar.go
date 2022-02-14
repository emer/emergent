// Copyright (c) 2022, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package egui

type ToolbarItem struct {
	Label   string
	Icon    string
	Tooltip string
	Active  ToolGhosting
	Func    func()
}
