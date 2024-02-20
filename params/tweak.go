// Copyright (c) 2019, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package params

import (
	"fmt"
	"slices"
	"strconv"
	"strings"

	"cogentcore.org/core/laser"
	"cogentcore.org/core/mat32"
	"golang.org/x/exp/maps"
)

func tweakVal(msd, fact, exp10 float32, isRmdr bool) float32 {
	if isRmdr {
		return mat32.Truncate(msd+fact*mat32.Pow(10, exp10), 3)
	}
	return mat32.Truncate(fact*mat32.Pow(10, exp10), 3)
}

// Tweak returns parameters to try below and above the given value.
// log: use the quasi-log scheme: 1, 2, 5, 10 etc.  Only if val is one of these vals.
// incr: use increments around current value: e.g., if .5, returns .4 and .6.
// These apply to the 2nd significant digit (remainder after most significant digit)
// if that is present in the original value.
func Tweak(v float32, log, incr bool) []float32 {
	ex := mat32.Floor(mat32.Log10(v))
	base := mat32.Pow(10, ex)
	basem1 := mat32.Pow(10, ex-1)
	fact := mat32.Round(v / base)
	msd := tweakVal(0, fact, ex, false)
	rmdr := mat32.Round((v - msd) / basem1)
	var vals []float32
	sv := fact
	isRmdr := false
	if rmdr != 0 {
		if rmdr < 0 {
			msd = tweakVal(0, fact-1, ex, false)
			rmdr = mat32.Round((v - msd) / basem1)
		}
		sv = rmdr
		ex--
		isRmdr = true
	}
	switch sv {
	case 1:
		if log {
			vals = append(vals, tweakVal(msd, 5, ex-1, isRmdr), tweakVal(msd, 2, ex, isRmdr))
		}
		if incr {
			vals = append(vals, tweakVal(msd, 9, ex-1, isRmdr))
			vals = append(vals, tweakVal(msd, 1.1, ex, isRmdr))
		}
	case 2:
		if log {
			vals = append(vals, tweakVal(msd, 1, ex, isRmdr), tweakVal(msd, 5, ex, isRmdr))
		}
		if incr {
			if !log {
				vals = append(vals, tweakVal(msd, 1, ex, isRmdr))
			}
			vals = append(vals, tweakVal(msd, 3, ex, isRmdr))
		}
	case 5:
		if log {
			vals = append(vals, tweakVal(msd, 2, ex, isRmdr), tweakVal(msd, 1, ex+1, isRmdr))
		}
		if incr {
			vals = append(vals, tweakVal(msd, 4, ex, isRmdr), tweakVal(msd, 6, ex, isRmdr))
		}
	case 9:
		vals = append(vals, tweakVal(msd, 8, ex, isRmdr), tweakVal(msd, 1, ex+1, isRmdr))
	default:
		vals = append(vals, tweakVal(msd, sv-1, ex, isRmdr), tweakVal(msd, sv+1, ex, isRmdr))
	}
	return vals
}

// Search is one parameter value to search, for float-valued params
type Search struct {
	// name of object with the parameter, from FlexVal name
	Name string

	// path to the parameter within the object
	Path string

	// value of the parameter
	Value float32
}

// TweakFromHypers uses given hyper parameters to generate a list of
// parameter values to search, based on simple Tweak values relative to the current
// default value, as specified by the .Hypers params. "Tweak" options:
// * log = logarithmic 1, 2, 5, 10 intervals
// * incr = increment by +/- ".1" (e.g., if .5, then .4, .6)
// * list of comma-delimited set of values in square brackets, e.g.: "[1.5, 1.2, 1.8]"
func TweakFromHypers(hypers Flex) []Search {
	var srch []Search
	fkeys := maps.Keys(hypers)
	slices.Sort(fkeys)
	for _, fk := range fkeys {
		fv := hypers[fk]
		hyp := fv.Obj.(Hypers)
		hkeys := maps.Keys(hyp)
		slices.Sort(hkeys)
		for _, ppath := range hkeys {
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
				fmt.Printf("TweakFromHypers float parse error: only works for float type params. Obj: %s  Param: %s  val: %s  parse error: %v\n", fv.Nm, ppath, val, err)
				continue
			}
			start := float32(f64)

			var pars []float32 // param vals to search
			if tweak[0] == '[' {
				err := laser.SetRobust(&pars, tweak)
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

			for _, par := range pars {
				srch = append(srch, Search{Name: fv.Nm, Path: ppath, Value: par})
			}
		}
	}
	return srch
}
