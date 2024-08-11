// Copyright (c) 2019, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package emer

import (
	"bufio"
	"compress/gzip"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"

	"cogentcore.org/core/base/indent"
	"cogentcore.org/core/core"
	"github.com/emer/emergent/v2/weights"
)

// SaveWeightsJSON saves network weights (and any other state that adapts with learning)
// to a JSON-formatted file.  If filename has .gz extension, then file is gzip compressed.
func (nt *NetworkBase) SaveWeightsJSON(filename core.Filename) error { //types:add
	fp, err := os.Create(string(filename))
	defer fp.Close()
	if err != nil {
		log.Println(err)
		return err
	}
	ext := filepath.Ext(string(filename))
	if ext == ".gz" {
		gzr := gzip.NewWriter(fp)
		err = nt.WriteWeightsJSON(gzr)
		gzr.Close()
	} else {
		bw := bufio.NewWriter(fp)
		err = nt.WriteWeightsJSON(bw)
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
		log.Println(err)
		return err
	}
	ext := filepath.Ext(string(filename))
	if ext == ".gz" {
		gzr, err := gzip.NewReader(fp)
		defer gzr.Close()
		if err != nil {
			log.Println(err)
			return err
		}
		return nt.ReadWeightsJSON(gzr)
	} else {
		return nt.ReadWeightsJSON(bufio.NewReader(fp))
	}
}

// todo: proper error handling here!

// WriteWeightsJSON writes the weights from this layer from the receiver-side perspective
// in a JSON text format.  We build in the indentation logic to make it much faster and
// more efficient.
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
		log.Println(err)
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
