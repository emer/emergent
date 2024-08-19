// Code generated by "core generate -add-types"; DO NOT EDIT.

package egui

import (
	"cogentcore.org/core/types"
)

var _ = types.AddType(&types.Type{Name: "github.com/emer/emergent/v2/egui.GUI", IDName: "gui", Doc: "GUI manages all standard elements of a simulation Graphical User Interface", Fields: []types.Field{{Name: "CycleUpdateInterval", Doc: "how many cycles between updates of cycle-level plots"}, {Name: "Active", Doc: "true if the GUI is configured and running"}, {Name: "IsRunning", Doc: "true if sim is running"}, {Name: "StopNow", Doc: "flag to stop running"}, {Name: "Plots", Doc: "plots by scope"}, {Name: "TableViews", Doc: "plots by scope"}, {Name: "Grids", Doc: "tensor grid views by name -- used e.g., for Rasters or ActRFs -- use Grid(name) to access"}, {Name: "ViewUpdate", Doc: "the view update for managing updates of netview"}, {Name: "NetData", Doc: "net data for recording in nogui mode, if !nil"}, {Name: "SimForm", Doc: "displays Sim fields on left"}, {Name: "Tabs", Doc: "tabs for different view elements: plots, rasters"}, {Name: "Body", Doc: "Body is the content of the sim window"}}})

var _ = types.AddType(&types.Type{Name: "github.com/emer/emergent/v2/egui.ToolbarItem", IDName: "toolbar-item", Doc: "ToolbarItem holds the configuration values for a toolbar item", Fields: []types.Field{{Name: "Label"}, {Name: "Icon"}, {Name: "Tooltip"}, {Name: "Active"}, {Name: "Func"}}})

var _ = types.AddType(&types.Type{Name: "github.com/emer/emergent/v2/egui.ToolGhosting", IDName: "tool-ghosting", Doc: "ToolGhosting the mode enum"})
