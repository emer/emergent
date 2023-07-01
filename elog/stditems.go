// Copyright (c) 2022, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package elog

import (
	"fmt"

	"github.com/emer/emergent/emer"
	"github.com/emer/emergent/etime"
	"github.com/emer/etable/agg"
	"github.com/emer/etable/etable"
	"github.com/emer/etable/etensor"
	"github.com/emer/etable/minmax"
	"github.com/emer/etable/split"
)

// AddCounterItems adds given Int counters from Stats,
// typically by recording looper counter values to Stats.
func (lg *Logs) AddCounterItems(ctrs ...etime.Times) {
	for ci, ctr := range ctrs {
		ctrName := ctr.String() // closure
		tm := etime.AllTimes
		if ctr < etime.Epoch {
			tm = ctr
		}
		itm := lg.AddItem(&Item{
			Name: ctrName,
			Type: etensor.INT64,
			Write: WriteMap{
				etime.Scope(etime.AllModes, tm): func(ctx *Context) {
					ctx.SetStatInt(ctrName)
				}}})
		if ctr < etime.Epoch {
			for ti := ci + 1; ti < len(ctrs); ti++ {
				itm.Write[etime.Scope(etime.AllModes, ctrs[ti])] = func(ctx *Context) {
					ctx.SetStatInt(ctrName)
				}
			}
		}
	}
}

// AddStatAggItem adds a Float64 stat that is aggregated across the given time scales,
// ordered from higher to lower, e.g., Run, Epoch, Trial.
// The statName is the source statistic in stats at the lowest level,
// and is also used for the log item name.
func (lg *Logs) AddStatAggItem(statName string, times ...etime.Times) *Item {
	ntimes := len(times)
	itm := lg.AddItem(&Item{
		Name:   statName,
		Type:   etensor.FLOAT64,
		FixMin: true,
		// FixMax: true,
		Range: minmax.F64{Max: 1},
		Write: WriteMap{
			etime.Scope(etime.AllModes, times[ntimes-1]): func(ctx *Context) {
				ctx.SetFloat64(ctx.Stats.Float(statName))
			}}})
	for i := ntimes - 2; i >= 1; i-- {
		i := i
		itm.Write[etime.Scope(etime.AllModes, times[i])] = func(ctx *Context) {
			ctx.SetAgg(ctx.Mode, times[i+1], agg.AggMean)
		}
	}
	itm.Write[etime.Scope(etime.Train, times[0])] = func(ctx *Context) {
		ix := ctx.LastNRows(etime.Train, times[1], 5) // cached
		ctx.SetFloat64(agg.Mean(ix, ctx.Item.Name)[0])
	}
	return itm
}

// AddStatFloatNoAggItem adds float statistic(s) of given names
// for just one mode, time, with no aggregation.
// If another item already exists for a different mode / time, this is added
// to it so there aren't any duplicate items.
func (lg *Logs) AddStatFloatNoAggItem(mode etime.Modes, etm etime.Times, stats ...string) {
	for _, st := range stats {
		stName := st // closure
		itm, has := lg.ItemByName(stName)
		if has {
			itm.Write[etime.Scope(mode, etm)] = func(ctx *Context) {
				ctx.SetStatFloat(stName)
			}
		} else {
			lg.AddItem(&Item{
				Name:  stName,
				Type:  etensor.FLOAT64,
				Range: minmax.F64{Min: -1},
				Write: WriteMap{
					etime.Scope(mode, etm): func(ctx *Context) {
						ctx.SetStatFloat(stName)
					}}})
		}
	}
}

// AddStatIntNoAggItem adds int statistic(s) of given names
// for just one mode, time, with no aggregation.
// If another item already exists for a different mode / time, this is added
// to it so there aren't any duplicate items.
func (lg *Logs) AddStatIntNoAggItem(mode etime.Modes, etm etime.Times, stats ...string) {
	for _, st := range stats {
		stName := st // closure
		itm, has := lg.ItemByName(stName)
		if has {
			itm.Write[etime.Scope(mode, etm)] = func(ctx *Context) {
				ctx.SetStatInt(stName)
			}
		} else {
			lg.AddItem(&Item{
				Name:  stName,
				Type:  etensor.INT,
				Range: minmax.F64{Min: -1},
				Write: WriteMap{
					etime.Scope(mode, etm): func(ctx *Context) {
						ctx.SetStatInt(stName)
					}}})
		}
	}
}

// AddStatStringItem adds string stat item(s) to given mode and time (e.g., Allmodes, Trial).
// If another item already exists for a different mode / time, this is added
// to it so there aren't any duplicate items.
func (lg *Logs) AddStatStringItem(mode etime.Modes, etm etime.Times, stats ...string) {
	for _, st := range stats {
		stName := st // closure
		itm, has := lg.ItemByName(stName)
		if has {
			itm.Write[etime.Scope(mode, etm)] = func(ctx *Context) {
				ctx.SetStatString(stName)
			}
		} else {
			lg.AddItem(&Item{
				Name: stName,
				Type: etensor.STRING,
				Write: WriteMap{
					etime.Scope(mode, etm): func(ctx *Context) {
						ctx.SetStatString(stName)
					}}})
		}
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
		Name:   "Err",
		Type:   etensor.FLOAT64,
		FixMin: true,
		FixMax: true,
		Range:  minmax.F64{Max: 1},
		Write: WriteMap{
			etime.Scope(etime.AllModes, times[2]): func(ctx *Context) {
				ctx.SetStatFloat(statName)
			}}})
	lg.AddItem(&Item{
		Name:   "PctErr",
		Type:   etensor.FLOAT64,
		FixMin: true,
		FixMax: true,
		Range:  minmax.F64{Max: 1},
		Write: WriteMap{
			etime.Scope(etime.Train, times[1]): func(ctx *Context) {
				pcterr := ctx.SetAggItem(ctx.Mode, times[2], "Err", agg.AggMean)[0]
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
		FixMin: true,
		FixMax: true,
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
		Range: minmax.F64{Min: -1},
		Write: WriteMap{
			etime.Scope(etime.Train, times[0]): func(ctx *Context) {
				ctx.SetStatInt("FirstZero")
			}}})

	lg.AddItem(&Item{
		Name:  "LastZero",
		Type:  etensor.FLOAT64,
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
func (lg *Logs) AddPerTrlMSec(itemName string, times ...etime.Times) *Item {
	return lg.AddItem(&Item{
		Name: itemName,
		Type: etensor.FLOAT64,
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

// AddLayerTensorItems adds tensor recording items for given variable,
// classes of layers, mode and time (e.g., Test, Trial).
// If another item already exists for a different mode / time, this is added
// to it so there aren't any duplicate items.
// di is a data parallel index di, for networks capable of processing input patterns in parallel.
func (lg *Logs) AddLayerTensorItems(net emer.Network, varNm string, mode etime.Modes, etm etime.Times, layClasses ...string) {
	layers := net.LayersByClass(layClasses...)
	for _, lnm := range layers {
		clnm := lnm
		cly := net.LayerByName(clnm)
		itmNm := clnm + "_" + varNm
		itm, has := lg.ItemByName(itmNm)
		if has {
			itm.Write[etime.Scope(mode, etm)] = func(ctx *Context) {
				ctx.SetLayerRepTensor(clnm, varNm)
			}
		} else {
			lg.AddItem(&Item{
				Name:      itmNm,
				Type:      etensor.FLOAT32,
				CellShape: cly.RepShape().Shp,
				FixMin:    true,
				Range:     minmax.F64{Max: 1},
				Write: WriteMap{
					etime.Scope(mode, etm): func(ctx *Context) {
						ctx.SetLayerRepTensor(clnm, varNm)
					}}})
		}
	}
}

// AddCopyFromFloatItems adds items that copy from one log to another,
// adding the given prefix string to each.
// if toTimes has more than 1 item, subsequent times are AggMean aggregates of first one.
// float64 type.
func (lg *Logs) AddCopyFromFloatItems(toMode etime.Modes, toTimes []etime.Times, fmMode etime.Modes, fmTime etime.Times, prefix string, itemNames ...string) {
	for _, st := range itemNames {
		stnm := st
		tonm := prefix + st
		itm := lg.AddItem(&Item{
			Name: tonm,
			Type: etensor.FLOAT64,
			Write: WriteMap{
				etime.Scope(toMode, toTimes[0]): func(ctx *Context) {
					ctx.SetFloat64(ctx.ItemFloat(fmMode, fmTime, stnm))
				}}})
		for i := 1; i < len(toTimes); i++ {
			i := i
			itm.Write[etime.Scope(toMode, toTimes[i])] = func(ctx *Context) {
				ctx.SetAgg(ctx.Mode, toTimes[i-1], agg.AggMean)
			}
		}
	}
}

// PlotItems turns on Plot flag for given items
func (lg *Logs) PlotItems(itemNames ...string) {
	for _, nm := range itemNames {
		itm, has := lg.ItemByName(nm)
		if !has {
			fmt.Printf("elog.PlotItems: item named: %s not found\n", nm)
			continue
		}
		itm.Plot = true
	}
}

// SetFloatMinItems turns off the FixMin flag for given items
func (lg *Logs) SetFloatMinItems(itemNames ...string) {
	for _, nm := range itemNames {
		itm, has := lg.ItemByName(nm)
		if !has {
			fmt.Printf("elog.SetFloatMinItems: item named: %s not found\n", nm)
			continue
		}
		itm.FixMin = false
	}
}

// SetFloatMaxItems turns off the FixMax flag for given items
func (lg *Logs) SetFloatMaxItems(itemNames ...string) {
	for _, nm := range itemNames {
		itm, has := lg.ItemByName(nm)
		if !has {
			fmt.Printf("elog.SetFloatMaxItems: item named: %s not found\n", nm)
			continue
		}
		itm.FixMax = false
	}
}

// SetFixMaxItems sets the FixMax flag and Range Max val for given items
func (lg *Logs) SetFixMaxItems(max float64, itemNames ...string) {
	for _, nm := range itemNames {
		itm, has := lg.ItemByName(nm)
		if !has {
			fmt.Printf("elog.SetFixMaxItems: item named: %s not found\n", nm)
			continue
		}
		itm.FixMax = true
		itm.Range.Max = max
	}
}

// SetFixMinItems sets the FixMin flag and Range Min val for given items
func (lg *Logs) SetFixMinItems(min float64, itemNames ...string) {
	for _, nm := range itemNames {
		itm, has := lg.ItemByName(nm)
		if !has {
			fmt.Printf("elog.SetFixMinItems: item named: %s not found\n", nm)
			continue
		}
		itm.FixMin = true
		itm.Range.Min = min
	}
}
