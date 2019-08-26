// Copyright (c) 2019, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package weights

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"testing"
)

func TestCppOpenWts(t *testing.T) {
	fp, err := os.Open("FaceNetworkCpp.wts")
	defer fp.Close()
	if err != nil {
		t.Error(err)
	}
	nw, err := NetReadCpp(fp)
	if err != nil {
		t.Error(err)
	}
	nb, err := json.MarshalIndent(nw, "", "\t")
	if err != nil {
		t.Error(err)
	}
	err = ioutil.WriteFile("FaceNetworkJSON.wts", nb, 0644)
	if err != nil {
		t.Error(err)
	}
}
