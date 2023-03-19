// Copyright (c) 2019, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package params

// The params.History interface records history of parameters applied
// to a given object.
type History interface {
	// ParamsHistoryReset resets parameter application history
	ParamsHistoryReset()

	// ParamsApplied is called when a parameter is successfully applied for given selector
	ParamsApplied(sel *Sel)
}

// HistoryImpl implements the History interface.  Implementing object can
// just pass calls to a HistoryImpl field.
type HistoryImpl []*Sel

// ParamsHistoryReset resets parameter application history
func (hi *HistoryImpl) ParamsHistoryReset() {
	*hi = nil
}

// ParamsApplied is called when a parameter is successfully applied for given selector
func (hi *HistoryImpl) ParamsApplied(sel *Sel) {
	*hi = append(*hi, sel)
}

// ParamsHistory returns the sequence of params applied for each parameter
// from all Sel's applied, in reverse order
func (hi *HistoryImpl) ParamsHistory() Params {
	pr := make(Params)
	lastSet := ""
	for _, sl := range *hi {
		for pt, v := range sl.Params {
			nmv := sl.Sel + ": " + v
			if sl.SetName != lastSet {
				nmv = sl.SetName + ":" + nmv
				lastSet = sl.SetName
			}
			ev, has := pr[pt]
			if has {
				pr[pt] = nmv + " | " + ev
			} else {
				pr[pt] = nmv
			}
		}
	}
	return pr
}
