// Code generated by "goki generate"; DO NOT EDIT.

package erand

import (
	"errors"
	"log"
	"strconv"
	"strings"

	"goki.dev/enums"
)

var _RndDistsValues = []RndDists{0, 1, 2, 3, 4, 5, 6}

// RndDistsN is the highest valid value
// for type RndDists, plus one.
const RndDistsN RndDists = 7

// An "invalid array index" compiler error signifies that the constant values have changed.
// Re-run the enumgen command to generate them again.
func _RndDistsNoOp() {
	var x [1]struct{}
	_ = x[Uniform-(0)]
	_ = x[Binomial-(1)]
	_ = x[Poisson-(2)]
	_ = x[Gamma-(3)]
	_ = x[Gaussian-(4)]
	_ = x[Beta-(5)]
	_ = x[Mean-(6)]
}

var _RndDistsNameToValueMap = map[string]RndDists{
	`Uniform`:  0,
	`uniform`:  0,
	`Binomial`: 1,
	`binomial`: 1,
	`Poisson`:  2,
	`poisson`:  2,
	`Gamma`:    3,
	`gamma`:    3,
	`Gaussian`: 4,
	`gaussian`: 4,
	`Beta`:     5,
	`beta`:     5,
	`Mean`:     6,
	`mean`:     6,
}

var _RndDistsDescMap = map[RndDists]string{
	0: `Uniform has a uniform probability distribution over Var = range on either side of the Mean`,
	1: `Binomial represents number of 1&#39;s in n (Par) random (Bernouli) trials of probability p (Var)`,
	2: `Poisson represents number of events in interval, with event rate (lambda = Var) plus Mean`,
	3: `Gamma represents maximum entropy distribution with two parameters: scaling parameter (Var) and shape parameter k (Par) plus Mean`,
	4: `Gaussian normal with Var = stddev plus Mean`,
	5: `Beta with Var = alpha and Par = beta shape parameters`,
	6: `Mean is just the constant Mean, no randomness`,
}

var _RndDistsMap = map[RndDists]string{
	0: `Uniform`,
	1: `Binomial`,
	2: `Poisson`,
	3: `Gamma`,
	4: `Gaussian`,
	5: `Beta`,
	6: `Mean`,
}

// String returns the string representation
// of this RndDists value.
func (i RndDists) String() string {
	if str, ok := _RndDistsMap[i]; ok {
		return str
	}
	return strconv.FormatInt(int64(i), 10)
}

// SetString sets the RndDists value from its
// string representation, and returns an
// error if the string is invalid.
func (i *RndDists) SetString(s string) error {
	if val, ok := _RndDistsNameToValueMap[s]; ok {
		*i = val
		return nil
	}
	if val, ok := _RndDistsNameToValueMap[strings.ToLower(s)]; ok {
		*i = val
		return nil
	}
	return errors.New(s + " is not a valid value for type RndDists")
}

// Int64 returns the RndDists value as an int64.
func (i RndDists) Int64() int64 {
	return int64(i)
}

// SetInt64 sets the RndDists value from an int64.
func (i *RndDists) SetInt64(in int64) {
	*i = RndDists(in)
}

// Desc returns the description of the RndDists value.
func (i RndDists) Desc() string {
	if str, ok := _RndDistsDescMap[i]; ok {
		return str
	}
	return i.String()
}

// RndDistsValues returns all possible values
// for the type RndDists.
func RndDistsValues() []RndDists {
	return _RndDistsValues
}

// Values returns all possible values
// for the type RndDists.
func (i RndDists) Values() []enums.Enum {
	res := make([]enums.Enum, len(_RndDistsValues))
	for i, d := range _RndDistsValues {
		res[i] = d
	}
	return res
}

// IsValid returns whether the value is a
// valid option for type RndDists.
func (i RndDists) IsValid() bool {
	_, ok := _RndDistsMap[i]
	return ok
}

// MarshalText implements the [encoding.TextMarshaler] interface.
func (i RndDists) MarshalText() ([]byte, error) {
	return []byte(i.String()), nil
}

// UnmarshalText implements the [encoding.TextUnmarshaler] interface.
func (i *RndDists) UnmarshalText(text []byte) error {
	if err := i.SetString(string(text)); err != nil {
		log.Println(err)
	}
	return nil
}