// Copyright (c) 2019, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package params

import (
	"fmt"
	"log"

	"github.com/goki/ki/kit"
)

// HyperVals is a string-value map for storing hyperparameter values
type HyperVals map[string]string

// Hypers is a parallel structure to Params which stores information relevant
// to hyperparameter search as well as the values.
// Use the key "Val" for the default value. This is equivalant to the value in
// Params. "Min" and "Max" guid the range, and "Sigma" describes a Gaussian.
type Hypers map[string]HyperVals

// ParamByNameTry returns given parameter, by name.
// Returns error if not found.
func (pr *Hypers) ParamByNameTry(name string) (map[string]string, error) {
	vl, ok := (*pr)[name]
	if !ok {
		err := fmt.Errorf("params.Params: parameter named %v not found", name)
		log.Println(err)
		return map[string]string{}, err
	}
	return vl, nil
}

// ParamByName returns given parameter by name (just does the map access)
// Returns "" if not found -- use Try version for error
func (pr *Hypers) ParamByName(name string) map[string]string {
	return (*pr)[name]
}

// SetParamByName sets given parameter by name to given value.
// (just a wrapper around map set function)
func (pr *Hypers) SetParamByName(name string, value map[string]string) {
	(*pr)[name] = value
}

var KiT_Hypers = kit.Types.AddType(&Hypers{}, HypersProps)
