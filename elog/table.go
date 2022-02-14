// Copyright (c) 2022, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package elog

import (
	"os"

	"github.com/emer/etable/etable"
)

// LogTable contains all the data for one log table
type LogTable struct {
	Table        *etable.Table              `desc:"Actual data stored."`
	Meta         map[string]string          `desc:"arbitrary meta-data for each table, e.g., hints for plotting: Plot = false to not plot, XAxisCol, LegendCol"`
	IdxView      *etable.IdxView            `view:"-" desc:"Index View of the table -- automatically updated when a new row of data is logged to the table."`
	NamedViews   map[string]*etable.IdxView `view:"-" desc:"named index views onto the table that can be saved and used across multiple items -- these are reset to nil after a new row is written -- see NamedIdxView funtion for more details."`
	File         *os.File                   `view:"-" desc:"File to store the log into."`
	WroteHeaders bool                       `view:"-" desc:"true if headers for File have already been written"`
}

// NewLogTable returns a new LogTable entry for given table, initializing values
func NewLogTable(table *etable.Table) *LogTable {
	lt := &LogTable{Table: table}
	lt.Meta = make(map[string]string)
	lt.NamedViews = make(map[string]*etable.IdxView)
	return lt
}

// GetIdxView returns the index view for the whole table.
// It is reset to nil after log row is written, and if nil
// then it is initialized to reflect current rows.
func (lt *LogTable) GetIdxView() *etable.IdxView {
	if lt.IdxView == nil {
		lt.IdxView = etable.NewIdxView(lt.Table)
	}
	return lt.IdxView
}

// NamedIdxView returns a named Index View of the table, and true
// if this index view was newly created to show entire table (else false).
// This is used for additional data aggregation, filtering etc.
// It is reset to nil after log row is written, and if nil
// then it is initialized to reflect current rows as a starting point (returning true).
// Thus, the bool return value can be used for re-using cached indexes.
func (lt *LogTable) NamedIdxView(name string) (*etable.IdxView, bool) {
	ix, has := lt.NamedViews[name]
	isnew := false
	if !has || ix == nil {
		ix = etable.NewIdxView(lt.Table)
		lt.NamedViews[name] = ix
		isnew = true
	}
	return ix, isnew
}

// ResetIdxViews resets all IdxViews -- after log row is written
func (lt *LogTable) ResetIdxViews() {
	lt.IdxView = nil
	for nm := range lt.NamedViews {
		lt.NamedViews[nm] = nil
	}
}
