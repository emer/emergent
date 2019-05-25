// Copyright (c) 2019, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package env

import (
	"github.com/emer/etable/etable"
)

// Channel specifies one channel of input / output data in an environment
type Channel struct {
	Name     string   `desc:"name of channel -- must be unique"`
	Shape    []int    `desc:"shape of a single cell in the column (i.e., without the row dimension) -- for scalars this is nil -- tensor column will add the outer row dimension to this shape"`
	DimNames []string `desc:"names of the dimensions within the CellShape -- 'Row' will be added to outer dimension"`
}

// Channels is a list of Channel info
type Channels []Channel

// FromSchema copies channel data from a etable Schema that describes an
// etable.Table
func (ch *Channels) FromSchema(sc etable.Schema) {
	*ch = make(Channels, len(sc))
	for i, cl := range sc {
		(*ch)[i].FromColumn(&cl)
	}
}

// FromColumn copies channel data from etable Column that describes an
// etable.Table
func (ch *Channel) FromColumn(sc *etable.Column) {
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
