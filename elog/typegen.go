// Code generated by "core generate -add-types"; DO NOT EDIT.

package elog

import (
	"reflect"

	"cogentcore.org/core/math32/minmax"
	"cogentcore.org/core/types"
)

var _ = types.AddType(&types.Type{Name: "github.com/emer/emergent/v2/elog.WriteFunc", IDName: "write-func", Doc: "WriteFunc function that computes and sets log values\nThe Context provides information typically needed for logging"})

var _ = types.AddType(&types.Type{Name: "github.com/emer/emergent/v2/elog.Context", IDName: "context", Doc: "Context provides the context for logging Write functions.\nSetContext must be called on Logs to set the Stats and Net values\nProvides various convenience functions for setting log values\nand other commonly used operations.", Fields: []types.Field{{Name: "Logs", Doc: "pointer to the Logs object with all log data"}, {Name: "Stats", Doc: "pointer to stats"}, {Name: "Net", Doc: "network"}, {Name: "Di", Doc: "data parallel index for accessing data from network"}, {Name: "Item", Doc: "current log Item"}, {Name: "Scope", Doc: "current scope key"}, {Name: "Mode", Doc: "current scope eval mode (if standard)"}, {Name: "Time", Doc: "current scope timescale (if standard)"}, {Name: "LogTable", Doc: "LogTable with extra data for the table"}, {Name: "Table", Doc: "current table to record value to"}, {Name: "Row", Doc: "current row in table to write to"}}})

var _ = types.AddType(&types.Type{Name: "github.com/emer/emergent/v2/elog.WriteMap", IDName: "write-map", Doc: "WriteMap holds log writing functions for scope keys"})

var _ = types.AddType(&types.Type{Name: "github.com/emer/emergent/v2/elog.Item", IDName: "item", Doc: "Item describes one item to be logged -- has all the info\nfor this item, across all scopes where it is relevant.", Directives: []types.Directive{{Tool: "types", Directive: "add", Args: []string{"-setters"}}}, Fields: []types.Field{{Name: "Name", Doc: "name of column -- must be unique for a table"}, {Name: "Type", Doc: "data type, using tensor types which are isomorphic with arrow.Type"}, {Name: "CellShape", Doc: "shape of a single cell in the column (i.e., without the row dimension) -- for scalars this is nil -- tensor column will add the outer row dimension to this shape"}, {Name: "DimNames", Doc: "names of the dimensions within the CellShape -- 'Row' will be added to outer dimension"}, {Name: "Write", Doc: "holds Write functions for different scopes.  After processing, the scope key will be a single mode and time, from Scope(mode, time), but the initial specification can lists for each, or the All* option, if there is a Write function that works across scopes"}, {Name: "Plot", Doc: "Whether or not to plot it"}, {Name: "Range", Doc: "The minimum and maximum values, for plotting"}, {Name: "FixMin", Doc: "Whether to fix the minimum in the display"}, {Name: "FixMax", Doc: "Whether to fix the maximum in the display"}, {Name: "ErrCol", Doc: "Name of other item that has the error bar values for this item -- for plotting"}, {Name: "TensorIndex", Doc: "index of tensor to plot -- defaults to 0 -- use -1 to plot all"}, {Name: "Color", Doc: "specific color for plot -- uses default ordering of colors if empty"}, {Name: "Modes", Doc: "map of eval modes that this item has a Write function for"}, {Name: "Times", Doc: "map of times that this item has a Write function for"}}})

// SetName sets the [Item.Name]:
// name of column -- must be unique for a table
func (t *Item) SetName(v string) *Item { t.Name = v; return t }

// SetType sets the [Item.Type]:
// data type, using tensor types which are isomorphic with arrow.Type
func (t *Item) SetType(v reflect.Kind) *Item { t.Type = v; return t }

// SetCellShape sets the [Item.CellShape]:
// shape of a single cell in the column (i.e., without the row dimension) -- for scalars this is nil -- tensor column will add the outer row dimension to this shape
func (t *Item) SetCellShape(v ...int) *Item { t.CellShape = v; return t }

// SetDimNames sets the [Item.DimNames]:
// names of the dimensions within the CellShape -- 'Row' will be added to outer dimension
func (t *Item) SetDimNames(v ...string) *Item { t.DimNames = v; return t }

// SetWrite sets the [Item.Write]:
// holds Write functions for different scopes.  After processing, the scope key will be a single mode and time, from Scope(mode, time), but the initial specification can lists for each, or the All* option, if there is a Write function that works across scopes
func (t *Item) SetWrite(v WriteMap) *Item { t.Write = v; return t }

// SetPlot sets the [Item.Plot]:
// Whether or not to plot it
func (t *Item) SetPlot(v bool) *Item { t.Plot = v; return t }

// SetRange sets the [Item.Range]:
// The minimum and maximum values, for plotting
func (t *Item) SetRange(v minmax.F32) *Item { t.Range = v; return t }

// SetFixMin sets the [Item.FixMin]:
// Whether to fix the minimum in the display
func (t *Item) SetFixMin(v bool) *Item { t.FixMin = v; return t }

// SetFixMax sets the [Item.FixMax]:
// Whether to fix the maximum in the display
func (t *Item) SetFixMax(v bool) *Item { t.FixMax = v; return t }

// SetErrCol sets the [Item.ErrCol]:
// Name of other item that has the error bar values for this item -- for plotting
func (t *Item) SetErrCol(v string) *Item { t.ErrCol = v; return t }

// SetTensorIndex sets the [Item.TensorIndex]:
// index of tensor to plot -- defaults to 0 -- use -1 to plot all
func (t *Item) SetTensorIndex(v int) *Item { t.TensorIndex = v; return t }

// SetColor sets the [Item.Color]:
// specific color for plot -- uses default ordering of colors if empty
func (t *Item) SetColor(v string) *Item { t.Color = v; return t }

// SetModes sets the [Item.Modes]:
// map of eval modes that this item has a Write function for
func (t *Item) SetModes(v map[string]bool) *Item { t.Modes = v; return t }

// SetTimes sets the [Item.Times]:
// map of times that this item has a Write function for
func (t *Item) SetTimes(v map[string]bool) *Item { t.Times = v; return t }

var _ = types.AddType(&types.Type{Name: "github.com/emer/emergent/v2/elog.Logs", IDName: "logs", Doc: "Logs contains all logging state and API for doing logging.\ndo AddItem to add any number of items, at different eval mode, time scopes.\nEach Item has its own Write functions, at each scope as neeeded.\nThen call CreateTables to generate log Tables from those items.\nCall Log with a scope to add a new row of data to the log\nand ResetLog to reset the log to empty.", Fields: []types.Field{{Name: "Tables", Doc: "Tables storing log data, auto-generated from Items."}, {Name: "MiscTables", Doc: "holds additional tables not computed from items -- e.g., aggregation results, intermediate computations, etc"}, {Name: "Items", Doc: "A list of the items that should be logged. Each item should describe one column that you want to log, and how.  Order in list determines order in logs."}, {Name: "Context", Doc: "context information passed to logging Write functions -- has all the information needed to compute and write log values -- is updated for each item in turn"}, {Name: "Modes", Doc: "All the eval modes that appear in any of the items of this log."}, {Name: "Times", Doc: "All the timescales that appear in any of the items of this log."}, {Name: "ItemIndexMap", Doc: "map of item indexes by name, for rapid access to items if they need to be modified after adding."}, {Name: "TableOrder", Doc: "sorted order of table scopes"}}})

var _ = types.AddType(&types.Type{Name: "github.com/emer/emergent/v2/elog.LogTable", IDName: "log-table", Doc: "LogTable contains all the data for one log table", Fields: []types.Field{{Name: "Table", Doc: "Actual data stored."}, {Name: "Meta", Doc: "arbitrary meta-data for each table, e.g., hints for plotting: Plot = false to not plot, XAxis, LegendCol"}, {Name: "IndexView", Doc: "Index View of the table -- automatically updated when a new row of data is logged to the table."}, {Name: "NamedViews", Doc: "named index views onto the table that can be saved and used across multiple items -- these are reset to nil after a new row is written -- see NamedIndexView funtion for more details."}, {Name: "File", Doc: "File to store the log into."}, {Name: "WroteHeaders", Doc: "true if headers for File have already been written"}}})
