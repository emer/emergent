// Copyright (c) 2022, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package estats

import "fmt"

// LinearDecodeTrain does decoding and training on the decoder
// of the given name, using given training value, saving
// the results to Float stats named with the decoder + Out and SSE.
// returns SSE.
// di is a data parallel index di, for networks capable
// of processing input patterns in parallel.
func (st *Stats) LinearDecodeTrain(decName, varNm string, di int, trainVal float32) (float32, error) {
	dec, ok := st.LinDecoders[decName]
	if !ok {
		err := fmt.Errorf("Linear Decoder named: %s not found", decName)
		fmt.Println(err)
		return 0, err
	}
	dec.Decode(varNm, di)
	out := []float32{0} // save alloc
	dec.Output(&out)
	st.SetFloat32(decName+"Out", out[0])
	out[0] = trainVal
	sse, err := dec.Train(out)
	if err != nil {
		fmt.Println(err)
		return sse, err
	}
	st.SetFloat32(decName+"SSE", sse)
	return sse, nil
}

// SoftLinearDecodeTrain does decoding and training on the decoder
// of the given name, using given training index value, saving
// the results to Float stats named with the decoder + Out and Err.
// Returns Err which is 1 if output != trainIndex, 0 otherwise.
// di is a data parallel index di, for networks capable
// of processing input patterns in parallel.
func (st *Stats) SoftMaxDecodeTrain(decName, varNm string, di int, trainIndex int) (float32, error) {
	dec, ok := st.SoftMaxDecoders[decName]
	if !ok {
		err := fmt.Errorf("SoftMax Decoder named: %s not found", decName)
		fmt.Println(err)
		return 0, err
	}
	out := dec.Decode(varNm, di)
	st.SetInt(decName+"Out", out)
	derr := float32(0)
	if out != trainIndex {
		derr = 1
	}
	st.SetFloat32(decName+"Err", derr)
	dec.Train(trainIndex)
	return derr, nil
}
