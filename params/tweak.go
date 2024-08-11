// Copyright (c) 2024, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package params

import (
	"fmt"
	"slices"
	"strconv"
	"strings"

	"cogentcore.org/core/base/reflectx"
	"cogentcore.org/core/math32"
	"golang.org/x/exp/maps"
)

func tweakValue(msd, fact, exp10 float32, isRmdr bool) float32 {
	if isRmdr {
		return math32.Truncate(msd+fact*math32.Pow(10, exp10), 3)
	}
	return math32.Truncate(fact*math32.Pow(10, exp10), 3)
}

// Tweak returns parameters to try below and above the given value.
// log: use the quasi-log scheme: 1, 2, 5, 10 etc.  Only if val is one of these vals.
// incr: use increments around current value: e.g., if .5, returns .4 and .6.
// These apply to the 2nd significant digit (remainder after most significant digit)
// if that is present in the original value.
func Tweak(v float32, log, incr bool) []float32 {
	ex := math32.Floor(math32.Log10(v))
	base := math32.Pow(10, ex)
	basem1 := math32.Pow(10, ex-1)
	fact := math32.Round(v / base)
	msd := tweakValue(0, fact, ex, false)
	rmdr := math32.Round((v - msd) / basem1)
	var vals []float32
	sv := fact
	isRmdr := false
	if rmdr != 0 {
		if rmdr < 0 {
			msd = tweakValue(0, fact-1, ex, false)
			rmdr = math32.Round((v - msd) / basem1)
		}
		sv = rmdr
		ex--
		isRmdr = true
	}
	switch sv {
	case 1:
		if log {
			vals = append(vals, tweakValue(msd, 5, ex-1, isRmdr), tweakValue(msd, 2, ex, isRmdr))
		}
		if incr {
			vals = append(vals, tweakValue(msd, 9, ex-1, isRmdr))
			vals = append(vals, tweakValue(msd, 1.1, ex, isRmdr))
		}
	case 2:
		if log {
			vals = append(vals, tweakValue(msd, 1, ex, isRmdr), tweakValue(msd, 5, ex, isRmdr))
		}
		if incr {
			if !log {
				vals = append(vals, tweakValue(msd, 1, ex, isRmdr))
			}
			vals = append(vals, tweakValue(msd, 3, ex, isRmdr))
		}
	case 5:
		if log {
			vals = append(vals, tweakValue(msd, 2, ex, isRmdr), tweakValue(msd, 1, ex+1, isRmdr))
		}
		if incr {
			vals = append(vals, tweakValue(msd, 4, ex, isRmdr), tweakValue(msd, 6, ex, isRmdr))
		}
	case 9:
		vals = append(vals, tweakValue(msd, 8, ex, isRmdr), tweakValue(msd, 1, ex+1, isRmdr))
	default:
		vals = append(vals, tweakValue(msd, sv-1, ex, isRmdr), tweakValue(msd, sv+1, ex, isRmdr))
	}
	return vals
}

// Tweaks holds parameter tweak values associated with one parameter selector.
// Has all the object values affected for a given parameter within one
// selector, that has a tweak hyperparameter set.
type Tweaks struct {
	// the parameter path for this param
	Param string

	// the param selector that set the specific value upon which tweak is based
	Sel *Sel

	// the search values for all objects covered by this selector
	Search []SearchValues
}

// TweaksFromHypers uses given hyper parameters to generate a list of
// parameter values to search, based on simple Tweak values relative to the current
// param starting value, as specified by the .Hypers params. "Tweak" options:
// * log = logarithmic 1, 2, 5, 10 intervals
// * incr = increment by +/- ".1" (e.g., if .5, then .4, .6)
// * list of comma-delimited set of values in square brackets, e.g.: "[1.5, 1.2, 1.8]"
// The resulting search values are organized by the most specific selector that generated
// the final parameter value, upon which the param tweak was based.
func TweaksFromHypers(hypers Flex) []*Tweaks {
	var tweaks []*Tweaks
	sels := make(map[*Sel]Flex)

	fkeys := maps.Keys(hypers)
	slices.Sort(fkeys)
	for _, fnm := range fkeys {
		flx := hypers[fnm]
		hyps := flx.Object.(Hypers)
		hkeys := maps.Keys(hyps)
		slices.Sort(hkeys)
		for _, ppath := range hkeys {
			hv := hyps[ppath]
			vl := hv["Val"]
			for _, sel := range flx.History {
				pv, has := sel.Params[ppath]
				if !has {
					continue
				}
				if vl != pv {
					continue
				}
				fm, ok := sels[sel]
				if !ok {
					fm = make(Flex)
				}
				sflex, ok := fm[fnm]
				if !ok {
					sflex = &FlexVal{}
					sflex.CopyFrom(flx)
					sflex.Object = make(Hypers)
					fm[fnm] = sflex
				}
				shyps := sflex.Object.(Hypers)
				_, ok = shyps[ppath]
				if !ok {
					shyps[ppath] = hv
				}
				sflex.Object = shyps
				sels[sel] = fm
			}
		}
	}

	slnms := make(map[string]*Sel)
	for sel := range sels {
		slnms[sel.Sel] = sel
	}
	slsort := maps.Keys(slnms)
	slices.Sort(slsort)

	for _, slnm := range slsort {
		sel := slnms[slnm]
		flx := sels[sel]
		// fmt.Println(reflectx.StringJSON(sel), "\n", reflectx.StringJSON(flx))
		var f0 *FlexVal
		for _, fv := range flx {
			if f0 == nil {
				f0 = fv
				break
			}
		}
		hyps := f0.Object.(Hypers)
		hkeys := maps.Keys(hyps)
		slices.Sort(hkeys)
		for _, ppath := range hkeys {
			twk := &Tweaks{Param: ppath, Sel: sel}
			var svals []SearchValues

			fkeys := maps.Keys(flx)
			slices.Sort(fkeys)
			for _, fk := range fkeys {
				fv := flx[fk]
				hyp := fv.Object.(Hypers)
				vals := hyp[ppath]
				tweak, ok := vals["Tweak"]
				tweak = strings.ToLower(strings.TrimSpace(tweak))
				if !ok || tweak == "" || tweak == "false" || tweak == "-" {
					continue
				}

				val, ok := vals["Val"]
				if !ok {
					continue
				}
				f64, err := strconv.ParseFloat(val, 32)
				if err != nil {
					fmt.Printf("TweakFromHypers float parse error: only works for float type params. Obj: %s  Param: %s  val: %s  parse error: %v\n", fv.Name, ppath, val, err)
					continue
				}
				start := float32(f64)

				sval := SearchValues{Name: fv.Name, Type: fv.Type, Path: ppath, Start: start}

				var pars []float32 // param vals to search
				if tweak[0] == '[' {
					err := reflectx.SetRobust(&pars, tweak)
					if err != nil {
						fmt.Println("Error processing tweak value list:", tweak, "error:", err)
						continue
					}
				} else {
					log := false
					incr := false
					if strings.Contains(tweak, "log") {
						log = true
					}
					if strings.Contains(tweak, "incr") {
						incr = true
					}
					if !log && !incr {
						fmt.Printf("Tweak value not recognized: %q\n", tweak)
						continue
					}
					pars = Tweak(start, log, incr)
				}
				if len(pars) > 0 {
					sval.Values = pars
					svals = append(svals, sval)
				}
			}
			if len(svals) > 0 {
				twk.Search = svals
				tweaks = append(tweaks, twk)
			}
		}
	}
	return tweaks
}
