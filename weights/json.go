// Copyright (c) 2019, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package weights

import (
	"io"

	"cogentcore.org/core/base/errors"
	"cogentcore.org/core/base/iox/jsonx"
)

// Prec is the precision for weight output in text formats.
// The default is aggressive for Leabra models.
// May need to increase for other models.
var Prec = 4

// NetReadJSON reads weights for entire network in a JSON format into Network structure
func NetReadJSON(r io.Reader) (*Network, error) {
	nw := &Network{}
	err := errors.Log(jsonx.Read(nw, r))
	return nw, err
}

// LayReadJSON reads weights for layer in a JSON format into Layer structure
func LayReadJSON(r io.Reader) (*Layer, error) {
	lw := &Layer{}
	err := errors.Log(jsonx.Read(lw, r))
	return lw, err
}

// PathReadJSON reads weights for path in a JSON format into Path structure
func PathReadJSON(r io.Reader) (*Path, error) {
	pw := &Path{}
	err := errors.Log(jsonx.Read(pw, r))
	return pw, err
}
