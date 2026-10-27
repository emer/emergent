// Copyright (c) 2026, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package mvmemergent

import (
	"bytes"
	"testing"

	"cogentcore.org/core/mvmcore"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestSymbols checks that Go code in the mvm interpreter can use the
// emergent packages without importing them.
func TestSymbols(t *testing.T) {
	var out bytes.Buffer
	in, err := mvmcore.NewGo(&out)
	require.NoError(t, err)
	Load()
	in.ImportPackageValues(Values)
	in.AutoImportPackages()
	_, err = in.Eval(`fmt.Println(egui.ActiveRunning, netview.MinUnitHeight > 0)
cfg := &egui.BaseConfig{}
cfg.Name = "test"
fmt.Println(cfg.Name)`)
	require.NoError(t, err)
	assert.Equal(t, "ActiveRunning true\ntest\n", out.String())
}
