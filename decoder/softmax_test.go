// Copyright (c) 2021, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package decoder

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSoftMax(t *testing.T) {
	dec := SoftMax{}
	dec.Init(2, 2)
	dec.Lrate = .1
	for i := 0; i < 100; i++ {
		trg := 0
		if i%2 == 0 {
			dec.Inputs[0] = 1
			dec.Inputs[1] = 0
		} else {
			trg = 1
			dec.Inputs[0] = 0
			dec.Inputs[1] = 1
		}
		dec.Forward()
		dec.Sort()
		if i > 2 {
			if dec.Sorted[0] != trg {
				t.Errorf("err: %d\t%d\t%v\n", i, trg, dec.Sorted)
			}
		}
		dec.Train(trg)
	}
}

func TestSoftMaxSaveLoad(t *testing.T) {
	dec := SoftMax{}
	dec.Init(2, 2)
	// Train it.
	dec.Lrate = .1
	for i := 0; i < 100; i++ {
		target := 0
		dec.Inputs[0] = 1
		dec.Inputs[1] = 0
		if i%2 == 0 {
			target = 1
			dec.Inputs[0] = 0
			dec.Inputs[1] = 1
		}
		dec.Forward()
		dec.Train(target)
	}
	zeroWeights := make([]float32, len(dec.Weights.Values))
	assert.NotEqual(t, zeroWeights, dec.Weights.Values)
	// Save and load it.
	tempDir := t.TempDir()
	for _, suffix := range []string{".gz", ".json"} {
		path := filepath.Join(tempDir, "test"+suffix)
		assert.NoError(t, dec.Save(path))
		dec2 := SoftMax{}
		assert.ErrorContains(t, dec2.Load(path), "length")
		dec2.Init(2, 2)
		assert.NoError(t, dec2.Load(path))
		assert.Equal(t, dec.Weights.Values, dec2.Weights.Values)
	}
}
