// Copyright (c) 2019, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package relpos

import (
	"fmt"
	"testing"

	"cogentcore.org/core/math32"
)

func TestRels(t *testing.T) {
	rp := Pos{}
	rp.Defaults()
	rp.Rel = RightOf
	rp.YAlign = Center
	rp.SetPos(math32.Vector3{}, math32.Vec2(10, 10), math32.Vec2(4, 4))
	fmt.Printf("rp: %v rs: %v\n", rp, rp.Pos)
	rp.YAlign = Front
	rp.SetPos(math32.Vector3{}, math32.Vec2(10, 10), math32.Vec2(4, 4))
	fmt.Printf("rp: %v rs: %v\n", rp, rp.Pos)
	rp.YAlign = Back
	rp.SetPos(math32.Vector3{}, math32.Vec2(10, 10), math32.Vec2(4, 4))
	fmt.Printf("rp: %v rs: %v\n", rp, rp.Pos)
}
