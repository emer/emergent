// Copyright (c) 2019, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package params

import "fmt"

// DiffsAll reports all the cases where the same param path is being set
// to different values across different sets
func (ps *Sets) DiffsAll() string {
	pd := ""
	sz := len(*ps)
	for i, set := range *ps {
		for j := i + 1; j < sz; j++ {
			oset := (*ps)[j]
			spd := set.Diffs(oset)
			if spd != "" {
				pd += "//////////////////////////////////////\n"
				pd += spd
			}
		}
	}
	return pd
}

// DiffsFirst reports all the cases where the same param path is being set
// to different values between the first set (e.g., the "Base" set) and
// all other sets
func (ps *Sets) DiffsFirst() string {
	pd := ""
	sz := len(*ps)
	if sz < 2 {
		return ""
	}
	set := (*ps)[0]
	for j := 1; j < sz; j++ {
		oset := (*ps)[j]
		spd := set.Diffs(oset)
		if spd != "" {
			pd += "//////////////////////////////////////\n"
			pd += spd
		}
	}
	return pd
}

// DiffsWithin reports all the cases where the same param path is being set
// to different values within different sheets in given set
func (ps *Sets) DiffsWithin(setName string) string {
	set, err := ps.SetByNameTry(setName)
	if err != nil {
		return err.Error()
	}
	return set.DiffsWithin()
}

/////////////////////////////////////////////////////////
//   Set

// Diffs reports all the cases where the same param path is being set
// to different values between this set and the other set.
func (ps *Set) Diffs(ops *Set) string {
	pd := ""
	for snm, sht := range ps.Sheets {
		for osnm, osht := range ops.Sheets {
			spd := sht.Diffs(osht, ps.Name+"."+snm, ops.Name+"."+osnm)
			pd += spd
		}
	}
	return pd
}

// DiffsWithin reports all the cases where the same param path is being set
// to different values within different sheets
func (ps *Set) DiffsWithin() string {
	return ps.Sheets.DiffsWithin()
}

/////////////////////////////////////////////////////////
//   Sheets

// DiffsWithin reports all the cases where the same param path is being set
// to different values within different sheets
func (ps *Sheets) DiffsWithin() string {
	pd := "Within Sheet Diffs (Same param path set differentially within a Sheet):\n\n"
	for snm, sht := range *ps {
		spd := sht.DiffsWithin(snm)
		pd += spd
	}
	got := false
	for snm, sht := range *ps {
		for osnm, osht := range *ps {
			spd := sht.Diffs(osht, snm, osnm)
			if !got {
				pd += "////////////////////////////////////////////////////////////////////////////////////\n"
				pd += "Between Sheet Diffs (Same param path set differentially between two Sheets):\n\n"
				got = true
			}
			pd += spd
		}
	}
	return pd
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
func (ps *Sheet) DiffsWithin(shtNm string) string {
	pd := ""
	sz := len(*ps)
	for i, sel := range *ps {
		for j := i + 1; j < sz; j++ {
			osel := (*ps)[j]
			spd := sel.Params.Diffs(&sel.Params, shtNm+":"+sel.Sel, shtNm+":"+osel.Sel)
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
