// Copyright (c) 2022, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package looper

import (
	"fmt"
	"log"
	"strings"

	"github.com/goki/ki/indent"
)

// NamedFunc is a named function
type NamedFunc struct {
	Name string `desc:"name of function -- use a Context:Function naming convention, where Context is the overall context of the function, which could be a type name, e.g., Env:Step"`
	Func func() `desc:"function -- note that you can pass a method to a type as well as a closure here"`
}

// Funcs is a list of named plain functions
type Funcs []*NamedFunc

// Run runs the list of functions in order
func (fs *Funcs) Run() {
	for _, fn := range *fs {
		fn.Func()
	}
}

// RunTrace runs the list of functions in order, printing a trace before each as they run
func (fs *Funcs) RunTrace(level int) {
	for _, fn := range *fs {
		fmt.Printf("%s%s:\n", indent.Spaces(level, indentSize), fn.Name)
		fn.Func()
	}
}

// Names returns the list of function names
func (fs *Funcs) Names() []string {
	nms := make([]string, len(*fs))
	for i, fn := range *fs {
		nms[i] = fn.Name
	}
	return nms
}

// String returns the list of function names
func (fs *Funcs) String() string {
	return strings.Join(fs.Names(), ", ")
}

// FindName finds index of function by name, returns not found err message if not found
func (fs *Funcs) FindName(nm string) (int, error) {
	for i, fn := range *fs {
		if fn.Name == nm {
			return i, nil
		}
	}
	err := fmt.Errorf("looper.Funcs:FindName: named function %s not found", nm)
	return -1, err
}

// Add adds a function to the end of the list
func (fs *Funcs) Add(nm string, f func()) {
	*fs = append(*fs, &NamedFunc{Name: nm, Func: f})
}

// Prepend adds a function to the start of the list
func (fs *Funcs) Prepend(nm string, f func()) {
	fs.InsertAt(0, nm, f)
}

// InsertAt inserts function at given index
func (fs *Funcs) InsertAt(i int, nm string, f func()) {
	sz := len(*fs)
	*fs = append(*fs, nil)
	if i < sz {
		copy((*fs)[i+1:], (*fs)[i:sz])
	}
	(*fs)[i] = &NamedFunc{Name: nm, Func: f}
}

// InsertBefore inserts function before other function of given name
func (fs *Funcs) InsertBefore(before, nm string, f func()) error {
	i, err := fs.FindName(before)
	if err != nil {
		err = fmt.Errorf("InsertBefore of %s: %s", nm, err)
		log.Println(err)
		return err
	}
	fs.InsertAt(i, nm, f)
	return nil
}

// InsertAfter inserts function after other function of given name
func (fs *Funcs) InsertAfter(after, nm string, f func()) error {
	i, err := fs.FindName(after)
	if err != nil {
		err = fmt.Errorf("InsertAfter of %s: %s", nm, err)
		log.Println(err)
		return err
	}
	fs.InsertAt(i+1, nm, f)
	return nil
}

// Replace replaces function with other function of given name
func (fs *Funcs) Replace(nm string, f func()) error {
	i, err := fs.FindName(nm)
	if err != nil {
		err = fmt.Errorf("Replace of %s: %s", nm, err)
		log.Println(err)
		return err
	}
	(*fs)[i] = &NamedFunc{Name: nm, Func: f}
	return nil
}

// Delete deletes function of given name
func (fs *Funcs) Delete(nm string) error {
	i, err := fs.FindName(nm)
	if err != nil {
		err = fmt.Errorf("Delete: %s", err)
		log.Println(err)
		return err
	}
	sz := len(*fs)
	copy((*fs)[i:], (*fs)[i+1:])
	(*fs)[sz-1] = nil
	(*fs) = (*fs)[:sz-1]
	return nil
}

////////////////////////////////////////////////////////////////

// NamedBoolFunc is a named function returning bool
type NamedBoolFunc struct {
	Name string      `desc:"name of function -- use a Context:Function naming convention, where Context is the overall context of the function, which could be a type name, e.g., Env:Step"`
	Func func() bool `desc:"function -- note that you can pass a method to a type as well as a closure here"`
}

// BoolFuncs is a list of named plain functions
type BoolFuncs []*NamedBoolFunc

// Run runs the list of functions in order, returning true
// as soon as any of the functions return true, else false
func (fs *BoolFuncs) Run() bool {
	for _, fn := range *fs {
		bv := fn.Func()
		if bv {
			return true
		}
	}
	return false
}

// RunTrace runs the list of functions in order, returning true
// as soon as any of the functions return true, else false.
// Prints a trace before each as they run
func (fs *BoolFuncs) RunTrace(level int) bool {
	for _, fn := range *fs {
		fmt.Printf("%s%s:\n", indent.Spaces(level, indentSize), fn.Name)
		bv := fn.Func()
		if bv {
			return true
		}
	}
	return false
}

// Names returns the list of function names
func (fs *BoolFuncs) Names() []string {
	nms := make([]string, len(*fs))
	for i, fn := range *fs {
		nms[i] = fn.Name
	}
	return nms
}

// String returns the list of function names
func (fs *BoolFuncs) String() string {
	return strings.Join(fs.Names(), ", ")
}

// FindName finds index of function by name, returns not found err message if not found
func (fs *BoolFuncs) FindName(nm string) (int, error) {
	for i, fn := range *fs {
		if fn.Name == nm {
			return i, nil
		}
	}
	err := fmt.Errorf("looper.BoolFuncs:FindName: named function %s not found", nm)
	return -1, err
}

// Add adds a function to the end of the list
func (fs *BoolFuncs) Add(nm string, f func() bool) {
	*fs = append(*fs, &NamedBoolFunc{Name: nm, Func: f})
}

// Prepend adds a function to the start of the list
func (fs *BoolFuncs) Prepend(nm string, f func() bool) {
	fs.InsertAt(0, nm, f)
}

// InsertAt inserts function at given index
func (fs *BoolFuncs) InsertAt(i int, nm string, f func() bool) {
	sz := len(*fs)
	*fs = append(*fs, nil)
	if i < sz {
		copy((*fs)[i+1:], (*fs)[i:sz])
	}
	(*fs)[i] = &NamedBoolFunc{Name: nm, Func: f}
}

// InsertBefore inserts function before other function of given name
func (fs *BoolFuncs) InsertBefore(before, nm string, f func() bool) error {
	i, err := fs.FindName(before)
	if err != nil {
		err = fmt.Errorf("InsertBefore of %s: %s", nm, err)
		log.Println(err)
		return err
	}
	fs.InsertAt(i, nm, f)
	return nil
}

// InsertAfter inserts function after other function of given name
func (fs *BoolFuncs) InsertAfter(after, nm string, f func() bool) error {
	i, err := fs.FindName(after)
	if err != nil {
		err = fmt.Errorf("InsertAfter of %s: %s", nm, err)
		log.Println(err)
		return err
	}
	fs.InsertAt(i+1, nm, f)
	return nil
}

// Replace replaces function with other function of given name
func (fs *BoolFuncs) Replace(nm string, f func() bool) error {
	i, err := fs.FindName(nm)
	if err != nil {
		err = fmt.Errorf("Replace of %s: %s", nm, err)
		log.Println(err)
		return err
	}
	(*fs)[i] = &NamedBoolFunc{Name: nm, Func: f}
	return nil
}

// Delete deletes function of given name
func (fs *BoolFuncs) Delete(nm string) error {
	i, err := fs.FindName(nm)
	if err != nil {
		err = fmt.Errorf("Delete: %s", err)
		log.Println(err)
		return err
	}
	sz := len(*fs)
	copy((*fs)[i:], (*fs)[i+1:])
	(*fs)[sz-1] = nil
	(*fs) = (*fs)[:sz-1]
	return nil
}
