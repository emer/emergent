// Copyright (c) 2019, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package netparams

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"

	"github.com/emer/emergent/params"
	"github.com/goki/gi/gi"
	"github.com/goki/ki/indent"
	"github.com/goki/ki/ki"
	"github.com/goki/ki/toml"
)

// OpenJSON opens params from a JSON-formatted file.
func (pr *Sets) OpenJSON(filename gi.FileName) error {
	*pr = make(Sets) // reset
	b, err := ioutil.ReadFile(string(filename))
	if err != nil {
		log.Println(err)
		return err
	}
	return json.Unmarshal(b, pr)
}

// SaveJSON saves params to a JSON-formatted file.
func (pr *Sets) SaveJSON(filename gi.FileName) error {
	b, err := json.MarshalIndent(pr, "", "  ")
	if err != nil {
		log.Println(err) // unlikely
		return err
	}
	err = ioutil.WriteFile(string(filename), b, 0644)
	if err != nil {
		log.Println(err)
	}
	return err
}

// OpenTOML opens params from a TOML-formatted file.
func (pr *Sets) OpenTOML(filename gi.FileName) error {
	*pr = make(Sets) // reset
	return toml.Open(pr, string(filename))
}

// SaveTOML saves params to a TOML-formatted file.
func (pr *Sets) SaveTOML(filename gi.FileName) error {
	return toml.Save(pr, string(filename))
}

// WriteGoCode writes params to corresponding Go initializer code.
func (pr *Sets) WriteGoCode(w io.Writer, depth int) {
	w.Write([]byte(fmt.Sprintf("params.Sets{\n")))
	depth++
	for nm, st := range *pr {
		w.Write(indent.TabBytes(depth))
		w.Write([]byte(nm + ": "))
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
func (pr *Sets) SaveGoCode(filename gi.FileName) error {
	fp, err := os.Create(string(filename))
	defer fp.Close()
	if err != nil {
		log.Println(err)
		return err
	}
	params.WriteGoPrelude(fp, "SavedParamsSets")
	pr.WriteGoCode(fp, 0)
	return nil
}

var SetsProps = ki.Props{
	"ToolBar": ki.PropSlice{
		{"Save", ki.PropSlice{
			{"SaveTOML", ki.Props{
				"label": "Save As TOML...",
				"desc":  "save to TOML formatted file",
				"icon":  "file-save",
				"Args": ki.PropSlice{
					{"File Name", ki.Props{
						"ext": ".toml",
					}},
				},
			}},
			{"SaveJSON", ki.Props{
				"label": "Save As JSON...",
				"desc":  "save to JSON formatted file",
				"icon":  "file-save",
				"Args": ki.PropSlice{
					{"File Name", ki.Props{
						"ext": ".json",
					}},
				},
			}},
			{"SaveGoCode", ki.Props{
				"label": "Save Code As...",
				"desc":  "save to Go-formatted initializer code in file",
				"icon":  "go",
				"Args": ki.PropSlice{
					{"File Name", ki.Props{
						"ext": ".go",
					}},
				},
			}},
		}},
		{"Open", ki.PropSlice{
			{"OpenTOML", ki.Props{
				"label": "Open...",
				"desc":  "open from TOML formatted file",
				"icon":  "file-open",
				"Args": ki.PropSlice{
					{"File Name", ki.Props{
						"ext": ".toml",
					}},
				},
			}},
			{"OpenJSON", ki.Props{
				"label": "Open...",
				"desc":  "open from JSON formatted file",
				"icon":  "file-open",
				"Args": ki.PropSlice{
					{"File Name", ki.Props{
						"ext": ".json",
					}},
				},
			}},
		}},
		{"StringGoCode", ki.Props{
			"label":       "Show Code",
			"desc":        "shows the Go-formatted initializer code, can be copy / pasted into program",
			"icon":        "go",
			"show-return": true,
		}},
		{"sep-diffs", ki.BlankProp{}},
		{"DiffsAll", ki.Props{
			"desc":        "between all sets, reports where the same param path is being set to different values",
			"icon":        "search",
			"show-return": true,
		}},
		{"DiffsFirst", ki.Props{
			"desc":        "between first set (e.g., the Base set) and rest of sets, reports where the same param path is being set to different values",
			"icon":        "search",
			"show-return": true,
		}},
		{"DiffsWithin", ki.Props{
			"desc":        "reports all the cases where the same param path is being set to different values within different sheets in given set",
			"icon":        "search",
			"show-return": true,
			"Args": ki.PropSlice{
				{"Set Name", ki.Props{}},
			},
		}},
	},
}
