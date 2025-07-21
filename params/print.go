// Copyright (c) 2019, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package params

import (
	"reflect"
	"strings"

	"cogentcore.org/core/base/indent"
	"cogentcore.org/core/base/reflectx"
)

// PrintStruct returns a string representation of a struct
// for printing out parameter values. It uses standard Cogent Core
// display tags to produce results that resemble the GUI interface,
// and only includes exported fields.
// The optional filter function determines whether a field is included, based
// on the full path to the field (using . separators) and the field value.
// Indent provides the starting indent level (2 spaces).
// The optional format function returns a string representation of the value,
// if you want to override the default, which just uses
// reflectx.ToString (returning an empty string means use the default).
func PrintStruct(v any, indent int,
	filter func(path string, ft reflect.StructField, fv any) bool,
	format func(path string, ft reflect.StructField, fv any) string) string {
	return printStruct("", indent, v, filter, format)
}

func addPath(par, field string) string {
	if par == "" {
		return field
	}
	return par + "." + field
}

func printStruct(parPath string, ident int, v any,
	filter func(path string, ft reflect.StructField, fv any) bool,
	format func(path string, ft reflect.StructField, fv any) string) string {
	rv := reflectx.Underlying(reflect.ValueOf(v))
	if reflectx.IsNil(rv) {
		return ""
	}
	var b strings.Builder
	rt := rv.Type()
	nf := rt.NumField()
	var fis []int
	maxFieldW := 0
	for i := range nf {
		ft := rt.Field(i)
		if !ft.IsExported() {
			continue
		}
		fv := rv.Field(i)
		fvi := fv.Interface()
		pp := addPath(parPath, ft.Name)
		if filter != nil && !filter(pp, ft, fvi) {
			continue
		}
		fis = append(fis, i)
		maxFieldW = max(maxFieldW, len(ft.Name))
	}
	for _, i := range fis {
		fv := rv.Field(i)
		ft := rt.Field(i)
		fvi := fv.Interface()
		pp := addPath(parPath, ft.Name)
		is := indent.Spaces(ident, 2)
		printName := func() {
			b.WriteString(is)
			b.WriteString(ft.Name)
			b.WriteString(strings.Repeat(" ", 1+maxFieldW-len(ft.Name)))
		}
		ps := ""
		if reflectx.NonPointerType(ft.Type).Kind() == reflect.Struct {
			if ft.Tag.Get("display") == "inline" {
				ps = printStructInline(pp, ident+1, fvi, filter, format)
				if ps != "{  }" {
					printName()
					b.WriteString(ps)
					b.WriteString("\n")
				}
			} else {
				ps := printStruct(pp, ident+1, fvi, filter, format)
				if ps != "" {
					printName()
					b.WriteString("{\n")
					b.WriteString(ps)
					b.WriteString(is)
					b.WriteString("}\n")
				}
			}
			continue
		}
		if ps == "" && format != nil {
			ps = format(pp, ft, fvi)
			if ps != "" {
				printName()
				b.WriteString(ps + "\n")
				continue
			}
		}
		printName()
		ps = reflectx.ToString(fvi)
		b.WriteString(ps + "\n")
	}
	return b.String()
}

func printStructInline(parPath string, ident int, v any,
	filter func(path string, ft reflect.StructField, fv any) bool,
	format func(path string, ft reflect.StructField, fv any) string) string {
	rv := reflectx.Underlying(reflect.ValueOf(v))
	if reflectx.IsNil(rv) {
		return ""
	}
	var b strings.Builder
	b.WriteString("{ ")
	rt := rv.Type()
	nf := rt.NumField()
	didPrint := false
	for i := range nf {
		ft := rt.Field(i)
		if !ft.IsExported() {
			continue
		}
		fv := rv.Field(i)
		fvi := fv.Interface()
		pp := addPath(parPath, ft.Name)
		if filter != nil && !filter(pp, ft, fvi) {
			continue
		}
		printName := func() {
			if didPrint {
				b.WriteString(", ")
			}
			b.WriteString(ft.Name)
			b.WriteString(": ")
			didPrint = true
		}
		ps := ""
		if reflectx.NonPointerType(ft.Type).Kind() == reflect.Struct {
			ps = printStructInline(pp, ident+1, fvi, filter, format)
			if ps != "{  }" {
				printName()
				b.WriteString(ps)
			}
			continue
		}
		if ps == "" && format != nil {
			ps = format(pp, ft, fvi)
			if ps != "" {
				printName()
				b.WriteString(ps)
				continue
			}
		}
		printName()
		ps = reflectx.ToString(fvi)
		b.WriteString(ps)
	}
	b.WriteString(" }")
	return b.String()
}
