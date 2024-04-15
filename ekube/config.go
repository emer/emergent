// Copyright (c) 2024, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

type Config struct { //types:add

	// Dir is the directory of the model to build.
	Dir string `posarg:"0"`
}
