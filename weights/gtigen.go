// Code generated by "core generate -add-types"; DO NOT EDIT.

package weights

import (
	"cogentcore.org/core/gti"
)

var _ = gti.AddType(&gti.Type{Name: "github.com/emer/emergent/v2/weights.Network", IDName: "network", Doc: "Network is temp structure for holding decoded weights", Directives: []gti.Directive{{Tool: "go", Directive: "generate", Args: []string{"core", "generate", "-add-types"}}}, Fields: []gti.Field{{Name: "Network"}, {Name: "MetaData"}, {Name: "Layers"}}})

var _ = gti.AddType(&gti.Type{Name: "github.com/emer/emergent/v2/weights.Layer", IDName: "layer", Doc: "Layer is temp structure for holding decoded weights, one for each layer", Fields: []gti.Field{{Name: "Layer"}, {Name: "MetaData"}, {Name: "Units"}, {Name: "Prjns"}}})

var _ = gti.AddType(&gti.Type{Name: "github.com/emer/emergent/v2/weights.Prjn", IDName: "prjn", Doc: "Prjn is temp structure for holding decoded weights, one for each projection", Fields: []gti.Field{{Name: "From"}, {Name: "MetaData"}, {Name: "MetaVals"}, {Name: "Rs"}}})

var _ = gti.AddType(&gti.Type{Name: "github.com/emer/emergent/v2/weights.Recv", IDName: "recv", Doc: "Recv is temp structure for holding decoded weights, one for each recv unit", Fields: []gti.Field{{Name: "Ri"}, {Name: "N"}, {Name: "Si"}, {Name: "Wt"}, {Name: "Wt1"}, {Name: "Wt2"}}})
