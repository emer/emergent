// Copyright (c) 2019, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package esg

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
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
	rls.RunRandSeed = 10
	rls.Init(0)
	// rls.Trace = true
	genstr := ""
	for i := 0; i < 50; i++ {
		str := rls.Gen()
		genstr += fmt.Sprintf("%v\n", str)
		// fmt.Println(str)
	}

	ex := `[schoolgirl ate food with daintiness]
[someone ate steak with utensil]
[child consumed food with knife]
[teacher ate icecream with spoon]
[teacher consumed soup with daintiness]
[pitcherpers consumed soup in kitchen]
[child ate food with finger]
[pitcherpers ate steak with knife]
[pitcherpers ate icecream in park]
[pitcherpers ate crackers with gusto]
[pitcherpers ate icecream with pleasure]
[someone ate soup with spoon]
[someone ate soup in kitchen]
[pitcherpers ate icecream with pleasure]
[child ate something with finger]
[teacher ate soup with crackers]
[teacher ate something in park]
[teacher ate soup with utensil]
[pitcherpers ate food with finger]
[pitcherpers ate steak in kitchen]
[pitcherpers ate steak with gusto]
[pitcherpers ate crackers with something]
[pitcherpers ate steak with schoolgirl]
[pitcherpers ate icecream in park]
[someone consumed something with jelly]
[teacher ate steak in kitchen]
[teacher ate icecream with spoon]
[teacher ate soup with daintiness]
[adult ate something in kitchen]
[someone ate food with crackers]
[adult ate food with utensil]
[teacher consumed something with adult]
[teacher consumed soup with daintiness]
[teacher consumed soup with daintiness]
[schoolgirl ate food with pitcherpers]
[someone consumed crackers with jelly]
[schoolgirl ate icecream with someone]
[schoolgirl consumed crackers with finger]
[adult ate soup with crackers]
[teacher ate food with daintiness]
[someone ate food with crackers]
[someone ate something with something]
[teacher ate soup in kitchen]
[teacher ate crackers with daintiness]
[teacher ate soup in kitchen]
[teacher ate crackers with daintiness]
[pitcherpers ate food in kitchen]
[pitcherpers ate icecream in park]
[pitcherpers ate soup with gusto]
[teacher ate food with daintiness]
`

	assert.Equal(t, ex, genstr)

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
