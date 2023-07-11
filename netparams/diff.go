// Copyright (c) 2019, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package netparams

import (
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
		sheet := (*ps)[sNm]
		for j := i + 1; j < sz; j++ {
			osNm := keys[j]
			osheet := (*ps)[osNm]
			spd := sheet.Diffs(osheet, sNm, osNm)
			if spd != "" {
				pd += "//////////////////////////////////////\n"
				pd += spd
			}
		}
	}
	return pd
}

// DiffsFirst reports all the cases where the same param path is being set
// to different values between the "Base" sheet and all other sheets.
// Only works if there is a sheet named "Base".
func (ps *Sets) DiffsFirst() string {
	pd := ""
	sz := len(*ps)
	if sz < 2 {
		return ""
	}
	sheet, ok := (*ps)["Base"]
	if !ok {
		return "params.DiffsFirst: Sheet named 'Base' not found\n"
	}
	keys := maps.Keys(*ps)
	sort.Strings(keys)
	for _, sNm := range keys {
		if sNm == "Base" {
			continue
		}
		osheet := (*ps)[sNm]
		spd := sheet.Diffs(osheet, "Base", sNm)
		if spd != "" {
			pd += "//////////////////////////////////////\n"
			pd += spd
		}
	}
	return pd
}

// DiffsWithin reports all the cases where the same param path is being set
// to different values within different sheets in given sheet
func (ps *Sets) DiffsWithin(sheetName string) string {
	sheet, err := ps.SheetByNameTry(sheetName)
	if err != nil {
		return err.Error()
	}
	return sheet.DiffsWithin(sheetName)
}
