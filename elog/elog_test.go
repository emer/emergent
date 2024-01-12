// Copyright (c) 2022, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package elog

import (
	"testing"

	"github.com/emer/emergent/v2/etime"
	"github.com/emer/etable/v2/etensor"
)

func TestScopeKeyStringing(t *testing.T) {
	sk := etime.Scope(etime.Train, etime.Epoch)
	if sk != "Train&Epoch" {
		t.Errorf("Got unexpected scopekey " + string(sk))
	}
	sk2 := etime.Scopes([]etime.Modes{etime.Train, etime.Test}, []etime.Times{etime.Epoch, etime.Cycle})
	if sk2 != "Train|Test&Epoch|Cycle" {
		t.Errorf("Got unexpected scopekey " + string(sk2))
	}
	modes, times := sk2.ModesAndTimes()
	if len(modes) != 2 || len(times) != 2 {
		t.Errorf("Error parsing scopekey")
	}
}

func TestItem(t *testing.T) {
	item := Item{
		Name: "Testo",
		Type: etensor.STRING,
		Write: WriteMap{"Train|Test&Epoch|Cycle": func(ctx *Context) {
			// DO NOTHING
		}},
	}
	item.SetEachScopeKey()
	_, ok := item.WriteFunc("Train", "Epoch")
	if !ok {
		t.Errorf("Error getting compute function")
	}
	if item.HasMode(etime.Validate) || item.HasTime(etime.Run) {
		t.Errorf("Item has mode or time it shouldn't")
	}
}
