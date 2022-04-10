// Copyright (c) 2019, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package envlp

import (
	"fmt"

	"github.com/emer/etable/etensor"
)

// States32 is a map of named *etensor.Float32 tensors, used by Envs to manage states.
type States32 map[string]*etensor.Float32

// NewStates32 returns a new States32 map for given names
func NewStates32(names ...string) States32 {
	st := make(map[string]*etensor.Float32, len(names))
	for _, nm := range names {
		st[nm] = &etensor.Float32{}
	}
	return st
}

// ByNameTry accesses the tensor map by name, generating an
// error if the state is not found.
func (st *States32) ByNameTry(nm string) (*etensor.Float32, error) {
	t, ok := (*st)[nm]
	if ok {
		return t, nil
	}
	err := fmt.Errorf("envlp.States32.ByNameTry: state named %s not found", nm)
	return nil, err
}

// SetShape sets the Shape of given tensor(s) by name
// If strides is nil, row-major strides will be inferred.
// If names (dimension names) is nil, a slice of empty strings will be created.
func (st *States32) SetShape(shape, strides []int, names []string, nm ...string) {
	for _, nm := range names {
		t, err := st.ByNameTry(nm)
		if err != nil {
			fmt.Println(err)
			continue
		}
		t.SetShape(shape, strides, names)
	}
}

// SetZeros sets all states to zeros -- call prior to rendering new
func (st *States32) SetZeros() {
	for _, t := range *st {
		t.SetZeros()
	}
}
