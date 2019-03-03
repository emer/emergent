// Copyright (c) 2019, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package eplot

import (
	"bytes"
	"log"

	"github.com/goki/gi/svg"
	"gonum.org/v1/plot"
	"gonum.org/v1/plot/vg"
	"gonum.org/v1/plot/vg/draw"
	"gonum.org/v1/plot/vg/vgsvg"
)

// PlotViewSVG shows the given gonum Plot in given GoGi svg editor widget
// xSz and ySz are the SVG plot sizes, in inches -- in general 4-5 seems
// pretty good in terms of overall font size results.  Scale to fit
// your window -- e.g., 2-3 depending on sizes
func PlotViewSVG(plt *plot.Plot, svge *svg.Editor, xSz, ySz, scale float64) {
	updt := svge.UpdateStart()
	defer svge.UpdateEnd(updt)
	svge.SetFullReRender()

	// Create a Canvas for writing SVG images.
	c := vgsvg.New(vg.Length(xSz)*vg.Inch, vg.Length(ySz)*vg.Inch)

	// Draw to the Canvas.
	plt.Draw(draw.New(c))

	var buf bytes.Buffer
	if _, err := c.WriteTo(&buf); err != nil {
		log.Println(err)
		return
	}
	svge.ReadXML(&buf)

	svge.SetNormXForm()
	svge.Scale = float32(scale) * (svge.Viewport.Win.LogicalDPI() / 96.0)
	svge.SetTransform()
}

// StringViewSVG shows the given svg string in given GoGi svg editor widget
// Scale to fit your window -- e.g., 2-3 depending on sizes
func StringViewSVG(svgstr string, svge *svg.Editor, scale float64) {
	updt := svge.UpdateStart()
	defer svge.UpdateEnd(updt)
	svge.SetFullReRender()

	var buf bytes.Buffer
	buf.Write([]byte(svgstr))
	svge.ReadXML(&buf)

	svge.SetNormXForm()
	svge.Scale = float32(scale) * (svge.Viewport.Win.LogicalDPI() / 96.0)
	svge.SetTransform()
}
