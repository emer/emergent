// Copyright (c) 2019, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package emer

import (
	"log"
)

// LayNames is a list of layer names.
// Has convenience methods for adding, validating.
type LayNames []string

// Validate ensures that LayNames layers are valid.
// ctxt is string for error message to provide context.
func (ln *LayNames) Validate(net Network, ctxt string) error {
	var lasterr error
	for _, lnm := range *ln {
		_, err := net.LayerByNameTry(lnm)
		if err != nil {
			log.Printf("%s LayNames.Validate: %v\n", ctxt, err)
			lasterr = err
		}
	}
	return lasterr
}

// Add adds given layer name(s) to list
func (ln *LayNames) Add(laynm ...string) {
	*ln = append(*ln, laynm...)
}

// AddAllBut adds all layers in network except those in exlude list
func (ln *LayNames) AddAllBut(net Network, excl []string) {
	exmap := make(map[string]struct{})
	for _, ex := range excl {
		exmap[ex] = struct{}{}
	}
	*ln = nil
	nl := net.NLayers()
	for li := 0; li < nl; li++ {
		aly := net.Layer(li)
		nm := aly.Name()
		if _, on := exmap[nm]; on {
			continue
		}
		ln.Add(nm)
	}
}
