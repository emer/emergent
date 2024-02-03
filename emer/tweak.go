// Copyright (c) 2019, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package emer

import (
	"fmt"
	"strconv"
	"strings"

	"cogentcore.org/core/laser"
	"github.com/emer/emergent/v2/params"
)

// ParamTweakFunc runs through given hyper parameters and calls given function,
// for Tweak values relative to the current default value, as specified
// by the .Hypers params, "Tweak" option: log = logarithmic 1, 2, 5, 10 intervals
// incr = increment by +/- ".1" (e.g., if .5, then .4, .6), or
// just list a comma-delimited set of values in square brackets, e.g.: "[1.5, 1.2, 1.8]"
// This is useful when the model is  basically working,  and you want to
// explore whether changing any given parameter has an effect.
func ParamTweakFunc(hypers params.Flex, net Network, fun func(name, ppath string, val float32)) {
	for _, fv := range hypers {
		hyp := fv.Obj.(params.Hypers)
		for ppath, vals := range hyp {
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
				fmt.Printf("Obj: %s  Param: %s  val: %s  parse error: %v\n", fv.Nm, ppath, val, err)
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
				pars = params.Tweak(start, log, incr)
			}

			var ly Layer
			var pj Prjn
			switch fv.Type {
			case "Layer":
				ly, err = net.LayerByNameTry(fv.Nm)
				if err != nil {
					fmt.Println(err)
					continue
				}
			case "Prjn":
				pj, err = net.PrjnByNameTry(fv.Nm)
				if err != nil {
					fmt.Println(err)
					continue
				}
			}
			path := params.PathAfterType(ppath)
			for _, par := range pars {
				prs := fmt.Sprintf("%g", par)
				if ly != nil {
					err = ly.SetParam(path, prs)
				} else {
					err = pj.SetParam(path, prs)
				}
				if err != nil {
					fmt.Println(err)
					break
				}
				fun(fv.Nm, ppath, par)
			}
			prs := fmt.Sprintf("%g", start)
			if ly != nil {
				err = ly.SetParam(path, prs)
			} else {
				err = pj.SetParam(path, prs)
			}
			if err != nil {
				fmt.Println(err)
			}
		}
	}
}
