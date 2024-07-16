// Copyright (c) 2022, The Emergent Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package estats

import (
	"cogentcore.org/core/plot/plotcore"
	"cogentcore.org/core/tensor/stats/clust"
	"cogentcore.org/core/tensor/stats/metric"
	"cogentcore.org/core/tensor/stats/simat"
	"cogentcore.org/core/tensor/table"
)

func ConfigPCAPlot(plt *plotcore.PlotEditor, dt *table.Table, nm string) {
	plt.Options.Title = nm
	col1 := dt.ColumnName(1)
	plt.Options.XAxisColumn = col1
	plt.SetTable(dt)
	plt.Options.Lines = false
	plt.Options.Points = true
	// order of params: on, fixMin, min, fixMax, max
	plt.SetColumnOptions(dt.ColumnName(0), plotcore.On, plotcore.FloatMin, 0, plotcore.FloatMax, 0)
	plt.SetColumnOptions(col1, plotcore.Off, plotcore.FloatMin, -3, plotcore.FloatMax, 3)
	plt.SetColumnOptions(dt.ColumnName(2), plotcore.On, plotcore.FloatMin, -3, plotcore.FloatMax, 3)
}

// ClustPlot does one cluster plot on given table column name
// and label name
func ClustPlot(plt *plotcore.PlotEditor, ix *table.IndexView, colNm, lblNm string) {
	nm, _ := ix.Table.MetaData["name"]
	smat := &simat.SimMat{}
	smat.TableCol(ix, colNm, lblNm, false, metric.Euclidean64)
	pt := &table.Table{}
	clust.Plot(pt, clust.Glom(smat, clust.ContrastDist), smat)
	plt.Name = colNm
	plt.Options.Title = "Cluster Plot of: " + nm + " " + colNm
	plt.Options.XAxisColumn = "X"
	plt.SetTable(pt)
	// order of params: on, fixMin, min, fixMax, max
	plt.SetColumnOptions("X", plotcore.Off, plotcore.FixMin, 0, plotcore.FloatMax, 0)
	plt.SetColumnOptions("Y", plotcore.On, plotcore.FixMin, 0, plotcore.FloatMax, 0)
	plt.SetColumnOptions("Label", plotcore.On, plotcore.FixMin, 0, plotcore.FloatMax, 0)
}
