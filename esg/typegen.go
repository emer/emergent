// Code generated by "core generate -add-types"; DO NOT EDIT.

package esg

import (
	"cogentcore.org/core/types"
)

var _ = types.AddType(&types.Type{Name: "github.com/emer/emergent/v2/esg.Conds", IDName: "conds", Doc: "Conds are conditionals"})

var _ = types.AddType(&types.Type{Name: "github.com/emer/emergent/v2/esg.Cond", IDName: "cond", Doc: "Cond is one element of a conditional", Fields: []types.Field{{Name: "El", Doc: "what type of conditional element is this"}, {Name: "Rule", Doc: "name of rule or token to evaluate for CRule"}, {Name: "Conds", Doc: "sub-conditions for SubCond"}}})

var _ = types.AddType(&types.Type{Name: "github.com/emer/emergent/v2/esg.CondEls", IDName: "cond-els", Doc: "CondEls are different types of conditional elements"})

var _ = types.AddType(&types.Type{Name: "github.com/emer/emergent/v2/esg.Item", IDName: "item", Doc: "Item is one item within a rule", Directives: []types.Directive{{Tool: "git", Directive: "add"}}, Fields: []types.Field{{Name: "Prob", Doc: "probability for choosing this item -- 0 if uniform random"}, {Name: "Elems", Doc: "elements of the rule -- for non-Cond rules"}, {Name: "Cond", Doc: "conditions for this item -- specified by ?"}, {Name: "SubRule", Doc: "for conditional, this is the sub-rule that is run with sub-items"}, {Name: "State", Doc: "state update name=value to set for rule"}}})

var _ = types.AddType(&types.Type{Name: "github.com/emer/emergent/v2/esg.Elem", IDName: "elem", Doc: "Elem is one elemenent in a concrete Item: either rule or token", Directives: []types.Directive{{Tool: "git", Directive: "add"}}, Fields: []types.Field{{Name: "El", Doc: "type of element: Rule, Token, or SubItems"}, {Name: "Value", Doc: "value of the token: name of Rule or Token"}}})

var _ = types.AddType(&types.Type{Name: "github.com/emer/emergent/v2/esg.Elements", IDName: "elements", Doc: "Elements are different types of elements"})

var _ = types.AddType(&types.Type{Name: "github.com/emer/emergent/v2/esg.State", IDName: "state", Doc: "State holds the name=value state settings associated with rule or item\nas a string, string map"})

var _ = types.AddType(&types.Type{Name: "github.com/emer/emergent/v2/esg.RuleTypes", IDName: "rule-types", Doc: "RuleTypes are different types of rules (i.e., how the items are selected)"})

var _ = types.AddType(&types.Type{Name: "github.com/emer/emergent/v2/esg.Rule", IDName: "rule", Doc: "Rule is one rule containing some number of items", Directives: []types.Directive{{Tool: "git", Directive: "add"}}, Fields: []types.Field{{Name: "Name", Doc: "name of rule"}, {Name: "Desc", Doc: "description / notes on rule"}, {Name: "Type", Doc: "type of rule -- how to choose the items"}, {Name: "Items", Doc: "items in rule"}, {Name: "State", Doc: "state update for rule"}, {Name: "PrevIndex", Doc: "previously selected item (from perspective of current rule)"}, {Name: "CurIndex", Doc: "current index in Items (what will be used next)"}, {Name: "RepeatP", Doc: "probability of repeating same item -- signaled by =%p"}, {Name: "Order", Doc: "permuted order if doing that"}}})

var _ = types.AddType(&types.Type{Name: "github.com/emer/emergent/v2/esg.Rules", IDName: "rules", Doc: "Rules is a collection of rules", Directives: []types.Directive{{Tool: "git", Directive: "add"}}, Fields: []types.Field{{Name: "Name", Doc: "name of this rule collection"}, {Name: "Desc", Doc: "description of this rule collection"}, {Name: "Trace", Doc: "if true, will print out a trace during generation"}, {Name: "Top", Doc: "top-level rule -- this is where to start generating"}, {Name: "Map", Doc: "map of each rule"}, {Name: "Fired", Doc: "map of names of all the rules that have fired"}, {Name: "Output", Doc: "array of output strings -- appended as the rules generate output"}, {Name: "States", Doc: "user-defined state map optionally created during generation"}, {Name: "ParseErrs", Doc: "errors from parsing"}, {Name: "ParseLn", Doc: "current line number during parsing"}}})
