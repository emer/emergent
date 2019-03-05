// Copyright (c) 2019, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package dtable

import (
	"fmt"

	"github.com/emer/emergent/etensor"
	"github.com/goki/ki/ints"
	"github.com/goki/ki/ki"
	"github.com/goki/ki/kit"
)

// dtable.Table is the DataTable structure, containing columns of etensor tensors.
// All tensors MUST have RowMajor stride layout!
type Table struct {
	Cols       []etensor.Tensor `view:"no-inline" desc:"columns of data, as etensor.Tensor tensors"`
	ColNames   []string         `desc:"the names of the columns"`
	Rows       int              `inactive:"+" desc:"number of rows, which is enforced to be the size of the outer-most dimension of the column tensors"`
	ColNameMap map[string]int   `view:"-" desc:"the map of column names to column numbers"`
}

var KiT_Table = kit.Types.AddType(&Table{}, TableProps)

// NumRows returns the number of rows (arrow / dframe api)
func (dt *Table) NumRows() int {
	return dt.Rows
}

// NumCols returns the number of columns (arrow / dframe api)
func (dt *Table) NumCols() int {
	return len(dt.Cols)
}

// Col returns the tensor at given column index
func (dt *Table) Col(i int) etensor.Tensor {
	return dt.Cols[i]
}

// Schema returns the Schema (column properties) for this table
func (dt *Table) Schema() Schema {
	nc := dt.NumCols()
	sc := make(Schema, nc)
	for i := range dt.Cols {
		cl := &sc[i]
		tsr := dt.Cols[i]
		cl.Name = dt.ColNames[i]
		cl.Type = etensor.Type(tsr.DataType().ID())
		cl.CellShape = tsr.Shapes()[1:]
		cl.DimNames = tsr.DimNames()[1:]
	}
	return sc
}

// ColNameIndex returns the index of the given column name, along with an error if not found
func (dt *Table) ColNameIndex(name string) (int, error) {
	i, ok := dt.ColNameMap[name]
	if !ok {
		return 0, fmt.Errorf("dtable.Table ColNameIndex: column named: %v not found", name)
	}
	return i, nil
}

// ColByName returns the tensor at given column name without any error messages -- just
// returns nil if not found
func (dt *Table) ColByName(name string) etensor.Tensor {
	i, ok := dt.ColNameMap[name]
	if !ok {
		return nil
	}
	return dt.Cols[i]
}

// ColByNameTry returns the tensor at given column name, if not found, returns error
func (dt *Table) ColByNameTry(name string) (etensor.Tensor, error) {
	i, err := dt.ColNameIndex(name)
	if err != nil {
		return nil, err
	}
	return dt.Cols[i], nil
}

// ColName returns the name of given column
func (dt *Table) ColName(i int) string {
	return dt.ColNames[i]
}

// UpdateColNameMap updates the column name map
func (dt *Table) UpdateColNameMap() {
	nc := dt.NumCols()
	dt.ColNameMap = make(map[string]int, nc)
	for i := range dt.ColNames {
		dt.ColNameMap[dt.ColNames[i]] = i
	}
}

// AddCol adds the given tensor as a column to the table.
// returns error if it is not a RowMajor organized tensor, and automatically
// adjusts the shape to fit the current number of rows.
func (dt *Table) AddCol(tsr etensor.Tensor, name string) error {
	if !tsr.IsRowMajor() {
		return fmt.Errorf("tensor must be RowMajor organized")
	}
	dt.Cols = append(dt.Cols, tsr)
	dt.ColNames = append(dt.ColNames, name)
	dt.UpdateColNameMap()
	tsr.SetNumRows(dt.Rows)
	return nil
}

// AddRows adds n rows to each of the columns
func (dt *Table) AddRows(n int) {
	for _, tsr := range dt.Cols {
		tsr.AddRows(n)
	}
	dt.Rows += n
}

// SetNumRows sets the number of rows in the table, across all columns
// if rows = 0 then effective number of rows in tensors is 1, as this dim cannot be 0
func (dt *Table) SetNumRows(rows int) {
	dt.Rows = rows // can be 0
	rows = ints.MaxInt(1, rows)
	for _, tsr := range dt.Cols {
		tsr.SetNumRows(rows)
	}
}

// SetFromSchema configures table from given Schema.
// The actual tensor number of rows is enforced to be > 0, because we
// cannot have a null dimension in tensor shape.
// does not preserve any existing columns / data.
func (dt *Table) SetFromSchema(sc Schema, rows int) {
	nc := len(sc)
	dt.Cols = make([]etensor.Tensor, nc)
	dt.ColNames = make([]string, nc)
	dt.Rows = rows // can be 0
	rows = ints.MaxInt(1, rows)
	for i := range dt.Cols {
		cl := &sc[i]
		dt.ColNames[i] = cl.Name
		sh := append([]int{rows}, cl.CellShape...)
		dn := append([]string{"row"}, cl.DimNames...)
		tsr := etensor.New(cl.Type, sh, nil, dn)
		dt.Cols[i] = tsr
	}
	dt.UpdateColNameMap()
}

// New returns a new Table constructed from given Schema.
// The actual tensor number of rows is enforced to be > 0, because we
// cannot have a null dimension in tensor shape
func New(sc Schema, rows int) *Table {
	dt := &Table{}
	dt.SetFromSchema(sc, rows)
	return dt
}

//////////////////////////////////////////////////////////////////////////////////////
//  Table props for gui

var TableProps = ki.Props{
	"ToolBar": ki.PropSlice{
		{"OpenCSV", ki.Props{
			"label": "Open CSV File...",
			"icon":  "file-open",
			"desc":  "Open CSV-formatted data (or any delimeter -- default is tab (9), comma = 44) -- also recognizes emergent-style headers",
			"Args": ki.PropSlice{
				{"File Name", ki.Props{}},
				{"Delimiter", ki.Props{
					"default": ',',
					"desc":    "can use any single-character rune here -- default is tab (9) b/c otherwise hard to type, comma = 44",
				}},
			},
		}},
		{"SaveCSV", ki.Props{
			"label": "Save CSV File...",
			"icon":  "file-save",
			"desc":  "Save CSV-formatted data (or any delimiter -- default is tab (9), comma = 44) -- header outputs emergent-style header data",
			"Args": ki.PropSlice{
				{"File Name", ki.Props{}},
				{"Delimiter", ki.Props{
					"default": '\t',
					"desc":    "can use any single-character rune here -- default is tab (9) b/c otherwise hard to type, comma = 44",
				}},
				{"Headers", ki.Props{
					"desc": "output C++ emergent-style headers that have type and tensor geometry information",
				}},
			},
		}},
		{"sep-file", ki.BlankProp{}},
		{"AddRows", ki.Props{
			"icon": "new",
			"Args": ki.PropSlice{
				{"N Rows", ki.Props{
					"default": 1,
				}},
			},
		}},
		{"SetNumRows", ki.Props{
			"label": "Set N Rows",
			"icon":  "new",
			"Args": ki.PropSlice{
				{"N Rows", ki.Props{
					"default-field": "Rows",
				}},
			},
		}},
	},
}
