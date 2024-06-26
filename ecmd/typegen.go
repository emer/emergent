// Code generated by "core generate -add-types"; DO NOT EDIT.

package ecmd

import (
	"cogentcore.org/core/types"
)

var _ = types.AddType(&types.Type{Name: "github.com/emer/emergent/v2/ecmd.Args", IDName: "args", Doc: "Args provides maps for storing commandline args.", Fields: []types.Field{{Name: "Ints"}, {Name: "Bools"}, {Name: "Strings"}, {Name: "Floats"}, {Name: "Flagged", Doc: "true when all args have been set to flag package"}}})

var _ = types.AddType(&types.Type{Name: "github.com/emer/emergent/v2/ecmd.Int", IDName: "int", Doc: "Int represents a int valued arg", Fields: []types.Field{{Name: "Name", Doc: "name of arg -- must be unique"}, {Name: "Desc", Doc: "description of arg"}, {Name: "Val", Doc: "value as parsed"}, {Name: "Def", Doc: "default initial value"}}})

var _ = types.AddType(&types.Type{Name: "github.com/emer/emergent/v2/ecmd.Bool", IDName: "bool", Doc: "Bool represents a bool valued arg", Fields: []types.Field{{Name: "Name", Doc: "name of arg -- must be unique"}, {Name: "Desc", Doc: "description of arg"}, {Name: "Val", Doc: "value as parsed"}, {Name: "Def", Doc: "default initial value"}}})

var _ = types.AddType(&types.Type{Name: "github.com/emer/emergent/v2/ecmd.String", IDName: "string", Doc: "String represents a string valued arg", Fields: []types.Field{{Name: "Name", Doc: "name of arg -- must be unique"}, {Name: "Desc", Doc: "description of arg"}, {Name: "Val", Doc: "value as parsed"}, {Name: "Def", Doc: "default initial value"}}})

var _ = types.AddType(&types.Type{Name: "github.com/emer/emergent/v2/ecmd.Float", IDName: "float", Doc: "Float represents a float64 valued arg", Fields: []types.Field{{Name: "Name", Doc: "name of arg -- must be unique"}, {Name: "Desc", Doc: "description of arg"}, {Name: "Val", Doc: "value as parsed"}, {Name: "Def", Doc: "default initial value"}}})
