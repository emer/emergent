// Code generated by "core generate -add-types"; DO NOT EDIT.

package stepper

import (
	"cogentcore.org/core/enums"
)

var _RunStateValues = []RunState{0, 1, 2, 3}

// RunStateN is the highest valid value for type RunState, plus one.
const RunStateN RunState = 4

var _RunStateValueMap = map[string]RunState{`Stopped`: 0, `Paused`: 1, `Stepping`: 2, `Running`: 3}

var _RunStateDescMap = map[RunState]string{0: ``, 1: ``, 2: ``, 3: ``}

var _RunStateMap = map[RunState]string{0: `Stopped`, 1: `Paused`, 2: `Stepping`, 3: `Running`}

// String returns the string representation of this RunState value.
func (i RunState) String() string { return enums.String(i, _RunStateMap) }

// SetString sets the RunState value from its string representation,
// and returns an error if the string is invalid.
func (i *RunState) SetString(s string) error {
	return enums.SetString(i, s, _RunStateValueMap, "RunState")
}

// Int64 returns the RunState value as an int64.
func (i RunState) Int64() int64 { return int64(i) }

// SetInt64 sets the RunState value from an int64.
func (i *RunState) SetInt64(in int64) { *i = RunState(in) }

// Desc returns the description of the RunState value.
func (i RunState) Desc() string { return enums.Desc(i, _RunStateDescMap) }

// RunStateValues returns all possible values for the type RunState.
func RunStateValues() []RunState { return _RunStateValues }

// Values returns all possible values for the type RunState.
func (i RunState) Values() []enums.Enum { return enums.Values(_RunStateValues) }

// MarshalText implements the [encoding.TextMarshaler] interface.
func (i RunState) MarshalText() ([]byte, error) { return []byte(i.String()), nil }

// UnmarshalText implements the [encoding.TextUnmarshaler] interface.
func (i *RunState) UnmarshalText(text []byte) error { return enums.UnmarshalText(i, text, "RunState") }
