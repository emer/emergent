// Copyright (c) 2019, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package esg

import (
	"fmt"
	"math/rand"
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
	t.SkipNow()
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
	rand.Seed(10)
	genstr := ""
	for i := 0; i < 50; i++ {
		str := rls.Gen()
		genstr += fmt.Sprintf("%v\n", str)
		// fmt.Println(str)
	}

	ex := `[schoolgirl consumed food in park]
[someone ate crackers with finger]
[busdriver ate soup in kitchen]
[busdriver consumed steak in kitchen]
[busdriver ate steak with something]
[pitcherpers ate something with pleasure]
[someone consumed soup with gusto]
[pitcherpers ate steak with gusto]
[child ate food in kitchen]
[child ate food in park]
[pitcherpers ate crackers with finger]
[pitcherpers ate soup with crackers]
[pitcherpers ate icecream in park]
[pitcherpers consumed food with jelly]
[adult ate something with teacher]
[busdriver ate steak in kitchen]
[busdriver consumed steak in kitchen]
[busdriver consumed food in kitchen]
[adult ate food with gusto]
[busdriver ate steak in kitchen]
[busdriver ate steak with teacher]
[adult ate icecream in park]
[busdriver ate steak in kitchen]
[busdriver ate steak with teacher]
[adult consumed steak with utensil]
[busdriver ate soup in kitchen]
[busdriver ate icecream with spoon]
[busdriver ate steak with gusto]
[adult ate something in kitchen]
[someone ate food with teacher]
[busdriver ate icecream in park]
[adult ate crackers in kitchen]
[busdriver consumed food in kitchen]
[adult ate food with gusto]
[adult consumed soup with crackers]
[teacher ate something in kitchen]
[teacher ate soup with crackers]
[teacher consumed crackers with finger]
[someone ate food with utensil]
[teacher ate food in kitchen]
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
