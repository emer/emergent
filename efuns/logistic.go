// Copyright (c) 2019, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package efuns

import "github.com/goki/mat32"

// Logistic is the logistic (sigmoid) function of x: 1/(1 + e^(-gain*(x-off)))
func Logistic(x, gain, off float32) float32 {
	return 1 / (1 + mat32.FastExp(-gain*(x-off)))
}
