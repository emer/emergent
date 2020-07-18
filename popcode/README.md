[GoDoc](https://godoc.org/github.com/emer/emergent/popcode)

Package `popcode` provides population code encoding and decoding support functionality, in 1D and 2D.

# popcode.OneD

`popcode.OneD` `Encode` method turns a single scalar value into a 1D population code according to a set of parameters about the nature of the population code, range of values to encode, etc.

`Decode` takes a distributed pattern of activity and decodes a scalar value from it, using activation-weighted average based on tuning value of individual units.

# popcode.TwoD

`popcode.TwoD` likewise has `Encode` and `Decode` methods for 2D gaussian-bumps that simultaneously encode a 2D value such as a 2D position.

# popcode.Ring

`popcode.Ring` is a version of `popcode.OneD` for values that wrap-around, such as an angle -- set the Min and Max to the exact values with no extra (e.g., 0, 360 for angle).

