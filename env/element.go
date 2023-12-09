// Copyright (c) 2019, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package env

import (
	"goki.dev/etable/v2/etable"
)

// Element specifies one element of State or Action in an environment
type Element struct {

	// name of this element -- must be unique
	Name string

	// shape of the tensor for this element -- each element should generally have a well-defined consistent shape to enable the model to process it consistently
	Shape []int

	// names of the dimensions within the Shape -- optional but useful for ensuring correct usage
	DimNames []string
}

// Elements is a list of Element info
type Elements []Element

// FromSchema copies element data from a etable Schema that describes an
// etable.Table
func (ch *Elements) FromSchema(sc etable.Schema) {
	*ch = make(Elements, len(sc))
	for i, cl := range sc {
		(*ch)[i].FromColumn(&cl)
	}
}

// FromColumn copies element data from etable Column that describes an
// etable.Table
func (ch *Element) FromColumn(sc *etable.Column) {
	ch.Name = sc.Name
	ch.Shape = make([]int, len(sc.CellShape))
	copy(ch.Shape, sc.CellShape)
	if sc.DimNames != nil {
		ch.DimNames = make([]string, len(sc.DimNames))
		copy(ch.DimNames, sc.DimNames)
	} else {
		ch.DimNames = nil
	}
}
