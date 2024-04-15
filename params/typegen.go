// Code generated by "core generate -add-types"; DO NOT EDIT.

package params

import (
	"cogentcore.org/core/types"
)

var _ = types.AddType(&types.Type{Name: "github.com/emer/emergent/v2/params.FlexVal", IDName: "flex-val", Doc: "FlexVal is a specific flexible value for the Flex parameter map\nthat implements the StylerObj interface for CSS-style selection logic.\nThe field names are abbreviated because full names are used in StylerObj.", Fields: []types.Field{{Name: "Nm", Doc: "name of this specific object, matches #Name selections"}, {Name: "Type", Doc: "type name of this object, matches plain TypeName selections"}, {Name: "Cls", Doc: "space-separated list of class name(s), match the .Class selections"}, {Name: "Obj", Doc: "actual object with data that is set by the parameters"}, {Name: "History", Doc: "History of params applied"}}})

var _ = types.AddType(&types.Type{Name: "github.com/emer/emergent/v2/params.Flex", IDName: "flex", Doc: "Flex supports arbitrary named parameter values that can be set\nby a Set of parameters, as a map of any objects.\nFirst initialize the map with set of names and a type to create\nblank values, then apply the Set to it."})

var _ = types.AddType(&types.Type{Name: "github.com/emer/emergent/v2/params.History", IDName: "history", Doc: "The params.History interface records history of parameters applied\nto a given object."})

var _ = types.AddType(&types.Type{Name: "github.com/emer/emergent/v2/params.HistoryImpl", IDName: "history-impl", Doc: "HistoryImpl implements the History interface.  Implementing object can\njust pass calls to a HistoryImpl field."})

var _ = types.AddType(&types.Type{Name: "github.com/emer/emergent/v2/params.HyperValues", IDName: "hyper-values", Doc: "HyperValues is a string-value map for storing hyperparameter values"})

var _ = types.AddType(&types.Type{Name: "github.com/emer/emergent/v2/params.Hypers", IDName: "hypers", Doc: "Hypers is a parallel structure to Params which stores information relevant\nto hyperparameter search as well as the values.\nUse the key \"Val\" for the default value. This is equivalant to the value in\nParams. \"Min\" and \"Max\" guid the range, and \"Sigma\" describes a Gaussian."})

var _ = types.AddType(&types.Type{Name: "github.com/emer/emergent/v2/params.Params", IDName: "params", Doc: "Params is a name-value map for parameter values that can be applied\nto any numeric type in any object.\nThe name must be a dot-separated path to a specific parameter, e.g., Prjn.Learn.Lrate\nThe first part of the path is the overall target object type, e.g., \"Prjn\" or \"Layer\",\nwhich is used for determining if the parameter applies to a given object type.\n\nAll of the params in one map must apply to the same target type because\nonly the first item in the map (which could be any due to order randomization)\nis used for checking the type of the target.  Also, they all fall within the same\nSel selector scope which is used to determine what specific objects to apply the\nparameters to."})

var _ = types.AddType(&types.Type{Name: "github.com/emer/emergent/v2/params.Sel", IDName: "sel", Doc: "params.Sel specifies a selector for the scope of application of a set of\nparameters, using standard css selector syntax (. prefix = class, # prefix = name,\nand no prefix = type)", Directives: []types.Directive{{Tool: "types", Directive: "add"}}, Fields: []types.Field{{Name: "Sel", Doc: "selector for what to apply the parameters to, using standard css selector syntax: .Example applies to anything with a Class tag of 'Example', #Example applies to anything with a Name of 'Example', and Example with no prefix applies to anything of type 'Example'"}, {Name: "Desc", Doc: "description of these parameter values -- what effect do they have?  what range was explored?  it is valuable to record this information as you explore the params."}, {Name: "Params", Doc: "parameter values to apply to whatever matches the selector"}, {Name: "Hypers", Doc: "Put your hyperparams here"}, {Name: "NMatch", Doc: "number of times this selector matched a target during the last Apply process -- a warning is issued for any that remain at 0 -- see Sheet SelMatchReset and SelNoMatchWarn methods"}, {Name: "SetName", Doc: "name of current Set being applied"}}})

var _ = types.AddType(&types.Type{Name: "github.com/emer/emergent/v2/params.Sheet", IDName: "sheet", Doc: "Sheet is a CSS-like style-sheet of params.Sel values, each of which represents\na different set of specific parameter values applied according to the Sel selector:\n.Class #Name or Type.\n\nThe order of elements in the Sheet list is critical, as they are applied\nin the order given by the list (slice), and thus later Sel's can override\nthose applied earlier.  Thus, you generally want to have more general Type-level\nparameters listed first, and then subsequently more specific ones (.Class and #Name)\n\nThis is the highest level of params that has an Apply method -- above this level\napplication must be done under explicit program control."})

var _ = types.AddType(&types.Type{Name: "github.com/emer/emergent/v2/params.Sheets", IDName: "sheets", Doc: "Sheets is a map of named sheets -- used in the Set"})

var _ = types.AddType(&types.Type{Name: "github.com/emer/emergent/v2/params.Set", IDName: "set", Doc: "Set is a collection of Sheets that constitute a coherent set of parameters --\na particular specific configuration of parameters, which the user selects to use.\nThe Set name is stored in the Sets map from which it is typically accessed.\nA good strategy is to have a \"Base\" set that has all the best parameters so far,\nand then other sets can modify relative to that one.  It is up to the Sim code to\napply parameter sets in whatever order is desired.\n\nWithin a params.Set, multiple different params.Sheets can be organized,\nwith each CSS-style sheet achieving a relatively complete parameter styling\nof a given element of the overal model, e.g., \"Network\", \"Sim\", \"Env\".\nOr Network could be further broken down into \"Learn\" vs. \"Act\" etc,\nor according to different brain areas (\"Hippo\", \"PFC\", \"BG\", etc).\nAgain, this is entirely at the discretion of the modeler and must be\nperformed under explict program control, especially because order is so critical.\n\nNote that there is NO deterministic ordering of the Sheets due to the use of\na Go map structure, which specifically randomizes order, so simply iterating over them\nand applying may produce unexpected results -- it is better to lookup by name.", Directives: []types.Directive{{Tool: "types", Directive: "add"}}, Fields: []types.Field{{Name: "Desc", Doc: "description of this param set -- when should it be used?  how is it different from the other sets?"}, {Name: "Sheets", Doc: "Sheet's grouped according to their target and / or function. For example,\n\"Network\" for all the network params (or \"Learn\" vs. \"Act\" for more fine-grained), and \"Sim\" for overall simulation control parameters, \"Env\" for environment parameters, etc.  It is completely up to your program to lookup these names and apply them as appropriate."}}})

var _ = types.AddType(&types.Type{Name: "github.com/emer/emergent/v2/params.Sets", IDName: "sets", Doc: "Sets is a collection of Set's that can be chosen among\ndepending on different desired configurations etc.  Thus, each Set\nrepresents a collection of different possible specific configurations,\nand different such configurations can be chosen by name to apply as desired."})

var _ = types.AddType(&types.Type{Name: "github.com/emer/emergent/v2/params.SearchValues", IDName: "search-values", Doc: "SearchValues is a list of parameter values to search for one parameter\non a given object (specified by Name), for float-valued params.", Fields: []types.Field{{Name: "Name", Doc: "name of object with the parameter"}, {Name: "Type", Doc: "type of object with the parameter. This is a Base type name (e.g., Layer, Prjn),\nthat is at the start of the path in Network params."}, {Name: "Path", Doc: "path to the parameter within the object"}, {Name: "Start", Doc: "starting value, e.g., for restoring after searching\nbefore moving on to another parameter, for grid search."}, {Name: "Values", Doc: "values of the parameter to search"}}})

var _ = types.AddType(&types.Type{Name: "github.com/emer/emergent/v2/params.Styler", IDName: "styler", Doc: "The params.Styler interface exposes TypeName, Class, and Name methods\nthat allow the params.Sel CSS-style selection specifier to determine\nwhether a given parameter applies.\nAdding Set versions of Name and Class methods is a good idea but not\nneeded for this interface, so they are not included here."})

var _ = types.AddType(&types.Type{Name: "github.com/emer/emergent/v2/params.StylerObj", IDName: "styler-obj", Doc: "The params.StylerObj interface extends Styler to include an arbitary\nfunction to access the underlying object type."})

var _ = types.AddType(&types.Type{Name: "github.com/emer/emergent/v2/params.Tweaks", IDName: "tweaks", Doc: "Tweaks holds parameter tweak values associated with one parameter selector.\nHas all the object values affected for a given parameter within one\nselector, that has a tweak hyperparameter set.", Fields: []types.Field{{Name: "Param", Doc: "the parameter path for this param"}, {Name: "Sel", Doc: "the param selector that set the specific value upon which tweak is based"}, {Name: "Search", Doc: "the search values for all objects covered by this selector"}}})