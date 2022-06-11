// Copyright (c) 2022, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package elog

import (
	"github.com/emer/emergent/etime"
	"github.com/emer/etable/agg"
	"github.com/emer/etable/etable"
	"github.com/emer/etable/etensor"
	"github.com/emer/etable/minmax"
	"github.com/emer/etable/split"
)

// AddCounterItems adds given Int counters, and string value stats
func (lg *Logs) AddCounterItems(ctrs []etime.Times, strNames []string) {
	for _, ctr := range ctrs {
		ctrName := ctr.String() // closure
		tm := etime.AllTimes
		if ctr < etime.Epoch {
			tm = ctr
		}
		lg.AddItem(&Item{
			Name: ctrName,
			Type: etensor.INT64,
			Plot: DFalse,
			Write: WriteMap{
				etime.Scope(etime.AllModes, tm): func(ctx *Context) {
					ctx.SetStatInt(ctrName)
				}}})
	}
	for _, str := range strNames {
		strName := str // closure
		lg.AddItem(&Item{
			Name: strName,
			Type: etensor.STRING,
			Plot: DFalse,
			Write: WriteMap{
				etime.Scope(etime.AllModes, etime.AllTimes): func(ctx *Context) {
					ctx.SetStatString(strName)
				}}})
	}
}

// AddStatAggItem adds a Float64 stat that is aggregated across the 3 time scales,
// ordered from higher to lower, e.g., Run, Epoch, Trial.
// The itemName is what is saved in the table, and statName is the source
// statistic in stats at the lowest level.
func (lg *Logs) AddStatAggItem(itemName, statName string, plot DefaultBool, times ...etime.Times) {
	lg.AddItem(&Item{
		Name:   itemName,
		Type:   etensor.FLOAT64,
		Plot:   plot,
		FixMax: DTrue,
		Range:  minmax.F64{Max: 1},
		Write: WriteMap{
			etime.Scope(etime.AllModes, times[2]): func(ctx *Context) {
				ctx.SetFloat64(ctx.Stats.Float(statName))
			}, etime.Scope(etime.AllModes, times[1]): func(ctx *Context) {
				ctx.SetAgg(ctx.Mode, times[2], agg.AggMean)
			}, etime.Scope(etime.Train, times[0]): func(ctx *Context) {
				ix := ctx.LastNRows(etime.Train, times[1], 5) // cached
				ctx.SetFloat64(agg.Mean(ix, ctx.Item.Name)[0])
			}}})
}

// AddStatFloatNoAggItem adds int statistic(s) of given names
// for just one mode, time, with no aggregation.
func (lg *Logs) AddStatFloatNoAggItem(mode etime.Modes, etm etime.Times, stats ...string) {
	for _, st := range stats {
		lg.AddItem(&Item{
			Name:  st,
			Type:  etensor.FLOAT64,
			Plot:  DTrue,
			Range: minmax.F64{Min: -1},
			Write: WriteMap{
				etime.Scope(mode, etm): func(ctx *Context) {
					ctx.SetStatFloat(st)
				}}})
	}
}

// AddStatIntNoAggItem adds int statistic(s) of given names
// for just one mode, time, with no aggregation.
func (lg *Logs) AddStatIntNoAggItem(mode etime.Modes, etm etime.Times, stats ...string) {
	for _, st := range stats {
		lg.AddItem(&Item{
			Name:  st,
			Type:  etensor.FLOAT64,
			Plot:  DTrue,
			Range: minmax.F64{Min: -1},
			Write: WriteMap{
				etime.Scope(mode, etm): func(ctx *Context) {
					ctx.SetStatInt(st)
				}}})
	}
}

// InitErrStats initializes the base stats variables used for
// AddErrStatAggItems: TrlErr, FirstZero, LastZero, NZero
func (lg *Logs) InitErrStats() {
	stats := lg.Context.Stats
	if stats == nil {
		return
	}
	stats.SetFloat("TrlErr", 0.0)
	stats.SetInt("FirstZero", -1) // critical to reset to -1
	stats.SetInt("LastZero", -1)  // critical to reset to -1
	stats.SetInt("NZero", 0)
}

// AddErrStatAggItems adds Err, PctErr, PctCor items recording overall performance
// from the given statName statistic (e.g., "TrlErr") across the 3 time scales,
// ordered from higher to lower, e.g., Run, Epoch, Trial.
func (lg *Logs) AddErrStatAggItems(statName string, times ...etime.Times) {
	lg.AddItem(&Item{
		Name: "Err",
		Type: etensor.FLOAT64,
		Plot: DTrue,
		Write: WriteMap{
			etime.Scope(etime.AllModes, times[2]): func(ctx *Context) {
				ctx.SetStatFloat(statName)
			}}})
	lg.AddItem(&Item{
		Name:   "PctErr",
		Type:   etensor.FLOAT64,
		Plot:   DFalse,
		FixMax: DTrue,
		Range:  minmax.F64{Max: 1},
		Write: WriteMap{
			etime.Scope(etime.Train, times[1]): func(ctx *Context) {
				pcterr := ctx.SetAggItem(ctx.Mode, times[2], "Err", agg.AggMean)
				epc := ctx.Stats.Int("Epoch")
				if ctx.Stats.Int("FirstZero") < 0 && pcterr == 0 {
					ctx.Stats.SetInt("FirstZero", epc)
				}
				if pcterr == 0 {
					nzero := ctx.Stats.Int("NZero")
					ctx.Stats.SetInt("NZero", nzero+1)
					ctx.Stats.SetInt("LastZero", epc)
				} else {
					ctx.Stats.SetInt("NZero", 0)
				}
			}, etime.Scope(etime.Test, times[1]): func(ctx *Context) {
				ctx.SetAggItem(ctx.Mode, times[2], "Err", agg.AggMean)
			}, etime.Scope(etime.AllModes, times[0]): func(ctx *Context) {
				ix := ctx.LastNRows(ctx.Mode, times[1], 5) // cached
				ctx.SetFloat64(agg.Mean(ix, ctx.Item.Name)[0])
			}}})
	lg.AddItem(&Item{
		Name:   "PctCor",
		Type:   etensor.FLOAT64,
		Plot:   DTrue,
		FixMax: DTrue,
		Range:  minmax.F64{Max: 1},
		Write: WriteMap{
			etime.Scope(etime.AllModes, times[1]): func(ctx *Context) {
				ctx.SetFloat64(1 - ctx.ItemFloatScope(ctx.Scope, "PctErr"))
			}, etime.Scope(etime.AllModes, times[0]): func(ctx *Context) {
				ix := ctx.LastNRows(ctx.Mode, times[1], 5) // cached
				ctx.SetFloat64(agg.Mean(ix, ctx.Item.Name)[0])
			}}})

	lg.AddItem(&Item{
		Name:  "FirstZero",
		Type:  etensor.FLOAT64,
		Plot:  DTrue,
		Range: minmax.F64{Min: -1},
		Write: WriteMap{
			etime.Scope(etime.Train, times[0]): func(ctx *Context) {
				ctx.SetStatInt("FirstZero")
			}}})

	lg.AddItem(&Item{
		Name:  "LastZero",
		Type:  etensor.FLOAT64,
		Plot:  DTrue,
		Range: minmax.F64{Min: -1},
		Write: WriteMap{
			etime.Scope(etime.Train, times[0]): func(ctx *Context) {
				ctx.SetStatInt("LastZero")
			}}})

}

// AddPerTrlMSec adds a log item that records PerTrlMSec log item that records
// the time taken to process one trial.  itemName is PerTrlMSec by default.
// and times are relevant 3 times to record, ordered higher to lower,
// e.g., Run, Epoch, Trial
func (lg *Logs) AddPerTrlMSec(itemName string, times ...etime.Times) {
	lg.AddItem(&Item{
		Name: itemName,
		Type: etensor.FLOAT64,
		Plot: DFalse,
		Write: WriteMap{
			etime.Scope(etime.Train, times[1]): func(ctx *Context) {
				nm := ctx.Item.Name
				tmr := ctx.Stats.StopTimer(nm)
				trls := ctx.Logs.Table(ctx.Mode, times[2])
				tmr.N = trls.Rows
				pertrl := tmr.AvgMSecs()
				if ctx.Row == 0 {
					pertrl = 0 // first one is always inaccruate
				}
				ctx.Stats.SetFloat(nm, pertrl)
				ctx.SetFloat64(pertrl)
				tmr.ResetStart()
			}, etime.Scope(etime.AllModes, times[0]): func(ctx *Context) {
				ix := ctx.LastNRows(ctx.Mode, times[1], 5)
				ctx.SetFloat64(agg.Mean(ix, ctx.Item.Name)[0])
			}}})
}

// RunStats records descriptive values for given stats across all runs,
// at Train Run scope, saving to RunStats misc table
func (lg *Logs) RunStats(stats ...string) {
	sk := etime.Scope(etime.Train, etime.Run)
	lt := lg.TableDetailsScope(sk)
	ix, _ := lt.NamedIdxView("RunStats")

	spl := split.GroupBy(ix, []string{"RunName"})
	for _, st := range stats {
		split.Desc(spl, st)
	}
	lg.MiscTables["RunStats"] = spl.AggsToTable(etable.AddAggName)
}
