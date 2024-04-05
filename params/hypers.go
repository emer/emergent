// Copyright (c) 2019, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package params

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
)

// HyperValues is a string-value map for storing hyperparameter values
type HyperValues map[string]string //gti:add

// JSONString returns hyper values as a JSON formatted string
func (hv *HyperValues) JSONString() string {
	var buf bytes.Buffer
	b, _ := json.Marshal(hv)
	buf.Write(b)
	return buf.String()
}

// SetJSONString sets from a JSON_formatted string
func (hv *HyperValues) SetJSONString(str string) error {
	return json.Unmarshal([]byte(str), hv)
}

// CopyFrom copies from another HyperValues
func (hv *HyperValues) CopyFrom(cp HyperValues) {
	if *hv == nil {
		*hv = make(HyperValues, len(cp))
	}
	for k, v := range cp {
		(*hv)[k] = v
	}
}

// Hypers is a parallel structure to Params which stores information relevant
// to hyperparameter search as well as the values.
// Use the key "Val" for the default value. This is equivalant to the value in
// Params. "Min" and "Max" guid the range, and "Sigma" describes a Gaussian.
type Hypers map[string]HyperValues //gti:add

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

// SetByName sets given parameter by name to given value.
// (just a wrapper around map set function)
func (pr *Hypers) SetByName(name string, value map[string]string) {
	(*pr)[name] = value
}

// CopyFrom copies hyper vals from source
func (pr *Hypers) CopyFrom(cp Hypers) {
	if *pr == nil {
		*pr = make(Hypers, len(cp))
	}
	for path, hv := range cp {
		if shv, has := (*pr)[path]; has {
			shv.CopyFrom(hv)
		} else {
			shv := HyperValues{}
			shv.CopyFrom(hv)
			(*pr)[path] = shv
		}
	}
}

// DeleteValOnly deletes entries that only have a "Val" entry.
// This happens when applying a param Sheet using Flex params
// to compile values using styling logic
func (pr *Hypers) DeleteValOnly() {
	for path, hv := range *pr {
		if len(hv) == 1 {
			if _, has := (hv)["Val"]; has {
				delete(*pr, path)
			}
		}
	}
}
