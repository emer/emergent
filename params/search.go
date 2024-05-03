// Copyright (c) 2024, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package params

// SearchValues is a list of parameter values to search for one parameter
// on a given object (specified by Name), for float-valued params.
type SearchValues struct {
	// name of object with the parameter
	Name string

	// type of object with the parameter. This is a Base type name (e.g., Layer, Path),
	// that is at the start of the path in Network params.
	Type string

	// path to the parameter within the object
	Path string

	// starting value, e.g., for restoring after searching
	// before moving on to another parameter, for grid search.
	Start float32

	// values of the parameter to search
	Values []float32
}
