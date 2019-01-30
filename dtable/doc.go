// Copyright (c) 2019, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package dtable provides a DataTable structure (also known as a DataFrame)
// which is a collection of columnar data all having the same number of rows.
// Each column is an etensor.Tensor.
//
// All tensors MUST have RowMajor stride layout, and the outer-most dimension
// is the row dimension, which is enforced to be the same across all columns.
//
// The tensor columns can be individually converted to / from arrow.Tensors
// and conversion between arrow.Table is planned, along with inter-conversion
// with relevant gonum structures including the planned dframe.Frame.
//
// Native support is provided for basic CSV, TSV I/O, including the
// C++ emergent standard TSV format with full type information in the first
// row column headers.
//
// Basic data manipulations are supported: GroupBy, SortIndex, FilterIndex, Join
//
// Other relevant examples of DataTable-like structures:
// https://github.com/apache/arrow/tree/master/go/arrow Table
// http://xarray.pydata.org/en/stable/index.html
// https://pandas.pydata.org/pandas-docs/stable/reference/frame.html
// https://www.rdocumentation.org/packages/base/versions/3.4.3/topics/data.frame
// https://github.com/tobgu/qframe
// https://github.com/kniren/gota
//
package dtable
