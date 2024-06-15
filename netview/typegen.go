// Code generated by "core generate -add-types"; DO NOT EDIT.

package netview

import (
	"sync"

	"cogentcore.org/core/colors/colormap"
	"cogentcore.org/core/core"
	"cogentcore.org/core/tree"
	"cogentcore.org/core/types"
)

var _ = types.AddType(&types.Type{Name: "github.com/emer/emergent/v2/netview.LayData", IDName: "lay-data", Doc: "LayData maintains a record of all the data for a given layer", Fields: []types.Field{{Name: "LayName", Doc: "the layer name"}, {Name: "NUnits", Doc: "cached number of units"}, {Name: "Data", Doc: "the full data, in that order"}, {Name: "RecvPaths", Doc: "receiving pathway data -- shared with SendPaths"}, {Name: "SendPaths", Doc: "sending pathway data -- shared with RecvPaths"}}})

var _ = types.AddType(&types.Type{Name: "github.com/emer/emergent/v2/netview.PathData", IDName: "path-data", Doc: "PathData holds display state for a pathway", Fields: []types.Field{{Name: "Send", Doc: "name of sending layer"}, {Name: "Recv", Doc: "name of recv layer"}, {Name: "Path", Doc: "source pathway"}, {Name: "SynData", Doc: "synaptic data, by variable in SynVars and number of data points"}}})

// SceneType is the [types.Type] for [Scene]
var SceneType = types.AddType(&types.Type{Name: "github.com/emer/emergent/v2/netview.Scene", IDName: "scene", Doc: "Scene is a Widget for managing the 3D Scene of the NetView", Embeds: []types.Field{{Name: "Scene"}}, Fields: []types.Field{{Name: "NetView"}}, Instance: &Scene{}})

// NewScene returns a new [Scene] with the given optional parent:
// Scene is a Widget for managing the 3D Scene of the NetView
func NewScene(parent ...tree.Node) *Scene { return tree.New[*Scene](parent...) }

// NodeType returns the [*types.Type] of [Scene]
func (t *Scene) NodeType() *types.Type { return SceneType }

// New returns a new [*Scene] value
func (t *Scene) New() tree.Node { return &Scene{} }

// SetNetView sets the [Scene.NetView]
func (t *Scene) SetNetView(v *NetView) *Scene { t.NetView = v; return t }

var _ = types.AddType(&types.Type{Name: "github.com/emer/emergent/v2/netview.LayMesh", IDName: "lay-mesh", Doc: "LayMesh is a xyz.Mesh that represents a layer -- it is dynamically updated using the\nUpdate method which only resets the essential Vertex elements.\nThe geometry is literal in the layer size: 0,0,0 lower-left corner and increasing X,Z\nfor the width and height of the layer, in unit (1) increments per unit..\nNetView applies an overall scaling to make it fit within the larger view.", Embeds: []types.Field{{Name: "MeshBase"}}, Fields: []types.Field{{Name: "Lay", Doc: "layer that we render"}, {Name: "Shape", Doc: "current shape that has been constructed -- if same, just update"}, {Name: "View", Doc: "netview that we're in"}}})

// LayObjType is the [types.Type] for [LayObj]
var LayObjType = types.AddType(&types.Type{Name: "github.com/emer/emergent/v2/netview.LayObj", IDName: "lay-obj", Doc: "LayObj is the Layer 3D object within the NetView", Directives: []types.Directive{{Tool: "types", Directive: "add"}}, Embeds: []types.Field{{Name: "Solid"}}, Fields: []types.Field{{Name: "LayName", Doc: "name of the layer we represent"}, {Name: "NetView", Doc: "our netview"}}, Instance: &LayObj{}})

// NewLayObj returns a new [LayObj] with the given optional parent:
// LayObj is the Layer 3D object within the NetView
func NewLayObj(parent ...tree.Node) *LayObj { return tree.New[*LayObj](parent...) }

// NodeType returns the [*types.Type] of [LayObj]
func (t *LayObj) NodeType() *types.Type { return LayObjType }

// New returns a new [*LayObj] value
func (t *LayObj) New() tree.Node { return &LayObj{} }

// SetLayName sets the [LayObj.LayName]:
// name of the layer we represent
func (t *LayObj) SetLayName(v string) *LayObj { t.LayName = v; return t }

// SetNetView sets the [LayObj.NetView]:
// our netview
func (t *LayObj) SetNetView(v *NetView) *LayObj { t.NetView = v; return t }

// LayNameType is the [types.Type] for [LayName]
var LayNameType = types.AddType(&types.Type{Name: "github.com/emer/emergent/v2/netview.LayName", IDName: "lay-name", Doc: "LayName is the Layer name as a Text2D within the NetView", Embeds: []types.Field{{Name: "Text2D"}}, Fields: []types.Field{{Name: "NetView", Doc: "our netview"}}, Instance: &LayName{}})

// NewLayName returns a new [LayName] with the given optional parent:
// LayName is the Layer name as a Text2D within the NetView
func NewLayName(parent ...tree.Node) *LayName { return tree.New[*LayName](parent...) }

// NodeType returns the [*types.Type] of [LayName]
func (t *LayName) NodeType() *types.Type { return LayNameType }

// New returns a new [*LayName] value
func (t *LayName) New() tree.Node { return &LayName{} }

// SetNetView sets the [LayName.NetView]:
// our netview
func (t *LayName) SetNetView(v *NetView) *LayName { t.NetView = v; return t }

var _ = types.AddType(&types.Type{Name: "github.com/emer/emergent/v2/netview.NetData", IDName: "net-data", Doc: "NetData maintains a record of all the network data that has been displayed\nup to a given maximum number of records (updates), using efficient ring index logic\nwith no copying to store in fixed-sized buffers.", Directives: []types.Directive{{Tool: "types", Directive: "add"}}, Methods: []types.Method{{Name: "OpenJSON", Doc: "OpenJSON opens colors from a JSON-formatted file.", Directives: []types.Directive{{Tool: "types", Directive: "add"}}, Args: []string{"filename"}, Returns: []string{"error"}}, {Name: "SaveJSON", Doc: "SaveJSON saves colors to a JSON-formatted file.", Directives: []types.Directive{{Tool: "types", Directive: "add"}}, Args: []string{"filename"}, Returns: []string{"error"}}}, Fields: []types.Field{{Name: "Net", Doc: "the network that we're viewing"}, {Name: "NoSynData", Doc: "copied from Params -- do not record synapse level data -- turn this on for very large networks where recording the entire synaptic state would be prohibitive"}, {Name: "PathLay", Doc: "name of the layer with unit for viewing pathways (connection / synapse-level values)"}, {Name: "PathUnIndex", Doc: "1D index of unit within PathLay for for viewing pathways"}, {Name: "PathType", Doc: "copied from NetView Params: if non-empty, this is the type pathway to show when there are multiple pathways from the same layer -- e.g., Inhib, Lateral, Forward, etc"}, {Name: "UnVars", Doc: "the list of unit variables saved"}, {Name: "UnVarIndexes", Doc: "index of each variable in the Vars slice"}, {Name: "SynVars", Doc: "the list of synaptic variables saved"}, {Name: "SynVarIndexes", Doc: "index of synaptic variable in the SynVars slice"}, {Name: "Ring", Doc: "the circular ring index -- Max here is max number of values to store, Len is number stored, and Index(Len-1) is the most recent one, etc"}, {Name: "MaxData", Doc: "max data parallel data per unit"}, {Name: "LayData", Doc: "the layer data -- map keyed by layer name"}, {Name: "UnMinPer", Doc: "unit var min values for each Ring.Max * variable"}, {Name: "UnMaxPer", Doc: "unit var max values for each Ring.Max * variable"}, {Name: "UnMinVar", Doc: "min values for unit variables"}, {Name: "UnMaxVar", Doc: "max values for unit variables"}, {Name: "SynMinVar", Doc: "min values for syn variables"}, {Name: "SynMaxVar", Doc: "max values for syn variables"}, {Name: "Counters", Doc: "counter strings"}, {Name: "RasterCtrs", Doc: "raster counter values"}, {Name: "RasterMap", Doc: "map of raster counter values to record numbers"}, {Name: "RastCtr", Doc: "dummy raster counter when passed a -1 -- increments and wraps around"}}})

// NetViewType is the [types.Type] for [NetView]
var NetViewType = types.AddType(&types.Type{Name: "github.com/emer/emergent/v2/netview.NetView", IDName: "net-view", Doc: "NetView is a Cogent Core Widget that provides a 3D network view using the Cogent Core gi3d\n3D framework.", Methods: []types.Method{{Name: "PlotSelectedUnit", Doc: "PlotSelectedUnit opens a window with a plot of all the data for the\ncurrently selected unit.\nUseful for replaying detailed trace for units of interest.", Directives: []types.Directive{{Tool: "types", Directive: "add"}}, Returns: []string{"Table", "PlotEditor"}}, {Name: "Current", Doc: "Current records the current state of the network, including synaptic values,\nand updates the display.  Use this when switching to NetView tab after network\nhas been running while viewing another tab, because the network state\nis typically not recored then.", Directives: []types.Directive{{Tool: "types", Directive: "add"}}}, {Name: "SaveWeights", Doc: "SaveWeights saves the network weights.", Directives: []types.Directive{{Tool: "types", Directive: "add"}}, Args: []string{"filename"}}, {Name: "OpenWeights", Doc: "OpenWeights opens the network weights.", Directives: []types.Directive{{Tool: "types", Directive: "add"}}, Args: []string{"filename"}}, {Name: "ShowNonDefaultParams", Doc: "ShowNonDefaultParams shows a dialog of all the parameters that\nare not at their default values in the network.  Useful for setting params.", Directives: []types.Directive{{Tool: "types", Directive: "add"}}, Returns: []string{"string"}}, {Name: "ShowAllParams", Doc: "ShowAllParams shows a dialog of all the parameters in the network.", Directives: []types.Directive{{Tool: "types", Directive: "add"}}, Returns: []string{"string"}}, {Name: "ShowKeyLayerParams", Doc: "ShowKeyLayerParams shows a dialog with a listing for all layers in the network,\nof the most important layer-level params (specific to each algorithm)", Directives: []types.Directive{{Tool: "types", Directive: "add"}}, Returns: []string{"string"}}, {Name: "ShowKeyPathParams", Doc: "ShowKeyPathParams shows a dialog with a listing for all Recv pathways in the network,\nof the most important pathway-level params (specific to each algorithm)", Directives: []types.Directive{{Tool: "types", Directive: "add"}}, Returns: []string{"string"}}}, Embeds: []types.Field{{Name: "Frame"}}, Fields: []types.Field{{Name: "Net", Doc: "the network that we're viewing"}, {Name: "Var", Doc: "current variable that we're viewing"}, {Name: "Di", Doc: "current data parallel index di, for networks capable of processing input patterns in parallel."}, {Name: "Vars", Doc: "the list of variables to view"}, {Name: "SynVars", Doc: "list of synaptic variables"}, {Name: "SynVarsMap", Doc: "map of synaptic variable names to index"}, {Name: "VarParams", Doc: "parameters for the list of variables to view"}, {Name: "CurVarParams", Doc: "current var params -- only valid during Update of display"}, {Name: "Params", Doc: "parameters controlling how the view is rendered"}, {Name: "ColorMap", Doc: "color map for mapping values to colors -- set by name in Params"}, {Name: "ColorMapButton", Doc: "color map value representing ColorMap"}, {Name: "RecNo", Doc: "record number to display -- use -1 to always track latest, otherwise in range"}, {Name: "LastCtrs", Doc: "last non-empty counters string provided -- re-used if no new one"}, {Name: "CurCtrs", Doc: "current counters"}, {Name: "Data", Doc: "contains all the network data with history"}, {Name: "DataMu", Doc: "mutex on data access"}}, Instance: &NetView{}})

// NewNetView returns a new [NetView] with the given optional parent:
// NetView is a Cogent Core Widget that provides a 3D network view using the Cogent Core gi3d
// 3D framework.
func NewNetView(parent ...tree.Node) *NetView { return tree.New[*NetView](parent...) }

// NodeType returns the [*types.Type] of [NetView]
func (t *NetView) NodeType() *types.Type { return NetViewType }

// New returns a new [*NetView] value
func (t *NetView) New() tree.Node { return &NetView{} }

// SetDi sets the [NetView.Di]:
// current data parallel index di, for networks capable of processing input patterns in parallel.
func (t *NetView) SetDi(v int) *NetView { t.Di = v; return t }

// SetVars sets the [NetView.Vars]:
// the list of variables to view
func (t *NetView) SetVars(v ...string) *NetView { t.Vars = v; return t }

// SetSynVars sets the [NetView.SynVars]:
// list of synaptic variables
func (t *NetView) SetSynVars(v ...string) *NetView { t.SynVars = v; return t }

// SetSynVarsMap sets the [NetView.SynVarsMap]:
// map of synaptic variable names to index
func (t *NetView) SetSynVarsMap(v map[string]int) *NetView { t.SynVarsMap = v; return t }

// SetVarParams sets the [NetView.VarParams]:
// parameters for the list of variables to view
func (t *NetView) SetVarParams(v map[string]*VarParams) *NetView { t.VarParams = v; return t }

// SetCurVarParams sets the [NetView.CurVarParams]:
// current var params -- only valid during Update of display
func (t *NetView) SetCurVarParams(v *VarParams) *NetView { t.CurVarParams = v; return t }

// SetParams sets the [NetView.Params]:
// parameters controlling how the view is rendered
func (t *NetView) SetParams(v Params) *NetView { t.Params = v; return t }

// SetColorMap sets the [NetView.ColorMap]:
// color map for mapping values to colors -- set by name in Params
func (t *NetView) SetColorMap(v *colormap.Map) *NetView { t.ColorMap = v; return t }

// SetColorMapButton sets the [NetView.ColorMapButton]:
// color map value representing ColorMap
func (t *NetView) SetColorMapButton(v *core.ColorMapButton) *NetView { t.ColorMapButton = v; return t }

// SetRecNo sets the [NetView.RecNo]:
// record number to display -- use -1 to always track latest, otherwise in range
func (t *NetView) SetRecNo(v int) *NetView { t.RecNo = v; return t }

// SetLastCtrs sets the [NetView.LastCtrs]:
// last non-empty counters string provided -- re-used if no new one
func (t *NetView) SetLastCtrs(v string) *NetView { t.LastCtrs = v; return t }

// SetCurCtrs sets the [NetView.CurCtrs]:
// current counters
func (t *NetView) SetCurCtrs(v string) *NetView { t.CurCtrs = v; return t }

// SetData sets the [NetView.Data]:
// contains all the network data with history
func (t *NetView) SetData(v NetData) *NetView { t.Data = v; return t }

// SetDataMu sets the [NetView.DataMu]:
// mutex on data access
func (t *NetView) SetDataMu(v sync.RWMutex) *NetView { t.DataMu = v; return t }

var _ = types.AddType(&types.Type{Name: "github.com/emer/emergent/v2/netview.RasterParams", IDName: "raster-params", Doc: "RasterParams holds parameters controlling the raster plot view", Directives: []types.Directive{{Tool: "types", Directive: "add"}}, Fields: []types.Field{{Name: "On", Doc: "if true, show a raster plot over time, otherwise units"}, {Name: "XAxis", Doc: "if true, the raster counter (time) is plotted across the X axis -- otherwise the Z depth axis"}, {Name: "Max", Doc: "maximum count for the counter defining the raster plot"}, {Name: "UnitSize", Doc: "size of a single unit, where 1 = full width and no space.. 1 default"}, {Name: "UnitHeight", Doc: "height multiplier for units, where 1 = full height.. 0.2 default"}}})

var _ = types.AddType(&types.Type{Name: "github.com/emer/emergent/v2/netview.Params", IDName: "params", Doc: "Params holds parameters controlling how the view is rendered", Directives: []types.Directive{{Tool: "types", Directive: "add"}}, Fields: []types.Field{{Name: "Raster", Doc: "raster plot parameters"}, {Name: "NoSynData", Doc: "do not record synapse level data -- turn this on for very large networks where recording the entire synaptic state would be prohibitive"}, {Name: "PathType", Doc: "if non-empty, this is the type pathway to show when there are multiple pathways from the same layer -- e.g., Inhib, Lateral, Forward, etc"}, {Name: "MaxRecs", Doc: "maximum number of records to store to enable rewinding through prior states"}, {Name: "NVarCols", Doc: "number of variable columns"}, {Name: "UnitSize", Doc: "size of a single unit, where 1 = full width and no space.. .9 default"}, {Name: "LayNmSize", Doc: "size of the layer name labels -- entire network view is unit sized"}, {Name: "ColorMap", Doc: "name of color map to use"}, {Name: "ZeroAlpha", Doc: "opacity (0-1) of zero values -- greater magnitude values become increasingly opaque on either side of this minimum"}, {Name: "NetView", Doc: "our netview, for update method"}, {Name: "NFastSteps", Doc: "the number of records to jump for fast forward/backward"}}})

var _ = types.AddType(&types.Type{Name: "github.com/emer/emergent/v2/netview.VarParams", IDName: "var-params", Doc: "VarParams holds parameters for display of each variable", Directives: []types.Directive{{Tool: "types", Directive: "add"}}, Fields: []types.Field{{Name: "Var", Doc: "name of the variable"}, {Name: "ZeroCtr", Doc: "keep Min - Max centered around 0, and use negative heights for units -- else use full min-max range for height (no negative heights)"}, {Name: "Range", Doc: "range to display"}, {Name: "MinMax", Doc: "if not using fixed range, this is the actual range of data"}}})

var _ = types.AddType(&types.Type{Name: "github.com/emer/emergent/v2/netview.ViewUpdate", IDName: "view-update", Doc: "ViewUpdate manages time scales for updating the NetView", Fields: []types.Field{{Name: "View", Doc: "the network view"}, {Name: "Testing", Doc: "whether in testing mode -- can be set in advance to drive appropriate updating"}, {Name: "Text", Doc: "text to display at the bottom of the view"}, {Name: "On", Doc: "toggles update of display on"}, {Name: "SkipInvis", Doc: "if true, do not record network data when the NetView is invisible -- this speeds up running when not visible, but the NetView display will not show the current state when switching back to it"}, {Name: "Train", Doc: "at what time scale to update the display during training?"}, {Name: "Test", Doc: "at what time scale to update the display during testing?"}}})
