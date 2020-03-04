// Copyright (c) 2019, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package esg

import (
	"fmt"
	"testing"
)

func TestParse(t *testing.T) {
	// t.SkipNow()
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
	// t.SkipNow()
	rls := &Rules{Name: "test"}
	errs := rls.OpenRules("testdata/testrules.txt")
	if errs != nil {
		t.Error("parsing errors occured as logged above")
	}
	errs = rls.Validate()
	if errs != nil {
		t.Error("validation errors occured as logged above")
	}
	// rls.Trace = true
	for i := 0; i < 50; i++ {
		str := rls.Gen()
		fmt.Println(str)
	}
}

// func TestGenIto(t *testing.T) {
// 	t.SkipNow()
// 	rls := &Rules{Name: "test"}
// 	errs := rls.OpenRules("testdata/ito.txt")
// 	if errs != nil {
// 		t.Error("parsing errors occured as logged above")
// 	}
// 	errs = rls.Validate()
// 	if errs != nil {
// 		t.Error("validation errors occured as logged above")
// 	}
// 	// rls.Trace = true
// 	for i := 0; i < 10; i++ {
// 		str := rls.Gen()
// 		fmt.Println(str)
// 	}
// }
