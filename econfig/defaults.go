// Copyright (c) 2023, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package econfig

import (
	"fmt"
	"log"
	"reflect"

	"github.com/goki/ki/kit"
)

// SetFromDefaults sets Config values from field tag `def:` values.
// Parsing errors are automatically logged.
func SetFromDefaults(cfg any) error {
	return SetFromDefaultsStruct(cfg)
}

// todo: move this to kit:

// SetFromDefaultsStruct sets values of fields in given struct based on
// `def:` default value field tags.
func SetFromDefaultsStruct(obj any) error {
	typ := kit.NonPtrType(reflect.TypeOf(obj))
	val := kit.NonPtrValue(reflect.ValueOf(obj))
	var err error
	for i := 0; i < typ.NumField(); i++ {
		f := typ.Field(i)
		fv := val.Field(i)
		if kit.NonPtrType(f.Type).Kind() == reflect.Struct {
			SetFromDefaultsStruct(kit.PtrValue(fv).Interface())
		}
		def, ok := f.Tag.Lookup("def")
		if !ok || def == "" {
			continue
		}
		ok = kit.SetRobust(kit.PtrValue(fv).Interface(), def) // overkill but whatever
		if !ok {
			err = fmt.Errorf("SetFromDefaultsStruct: was not able to set field: %s in object of type: %s from val: %s", f.Name, typ.Name(), def)
			log.Println(err)
		}
	}
	return err
}
