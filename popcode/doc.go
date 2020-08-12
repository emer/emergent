// Copyright (c) 2019, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

/*
Package `popcode` provides population code encoding and decoding
support functionality, in 1D and 2D.

`popcode.OneD` `Encode` method turns a scalar value into a 1D
population code according to a set of parameters about the nature
of the population code, range of values to encode, etc.

`Decode` takes a distributed pattern of activity and decodes a scalar
value from it, using activation-weighted average based on tuning
value of individual units.

`popcode.TwoD` likewise has `Encode` and `Decode` methods for 2D
gaussian-bumps that simultaneously encode a 2D value such as a 2D
position.

The `add` option to the Encode methods allows multiple values to be
encoded, and `DecodeNPeaks` allows multiple to be decoded, using a
neighborhood around local maxima.
*/
package popcode
