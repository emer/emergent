// Code generated by "goki generate ./..."; DO NOT EDIT.

package relpos

import (
	"goki.dev/gti"
	"goki.dev/ordmap"
)

var _ = gti.AddType(&gti.Type{
	Name:      "github.com/emer/emergent/v2/relpos.Rel",
	ShortName: "relpos.Rel",
	IDName:    "rel",
	Doc:       "Rel defines a position relationship among layers, in terms of X,Y width and height of layer\nand associated position within a given X-Y plane,\nand Z vertical stacking of layers above and below each other.",
	Directives: gti.Directives{
		&gti.Directive{Tool: "git", Directive: "add", Args: []string{}},
	},
	Fields: ordmap.Make([]ordmap.KeyVal[string, *gti.Field]{
		{"Rel", &gti.Field{Name: "Rel", Type: "github.com/emer/emergent/v2/relpos.Relations", LocalType: "Relations", Doc: "spatial relationship between this layer and the other layer", Directives: gti.Directives{}, Tag: ""}},
		{"XAlign", &gti.Field{Name: "XAlign", Type: "github.com/emer/emergent/v2/relpos.XAligns", LocalType: "XAligns", Doc: "] horizontal (x-axis) alignment relative to other", Directives: gti.Directives{}, Tag: "viewif:\"Rel=[FrontOf,Behind,Above,Below]\""}},
		{"YAlign", &gti.Field{Name: "YAlign", Type: "github.com/emer/emergent/v2/relpos.YAligns", LocalType: "YAligns", Doc: "] vertical (y-axis) alignment relative to other", Directives: gti.Directives{}, Tag: "viewif:\"Rel=[LeftOf,RightOf,Above,Below]\""}},
		{"Other", &gti.Field{Name: "Other", Type: "string", LocalType: "string", Doc: "name of the other layer we are in relationship to", Directives: gti.Directives{}, Tag: ""}},
		{"Scale", &gti.Field{Name: "Scale", Type: "float32", LocalType: "float32", Doc: "scaling factor applied to layer size for displaying", Directives: gti.Directives{}, Tag: ""}},
		{"Space", &gti.Field{Name: "Space", Type: "float32", LocalType: "float32", Doc: "number of unit-spaces between us", Directives: gti.Directives{}, Tag: ""}},
		{"XOffset", &gti.Field{Name: "XOffset", Type: "float32", LocalType: "float32", Doc: "for vertical (y-axis) alignment, amount we are offset relative to perfect alignment", Directives: gti.Directives{}, Tag: ""}},
		{"YOffset", &gti.Field{Name: "YOffset", Type: "float32", LocalType: "float32", Doc: "for horizontial (x-axis) alignment, amount we are offset relative to perfect alignment", Directives: gti.Directives{}, Tag: ""}},
	}),
	Embeds:  ordmap.Make([]ordmap.KeyVal[string, *gti.Field]{}),
	Methods: ordmap.Make([]ordmap.KeyVal[string, *gti.Method]{}),
})

var _ = gti.AddType(&gti.Type{
	Name:      "github.com/emer/emergent/v2/relpos.Relations",
	ShortName: "relpos.Relations",
	IDName:    "relations",
	Doc:       "Relations are different spatial relationships (of layers)",
	Directives: gti.Directives{
		&gti.Directive{Tool: "enums", Directive: "enum", Args: []string{}},
	},

	Methods: ordmap.Make([]ordmap.KeyVal[string, *gti.Method]{}),
})

var _ = gti.AddType(&gti.Type{
	Name:      "github.com/emer/emergent/v2/relpos.XAligns",
	ShortName: "relpos.XAligns",
	IDName:    "x-aligns",
	Doc:       "XAligns are different horizontal alignments",
	Directives: gti.Directives{
		&gti.Directive{Tool: "enums", Directive: "enum", Args: []string{}},
	},

	Methods: ordmap.Make([]ordmap.KeyVal[string, *gti.Method]{}),
})

var _ = gti.AddType(&gti.Type{
	Name:      "github.com/emer/emergent/v2/relpos.YAligns",
	ShortName: "relpos.YAligns",
	IDName:    "y-aligns",
	Doc:       "YAligns are different vertical alignments",
	Directives: gti.Directives{
		&gti.Directive{Tool: "enums", Directive: "enum", Args: []string{}},
	},

	Methods: ordmap.Make([]ordmap.KeyVal[string, *gti.Method]{}),
})