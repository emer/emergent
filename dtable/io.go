// Copyright (c) 2019, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package dtable

import (
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/apache/arrow/go/arrow"
	"github.com/emer/emergent/etensor"
	"github.com/goki/gi/gi"
)

// SaveCSV writes a table to a comma-separated-values (CSV) file (where comma = any delimiter,
// specified in the delim arg).
// If headers = true then generate C++ emergent-tyle column headers and add _H: to the header line
// and _D: to the data lines.  These headers have full configuration information for the tensor
// columns.  Otherwise, only the data is written.
func (dt *Table) SaveCSV(filename gi.FileName, delim rune, headers bool) error {
	fp, err := os.Create(string(filename))
	defer fp.Close()
	if err != nil {
		log.Println(err)
		return err
	}
	dt.WriteCSV(fp, delim, headers)
	return nil
}

// OpenCSV reads a table from a comma-separated-values (CSV) file (where comma = any delimiter,
// specified in the delim arg), using the Go standard encoding/csv reader conforming
// to the official CSV standard.
// If the table does not currently have any columns, the first row of the file is assumed to be
// headers, and columns are constructed therefrom.  We parse the C++ emergent column
// headers, if the first line starts with _H: -- these have full configuration information for tensor
// dimensionality, and are also supported for writing using WriteCSV.
// If the table DOES have existing columns, then those are used robustly for whatever information
// fits from each row of the file.
func (dt *Table) OpenCSV(filename gi.FileName, delim rune) error {
	fp, err := os.Open(string(filename))
	defer fp.Close()
	if err != nil {
		log.Println(err)
		return err
	}
	return dt.ReadCSV(fp, delim)
}

// ReadCSV reads a table from a comma-separated-values (CSV) file (where comma = any delimiter,
// specified in the delim arg), using the Go standard encoding/csv reader conforming
// to the official CSV standard.
// If the table does not currently have any columns, the first row of the file is assumed to be
// headers, and columns are constructed therefrom.  We parse the C++ emergent column
// headers, if the first line starts with _H: -- these have full configuration information for tensor
// dimensionality, and are also supported for writing using WriteCSV.
// If the table DOES have existing columns, then those are used robustly for whatever information
// fits from each row of the file.
func (dt *Table) ReadCSV(r io.Reader, delim rune) error {
	cr := csv.NewReader(r)
	if delim != 0 {
		cr.Comma = delim
	}
	rec, err := cr.ReadAll() // todo: lazy, avoid resizing
	if err != nil {
		return err
	}
	rows := len(rec)
	// cols := len(rec[0])
	strow := 0
	if dt.NumCols() == 0 {
		sc, err := SchemaFromHeaders(rec[0])
		if err != nil {
			log.Println(err.Error())
			return err
		}
		strow++
		rows--
		dt.SetFromSchema(sc, rows)
	}
	tc := dt.NumCols()
	dt.SetNumRows(rows)
rowloop:
	for ri := 0; ri < rows; ri++ {
		ci := 0
		rr := rec[ri+strow]
		if rr[0] == "_D:" { // emergent data row
			ci++
		}
		for j := 0; j < tc; j++ {
			tsr := dt.Cols[j]
			_, cells := tsr.RowCellSize()
			stoff := ri * cells
			for cc := 0; cc < cells; cc++ {
				str := rr[ci]
				tsr.SetFlatString(stoff+cc, str)
				ci++
				if ci >= len(rr) {
					continue rowloop
				}
			}
		}
	}
	return nil
}

// SchemaFromHeaders attempts to configure a Table Schema based on the headers
func SchemaFromHeaders(hdrs []string) (Schema, error) {
	if hdrs[0] == "_H:" {
		return SchemaFromEmerHeaders(hdrs)
	}
	return nil, fmt.Errorf("dtable.SchemaFromHeaders: only emergent header format currently supported")
}

// SchemaFromEmerHeaders attempts to configure a Table Schema based on emergent DataTable headers
func SchemaFromEmerHeaders(hdrs []string) (Schema, error) {
	nc := len(hdrs) - 1
	sc := Schema{}
	for ci := 0; ci < nc; ci++ {
		hd := hdrs[ci+1]
		if hd == "" {
			continue
		}
		var typ arrow.Type
		typ, hd = EmerColType(hd)
		dimst := strings.Index(hd, "]<")
		if dimst > 0 {
			dims := hd[dimst+2 : len(hd)-1]
			lbst := strings.Index(hd, "[")
			hd = hd[:lbst]
			csh := ShapeFromString(dims)
			// new tensor starting
			sc = append(sc, Column{Name: hd, Type: etensor.Type(typ), CellShape: csh})
			continue
		}
		dimst = strings.Index(hd, "[")
		if dimst > 0 {
			continue
		}
		sc = append(sc, Column{Name: hd, Type: etensor.Type(typ), CellShape: nil})
	}
	return sc, nil
}

var EmerHdrCharToType = map[byte]arrow.Type{
	'$': arrow.STRING,
	'%': arrow.FLOAT32,
	'#': arrow.FLOAT64,
	'|': arrow.INT64,
	'@': arrow.UINT8,
	'&': arrow.STRING,
	'^': arrow.BOOL,
}

var EmerHdrTypeToChar map[arrow.Type]byte

func init() {
	EmerHdrTypeToChar = make(map[arrow.Type]byte)
	for k, v := range EmerHdrCharToType {
		if k != '&' {
			EmerHdrTypeToChar[v] = k
		}
	}
	EmerHdrTypeToChar[arrow.INT8] = '@'
	EmerHdrTypeToChar[arrow.INT16] = '|'
	EmerHdrTypeToChar[arrow.UINT16] = '|'
	EmerHdrTypeToChar[arrow.INT32] = '|'
	EmerHdrTypeToChar[arrow.UINT32] = '|'
	EmerHdrTypeToChar[arrow.UINT64] = '|'
}

// EmerColType parses the column header for type information using the emergent naming convention
func EmerColType(nm string) (arrow.Type, string) {
	typ, ok := EmerHdrCharToType[nm[0]]
	if ok {
		nm = nm[1:]
	} else {
		typ = arrow.STRING // most general, default
	}
	return typ, nm
}

// ShapeFromString parses string representation of shape as N:d,d,..
func ShapeFromString(dims string) []int {
	clni := strings.Index(dims, ":")
	nd, _ := strconv.Atoi(dims[:clni])
	sh := make([]int, nd)
	ci := clni + 1
	for i := 0; i < nd; i++ {
		dstr := ""
		if i < nd-1 {
			nci := strings.Index(dims[ci:], ",")
			dstr = dims[ci : ci+nci]
			ci += nci + 1
		} else {
			dstr = dims[ci:]
		}
		d, _ := strconv.Atoi(dstr)
		sh[i] = d
	}
	return sh
}

//////////////////////////////////////////////////////////////////////////
// WriteCSV

// WriteCSV writes a table to a comma-separated-values (CSV) file (where comma = any delimiter,
//  specified in the delim arg).
// If headers = true then generate C++ emergent-tyle column headers and add _H: to the header line
// and _D: to the data lines.  These headers have full configuration information for the tensor
// columns.  Otherwise, only the data is written.
func (dt *Table) WriteCSV(w io.Writer, delim rune, headers bool) error {
	cw := csv.NewWriter(w)
	if delim != 0 {
		cw.Comma = delim
	}
	hcap := 0
	if headers {
		hdrs := dt.EmerHeaders()
		err := cw.Write(hdrs)
		if err != nil {
			return err
		}
		hcap = len(hdrs)
	} else {
		hcap = 100
	}
	rec := make([]string, 0, hcap)
	for ri := 0; ri < dt.Rows; ri++ {
		rc := 0
		if headers {
			vl := "_D:"
			if len(rec) <= rc {
				rec = append(rec, vl)
			} else {
				rec[rc] = vl
			}
			rc++
		}
		for i := range dt.Cols {
			tsr := dt.Cols[i]
			nd := tsr.NumDims()
			if nd == 1 {
				vl := tsr.FlatStringVal(ri)
				if len(rec) <= rc {
					rec = append(rec, vl)
				} else {
					rec[rc] = vl
				}
				rc++
			} else {
				csh := etensor.NewShape(tsr.Shapes()[1:], nil, nil) // cell shape
				tc := csh.Len()
				for ti := 0; ti < tc; ti++ {
					vl := tsr.FlatStringVal(ri*tc + ti)
					if len(rec) <= rc {
						rec = append(rec, vl)
					} else {
						rec[rc] = vl
					}
					rc++
				}
			}
		}
		err := cw.Write(rec)
		if err != nil {
			return err
		}
	}
	cw.Flush()
	return nil
}

// EmerHeaders generates emergent DataTable header strings from the table.
// These have full information about type and tensor cell dimensionality.
// Also includes the _H: header marker typically output to indicate a header row as first element.
func (dt *Table) EmerHeaders() []string {
	hdrs := []string{"_H:"}
	for i := range dt.Cols {
		tsr := dt.Cols[i]
		nm := dt.ColNames[i]
		nm = string([]byte{EmerHdrTypeToChar[tsr.DataType().ID()]}) + nm
		if tsr.NumDims() == 1 {
			hdrs = append(hdrs, nm)
		} else {
			csh := etensor.NewShape(tsr.Shapes()[1:], nil, nil) // cell shape
			tc := csh.Len()
			nd := csh.NumDims()
			fnm := nm + fmt.Sprintf("[%v:", nd)
			dn := fmt.Sprintf("<%v:", nd)
			ffnm := fnm
			for di := 0; di < nd; di++ {
				ffnm += "0"
				dn += fmt.Sprintf("%v", csh.Dim(di))
				if di < nd-1 {
					ffnm += ","
					dn += ","
				}
			}
			ffnm += "]" + dn + ">"
			hdrs = append(hdrs, ffnm)
			for ti := 1; ti < tc; ti++ {
				idx := csh.Index(ti)
				ffnm := fnm
				for di := 0; di < nd; di++ {
					ffnm += fmt.Sprintf("%v", idx[di])
					if di < nd-1 {
						ffnm += ","
					}
				}
				ffnm += "]"
				hdrs = append(hdrs, ffnm)
			}
		}
	}
	return hdrs
}
