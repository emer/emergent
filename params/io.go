// Copyright (c) 2019, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package params

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"sort"

	"cogentcore.org/core/base/indent"
	"cogentcore.org/core/base/iox"
	"cogentcore.org/core/base/iox/jsonx"
	"cogentcore.org/core/base/iox/tomlx"
	"cogentcore.org/core/core"
	"github.com/BurntSushi/toml"
	"golang.org/x/exp/maps"
)

// WriteGoPrelude writes the start of a go file in package main that starts a
// variable assignment to given variable -- for start of SaveGoCode methods.
func WriteGoPrelude(w io.Writer, varNm string) {
	w.Write([]byte("// File generated by params.SaveGoCode\n\n"))
	w.Write([]byte("package main\n\n"))
	w.Write([]byte(`import "github.com/emer/emergent/v2/params"`))
	w.Write([]byte("\n\nvar " + varNm + " = "))
}

// OpenJSON opens params from a JSON-formatted file.
func (pr *Params) OpenJSON(filename core.Filename) error {
	*pr = make(Params) // reset
	return jsonx.Open(pr, string(filename))
}

// SaveJSON saves params to a JSON-formatted file.
func (pr *Params) SaveJSON(filename core.Filename) error {
	return jsonx.Save(pr, string(filename))
}

// OpenTOML opens params from a TOML-formatted file.
func (pr *Params) OpenTOML(filename core.Filename) error {
	*pr = make(Params) // reset
	return tomlx.Open(pr, string(filename))
}

// SaveTOML saves params to a TOML-formatted file.
func (pr *Params) SaveTOML(filename core.Filename) error {
	// return tomlx.Save(pr, string(filename)) // pelletier/go-toml produces bad output on maps
	return iox.Save(pr, string(filename), func(w io.Writer) iox.Encoder {
		return toml.NewEncoder(w)
	})
}

// WriteGoCode writes params to corresponding Go initializer code.
func (pr *Params) WriteGoCode(w io.Writer, depth int) {
	w.Write([]byte("params.Params{\n"))
	depth++
	paths := make([]string, len(*pr)) // alpha-sort paths for consistent output
	ctr := 0
	for pt := range *pr {
		paths[ctr] = pt
		ctr++
	}
	sort.StringSlice(paths).Sort()
	for _, pt := range paths {
		pv := (*pr)[pt]
		w.Write(indent.TabBytes(depth))
		w.Write([]byte(fmt.Sprintf("%q: %q,\n", pt, pv)))
	}
	depth--
	w.Write(indent.TabBytes(depth))
	w.Write([]byte("}"))
}

// StringGoCode returns Go initializer code as a byte string.
func (pr *Params) StringGoCode() []byte {
	var buf bytes.Buffer
	pr.WriteGoCode(&buf, 0)
	return buf.Bytes()
}

// SaveGoCode saves params to corresponding Go initializer code.
func (pr *Params) SaveGoCode(filename core.Filename) error {
	fp, err := os.Create(string(filename))
	defer fp.Close()
	if err != nil {
		log.Println(err)
		return err
	}
	WriteGoPrelude(fp, "SavedParams")
	pr.WriteGoCode(fp, 0)
	return nil
}

/////////////////////////////////////////////////////////
//   Hypers

// OpenJSON opens hypers from a JSON-formatted file.
func (pr *Hypers) OpenJSON(filename core.Filename) error {
	*pr = make(Hypers) // reset
	b, err := ioutil.ReadFile(string(filename))
	if err != nil {
		log.Println(err)
		return err
	}
	return json.Unmarshal(b, pr)
}

// SaveJSON saves hypers to a JSON-formatted file.
func (pr *Hypers) SaveJSON(filename core.Filename) error {
	return jsonx.Save(pr, string(filename))
}

// OpenTOML opens params from a TOML-formatted file.
func (pr *Hypers) OpenTOML(filename core.Filename) error {
	*pr = make(Hypers) // reset
	return tomlx.Open(pr, string(filename))
}

// SaveTOML saves params to a TOML-formatted file.
func (pr *Hypers) SaveTOML(filename core.Filename) error {
	// return tomlx.Save(pr, string(filename))
	return iox.Save(pr, string(filename), func(w io.Writer) iox.Encoder {
		return toml.NewEncoder(w)
	})
}

// WriteGoCode writes hypers to corresponding Go initializer code.
func (pr *Hypers) WriteGoCode(w io.Writer, depth int) {
	w.Write([]byte("params.Hypers{\n"))
	depth++
	paths := maps.Keys(*pr)
	sort.StringSlice(paths).Sort()
	for _, pt := range paths {
		pv := (*pr)[pt]
		w.Write(indent.TabBytes(depth))
		w.Write([]byte(fmt.Sprintf("%q: {", pt)))
		ks := maps.Keys(pv)
		sort.StringSlice(ks).Sort()
		for _, k := range ks {
			v := pv[k]
			w.Write([]byte(fmt.Sprintf("%q: %q,", k, v)))
		}
		w.Write([]byte("},\n"))
	}
	depth--
	w.Write(indent.TabBytes(depth))
	w.Write([]byte("}"))
}

// StringGoCode returns Go initializer code as a byte string.
func (pr *Hypers) StringGoCode() []byte {
	var buf bytes.Buffer
	pr.WriteGoCode(&buf, 0)
	return buf.Bytes()
}

// SaveGoCode saves hypers to corresponding Go initializer code.
func (pr *Hypers) SaveGoCode(filename core.Filename) error {
	fp, err := os.Create(string(filename))
	defer fp.Close()
	if err != nil {
		log.Println(err)
		return err
	}
	WriteGoPrelude(fp, "SavedHypers")
	pr.WriteGoCode(fp, 0)
	return nil
}

/////////////////////////////////////////////////////////
//   Sel

// OpenJSON opens params from a JSON-formatted file.
func (pr *Sel) OpenJSON(filename core.Filename) error {
	b, err := ioutil.ReadFile(string(filename))
	if err != nil {
		log.Println(err)
		return err
	}
	return json.Unmarshal(b, pr)
}

// SaveJSON saves params to a JSON-formatted file.
func (pr *Sel) SaveJSON(filename core.Filename) error {
	return jsonx.Save(pr, string(filename))
}

// OpenTOML opens params from a TOML-formatted file.
func (pr *Sel) OpenTOML(filename core.Filename) error {
	return tomlx.Open(pr, string(filename))
}

// SaveTOML saves params to a TOML-formatted file.
func (pr *Sel) SaveTOML(filename core.Filename) error {
	// return tomlx.Save(pr, string(filename))
	return iox.Save(pr, string(filename), func(w io.Writer) iox.Encoder {
		return toml.NewEncoder(w)
	})
}

// WriteGoCode writes params to corresponding Go initializer code.
func (pr *Sel) WriteGoCode(w io.Writer, depth int) {
	w.Write([]byte(fmt.Sprintf("Sel: %q, Desc: %q,\n", pr.Sel, pr.Desc)))
	depth++
	w.Write(indent.TabBytes(depth))
	w.Write([]byte("Params: "))
	pr.Params.WriteGoCode(w, depth)
	if len(pr.Hypers) > 0 {
		w.Write([]byte(", Hypers: "))
		pr.Hypers.WriteGoCode(w, depth)
	}
}

// StringGoCode returns Go initializer code as a byte string.
func (pr *Sel) StringGoCode() []byte {
	var buf bytes.Buffer
	pr.WriteGoCode(&buf, 0)
	return buf.Bytes()
}

// SaveGoCode saves params to corresponding Go initializer code.
func (pr *Sel) SaveGoCode(filename core.Filename) error {
	fp, err := os.Create(string(filename))
	defer fp.Close()
	if err != nil {
		log.Println(err)
		return err
	}
	WriteGoPrelude(fp, "SavedParamsSel")
	pr.WriteGoCode(fp, 0)
	return nil
}

/////////////////////////////////////////////////////////
//   Sheet

// OpenJSON opens params from a JSON-formatted file.
func (pr *Sheet) OpenJSON(filename core.Filename) error {
	*pr = make(Sheet, 0) // reset
	b, err := ioutil.ReadFile(string(filename))
	if err != nil {
		log.Println(err)
		return err
	}
	return json.Unmarshal(b, pr)
}

// SaveJSON saves params to a JSON-formatted file.
func (pr *Sheet) SaveJSON(filename core.Filename) error {
	return jsonx.Save(pr, string(filename))
}

// OpenTOML opens params from a TOML-formatted file.
func (pr *Sheet) OpenTOML(filename core.Filename) error {
	*pr = make(Sheet, 0) // reset
	return tomlx.Open(pr, string(filename))
}

// SaveTOML saves params to a TOML-formatted file.
func (pr *Sheet) SaveTOML(filename core.Filename) error {
	// return tomlx.Save(pr, string(filename))
	return iox.Save(pr, string(filename), func(w io.Writer) iox.Encoder {
		return toml.NewEncoder(w)
	})
}

// WriteGoCode writes params to corresponding Go initializer code.
func (pr *Sheet) WriteGoCode(w io.Writer, depth int) {
	w.Write([]byte("{\n"))
	depth++
	for _, pv := range *pr {
		w.Write(indent.TabBytes(depth))
		w.Write([]byte("{"))
		pv.WriteGoCode(w, depth)
		w.Write([]byte("},\n"))
	}
	depth--
	w.Write(indent.TabBytes(depth))
	w.Write([]byte("},\n"))
}

// StringGoCode returns Go initializer code as a byte string.
func (pr *Sheet) StringGoCode() []byte {
	var buf bytes.Buffer
	pr.WriteGoCode(&buf, 0)
	return buf.Bytes()
}

// SaveGoCode saves params to corresponding Go initializer code.
func (pr *Sheet) SaveGoCode(filename core.Filename) error {
	fp, err := os.Create(string(filename))
	defer fp.Close()
	if err != nil {
		log.Println(err)
		return err
	}
	WriteGoPrelude(fp, "SavedParamsSheet")
	pr.WriteGoCode(fp, 0)
	return nil
}

/////////////////////////////////////////////////////////
//   Sets

// OpenJSON opens params from a JSON-formatted file.
func (pr *Sets) OpenJSON(filename core.Filename) error {
	*pr = make(Sets) // reset
	b, err := ioutil.ReadFile(string(filename))
	if err != nil {
		log.Println(err)
		return err
	}
	return json.Unmarshal(b, pr)
}

// SaveJSON saves params to a JSON-formatted file.
func (pr *Sets) SaveJSON(filename core.Filename) error {
	return jsonx.Save(pr, string(filename))
}

// OpenTOML opens params from a TOML-formatted file.
func (pr *Sets) OpenTOML(filename core.Filename) error {
	*pr = make(Sets) // reset
	return tomlx.Open(pr, string(filename))
}

// SaveTOML saves params to a TOML-formatted file.
func (pr *Sets) SaveTOML(filename core.Filename) error {
	// return tomlx.Save(pr, string(filename))
	return iox.Save(pr, string(filename), func(w io.Writer) iox.Encoder {
		return toml.NewEncoder(w)
	})
}

// WriteGoCode writes params to corresponding Go initializer code.
func (pr *Sets) WriteGoCode(w io.Writer, depth int) {
	w.Write([]byte("params.Sets{\n"))
	depth++
	for _, st := range *pr {
		w.Write(indent.TabBytes(depth))
		w.Write([]byte("{"))
		st.WriteGoCode(w, depth)
		w.Write([]byte("},\n"))
	}
	depth--
	w.Write(indent.TabBytes(depth))
	w.Write([]byte("}\n"))
}

// StringGoCode returns Go initializer code as a byte string.
func (pr *Sets) StringGoCode() []byte {
	var buf bytes.Buffer
	pr.WriteGoCode(&buf, 0)
	return buf.Bytes()
}

// SaveGoCode saves params to corresponding Go initializer code.
func (pr *Sets) SaveGoCode(filename core.Filename) error {
	fp, err := os.Create(string(filename))
	defer fp.Close()
	if err != nil {
		log.Println(err)
		return err
	}
	WriteGoPrelude(fp, "SavedParamsSets")
	pr.WriteGoCode(fp, 0)
	return nil
}
