// Copyright (c) 2019, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package emer

import (
	"bufio"
	"compress/gzip"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"sort"

	"cogentcore.org/core/base/errors"
	"cogentcore.org/core/base/indent"
	"cogentcore.org/core/core"
	"github.com/emer/emergent/v2/weights"
	"golang.org/x/exp/maps"
)

// SaveWeightsJSON saves network weights (and any other state that adapts with learning)
// to a JSON-formatted file.  If filename has .gz extension, then file is gzip compressed.
func (nt *NetworkBase) SaveWeightsJSON(filename core.Filename) error { //types:add
	fp, err := os.Create(string(filename))
	defer fp.Close()
	if err != nil {
		return errors.Log(err)
	}
	ext := filepath.Ext(string(filename))
	if ext == ".gz" {
		gzr := gzip.NewWriter(fp)
		err = nt.EmerNetwork.WriteWeightsJSON(gzr)
		gzr.Close()
	} else {
		bw := bufio.NewWriter(fp)
		err = nt.EmerNetwork.WriteWeightsJSON(bw)
		bw.Flush()
	}
	return err
}

// OpenWeightsJSON opens network weights (and any other state that adapts with learning)
// from a JSON-formatted file.  If filename has .gz extension, then file is gzip uncompressed.
func (nt *NetworkBase) OpenWeightsJSON(filename core.Filename) error { //types:add
	fp, err := os.Open(string(filename))
	defer fp.Close()
	if err != nil {
		return errors.Log(err)
	}
	ext := filepath.Ext(string(filename))
	if ext == ".gz" {
		gzr, err := gzip.NewReader(fp)
		defer gzr.Close()
		if err != nil {
			return errors.Log(err)
		}
		return nt.EmerNetwork.ReadWeightsJSON(gzr)
	} else {
		return nt.EmerNetwork.ReadWeightsJSON(bufio.NewReader(fp))
	}
}

// OpenWeightsFS opens network weights (and any other state that adapts with learning)
// from a JSON-formatted file, in filesystem.
// If filename has .gz extension, then file is gzip uncompressed.
func (nt *NetworkBase) OpenWeightsFS(fsys fs.FS, filename string) error {
	fp, err := fsys.Open(filename)
	defer fp.Close()
	if err != nil {
		return errors.Log(err)
	}
	ext := filepath.Ext(filename)
	if ext == ".gz" {
		gzr, err := gzip.NewReader(fp)
		defer gzr.Close()
		if err != nil {
			return errors.Log(err)
		}
		return nt.EmerNetwork.ReadWeightsJSON(gzr)
	} else {
		return nt.EmerNetwork.ReadWeightsJSON(bufio.NewReader(fp))
	}
}

// todo: proper error handling here!

// WriteWeightsJSON writes the weights from this network
// from the receiver-side perspective in a JSON text format.
func (nt *NetworkBase) WriteWeightsJSON(w io.Writer) error {
	en := nt.EmerNetwork
	nlay := en.NumLayers()

	depth := 0
	w.Write(indent.TabBytes(depth))
	w.Write([]byte("{\n"))
	depth++
	w.Write(indent.TabBytes(depth))
	w.Write([]byte(fmt.Sprintf("\"Network\": %q,\n", nt.Name))) // note: can't use \n in `` so need "
	w.Write(indent.TabBytes(depth))
	onls := make([]Layer, 0, nlay)
	for li := range nlay {
		ly := en.EmerLayer(li)
		if !ly.AsEmer().Off {
			onls = append(onls, ly)
		}
	}
	nl := len(onls)
	if nl == 0 {
		w.Write([]byte("\"Layers\": null\n"))
	} else {
		w.Write([]byte("\"Layers\": [\n"))
		depth++
		for li, ly := range onls {
			ly.WriteWeightsJSON(w, depth)
			if li == nl-1 {
				w.Write([]byte("\n"))
			} else {
				w.Write([]byte(",\n"))
			}
		}
		depth--
		w.Write(indent.TabBytes(depth))
		w.Write([]byte("]\n"))
	}
	depth--
	w.Write(indent.TabBytes(depth))
	_, err := w.Write([]byte("}\n"))
	return err
}

// ReadWeightsJSON reads network weights from the receiver-side perspective
// in a JSON text format.  Reads entire file into a temporary weights.Weights
// structure that is then passed to Layers etc using SetWeights method.
func (nt *NetworkBase) ReadWeightsJSON(r io.Reader) error {
	nw, err := weights.NetReadJSON(r)
	if err != nil {
		return err // note: already logged
	}
	err = nt.SetWeights(nw)
	if err != nil {
		errors.Log(err)
	}
	return err
}

// SetWeights sets the weights for this network from weights.Network decoded values
func (nt *NetworkBase) SetWeights(nw *weights.Network) error {
	var errs []error
	if nw.Network != "" {
		nt.Name = nw.Network
	}
	if nw.MetaData != nil {
		if nt.MetaData == nil {
			nt.MetaData = nw.MetaData
		} else {
			for mk, mv := range nw.MetaData {
				nt.MetaData[mk] = mv
			}
		}
	}
	for li := range nw.Layers {
		lw := &nw.Layers[li]
		ly, err := nt.EmerLayerByName(lw.Layer)
		if err != nil {
			errs = append(errs, err)
			continue
		}
		ly.SetWeights(lw)
	}
	return errors.Join(errs...)
}

// WriteWeightsJSONBase writes the weights from this layer
// in a JSON text format.  Any values in the layer MetaData
// will be written first, and unit-level variables in unitVars
// are saved as well.  Then, all the receiving path data is saved.
func (ly *LayerBase) WriteWeightsJSONBase(w io.Writer, depth int, unitVars ...string) {
	el := ly.EmerLayer
	w.Write(indent.TabBytes(depth))
	w.Write([]byte("{\n"))
	depth++
	w.Write(indent.TabBytes(depth))
	w.Write([]byte(fmt.Sprintf("\"Layer\": %q,\n", ly.Name)))
	if len(ly.MetaData) > 0 {
		w.Write(indent.TabBytes(depth))
		w.Write([]byte(fmt.Sprintf("\"MetaData\": {\n")))
		depth++
		kys := maps.Keys(ly.MetaData)
		sort.StringSlice(kys).Sort()
		for i, k := range kys {
			w.Write(indent.TabBytes(depth))
			comma := ","
			if i == len(kys)-1 { // note: last one has no comma
				comma = ""
			}
			w.Write([]byte(fmt.Sprintf("%q: %q%s\n", k, ly.MetaData[k], comma)))
		}
		depth--
		w.Write(indent.TabBytes(depth))
		w.Write([]byte("},\n"))
	}
	if len(unitVars) > 0 {
		w.Write(indent.TabBytes(depth))
		w.Write([]byte(fmt.Sprintf("\"Units\": {\n")))
		depth++
		for i, vname := range unitVars {
			vidx, err := el.UnitVarIndex(vname)
			if errors.Log(err) != nil {
				continue
			}
			w.Write(indent.TabBytes(depth))
			w.Write([]byte(fmt.Sprintf("%q: [ ", vname)))
			nu := ly.NumUnits()
			for ni := range nu {
				val := el.UnitValue1D(vidx, ni, 0)
				w.Write([]byte(fmt.Sprintf("%g", val)))
				if ni < nu-1 {
					w.Write([]byte(", "))
				}
			}
			comma := ","
			if i == len(unitVars)-1 { // note: last one has no comma
				comma = ""
			}
			w.Write([]byte(fmt.Sprintf(" ]%s\n", comma)))
		}
		depth--
		w.Write(indent.TabBytes(depth))
		w.Write([]byte("},\n"))
	}
	w.Write(indent.TabBytes(depth))
	onps := make([]Path, 0, el.NumRecvPaths())
	for pi := range el.NumRecvPaths() {
		pt := el.RecvPath(pi)
		if !pt.AsEmer().Off {
			onps = append(onps, pt)
		}
	}
	np := len(onps)
	if np == 0 {
		w.Write([]byte(fmt.Sprintf("\"Paths\": null\n")))
	} else {
		w.Write([]byte(fmt.Sprintf("\"Paths\": [\n")))
		depth++
		for pi := range el.NumRecvPaths() {
			pt := el.RecvPath(pi)
			pt.WriteWeightsJSON(w, depth) // this leaves path unterminated
			if pi == np-1 {
				w.Write([]byte("\n"))
			} else {
				w.Write([]byte(",\n"))
			}
		}
		depth--
		w.Write(indent.TabBytes(depth))
		w.Write([]byte(" ]\n"))
	}
	depth--
	w.Write(indent.TabBytes(depth))
	w.Write([]byte("}")) // note: leave unterminated as outer loop needs to add , or just \n depending
}

// ReadWeightsJSON reads the weights from this layer from the
// receiver-side perspective in a JSON text format.
// This is for a set of weights that were saved *for one layer only*
// and is not used for the network-level ReadWeightsJSON,
// which reads into a separate structure -- see SetWeights method.
func (ly *LayerBase) ReadWeightsJSON(r io.Reader) error {
	lw, err := weights.LayReadJSON(r)
	if err != nil {
		return err // note: already logged
	}
	return ly.EmerLayer.SetWeights(lw)
}

// ReadWeightsJSON reads the weights from this pathway from the
// receiver-side perspective in a JSON text format.
// This is for a set of weights that were saved *for one path only*
// and is not used for the network-level ReadWeightsJSON,
// which reads into a separate structure -- see SetWeights method.
func (pt *PathBase) ReadWeightsJSON(r io.Reader) error {
	pw, err := weights.PathReadJSON(r)
	if err != nil {
		return err // note: already logged
	}
	return pt.EmerPath.SetWeights(pw)
}
