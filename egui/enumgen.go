// Code generated by "core generate -add-types"; DO NOT EDIT.

package egui

import (
	"cogentcore.org/core/enums"
)

var _ToolGhostingValues = []ToolGhosting{0, 1, 2}

// ToolGhostingN is the highest valid value for type ToolGhosting, plus one.
const ToolGhostingN ToolGhosting = 3

var _ToolGhostingValueMap = map[string]ToolGhosting{`ActiveStopped`: 0, `ActiveRunning`: 1, `ActiveAlways`: 2}

var _ToolGhostingDescMap = map[ToolGhosting]string{0: ``, 1: ``, 2: ``}

var _ToolGhostingMap = map[ToolGhosting]string{0: `ActiveStopped`, 1: `ActiveRunning`, 2: `ActiveAlways`}

// String returns the string representation of this ToolGhosting value.
func (i ToolGhosting) String() string { return enums.String(i, _ToolGhostingMap) }

// SetString sets the ToolGhosting value from its string representation,
// and returns an error if the string is invalid.
func (i *ToolGhosting) SetString(s string) error {
	return enums.SetString(i, s, _ToolGhostingValueMap, "ToolGhosting")
}

// Int64 returns the ToolGhosting value as an int64.
func (i ToolGhosting) Int64() int64 { return int64(i) }

// SetInt64 sets the ToolGhosting value from an int64.
func (i *ToolGhosting) SetInt64(in int64) { *i = ToolGhosting(in) }

// Desc returns the description of the ToolGhosting value.
func (i ToolGhosting) Desc() string { return enums.Desc(i, _ToolGhostingDescMap) }

// ToolGhostingValues returns all possible values for the type ToolGhosting.
func ToolGhostingValues() []ToolGhosting { return _ToolGhostingValues }

// Values returns all possible values for the type ToolGhosting.
func (i ToolGhosting) Values() []enums.Enum { return enums.Values(_ToolGhostingValues) }

// MarshalText implements the [encoding.TextMarshaler] interface.
func (i ToolGhosting) MarshalText() ([]byte, error) { return []byte(i.String()), nil }

// UnmarshalText implements the [encoding.TextUnmarshaler] interface.
func (i *ToolGhosting) UnmarshalText(text []byte) error {
	return enums.UnmarshalText(i, text, "ToolGhosting")
}
