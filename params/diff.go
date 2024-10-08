// Copyright (c) 2019, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package params

import (
	"fmt"
	"sort"

	"golang.org/x/exp/maps"
)

// DiffsAll reports all the cases where the same param path is being set
// to different values across different sets
func (ps *Sets) DiffsAll() string {
	pd := ""
	sz := len(*ps)
	keys := maps.Keys(*ps)
	sort.Strings(keys)
	for i, sNm := range keys {
		set := (*ps)[sNm]
		for j := i + 1; j < sz; j++ {
			osNm := keys[j]
			oset := (*ps)[osNm]
			spd := set.Diffs(oset, sNm, osNm)
			if spd != "" {
				pd += "//////////////////////////////////////\n"
				pd += spd
			}
		}
	}
	return pd
}

// DiffsFirst reports all the cases where the same param path is being set
// to different values between the "Base" set and all other sets.
// Only works if there is a set named "Base".
func (ps *Sets) DiffsFirst() string {
	pd := ""
	sz := len(*ps)
	if sz < 2 {
		return ""
	}
	set, ok := (*ps)["Base"]
	if !ok {
		return "params.DiffsFirst: Set named 'Base' not found\n"
	}
	keys := maps.Keys(*ps)
	sort.Strings(keys)
	for _, sNm := range keys {
		if sNm == "Base" {
			continue
		}
		oset := (*ps)[sNm]
		spd := set.Diffs(oset, "Base", sNm)
		if spd != "" {
			pd += "//////////////////////////////////////\n"
			pd += spd
		}
	}
	return pd
}

// DiffsWithin reports all the cases where the same param path is being set
// to different values within given sheet.
func (ps *Sets) DiffsWithin(sheetName string) string {
	sht, err := ps.SheetByName(sheetName)
	if err != nil {
		return err.Error()
	}
	return sht.DiffsWithin(sheetName)
}

/////////////////////////////////////////////////////////
//   Sheet

// Diffs reports all the cases where the same param path is being set
// to different values between this sheeet and the other sheeet.
func (ps *Sheet) Diffs(ops *Sheet, setNm1, setNm2 string) string {
	pd := ""
	for _, sel := range *ps {
		for _, osel := range *ops {
			spd := sel.Params.Diffs(&sel.Params, setNm1+":"+sel.Sel, setNm2+":"+osel.Sel)
			pd += spd
		}
	}
	return pd
}

// DiffsWithin reports all the cases where the same param path is being set
// to different values within different Sel's in this Sheet.
func (ps *Sheet) DiffsWithin(sheetName string) string {
	pd := ""
	sz := len(*ps)
	for i, sel := range *ps {
		for j := i + 1; j < sz; j++ {
			osel := (*ps)[j]
			spd := sel.Params.Diffs(&sel.Params, sheetName+":"+sel.Sel, sheetName+":"+osel.Sel)
			pd += spd
		}
	}
	return pd
}

/////////////////////////////////////////////////////////
//   Params

// Diffs returns comparison between all params in this params
// versus the other params, where the path is the same but the
// parameter value is different.  Nm1 is the name / id of the
// 'this' Params, and nm2 is for the other params.
func (pr *Params) Diffs(op *Params, nm1, nm2 string) string {
	pd := ""
	for pt, pv := range *pr {
		for opt, opv := range *op {
			if pt == opt && pv != opv {
				pd += fmt.Sprintf("%s:\t\t %s = %v \t|\t %s = %v,\n", pt, nm1, pv, nm2, opv)
			}
		}
	}
	return pd
}
