// Copyright (c) 2022, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package elog

import (
	"fmt"
	"log"

	"github.com/emer/emergent/emer"
	"github.com/emer/emergent/estats"
	"github.com/emer/emergent/etime"
	"github.com/emer/etable/agg"
	"github.com/emer/etable/etable"
	"github.com/emer/etable/etensor"
	"github.com/emer/etable/metric"
)

// WriteFunc function that computes and sets log values
// The Context provides information typically needed for logging
type WriteFunc func(ctx *Context)

// Context provides the context for logging Write functions.
// SetContext must be called on Logs to set the Stats and Net values
// Provides various convenience functions for setting log values
// and other commonly-used operations.
type Context struct {

	// pointer to the Logs object with all log data
	Logs *Logs `desc:"pointer to the Logs object with all log data"`

	// pointer to stats
	Stats *estats.Stats `desc:"pointer to stats"`

	// network
	Net emer.Network `desc:"network"`

	// data parallel index for accessing data from network
	Di int `desc:"data parallel index for accessing data from network"`

	// current log Item
	Item *Item `desc:"current log Item"`

	// current scope key
	Scope etime.ScopeKey `desc:"current scope key"`

	// current scope eval mode (if standard)
	Mode etime.Modes `desc:"current scope eval mode (if standard)"`

	// current scope timescale (if standard)
	Time etime.Times `desc:"current scope timescale (if standard)"`

	// LogTable with extra data for the table
	LogTable *LogTable `desc:"LogTable with extra data for the table"`

	// current table to record value to
	Table *etable.Table `desc:"current table to record value to"`

	// current row in table to write to
	Row int `desc:"current row in table to write to"`
}

// SetTable sets the current table & scope -- called by WriteItems
func (ctx *Context) SetTable(sk etime.ScopeKey, lt *LogTable, row int) {
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

// SetFloat64Cells sets float64 values to tensor cell
// in current table, item, row
func (ctx *Context) SetFloat64Cells(vals []float64) {
	for i, v := range vals {
		ctx.Table.SetCellTensorFloat1D(ctx.Item.Name, ctx.Row, i, v)
	}
}

///////////////////////////////////////////////////
//  Aggregation, data access

// SetAgg sets an aggregated value computed from given eval mode
// and time scale with same Item name, to current item, row.
// Supports scalar or tensor cells.
// returns aggregated value(s).
func (ctx *Context) SetAgg(mode etime.Modes, time etime.Times, ag agg.Aggs) []float64 {
	return ctx.SetAggScope(etime.Scope(mode, time), ag)
}

// SetAggScope sets an aggregated value computed from
// another scope (ScopeKey) with same Item name, to current item, row.
// Supports scalar or tensor cells.
// returns aggregated value(s).
func (ctx *Context) SetAggScope(scope etime.ScopeKey, ag agg.Aggs) []float64 {
	return ctx.SetAggItemScope(scope, ctx.Item.Name, ag)
}

// SetAggItem sets an aggregated value computed from given eval mode
// and time scale with given Item name, to current item, row.
// Supports scalar or tensor cells.
// returns aggregated value(s).
func (ctx *Context) SetAggItem(mode etime.Modes, time etime.Times, itemNm string, ag agg.Aggs) []float64 {
	return ctx.SetAggItemScope(etime.Scope(mode, time), itemNm, ag)
}

// SetAggItemScope sets an aggregated value computed from
// another scope (ScopeKey) with given Item name, to current item, row.
// Supports scalar or tensor cells.
// returns aggregated value(s).
func (ctx *Context) SetAggItemScope(scope etime.ScopeKey, itemNm string, ag agg.Aggs) []float64 {
	ix := ctx.Logs.IdxViewScope(scope)
	vals := agg.Agg(ix, itemNm, ag)
	if len(vals) == 0 {
		fmt.Printf("elog.Context SetAggItemScope for item: %s in scope: %s -- could not aggregate item: %s from scope: %s -- check names\n", ctx.Item.Name, ctx.Scope, itemNm, scope)
		return nil
	}
	cl, err := ctx.Table.ColByNameTry(ctx.Item.Name)
	if err != nil {
		log.Println(err)
		return vals
	}
	if cl.NumDims() > 1 {
		ctx.SetFloat64Cells(vals)
	} else {
		ctx.SetFloat64(vals[0])
	}
	return vals
}

// ItemFloat returns a float64 value of the last row of given item name
// in log for given mode, time
func (ctx *Context) ItemFloat(mode etime.Modes, time etime.Times, itemNm string) float64 {
	return ctx.ItemFloatScope(etime.Scope(mode, time), itemNm)
}

// ItemFloatScope returns a float64 value of the last row of given item name
// in log for given scope.
func (ctx *Context) ItemFloatScope(scope etime.ScopeKey, itemNm string) float64 {
	dt := ctx.Logs.TableScope(scope)
	if dt.Rows == 0 {
		return 0
	}
	return dt.CellFloat(itemNm, dt.Rows-1)
}

// ItemString returns a string value of the last row of given item name
// in log for given mode, time
func (ctx *Context) ItemString(mode etime.Modes, time etime.Times, itemNm string) string {
	return ctx.ItemStringScope(etime.Scope(mode, time), itemNm)
}

// ItemStringScope returns a string value of the last row of given item name
// in log for given scope.
func (ctx *Context) ItemStringScope(scope etime.ScopeKey, itemNm string) string {
	dt := ctx.Logs.TableScope(scope)
	if dt.Rows == 0 {
		return ""
	}
	return dt.CellString(itemNm, dt.Rows-1)
}

// ItemTensor returns an etensor.Tensor of the last row of given item name
// in log for given mode, time
func (ctx *Context) ItemTensor(mode etime.Modes, time etime.Times, itemNm string) etensor.Tensor {
	return ctx.ItemTensorScope(etime.Scope(mode, time), itemNm)
}

// ItemTensorScope returns an etensor.Tensor of the last row of given item name
// in log for given scope.
func (ctx *Context) ItemTensorScope(scope etime.ScopeKey, itemNm string) etensor.Tensor {
	dt := ctx.Logs.TableScope(scope)
	if dt.Rows == 0 {
		return nil
	}
	return dt.CellTensor(itemNm, dt.Rows-1)
}

// ItemColTensor returns an etensor.Tensor of the entire column of given item name
// in log for given mode, time
func (ctx *Context) ItemColTensor(mode etime.Modes, time etime.Times, itemNm string) etensor.Tensor {
	return ctx.ItemColTensorScope(etime.Scope(mode, time), itemNm)
}

// ItemColTensorScope returns an etensor.Tensor of the entire column of given item name
// in log for given scope.
func (ctx *Context) ItemColTensorScope(scope etime.ScopeKey, itemNm string) etensor.Tensor {
	dt := ctx.Logs.TableScope(scope)
	return dt.ColByName(itemNm)
}

///////////////////////////////////////////////////
//  Network

// Layer returns layer by name as the emer.Layer interface --
// you may then need to convert to a concrete type depending.
func (ctx *Context) Layer(layNm string) emer.Layer {
	return ctx.Net.LayerByName(layNm)
}

// GetLayerTensor gets tensor of Unit values on a layer for given variable
// from current ctx.Di data parallel index.
func (ctx *Context) GetLayerTensor(layNm, unitVar string) *etensor.Float32 {
	ly := ctx.Layer(layNm)
	tsr := ctx.Stats.F32Tensor(layNm)
	ly.UnitValsTensor(tsr, unitVar, ctx.Di)
	return tsr
}

// GetLayerRepTensor gets tensor of representative Unit values on a layer for given variable
// from current ctx.Di data parallel index.
func (ctx *Context) GetLayerRepTensor(layNm, unitVar string) *etensor.Float32 {
	ly := ctx.Layer(layNm)
	tsr := ctx.Stats.F32Tensor(layNm)
	ly.UnitValsRepTensor(tsr, unitVar, ctx.Di)
	return tsr
}

// SetLayerTensor sets tensor of Unit values on a layer for given variable
// to current ctx.Di data parallel index.
func (ctx *Context) SetLayerTensor(layNm, unitVar string) *etensor.Float32 {
	tsr := ctx.GetLayerTensor(layNm, unitVar)
	ctx.SetTensor(tsr)
	return tsr
}

// SetLayerRepTensor sets tensor of representative Unit values on a layer for given variable
// to current ctx.Di data parallel index.
func (ctx *Context) SetLayerRepTensor(layNm, unitVar string) *etensor.Float32 {
	tsr := ctx.GetLayerRepTensor(layNm, unitVar)
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
func (ctx *Context) LastNRows(mode etime.Modes, time etime.Times, n int) *etable.IdxView {
	return ctx.LastNRowsScope(etime.Scope(mode, time), n)
}

// LastNRowsScope returns an IdxView onto table for given scope with the last
// n rows of the table (only valid rows, if less than n).
// This index view is available later with the "LastNRows" name via
// NamedIdxView functions.
func (ctx *Context) LastNRowsScope(sk etime.ScopeKey, n int) *etable.IdxView {
	ix, isnew := ctx.Logs.NamedIdxViewScope(sk, "LastNRows")
	if !isnew {
		return ix
	}
	if n > ix.Len()-1 {
		n = ix.Len() - 1
	}
	if ix.Idxs == nil {
		ix.Idxs = make([]int, ix.Table.Rows)
	}
	ix.Idxs = ix.Idxs[ix.Len()-n:]
	return ix
}
