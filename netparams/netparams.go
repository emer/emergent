// Copyright (c) 2019, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package netparams

//go:generate core generate -add-types

import (
	"fmt"
	"log"

	"github.com/emer/emergent/v2/params"
)

// Sets is a collection of Sheets that can be chosen among
// depending on different desired configurations etc.  Thus, each Set
// represents a collection of different possible specific configurations,
// and different such configurations can be chosen by name to apply as desired.
type Sets map[string]*params.Sheet //git:add

// SheetByNameTry tries to find given set by name, and returns error
// if not found (also logs the error)
func (ps *Sets) SheetByNameTry(name string) (*params.Sheet, error) {
	st, ok := (*ps)[name]
	if ok {
		return st, nil
	}
	err := fmt.Errorf("params.Sets: Param Sheet named %s not found", name)
	log.Println(err)
	return nil, err
}

// SheetByName returns given sheet by name -- for use when confident
// that it exists, as a nil will return if not found with no error
func (ps *Sets) SheetByName(name string) *params.Sheet {
	return (*ps)[name]
}

// SetFloat sets the value of given parameter, in selection sel,
// in sheet and set.
func (ps *Sets) SetFloat(sheet, sel, param string, val float64) error {
	sp, err := ps.SheetByNameTry(sheet)
	if err != nil {
		return err
	}
	return sp.SetFloat(sel, param, val)
}

// SetString sets the value of given parameter, in selection sel,
// in sheet and set.  Returns error if anything is not found.
func (ps *Sets) SetString(sheet, sel, param string, val string) error {
	sp, err := ps.SheetByNameTry(sheet)
	if err != nil {
		return err
	}
	return sp.SetString(sel, param, val)
}

// ParamVal returns the value of given parameter, in selection sel,
// in sheet and set.  Returns error if anything is not found.
func (ps *Sets) ParamVal(sheet, sel, param string) (string, error) {
	sp, err := ps.SheetByNameTry(sheet)
	if err != nil {
		return "", err
	}
	return sp.ParamVal(sel, param)
}
