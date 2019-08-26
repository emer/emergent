// Copyright (c) 2019, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

/*
Package weights provides weight loading routines that parse weight files into
a temporary structure that can then be used to set weight values in the network.
This is much simpler and allows use of the standard Go json Unmarshal routines.
*/
package weights
