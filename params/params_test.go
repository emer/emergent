// Copyright (c) 2019, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package params

import (
	"testing"
)

// // DefaultParams are the initial default parameters for this simulation
// var DefaultParams = ParamStyle{
// 	{"Prjn", Params{
// 		"Prjn.Learn.Norm.On":     1,
// 		"Prjn.Learn.Momentum.On": 1,
// 		"Prjn.Learn.WtBal.On":    0,
// 	}},
// 	// "Layer": {
// 	// 	"Layer.Inhib.Layer.Gi": 1.8, // this is the default
// 	// },
// 	{"#Output", Params{
// 		"Layer.Inhib.Layer.Gi": 1.4, // this turns out to be critical for small output layer
// 	}},
// 	{".Back", Params{
// 		"Prjn.WtScale.Rel": 0.2, // this is generally quite important
// 	}},
// }

func TestParamStyleWriteGo(t *testing.T) {
	// var buf bytes.Buffer
	// DefaultParams.WriteGoCode(&buf, 0)
	// dfb := buf.Bytes()
	// dfs := string(dfb)
	// fmt.Printf("%v", dfs)
	// trg := `emer.ParamStyle{
	// {"Prjn", emer.Params{
	// 	"Prjn.Learn.WtBal.On": 0,
	// 	"Prjn.Learn.Norm.On": 1,
	// 	"Prjn.Learn.Momentum.On": 1,
	// }},
	// {"#Output", emer.Params{
	// 	"Layer.Inhib.Layer.Gi": 1.4,
	// }},
	// {".Back", emer.Params{
	// 	"Prjn.WtScale.Rel": 0.2,
	// }},
	//// `
	// // cannot compare due to map random ordering!!
	// if dfs != trg {
	// 	t.Errorf("ParamStyle output incorrect at: %v or %v!\n", strings.Index(trg, dfs), strings.Index(dfs, trg))
	// }

}
