// Copyright (c) 2019, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package etensor

import (
	"github.com/apache/arrow/go/arrow"
	"github.com/goki/ki/kit"
)

// Type is a logical type -- the subset supported by dtable.
// This is copied directly from arrow.Type
// They can be expressed as either a primitive physical type
// (bytes or bits of some fixed size), a nested type consisting of other data types,
// or another data type (e.g. a timestamp encoded as an int64)
type Type int

const (
	// Null type having no physical storage
	NULL Type = Type(arrow.NULL)

	// Bool is a 1 bit, LSB bit-packed ordering
	BOOl Type = Type(arrow.BOOL)

	// UINT8 is an Unsigned 8-bit little-endian integer
	UINT8 Type = Type(arrow.UINT8)

	// INT8 is a Signed 8-bit little-endian integer
	INT8 Type = Type(arrow.INT8)

	// UINT16 is an Unsigned 16-bit little-endian integer
	UINT16 Type = Type(arrow.UINT16)

	// INT16 is a Signed 16-bit little-endian integer
	INT16 Type = Type(arrow.INT16)

	// UINT32 is an Unsigned 32-bit little-endian integer
	UINT32 Type = Type(arrow.UINT32)

	// INT32 is a Signed 32-bit little-endian integer
	INT32 Type = Type(arrow.INT32)

	// UINT64 is an Unsigned 64-bit little-endian integer
	UINT64 Type = Type(arrow.UINT64)

	// INT64 is a Signed 64-bit little-endian integer
	INT64 Type = Type(arrow.INT64)

	// FLOAT16 is a 2-byte floating point value
	FLOAT16 Type = Type(arrow.HALF_FLOAT)

	// FLOAT32 is a 4-byte floating point value
	FLOAT32 Type = Type(arrow.FLOAT32)

	// FLOAT64 is an 8-byte floating point value
	FLOAT64 Type = Type(arrow.FLOAT64)

	// STRING is a UTF8 variable-length string
	STRING Type = Type(arrow.STRING)

	TypeN = STRING + 1
)

//go:generate stringer -type=Type

var KiT_Type = kit.Enums.AddEnum(TypeN, false, nil)

func (ev Type) MarshalJSON() ([]byte, error)  { return kit.EnumMarshalJSON(ev) }
func (ev *Type) UnmarshalJSON(b []byte) error { return kit.EnumUnmarshalJSON(ev, b) }
