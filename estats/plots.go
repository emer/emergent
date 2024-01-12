// Copyright (c) 2022, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package estats

import (
	"github.com/emer/etable/v2/clust"
	"github.com/emer/etable/v2/eplot"
	"github.com/emer/etable/v2/etable"
	"github.com/emer/etable/v2/metric"
	"github.com/emer/etable/v2/simat"
)

func ConfigPCAPlot(plt *eplot.Plot2D, dt *etable.Table, nm string) {
	plt.Params.Title = nm
	col1 := dt.ColName(1)
	plt.Params.XAxisCol = col1
	plt.SetTable(dt)
	plt.Params.Lines = false
	plt.Params.Points = true
	// order of params: on, fixMin, min, fixMax, max
	plt.SetColParams(dt.ColName(0), eplot.On, eplot.FloatMin, 0, eplot.FloatMax, 0)
	plt.SetColParams(col1, eplot.Off, eplot.FloatMin, -3, eplot.FloatMax, 3)
	plt.SetColParams(dt.ColName(2), eplot.On, eplot.FloatMin, -3, eplot.FloatMax, 3)
}

// ClustPlot does one cluster plot on given table column name
// and label name
func ClustPlot(plt *eplot.Plot2D, ix *etable.IdxView, colNm, lblNm string) {
	nm, _ := ix.Table.MetaData["name"]
	smat := &simat.SimMat{}
	smat.TableCol(ix, colNm, lblNm, false, metric.Euclidean64)
	pt := &etable.Table{}
	clust.Plot(pt, clust.Glom(smat, clust.ContrastDist), smat)
	plt.InitName(plt, colNm)
	plt.Params.Title = "Cluster Plot of: " + nm + " " + colNm
	plt.Params.XAxisCol = "X"
	plt.SetTable(pt)
	// order of params: on, fixMin, min, fixMax, max
	plt.SetColParams("X", eplot.Off, eplot.FixMin, 0, eplot.FloatMax, 0)
	plt.SetColParams("Y", eplot.On, eplot.FixMin, 0, eplot.FloatMax, 0)
	plt.SetColParams("Label", eplot.On, eplot.FixMin, 0, eplot.FloatMax, 0)
}
