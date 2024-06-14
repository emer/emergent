// Copyright (c) 2022, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package elog

import (
	"os"

	"cogentcore.org/core/tensor/table"
)

// LogTable contains all the data for one log table
type LogTable struct {

	// Actual data stored.
	Table *table.Table

	// arbitrary meta-data for each table, e.g., hints for plotting: Plot = false to not plot, XAxisColumn, LegendCol
	Meta map[string]string

	// Index View of the table -- automatically updated when a new row of data is logged to the table.
	IndexView *table.IndexView `display:"-"`

	// named index views onto the table that can be saved and used across multiple items -- these are reset to nil after a new row is written -- see NamedIndexView funtion for more details.
	NamedViews map[string]*table.IndexView `display:"-"`

	// File to store the log into.
	File *os.File `display:"-"`

	// true if headers for File have already been written
	WroteHeaders bool `display:"-"`
}

// NewLogTable returns a new LogTable entry for given table, initializing values
func NewLogTable(dt *table.Table) *LogTable {
	lt := &LogTable{Table: dt}
	lt.Meta = make(map[string]string)
	lt.NamedViews = make(map[string]*table.IndexView)
	return lt
}

// GetIndexView returns the index view for the whole table.
// It is reset to nil after log row is written, and if nil
// then it is initialized to reflect current rows.
func (lt *LogTable) GetIndexView() *table.IndexView {
	if lt.IndexView == nil {
		lt.IndexView = table.NewIndexView(lt.Table)
	}
	return lt.IndexView
}

// NamedIndexView returns a named Index View of the table, and true
// if this index view was newly created to show entire table (else false).
// This is used for additional data aggregation, filtering etc.
// It is reset to nil after log row is written, and if nil
// then it is initialized to reflect current rows as a starting point (returning true).
// Thus, the bool return value can be used for re-using cached indexes.
func (lt *LogTable) NamedIndexView(name string) (*table.IndexView, bool) {
	ix, has := lt.NamedViews[name]
	isnew := false
	if !has || ix == nil {
		ix = table.NewIndexView(lt.Table)
		lt.NamedViews[name] = ix
		isnew = true
	}
	return ix, isnew
}

// ResetIndexViews resets all IndexViews -- after log row is written
func (lt *LogTable) ResetIndexViews() {
	lt.IndexView = nil
	for nm := range lt.NamedViews {
		lt.NamedViews[nm] = nil
	}
}
