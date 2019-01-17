// Copyright (c) 2019, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package emer

// note: using interface{} map here probably doesn't make sense, like it did in GoGi
// it requires managing the possible types that can be created in the interface{}
// and that adds a lot of complexity.  Simpler to just have basic fixed float32
// values and aggregations thereof.

// Params is a name-value map for floating point parameter values that can be applied
// to network layers or prjns, which is where the parameter values live
type Params map[string]float32

// ParamSet is a name-based map of Params values, each of which represents a different
// set of specific parameter values.
type ParamSet map[string]Params
