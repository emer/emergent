// Copyright (c) 2019, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package dtable

import "github.com/emer/emergent/etensor"

// Column specifies everything about a column -- can be used for constructing tables
type Column struct {
	Name      string       `desc:"name of column -- must be unique for a table"`
	Type      etensor.Type `desc:"data type, using etensor types which are isomorphic with arrow.Type"`
	CellShape []int        `desc:"shape of a single cell in the column (i.e., without the row dimension) -- for scalars this is nil -- tensor column will add the outer row dimension to this shape"`
	DimNames  []string     `desc:"names of the dimensions within the CellShape -- 'Row' will be added to outer dimension"`
	// todo: metadata / props
}

// Schema specifies all of the columns of a table, sufficient to create the table
// It is just a slice list of Columns
type Schema []Column
