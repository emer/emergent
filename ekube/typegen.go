// Code generated by "core generate"; DO NOT EDIT.

package main

import (
	"cogentcore.org/core/types"
)

var _ = types.AddType(&types.Type{Name: "main.Config", IDName: "config", Directives: []types.Directive{{Tool: "types", Directive: "add"}}, Fields: []types.Field{{Name: "Dir", Doc: "Dir is the directory of the model to build."}}})

var _ = types.AddFunc(&types.Func{Name: "main.Build", Doc: "Build builds a Docker image for the emergent model in the current directory.", Directives: []types.Directive{{Tool: "types", Directive: "add"}}, Args: []string{"c"}, Returns: []string{"error"}})
