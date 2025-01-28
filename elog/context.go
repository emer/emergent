// Copyright (c) 2022, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package elog

import (
	"fmt"

	"cogentcore.org/core/base/errors"
	"github.com/emer/emergent/v2/emer"
	"github.com/emer/emergent/v2/estats"
	"github.com/emer/emergent/v2/etime"
	"github.com/emer/etensor/tensor"
	"github.com/emer/etensor/tensor/stats/metric"
	"github.com/emer/etensor/tensor/stats/stats"
	"github.com/emer/etensor/tensor/table"
)

// WriteFunc function that computes and sets log values
// The Context provides information typically needed for logging
type WriteFunc func(ctx *Context)

// Context provides the context for logging Write functions.
// SetContext must be called on Logs to set the Stats and Net values
// Provides various convenience functions for setting log values
// and other commonly used operations.
type Context struct {

	// pointer to the Logs object with all log data
	Logs *Logs

	// pointer to stats
	Stats *estats.Stats

	// network
	Net emer.Network

	// data parallel index for accessing data from network
	Di int

	// current log Item
	Item *Item

	// current scope key
	Scope etime.ScopeKey

	// current scope eval mode (if standard)
	Mode etime.Modes

	// current scope timescale (if standard)
	Time etime.Times

	// LogTable with extra data for the table
	LogTable *LogTable

	// current table to record value to
	Table *table.Table

	// current row in table to write to
	Row int
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
	ctx.Table.SetFloat(ctx.Item.Name, ctx.Row, val)
}

// SetFloat32 sets a float32 to current table, item, row
func (ctx *Context) SetFloat32(val float32) {
	ctx.Table.SetFloat(ctx.Item.Name, ctx.Row, float64(val))
}

// SetInt sets an int to current table, item, row
func (ctx *Context) SetInt(val int) {
	ctx.Table.SetFloat(ctx.Item.Name, ctx.Row, float64(val))
}

// SetString sets a string to current table, item, row
func (ctx *Context) SetString(val string) {
	ctx.Table.SetString(ctx.Item.Name, ctx.Row, val)
}

// SetStatFloat sets a Stats Float of given name to current table, item, row
func (ctx *Context) SetStatFloat(name string) {
	ctx.Table.SetFloat(ctx.Item.Name, ctx.Row, ctx.Stats.Float(name))
}

// SetStatInt sets a Stats int of given name to current table, item, row
func (ctx *Context) SetStatInt(name string) {
	ctx.Table.SetFloat(ctx.Item.Name, ctx.Row, float64(ctx.Stats.Int(name)))
}

// SetStatString sets a Stats string of given name to current table, item, row
func (ctx *Context) SetStatString(name string) {
	ctx.Table.SetString(ctx.Item.Name, ctx.Row, ctx.Stats.String(name))
}

// SetTensor sets a Tensor to current table, item, row
func (ctx *Context) SetTensor(val tensor.Tensor) {
	ctx.Table.SetTensor(ctx.Item.Name, ctx.Row, val)
}

// SetFloat64Cells sets float64 values to tensor cell
// in current table, item, row
func (ctx *Context) SetFloat64Cells(vals []float64) {
	for i, v := range vals {
		ctx.Table.SetTensorFloat1D(ctx.Item.Name, ctx.Row, i, v)
	}
}

///////////////////////////////////////////////////
//  Aggregation, data access

// SetAgg sets an aggregated value computed from given eval mode
// and time scale with same Item name, to current item, row.
// Supports scalar or tensor cells.
// returns aggregated value(s).
func (ctx *Context) SetAgg(mode etime.Modes, time etime.Times, ag stats.Stats) []float64 {
	return ctx.SetAggScope(etime.Scope(mode, time), ag)
}

// SetAggScope sets an aggregated value computed from
// another scope (ScopeKey) with same Item name, to current item, row.
// Supports scalar or tensor cells.
// returns aggregated value(s).
func (ctx *Context) SetAggScope(scope etime.ScopeKey, ag stats.Stats) []float64 {
	return ctx.SetAggItemScope(scope, ctx.Item.Name, ag)
}

// SetAggItem sets an aggregated value computed from given eval mode
// and time scale with given Item name, to current item, row.
// Supports scalar or tensor cells.
// returns aggregated value(s).
func (ctx *Context) SetAggItem(mode etime.Modes, time etime.Times, itemNm string, ag stats.Stats) []float64 {
	return ctx.SetAggItemScope(etime.Scope(mode, time), itemNm, ag)
}

// SetAggItemScope sets an aggregated value computed from
// another scope (ScopeKey) with given Item name, to current item, row.
// Supports scalar or tensor cells.
// returns aggregated value(s).
func (ctx *Context) SetAggItemScope(scope etime.ScopeKey, itemNm string, ag stats.Stats) []float64 {
	ix := ctx.Logs.IndexViewScope(scope)
	vals, err := stats.StatColumn(ix, itemNm, ag)
	if err != nil {
		fmt.Printf("elog.Context SetAggItemScope for item: %s in scope: %s: could not aggregate item: %s from scope: %s: %s\n", ctx.Item.Name, ctx.Scope, itemNm, scope, err.Error())
		return nil
	}
	cl, err := ctx.Table.ColumnByName(ctx.Item.Name)
	if errors.Log(err) != nil {
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
	return dt.Float(itemNm, dt.Rows-1)
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
	return dt.StringValue(itemNm, dt.Rows-1)
}

// ItemTensor returns an tensor.Tensor of the last row of given item name
// in log for given mode, time
func (ctx *Context) ItemTensor(mode etime.Modes, time etime.Times, itemNm string) tensor.Tensor {
	return ctx.ItemTensorScope(etime.Scope(mode, time), itemNm)
}

// ItemTensorScope returns an tensor.Tensor of the last row of given item name
// in log for given scope.
func (ctx *Context) ItemTensorScope(scope etime.ScopeKey, itemNm string) tensor.Tensor {
	dt := ctx.Logs.TableScope(scope)
	if dt.Rows == 0 {
		return nil
	}
	return dt.Tensor(itemNm, dt.Rows-1)
}

// ItemColTensor returns an tensor.Tensor of the entire column of given item name
// in log for given mode, time
func (ctx *Context) ItemColTensor(mode etime.Modes, time etime.Times, itemNm string) tensor.Tensor {
	return ctx.ItemColTensorScope(etime.Scope(mode, time), itemNm)
}

// ItemColTensorScope returns an tensor.Tensor of the entire column of given item name
// in log for given scope.
func (ctx *Context) ItemColTensorScope(scope etime.ScopeKey, itemNm string) tensor.Tensor {
	dt := ctx.Logs.TableScope(scope)
	return errors.Log1(dt.ColumnByName(itemNm))
}

///////////////////////////////////////////////////
//  Network

// Layer returns layer by name as the emer.Layer interface.
// May then need to convert to a concrete type depending.
func (ctx *Context) Layer(layNm string) emer.Layer {
	return errors.Log1(ctx.Net.AsEmer().EmerLayerByName(layNm))
}

// GetLayerTensor gets tensor of Unit values on a layer for given variable
// from current ctx.Di data parallel index.
func (ctx *Context) GetLayerTensor(layNm, unitVar string) *tensor.Float32 {
	ly := ctx.Layer(layNm)
	tsr := ctx.Stats.F32Tensor(layNm)
	ly.AsEmer().UnitValuesTensor(tsr, unitVar, ctx.Di)
	return tsr
}

// GetLayerSampleTensor gets tensor of representative Unit values on a layer for given variable
// from current ctx.Di data parallel index.
func (ctx *Context) GetLayerSampleTensor(layNm, unitVar string) *tensor.Float32 {
	ly := ctx.Layer(layNm)
	tsr := ctx.Stats.F32Tensor(layNm)
	ly.AsEmer().UnitValuesSampleTensor(tsr, unitVar, ctx.Di)
	return tsr
}

// SetLayerTensor sets tensor of Unit values on a layer for given variable
// to current ctx.Di data parallel index.
func (ctx *Context) SetLayerTensor(layNm, unitVar string) *tensor.Float32 {
	tsr := ctx.GetLayerTensor(layNm, unitVar)
	ctx.SetTensor(tsr)
	return tsr
}

// SetLayerSampleTensor sets tensor of representative Unit values on a layer for given variable
// to current ctx.Di data parallel index.
func (ctx *Context) SetLayerSampleTensor(layNm, unitVar string) *tensor.Float32 {
	tsr := ctx.GetLayerSampleTensor(layNm, unitVar)
	ctx.SetTensor(tsr)
	return tsr
}

// ClosestPat finds the closest pattern in given column of given pats table to
// given layer activation pattern using given variable.  Returns the row number,
// correlation value, and value of a column named namecol for that row if non-empty.
// Column must be tensor.Float32
func (ctx *Context) ClosestPat(layNm, unitVar string, pats *table.Table, colnm, namecol string) (int, float32, string) {
	tsr := ctx.SetLayerTensor(layNm, unitVar)
	col := errors.Log1(pats.ColumnByName(colnm))
	// note: requires Increasing metric so using Inv
	row, cor := metric.ClosestRow32(tsr, col.(*tensor.Float32), metric.InvCorrelation32)
	cor = 1 - cor // convert back to correl
	nm := ""
	if namecol != "" {
		nm = pats.StringValue(namecol, row)
	}
	return row, cor, nm
}

///////////////////////////////////////////////////
//  IndexViews

// LastNRows returns an IndexView onto table for given scope with the last
// n rows of the table (only valid rows, if less than n).
// This index view is available later with the "LastNRows" name via
// NamedIndexView functions.
func (ctx *Context) LastNRows(mode etime.Modes, time etime.Times, n int) *table.IndexView {
	return ctx.LastNRowsScope(etime.Scope(mode, time), n)
}

// LastNRowsScope returns an IndexView onto table for given scope with the last
// n rows of the table (only valid rows, if less than n).
// This index view is available later with the "LastNRows" name via
// NamedIndexView functions.
func (ctx *Context) LastNRowsScope(sk etime.ScopeKey, n int) *table.IndexView {
	return ctx.Logs.LastNRowsScope(sk, n)
}
