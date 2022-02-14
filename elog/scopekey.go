// Copyright (c) 2022, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package elog

import (
	"sort"
	"strings"
)

// ScopeKey the associated string representation of a scope or scopes.
// They include one or more EvalModes and one or more Times.
// It is fully extensible with arbitrary mode and time strings --
// the enums are a convenience for standard cases.
// Ultimately a single mode, time pair is used concretely, but the
// All* cases and lists of multiple can be used as a convenience
// to specify ranges
type ScopeKey string

// Like "Train|Test&Epoch|Trial"
var (
	ScopeKeySeparator = "&" // between mode and time
	ScopeKeyList      = "|" // between multiple modes, times
)

// FromScopesStr creates an associated scope merging
// the modes and times that are specified as strings
// If you modify this, also modify ModesAndTimes, below.
func (sk *ScopeKey) FromScopesStr(modes, times []string) {
	var mstr string
	var tstr string
	for _, str := range modes {
		if mstr == "" {
			mstr = str
		} else {
			mstr += ScopeKeyList + str
		}
	}
	for _, str := range times {
		if tstr == "" {
			tstr = str
		} else {
			tstr += ScopeKeyList + str
		}
	}
	*sk = ScopeKey(mstr + ScopeKeySeparator + tstr)
}

// FromScopes creates an associated scope merging
// the modes and times that are specified
// If you modify this, also modify ModesAndTimes, below.
func (sk *ScopeKey) FromScopes(modes []EvalModes, times []Times) {
	mstr := make([]string, len(modes))
	for i, mode := range modes {
		mstr[i] = mode.String()
	}
	tstr := make([]string, len(times))
	for i, time := range times {
		tstr[i] = time.String()
	}
	sk.FromScopesStr(mstr, tstr)
}

// FromScope create an associated scope from given
// standard mode and time
func (sk *ScopeKey) FromScope(mode EvalModes, time Times) {
	sk.FromScopesStr([]string{mode.String()}, []string{time.String()})
}

// FromScopeStr create an associated scope from given
// mode and time as strings
func (sk *ScopeKey) FromScopeStr(mode, time string) {
	sk.FromScopesStr([]string{mode}, []string{time})
}

// ModesAndTimes returns the mode(s) and time(s) as strings
// from the current key value.  This must be the inverse
// of FromScopesStr
func (sk *ScopeKey) ModesAndTimes() (modes, times []string) {
	skstr := strings.Split(string(*sk), ScopeKeySeparator)
	modestr := skstr[0]
	timestr := skstr[1]
	modes = strings.Split(modestr, ScopeKeyList)
	times = strings.Split(timestr, ScopeKeyList)
	return
}

// ModeAndTime returns the singular mode and time as enums from a
// concrete scope key having one of each (No* cases if not standard)
func (sk *ScopeKey) ModeAndTime() (mode EvalModes, time Times) {
	modes, times := sk.ModesAndTimes()
	if len(modes) != 1 {
		mode = NoEvalMode
	} else {
		if mode.FromString(modes[0]) != nil {
			mode = NoEvalMode
		}
	}
	if len(times) != 1 {
		time = NoTime
	} else {
		if time.FromString(times[0]) != nil {
			time = NoTime
		}
	}
	return
}

// FromScopesMap creates an associated scope key merging
// the modes and times that are specified by map of strings.
func (sk *ScopeKey) FromScopesMap(modes, times map[string]bool) {
	ml := make([]string, len(modes))
	tl := make([]string, len(times))
	idx := 0
	for m := range modes {
		ml[idx] = m
		idx++
	}
	idx = 0
	for t := range times {
		tl[idx] = t
		idx++
	}
	sk.FromScopesStr(ml, tl)
}

// ModesAndTimesMap returns maps of modes and times as strings
// parsed from the current scopekey
func (sk *ScopeKey) ModesAndTimesMap() (modes, times map[string]bool) {
	ml, tl := sk.ModesAndTimes()
	modes = make(map[string]bool)
	times = make(map[string]bool)
	for _, m := range ml {
		modes[m] = true
	}
	for _, t := range tl {
		times[t] = true
	}
	return
}

//////////////////////////////////////////////////
// Standalone funcs

// Scope generates a scope key string from one mode and time
func Scope(mode EvalModes, time Times) ScopeKey {
	var ss ScopeKey
	ss.FromScope(mode, time)
	return ss
}

// ScopeStr generates a scope key string from string
// values for mode, time
func ScopeStr(mode, time string) ScopeKey {
	var ss ScopeKey
	ss.FromScopeStr(mode, time)
	return ss
}

// Scopes generates a scope key string from multiple modes, times
func Scopes(modes []EvalModes, times []Times) ScopeKey {
	var ss ScopeKey
	ss.FromScopes(modes, times)
	return ss
}

// ScopesStr generates a scope key string from multiple modes, times
func ScopesStr(modes, times []string) ScopeKey {
	var ss ScopeKey
	ss.FromScopesStr(modes, times)
	return ss
}

// ScopesMap generates a scope key from maps of modes and times (warning: ordering is random!)
func ScopesMap(modes, times map[string]bool) ScopeKey {
	var ss ScopeKey
	ss.FromScopesMap(modes, times)
	return ss
}

// ScopeName generates a string name as just the concatenation of mode + time
// e.g., used for naming log tables
func ScopeName(mode EvalModes, time Times) string {
	return mode.String() + time.String()
}

// SortScopes sorts a list of concrete mode, time
// scopes according to the EvalModes and Times enum ordering
func SortScopes(scopes []ScopeKey) []ScopeKey {
	sort.Slice(scopes, func(i, j int) bool {
		mi, ti := scopes[i].ModeAndTime()
		mj, tj := scopes[j].ModeAndTime()
		if mi < mj {
			return true
		}
		if mi > mj {
			return false
		}
		return ti < tj
	})
	return scopes
}
