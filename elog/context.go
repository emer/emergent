// Copyright (c) 2022, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package elog

import (
	"fmt"

	"github.com/Astera-org/models/library/estats"
	"github.com/emer/emergent/emer"
	"github.com/emer/etable/agg"
	"github.com/emer/etable/etable"
	"github.com/emer/etable/etensor"
	"github.com/emer/etable/metric"
)

// WriteFunc function that computes and sets log values
// The Context provides information typically needed for logging
type WriteFunc func(ctxt *Context)

// Context provides the context for logging Write functions.
// SetContext must be called on Logs to set the Stats and Net values
// Provides various convenience functions for setting log values
// and other commonly-used operations.
type Context struct {
	Logs     *Logs         `desc:"pointer to the Logs object with all log data"`
	Stats    *estats.Stats `desc:"pointer to stats"`
	Net      emer.Network  `desc:"network"`
	Item     *Item         `desc:"current log Item"`
	Scope    ScopeKey      `desc:"current scope key"`
	Mode     EvalModes     `desc:"current scope eval mode (if standard)"`
	Time     Times         `desc:"current scope timescale (if standard)"`
	LogTable *LogTable     `desc:"LogTable with extra data for the table"`
	Table    *etable.Table `desc:"current table to record value to"`
	Row      int           `desc:"current row in table to write to"`
}

// SetTable sets the current table & scope -- called by WriteItems
func (ctx *Context) SetTable(sk ScopeKey, lt *LogTable, row int) {
	ctx.Scope = sk
	ctx.LogTable = lt
	ctx.Table = lt.Table
	ctx.Row = row
	ctx.Mode, ctx.Time = sk.ModeAndTime()
}

// SetFloat64 sets a float64 to current table, item, row
func (ctx *Context) SetFloat64(val float64) {
	ctx.Table.SetCellFloat(ctx.Item.Name, ctx.Row, val)
}

// SetFloat32 sets a float32 to current table, item, row
func (ctx *Context) SetFloat32(val float32) {
	ctx.Table.SetCellFloat(ctx.Item.Name, ctx.Row, float64(val))
}

// SetInt sets an int to current table, item, row
func (ctx *Context) SetInt(val int) {
	ctx.Table.SetCellFloat(ctx.Item.Name, ctx.Row, float64(val))
}

// SetString sets a string to current table, item, row
func (ctx *Context) SetString(val string) {
	ctx.Table.SetCellString(ctx.Item.Name, ctx.Row, val)
}

// SetStatFloat sets a Stats Float of given name to current table, item, row
func (ctx *Context) SetStatFloat(name string) {
	ctx.Table.SetCellFloat(ctx.Item.Name, ctx.Row, ctx.Stats.Float(name))
}

// SetStatInt sets a Stats int of given name to current table, item, row
func (ctx *Context) SetStatInt(name string) {
	ctx.Table.SetCellFloat(ctx.Item.Name, ctx.Row, float64(ctx.Stats.Int(name)))
}

// SetStatString sets a Stats string of given name to current table, item, row
func (ctx *Context) SetStatString(name string) {
	ctx.Table.SetCellString(ctx.Item.Name, ctx.Row, ctx.Stats.String(name))
}

// SetTensor sets a Tensor to current table, item, row
func (ctx *Context) SetTensor(val etensor.Tensor) {
	ctx.Table.SetCellTensor(ctx.Item.Name, ctx.Row, val)
}

///////////////////////////////////////////////////
//  Aggregation, data access

// SetAgg sets an aggregated scalar value computed from given eval mode
// and time scale with same Item name, to current item, row.
// returns aggregated value
func (ctx *Context) SetAgg(mode EvalModes, time Times, ag agg.Aggs) float64 {
	return ctx.SetAggScope(Scope(mode, time), ag)
}

// SetAggScope sets an aggregated scalar value computed from
// another scope (ScopeKey) with same Item name, to current item, row
// returns aggregated value
func (ctx *Context) SetAggScope(scope ScopeKey, ag agg.Aggs) float64 {
	return ctx.SetAggItemScope(scope, ctx.Item.Name, ag)
}

// SetAggItem sets an aggregated scalar value computed from given eval mode
// and time scale with given Item name, to current item, row.
// returns aggregated value
func (ctx *Context) SetAggItem(mode EvalModes, time Times, itemNm string, ag agg.Aggs) float64 {
	return ctx.SetAggItemScope(Scope(mode, time), itemNm, ag)
}

// SetAggItemScope sets an aggregated scalar value computed from
// another scope (ScopeKey) with given Item name, to current item, row.
// returns aggregated value
func (ctx *Context) SetAggItemScope(scope ScopeKey, itemNm string, ag agg.Aggs) float64 {
	ix := ctx.Logs.IdxViewScope(scope)
	vals := agg.Agg(ix, itemNm, ag)
	if len(vals) == 0 {
		fmt.Printf("elog.Context SetAggItemScope for item: %s in scope: %s -- could not aggregate item: %s from scope: %s -- check names\n", ctx.Item.Name, ctx.Scope, itemNm, scope)
		return 0
	}
	ctx.SetFloat64(vals[0])
	return vals[0]
}

// ItemFloat returns a float64 value of the last row of given item name
// in log for given mode, time
func (ctx *Context) ItemFloat(mode EvalModes, time Times, itemNm string) float64 {
	return ctx.ItemFloatScope(Scope(mode, time), itemNm)
}

// ItemFloatScope returns a float64 value of the last row of given item name
// in log for given scope.
func (ctx *Context) ItemFloatScope(scope ScopeKey, itemNm string) float64 {
	dt := ctx.Logs.TableScope(scope)
	return dt.CellFloat(itemNm, dt.Rows-1)
}

// ItemString returns a string value of the last row of given item name
// in log for given mode, time
func (ctx *Context) ItemString(mode EvalModes, time Times, itemNm string) string {
	return ctx.ItemStringScope(Scope(mode, time), itemNm)
}

// ItemStringScope returns a string value of the last row of given item name
// in log for given scope.
func (ctx *Context) ItemStringScope(scope ScopeKey, itemNm string) string {
	dt := ctx.Logs.TableScope(scope)
	return dt.CellString(itemNm, dt.Rows-1)
}

///////////////////////////////////////////////////
//  Network

// Layer returns layer by name as the emer.Layer interface --
// you may then need to convert to a concrete type depending.
func (ctx *Context) Layer(layNm string) emer.Layer {
	return ctx.Net.LayerByName(layNm)
}

// SetLayerTensor sets tensor of Unit values on a layer for given variable
func (ctx *Context) SetLayerTensor(layNm, unitVar string) *etensor.Float32 {
	ly := ctx.Layer(layNm)
	tsr := ctx.Stats.F32Tensor(layNm)
	ly.UnitValsTensor(tsr, unitVar)
	ctx.SetTensor(tsr)
	return tsr
}

// ClosestPat finds the closest pattern in given column of given pats table to
// given layer activation pattern using given variable.  Returns the row number,
// correlation value, and value of a column named namecol for that row if non-empty.
// Column must be etensor.Float32
func (ctx *Context) ClosestPat(layNm, unitVar string, pats *etable.Table, colnm, namecol string) (int, float32, string) {
	tsr := ctx.SetLayerTensor(layNm, unitVar)
	col := pats.ColByName(colnm)
	// note: requires Increasing metric so using Inv
	row, cor := metric.ClosestRow32(tsr, col.(*etensor.Float32), metric.InvCorrelation32)
	cor = 1 - cor // convert back to correl
	nm := ""
	if namecol != "" {
		nm = pats.CellString(namecol, row)
	}
	return row, cor, nm
}

///////////////////////////////////////////////////
//  IdxViews

// LastNRows returns an IdxView onto table for given scope with the last
// n rows of the table (only valid rows, if less than n).
// This index view is available later with the "LastNRows" name via
// NamedIdxView functions.
func (ctx *Context) LastNRows(mode EvalModes, time Times, n int) *etable.IdxView {
	return ctx.LastNRowsScope(Scope(mode, time), n)
}

// LastNRowsScope returns an IdxView onto table for given scope with the last
// n rows of the table (only valid rows, if less than n).
// This index view is available later with the "LastNRows" name via
// NamedIdxView functions.
func (ctx *Context) LastNRowsScope(sk ScopeKey, n int) *etable.IdxView {
	ix, isnew := ctx.Logs.NamedIdxViewScope(sk, "LastNRows")
	if !isnew {
		return ix
	}
	if n > ix.Len()-1 {
		n = ix.Len() - 1
	}
	ix.Idxs = ix.Idxs[ix.Len()-n:]
	return ix
}
