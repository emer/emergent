// Code generated by "core generate -add-types"; DO NOT EDIT.

package elog

import (
	"cogentcore.org/core/types"
)

var _ = types.AddType(&types.Type{Name: "github.com/emer/emergent/v2/elog.WriteFunc", IDName: "write-func", Doc: "WriteFunc function that computes and sets log values\nThe Context provides information typically needed for logging"})

var _ = types.AddType(&types.Type{Name: "github.com/emer/emergent/v2/elog.Context", IDName: "context", Doc: "Context provides the context for logging Write functions.\nSetContext must be called on Logs to set the Stats and Net values\nProvides various convenience functions for setting log values\nand other commonly used operations.", Fields: []types.Field{{Name: "Logs", Doc: "pointer to the Logs object with all log data"}, {Name: "Stats", Doc: "pointer to stats"}, {Name: "Net", Doc: "network"}, {Name: "Di", Doc: "data parallel index for accessing data from network"}, {Name: "Item", Doc: "current log Item"}, {Name: "Scope", Doc: "current scope key"}, {Name: "Mode", Doc: "current scope eval mode (if standard)"}, {Name: "Time", Doc: "current scope timescale (if standard)"}, {Name: "LogTable", Doc: "LogTable with extra data for the table"}, {Name: "Table", Doc: "current table to record value to"}, {Name: "Row", Doc: "current row in table to write to"}}})

var _ = types.AddType(&types.Type{Name: "github.com/emer/emergent/v2/elog.WriteMap", IDName: "write-map", Doc: "WriteMap holds log writing functions for scope keys"})

var _ = types.AddType(&types.Type{Name: "github.com/emer/emergent/v2/elog.Item", IDName: "item", Doc: "Item describes one item to be logged -- has all the info\nfor this item, across all scopes where it is relevant.", Fields: []types.Field{{Name: "Name", Doc: "name of column -- must be unique for a table"}, {Name: "Type", Doc: "data type, using etensor types which are isomorphic with arrow.Type"}, {Name: "CellShape", Doc: "shape of a single cell in the column (i.e., without the row dimension) -- for scalars this is nil -- tensor column will add the outer row dimension to this shape"}, {Name: "DimNames", Doc: "names of the dimensions within the CellShape -- 'Row' will be added to outer dimension"}, {Name: "Write", Doc: "holds Write functions for different scopes.  After processing, the scope key will be a single mode and time, from Scope(mode, time), but the initial specification can lists for each, or the All* option, if there is a Write function that works across scopes"}, {Name: "Plot", Doc: "Whether or not to plot it"}, {Name: "Range", Doc: "The minimum and maximum values, for plotting"}, {Name: "FixMin", Doc: "Whether to fix the minimum in the display"}, {Name: "FixMax", Doc: "Whether to fix the maximum in the display"}, {Name: "ErrCol", Doc: "Name of other item that has the error bar values for this item -- for plotting"}, {Name: "TensorIndex", Doc: "index of tensor to plot -- defaults to 0 -- use -1 to plot all"}, {Name: "Color", Doc: "specific color for plot -- uses default ordering of colors if empty"}, {Name: "Modes", Doc: "map of eval modes that this item has a Write function for"}, {Name: "Times", Doc: "map of times that this item has a Write function for"}}})

var _ = types.AddType(&types.Type{Name: "github.com/emer/emergent/v2/elog.Logs", IDName: "logs", Doc: "Logs contains all logging state and API for doing logging.\ndo AddItem to add any number of items, at different eval mode, time scopes.\nEach Item has its own Write functions, at each scope as neeeded.\nThen call CreateTables to generate log Tables from those items.\nCall Log with a scope to add a new row of data to the log\nand ResetLog to reset the log to empty.", Fields: []types.Field{{Name: "Tables", Doc: "Tables storing log data, auto-generated from Items."}, {Name: "MiscTables", Doc: "holds additional tables not computed from items -- e.g., aggregation results, intermediate computations, etc"}, {Name: "Items", Doc: "A list of the items that should be logged. Each item should describe one column that you want to log, and how.  Order in list determines order in logs."}, {Name: "Context", Doc: "context information passed to logging Write functions -- has all the information needed to compute and write log values -- is updated for each item in turn"}, {Name: "Modes", Doc: "All the eval modes that appear in any of the items of this log."}, {Name: "Times", Doc: "All the timescales that appear in any of the items of this log."}, {Name: "ItemIndexMap", Doc: "map of item indexes by name, for rapid access to items if they need to be modified after adding."}, {Name: "TableOrder", Doc: "sorted order of table scopes"}}})

var _ = types.AddType(&types.Type{Name: "github.com/emer/emergent/v2/elog.LogTable", IDName: "log-table", Doc: "LogTable contains all the data for one log table", Fields: []types.Field{{Name: "Table", Doc: "Actual data stored."}, {Name: "Meta", Doc: "arbitrary meta-data for each table, e.g., hints for plotting: Plot = false to not plot, XAxisCol, LegendCol"}, {Name: "IndexView", Doc: "Index View of the table -- automatically updated when a new row of data is logged to the table."}, {Name: "NamedViews", Doc: "named index views onto the table that can be saved and used across multiple items -- these are reset to nil after a new row is written -- see NamedIndexView funtion for more details."}, {Name: "File", Doc: "File to store the log into."}, {Name: "WroteHeaders", Doc: "true if headers for File have already been written"}}})
