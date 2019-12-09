// Copyright (c) 2019, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package sg

import (
	"fmt"
	"testing"
)

func TestParse(t *testing.T) {
	rls := &Rules{Name: "test"}
	errs := rls.OpenRules("testdata/testrules.txt")
	if errs != nil {
		t.Error("parsing errors occured as logged above")
	}
	// str := rls.String()
	// fmt.Println(str)
	errs = rls.Validate()
	if errs != nil {
		t.Error("validation errors occured as logged above")
	}
}

func TestGen(t *testing.T) {
	rls := &Rules{Name: "test"}
	errs := rls.OpenRules("testdata/testrules.txt")
	if errs != nil {
		t.Error("parsing errors occured as logged above")
	}
	errs = rls.Validate()
	if errs != nil {
		t.Error("validation errors occured as logged above")
	}
	for i := 0; i < 100; i++ {
		str := rls.Gen()
		fmt.Println(str)
	}
}
