// Copyright (c) 2019, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package eplot

import (
	"errors"
	"log"

	"github.com/emer/etable/etable"
)

// TableXY selects two 1D columns (i.e., scalar cells) from a etable.Table
// data table to plot in a gonum plot, satisfying the plotter.XYer interface
type TableXY struct {
	Table      *etable.Table `desc:"the data table to plot from"`
	XCol, YCol int           `desc:"the indexes of the 1D tensor columns to use for the X and Y data, respectively"`
}

// NewTableXY returns a new XY plot view onto the given etable.Table, from given column indexes.
// Column indexes are enforced to be valid, with an error message if they are not.
func NewTableXY(dt *etable.Table, xcol, ycol int) (*TableXY, error) {
	txy := &TableXY{Table: dt, XCol: xcol, YCol: ycol}
	return txy, txy.Validate()
}

// NewTableXYNames returns a new XY plot view onto the given etable.Table, from given column names
// Column indexes are enforced to be valid, with an error message if they are not.
func NewTableXYNames(dt *etable.Table, xcol, ycol string) (*TableXY, error) {
	xi, err := dt.ColNameIndex(xcol)
	if err != nil {
		log.Println(err)
	}
	yi, err := dt.ColNameIndex(ycol)
	if err != nil {
		log.Println(err)
	}
	txy := &TableXY{Table: dt, XCol: xi, YCol: yi}
	return txy, err
}

// Validate returns error message if column indexes are invalid, else nil
// it also sets column indexes to 0 so nothing crashes.
func (txy *TableXY) Validate() error {
	if txy.Table == nil {
		return errors.New("eplot.TableXY table is nil")
	}
	nc := txy.Table.NumCols()
	if txy.XCol >= nc || txy.XCol < 0 {
		txy.XCol = 0
		return errors.New("eplot.TableXY XCol index invalid -- reset to 0")
	}
	if txy.YCol >= nc || txy.YCol < 0 {
		txy.YCol = 0
		return errors.New("eplot.TableXY YCol index invalid -- reset to 0")
	}
	return nil
}

// Len returns the number of rows in the table
func (txy *TableXY) Len() int {
	if txy.Table == nil {
		return 0
	}
	return txy.Table.NumRows()
}

// XY returns an x, y pair at given row in table
func (txy *TableXY) XY(row int) (x, y float64) {
	if txy.Table == nil {
		return 0, 0
	}
	return txy.Table.Cols[txy.XCol].FloatVal1D(row), txy.Table.Cols[txy.YCol].FloatVal1D(row)
}
