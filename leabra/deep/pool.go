// Copyright (c) 2019, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package deep

import "github.com/emer/emergent/emer"

// deep.Pool contains extra statistics used in DeepLeabra
type Pool struct {
	ActNoAttn  emer.AvgMax
	TRCBurstGe emer.AvgMax
	AttnGe     emer.AvgMax
}
