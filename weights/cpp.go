// Copyright (c) 2019, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package weights

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"log"
	"strconv"
	"strings"
)

// NetReadCpp reads weights for entire network from old emergent C++ format
func NetReadCpp(r io.Reader) (*Network, error) {
	nw := &Network{}
	var (
		lw       *Layer
		pw       *Prjn
		rw       *Recv
		ri       int
		pi       int
		skipnext bool
		cidx     int
		err      error
		errlist  []error
	)
	scan := bufio.NewScanner(r) // line at a time
	for scan.Scan() {
		if skipnext {
			skipnext = false
			continue
		}
		b := scan.Bytes()
		bs := string(b)
		switch {
		case strings.HasPrefix(bs, "</"): // don't care about any ending tags
			continue
		case strings.HasPrefix(bs, "<Fmt "):
			continue
		case strings.HasPrefix(bs, "<Name "):
			continue
		case strings.HasPrefix(bs, "<Epoch "):
			continue
		case bs == "<Network>":
			continue
		case bs == "<Ug>":
			continue
		case bs == "<Un>":
			skipnext = true // skip over bias weight
			continue
		case strings.HasPrefix(bs, "<Lay "):
			lnm := strings.TrimSuffix(strings.TrimPrefix(bs, "<Lay "), ">")
			nw.Layers = append(nw.Layers, Layer{Layer: lnm})
			lw = &nw.Layers[len(nw.Layers)-1]
			pw = nil
			continue
		case strings.HasPrefix(bs, "<UgUn "):
			us := strings.TrimSuffix(strings.TrimPrefix(bs, "<UgUn "), ">")
			uss := strings.Split(us, " ") // includes unit name
			ri, err = strconv.Atoi(uss[0])
			if err != nil {
				errlist = append(errlist, err)
			}
			continue
		case strings.HasPrefix(bs, "<Cg "):
			cs := strings.TrimSuffix(strings.TrimPrefix(bs, "<Cg "), ">")
			css := strings.Split(cs, " ")
			pi, err = strconv.Atoi(css[0])
			if err != nil {
				errlist = append(errlist, err)
			}
			fm := strings.TrimPrefix(css[1], "Fm:")
			if len(lw.Prjns) < pi+1 {
				lw.Prjns = append(lw.Prjns, Prjn{From: fm})
			}
			pw = &lw.Prjns[pi]
			continue
		case strings.HasPrefix(bs, "<Cn "):
			us := strings.TrimSuffix(strings.TrimPrefix(bs, "<Cn "), ">")
			nc, err := strconv.Atoi(us)
			if err != nil {
				errlist = append(errlist, err)
			}
			if len(pw.Rs) < ri+1 {
				pw.Rs = append(pw.Rs, Recv{Ri: ri, N: nc})
			}
			rw = &pw.Rs[ri]
			if len(rw.Si) != nc {
				rw.Si = make([]int, nc)
				rw.Wt = make([]float32, nc)
				rw.Wt1 = make([]float32, nc)
			}
			cidx = 0 // start reading on next ones
			continue
		case strings.HasPrefix(bs, "<"): // misc meta
			kvl := strings.Split(bs, " ")
			if len(kvl) != 2 {
				err = fmt.Errorf("NetReadCpp: unrecognized input: %v", bs)
				errlist = append(errlist, err)
				log.Println(err)
				continue
			}
			ky := strings.TrimPrefix(kvl[0], "<")
			vl := strings.TrimSuffix(kvl[1], ">")
			switch ky {
			case "acts_m_avg":
				ky = "ActMAvg"
			case "acts_p_avg":
				ky = "ActPAvg"
			}
			if lw == nil {
				nw.SetMetaData(ky, vl)
			} else if pw == nil {
				lw.SetMetaData(ky, vl)
			} else {
				pw.SetMetaData(ky, vl)
			}
			continue
		default: // weight values read into current rw
			siwts := strings.Split(bs, " ")
			switch len(siwts) {
			case 2:
				si, err := strconv.Atoi(siwts[0])
				if err != nil {
					errlist = append(errlist, err)
				}
				wt, err := strconv.ParseFloat(siwts[1], 32)
				if err != nil {
					errlist = append(errlist, err)
				}
				rw.Si[cidx] = si
				rw.Wt[cidx] = float32(wt)
				rw.Wt1[cidx] = float32(0)
				cidx++
			case 3:
				si, err := strconv.Atoi(siwts[0])
				if err != nil {
					errlist = append(errlist, err)
				}
				wt, err := strconv.ParseFloat(siwts[1], 32)
				if err != nil {
					errlist = append(errlist, err)
				}
				scale, err := strconv.ParseFloat(siwts[2], 32)
				if err != nil {
					errlist = append(errlist, err)
				}
				rw.Si[cidx] = si
				rw.Wt[cidx] = float32(wt)
				rw.Wt1[cidx] = float32(scale)
				cidx++
			default:
				err = fmt.Errorf("NetReadCpp: unrecognized input: %v", bs)
				errlist = append(errlist, err)
				log.Println(err)
				continue
			}
		}
	}
	var eall error
	if len(errlist) > 0 {
		es := ""
		for _, er := range errlist {
			es = es + er.Error() + "\n"
		}
		eall = errors.New(es)
	}
	return nw, eall
}
