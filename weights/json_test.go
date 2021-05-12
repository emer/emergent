// Copyright (c) 2019, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package weights

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/rand"
	"testing"
)

func TestSaveWts(t *testing.T) {
	nw := &Network{Network: "TestNet"}
	nw.SetMetaData("Epoch", "100")
	nw.SetMetaData("TrainEnv", "ra25")
	nw.Layers = make([]Layer, 2)
	l0 := &nw.Layers[0]
	l0.Layer = "Input"
	l1 := &nw.Layers[1]
	l1.Layer = "Hidden"
	l1.SetMetaData("ActMAvg", "0.15")
	l1.SetMetaData("ActPAvg", "0.18")
	l1.Units = make(map[string][]float32)
	un := make([]float32, 10)
	for i := range un {
		un[i] = rand.Float32()
	}
	l1.Units["TrgAvg"] = un
	l1.Prjns = make([]Prjn, 1)
	pj := &l1.Prjns[0]
	pj.From = "Input"
	pj.SetMetaData("GScale", "0.333")
	pj.Rs = make([]Recv, 3)
	for ri := range pj.Rs {
		rw := &pj.Rs[ri]
		rw.Ri = ri
		rw.N = 2
		rw.Si = make([]int, rw.N)
		rw.Wt = make([]float32, rw.N)
		for si := range rw.Si {
			rw.Si[si] = si
			rw.Wt[si] = rand.Float32()
		}
	}

	nb, err := json.MarshalIndent(nw, "", "\t")
	if err != nil {
		t.Error(err)
	}
	err = ioutil.WriteFile("TestNet.wts", nb, 0644)
	if err != nil {
		t.Error(err)
	}
}

func TestOpenWts(t *testing.T) {
	nw := &Network{}
	nb, err := ioutil.ReadFile("TestNet.wts")
	if err != nil {
		t.Error(err)
	}
	err = json.Unmarshal(nb, nw)
	if err != nil {
		t.Error(err)
	}
	sb, err := json.MarshalIndent(nw, "", "\t")
	if err != nil {
		t.Error(err)
	}
	if bytes.Compare(nb, sb) != 0 {
		t.Errorf("opened, saved bytes differ!\n")
		fmt.Printf("loaded: %v\n", string(sb))
	}
}
