// Code generated by "goki generate ./..."; DO NOT EDIT.

package env

import (
	"goki.dev/gti"
	"goki.dev/ordmap"
)

var _ = gti.AddType(&gti.Type{
	Name:       "github.com/emer/emergent/v2/env.Ctr",
	ShortName:  "env.Ctr",
	IDName:     "ctr",
	Doc:        "Ctr is a counter that counts increments at a given time scale.\nIt keeps track of when it has been incremented or not, and\nretains the previous value.",
	Directives: gti.Directives{},
	Fields: ordmap.Make([]ordmap.KeyVal[string, *gti.Field]{
		{"Cur", &gti.Field{Name: "Cur", Type: "int", LocalType: "int", Doc: "current counter value", Directives: gti.Directives{}, Tag: ""}},
		{"Prv", &gti.Field{Name: "Prv", Type: "int", LocalType: "int", Doc: "previous counter value, prior to last Incr() call (init to -1)", Directives: gti.Directives{}, Tag: "view:\"-\""}},
		{"Chg", &gti.Field{Name: "Chg", Type: "bool", LocalType: "bool", Doc: "did this change on the last Step() call or not?", Directives: gti.Directives{}, Tag: "view:\"-\""}},
		{"Max", &gti.Field{Name: "Max", Type: "int", LocalType: "int", Doc: "where relevant, this is a fixed maximum counter value, above which the counter will reset back to 0 -- only used if > 0", Directives: gti.Directives{}, Tag: ""}},
		{"Scale", &gti.Field{Name: "Scale", Type: "github.com/emer/emergent/v2/env.TimeScales", LocalType: "TimeScales", Doc: "the unit of time scale represented by this counter (just FYI)", Directives: gti.Directives{}, Tag: "view:\"-\""}},
	}),
	Embeds:  ordmap.Make([]ordmap.KeyVal[string, *gti.Field]{}),
	Methods: ordmap.Make([]ordmap.KeyVal[string, *gti.Method]{}),
})

var _ = gti.AddType(&gti.Type{
	Name:       "github.com/emer/emergent/v2/env.Ctrs",
	ShortName:  "env.Ctrs",
	IDName:     "ctrs",
	Doc:        "Ctrs contains an ordered slice of timescales,\nand a lookup map of counters by timescale\nused to manage counters in the Env.",
	Directives: gti.Directives{},
	Fields: ordmap.Make([]ordmap.KeyVal[string, *gti.Field]{
		{"Order", &gti.Field{Name: "Order", Type: "[]github.com/emer/emergent/v2/env.TimeScales", LocalType: "[]TimeScales", Doc: "ordered list of the counter timescales, from outer-most (highest) to inner-most (lowest)", Directives: gti.Directives{}, Tag: ""}},
		{"Ctrs", &gti.Field{Name: "Ctrs", Type: "map[github.com/emer/emergent/v2/env.TimeScales]*github.com/emer/emergent/v2/env.Ctr", LocalType: "map[TimeScales]*Ctr", Doc: "map of the counters by timescale", Directives: gti.Directives{}, Tag: ""}},
	}),
	Embeds:  ordmap.Make([]ordmap.KeyVal[string, *gti.Field]{}),
	Methods: ordmap.Make([]ordmap.KeyVal[string, *gti.Method]{}),
})

var _ = gti.AddType(&gti.Type{
	Name:       "github.com/emer/emergent/v2/env.CurPrvF32",
	ShortName:  "env.CurPrvF32",
	IDName:     "cur-prv-f-32",
	Doc:        "CurPrvF32 is basic state management for current and previous values, float32 values",
	Directives: gti.Directives{},
	Fields: ordmap.Make([]ordmap.KeyVal[string, *gti.Field]{
		{"Cur", &gti.Field{Name: "Cur", Type: "float32", LocalType: "float32", Doc: "current value", Directives: gti.Directives{}, Tag: ""}},
		{"Prv", &gti.Field{Name: "Prv", Type: "float32", LocalType: "float32", Doc: "previous value", Directives: gti.Directives{}, Tag: ""}},
	}),
	Embeds:  ordmap.Make([]ordmap.KeyVal[string, *gti.Field]{}),
	Methods: ordmap.Make([]ordmap.KeyVal[string, *gti.Method]{}),
})

var _ = gti.AddType(&gti.Type{
	Name:       "github.com/emer/emergent/v2/env.CurPrvInt",
	ShortName:  "env.CurPrvInt",
	IDName:     "cur-prv-int",
	Doc:        "CurPrvInt is basic state management for current and previous values, int values",
	Directives: gti.Directives{},
	Fields: ordmap.Make([]ordmap.KeyVal[string, *gti.Field]{
		{"Cur", &gti.Field{Name: "Cur", Type: "int", LocalType: "int", Doc: "current value", Directives: gti.Directives{}, Tag: ""}},
		{"Prv", &gti.Field{Name: "Prv", Type: "int", LocalType: "int", Doc: "previous value", Directives: gti.Directives{}, Tag: ""}},
	}),
	Embeds:  ordmap.Make([]ordmap.KeyVal[string, *gti.Field]{}),
	Methods: ordmap.Make([]ordmap.KeyVal[string, *gti.Method]{}),
})

var _ = gti.AddType(&gti.Type{
	Name:       "github.com/emer/emergent/v2/env.CurPrvString",
	ShortName:  "env.CurPrvString",
	IDName:     "cur-prv-string",
	Doc:        "CurPrvString is basic state management for current and previous values, string values",
	Directives: gti.Directives{},
	Fields: ordmap.Make([]ordmap.KeyVal[string, *gti.Field]{
		{"Cur", &gti.Field{Name: "Cur", Type: "string", LocalType: "string", Doc: "current value", Directives: gti.Directives{}, Tag: ""}},
		{"Prv", &gti.Field{Name: "Prv", Type: "string", LocalType: "string", Doc: "previous value", Directives: gti.Directives{}, Tag: ""}},
	}),
	Embeds:  ordmap.Make([]ordmap.KeyVal[string, *gti.Field]{}),
	Methods: ordmap.Make([]ordmap.KeyVal[string, *gti.Method]{}),
})

var _ = gti.AddType(&gti.Type{
	Name:       "github.com/emer/emergent/v2/env.Element",
	ShortName:  "env.Element",
	IDName:     "element",
	Doc:        "Element specifies one element of State or Action in an environment",
	Directives: gti.Directives{},
	Fields: ordmap.Make([]ordmap.KeyVal[string, *gti.Field]{
		{"Name", &gti.Field{Name: "Name", Type: "string", LocalType: "string", Doc: "name of this element -- must be unique", Directives: gti.Directives{}, Tag: ""}},
		{"Shape", &gti.Field{Name: "Shape", Type: "[]int", LocalType: "[]int", Doc: "shape of the tensor for this element -- each element should generally have a well-defined consistent shape to enable the model to process it consistently", Directives: gti.Directives{}, Tag: ""}},
		{"DimNames", &gti.Field{Name: "DimNames", Type: "[]string", LocalType: "[]string", Doc: "names of the dimensions within the Shape -- optional but useful for ensuring correct usage", Directives: gti.Directives{}, Tag: ""}},
	}),
	Embeds:  ordmap.Make([]ordmap.KeyVal[string, *gti.Field]{}),
	Methods: ordmap.Make([]ordmap.KeyVal[string, *gti.Method]{}),
})

var _ = gti.AddType(&gti.Type{
	Name:       "github.com/emer/emergent/v2/env.Elements",
	ShortName:  "env.Elements",
	IDName:     "elements",
	Doc:        "Elements is a list of Element info",
	Directives: gti.Directives{},

	Methods: ordmap.Make([]ordmap.KeyVal[string, *gti.Method]{}),
})

var _ = gti.AddType(&gti.Type{
	Name:       "github.com/emer/emergent/v2/env.Env",
	ShortName:  "env.Env",
	IDName:     "env",
	Doc:        "Env defines an interface for environments, which determine the nature and\nsequence of States that can be used as inputs to a model, and the Env\nalso can accept Action responses from the model that affect state evolution.\n\nThe Env encapsulates all of the counter management logic to advance\nthe temporal state of the environment, using TimeScales standard\nintervals.\n\nState is comprised of one or more Elements, each of which consists of an\netensor.Tensor chunk of values that can be obtained by the model.\nLikewise, Actions can also have Elements.  The Step method is the main\ninterface for advancing the Env state.  Counters should be queried\nafter calling Step to see if any relevant values have changed, to trigger\nfunctions in the model (e.g., logging of prior statistics, etc).\n\nTypically each specific implementation of this Env interface will have\nmultiple parameters etc that can be modified to control env behavior --\nall of this is paradigm-specific and outside the scope of this basic interface.",
	Directives: gti.Directives{},

	Methods: ordmap.Make([]ordmap.KeyVal[string, *gti.Method]{}),
})

var _ = gti.AddType(&gti.Type{
	Name:       "github.com/emer/emergent/v2/env.EnvDesc",
	ShortName:  "env.EnvDesc",
	IDName:     "env-desc",
	Doc:        "EnvDesc is an interface that defines methods that describe an Env.\nThese are optional for basic Env, but in cases where an Env\nshould be fully self-describing, these methods can be implemented.",
	Directives: gti.Directives{},

	Methods: ordmap.Make([]ordmap.KeyVal[string, *gti.Method]{}),
})

var _ = gti.AddType(&gti.Type{
	Name:       "github.com/emer/emergent/v2/env.Envs",
	ShortName:  "env.Envs",
	IDName:     "envs",
	Doc:        "Envs is a map of environments organized according\nto the evaluation mode string (recommended key value)",
	Directives: gti.Directives{},

	Methods: ordmap.Make([]ordmap.KeyVal[string, *gti.Method]{}),
})

var _ = gti.AddType(&gti.Type{
	Name:       "github.com/emer/emergent/v2/env.FixedTable",
	ShortName:  "env.FixedTable",
	IDName:     "fixed-table",
	Doc:        "FixedTable is a basic Env that manages patterns from an etable.Table, with\neither sequential or permuted random ordering, and uses standard Trial / Epoch\nTimeScale counters to record progress and iterations through the table.\nIt also records the outer loop of Run as provided by the model.\nIt uses an IdxView indexed view of the Table, so a single shared table\ncan be used across different environments, with each having its own unique view.",
	Directives: gti.Directives{},
	Fields: ordmap.Make([]ordmap.KeyVal[string, *gti.Field]{
		{"Nm", &gti.Field{Name: "Nm", Type: "string", LocalType: "string", Doc: "name of this environment", Directives: gti.Directives{}, Tag: ""}},
		{"Dsc", &gti.Field{Name: "Dsc", Type: "string", LocalType: "string", Doc: "description of this environment", Directives: gti.Directives{}, Tag: ""}},
		{"Table", &gti.Field{Name: "Table", Type: "*goki.dev/etable/v2/etable.IdxView", LocalType: "*etable.IdxView", Doc: "this is an indexed view of the table with the set of patterns to output -- the indexes are used for the *sequential* view so you can easily sort / split / filter the patterns to be presented using this view -- we then add the random permuted Order on top of those if !sequential", Directives: gti.Directives{}, Tag: ""}},
		{"Sequential", &gti.Field{Name: "Sequential", Type: "bool", LocalType: "bool", Doc: "present items from the table in sequential order (i.e., according to the indexed view on the Table)?  otherwise permuted random order", Directives: gti.Directives{}, Tag: ""}},
		{"Order", &gti.Field{Name: "Order", Type: "[]int", LocalType: "[]int", Doc: "permuted order of items to present if not sequential -- updated every time through the list", Directives: gti.Directives{}, Tag: ""}},
		{"Run", &gti.Field{Name: "Run", Type: "github.com/emer/emergent/v2/env.Ctr", LocalType: "Ctr", Doc: "current run of model as provided during Init", Directives: gti.Directives{}, Tag: "view:\"inline\""}},
		{"Epoch", &gti.Field{Name: "Epoch", Type: "github.com/emer/emergent/v2/env.Ctr", LocalType: "Ctr", Doc: "number of times through entire set of patterns", Directives: gti.Directives{}, Tag: "view:\"inline\""}},
		{"Trial", &gti.Field{Name: "Trial", Type: "github.com/emer/emergent/v2/env.Ctr", LocalType: "Ctr", Doc: "current ordinal item in Table -- if Sequential then = row number in table, otherwise is index in Order list that then gives row number in Table", Directives: gti.Directives{}, Tag: "view:\"inline\""}},
		{"TrialName", &gti.Field{Name: "TrialName", Type: "github.com/emer/emergent/v2/env.CurPrvString", LocalType: "CurPrvString", Doc: "if Table has a Name column, this is the contents of that", Directives: gti.Directives{}, Tag: ""}},
		{"GroupName", &gti.Field{Name: "GroupName", Type: "github.com/emer/emergent/v2/env.CurPrvString", LocalType: "CurPrvString", Doc: "if Table has a Group column, this is contents of that", Directives: gti.Directives{}, Tag: ""}},
		{"NameCol", &gti.Field{Name: "NameCol", Type: "string", LocalType: "string", Doc: "name of the Name column -- defaults to 'Name'", Directives: gti.Directives{}, Tag: ""}},
		{"GroupCol", &gti.Field{Name: "GroupCol", Type: "string", LocalType: "string", Doc: "name of the Group column -- defaults to 'Group'", Directives: gti.Directives{}, Tag: ""}},
	}),
	Embeds:  ordmap.Make([]ordmap.KeyVal[string, *gti.Field]{}),
	Methods: ordmap.Make([]ordmap.KeyVal[string, *gti.Method]{}),
})

var _ = gti.AddType(&gti.Type{
	Name:       "github.com/emer/emergent/v2/env.FreqTable",
	ShortName:  "env.FreqTable",
	IDName:     "freq-table",
	Doc:        "FreqTable is an Env that manages patterns from an etable.Table with frequency\ninformation so that items are presented according to their associated frequencies\nwhich are effectively probabilities of presenting any given input -- must have\na Freq column with these numbers in the table (actual col name in FreqCol).\nEither sequential or permuted random ordering is supported, with std Trial / Epoch\nTimeScale counters to record progress and iterations through the table.\nIt also records the outer loop of Run as provided by the model.\nIt uses an IdxView indexed view of the Table, so a single shared table\ncan be used across different environments, with each having its own unique view.",
	Directives: gti.Directives{},
	Fields: ordmap.Make([]ordmap.KeyVal[string, *gti.Field]{
		{"Nm", &gti.Field{Name: "Nm", Type: "string", LocalType: "string", Doc: "name of this environment", Directives: gti.Directives{}, Tag: ""}},
		{"Dsc", &gti.Field{Name: "Dsc", Type: "string", LocalType: "string", Doc: "description of this environment", Directives: gti.Directives{}, Tag: ""}},
		{"Table", &gti.Field{Name: "Table", Type: "*goki.dev/etable/v2/etable.IdxView", LocalType: "*etable.IdxView", Doc: "this is an indexed view of the table with the set of patterns to output -- the indexes are used for the *sequential* view so you can easily sort / split / filter the patterns to be presented using this view -- we then add the random permuted Order on top of those if !sequential", Directives: gti.Directives{}, Tag: ""}},
		{"NSamples", &gti.Field{Name: "NSamples", Type: "float64", LocalType: "float64", Doc: "number of samples to use in constructing the list of items to present according to frequency -- number per epoch ~ NSamples * Freq -- see RndSamp option", Directives: gti.Directives{}, Tag: ""}},
		{"RndSamp", &gti.Field{Name: "RndSamp", Type: "bool", LocalType: "bool", Doc: "if true, use random sampling of items NSamples times according to given Freq probability value -- otherwise just directly add NSamples * Freq items to the list", Directives: gti.Directives{}, Tag: ""}},
		{"Sequential", &gti.Field{Name: "Sequential", Type: "bool", LocalType: "bool", Doc: "present items from the table in sequential order (i.e., according to the indexed view on the Table)?  otherwise permuted random order.  All repetitions of given item will be sequential if Sequential", Directives: gti.Directives{}, Tag: ""}},
		{"Order", &gti.Field{Name: "Order", Type: "[]int", LocalType: "[]int", Doc: "list of items to present, with repetitions -- updated every time through the list", Directives: gti.Directives{}, Tag: ""}},
		{"Run", &gti.Field{Name: "Run", Type: "github.com/emer/emergent/v2/env.Ctr", LocalType: "Ctr", Doc: "current run of model as provided during Init", Directives: gti.Directives{}, Tag: "view:\"inline\""}},
		{"Epoch", &gti.Field{Name: "Epoch", Type: "github.com/emer/emergent/v2/env.Ctr", LocalType: "Ctr", Doc: "number of times through entire set of patterns", Directives: gti.Directives{}, Tag: "view:\"inline\""}},
		{"Trial", &gti.Field{Name: "Trial", Type: "github.com/emer/emergent/v2/env.Ctr", LocalType: "Ctr", Doc: "current ordinal item in Table -- if Sequential then = row number in table, otherwise is index in Order list that then gives row number in Table", Directives: gti.Directives{}, Tag: "view:\"inline\""}},
		{"TrialName", &gti.Field{Name: "TrialName", Type: "github.com/emer/emergent/v2/env.CurPrvString", LocalType: "CurPrvString", Doc: "if Table has a Name column, this is the contents of that", Directives: gti.Directives{}, Tag: ""}},
		{"GroupName", &gti.Field{Name: "GroupName", Type: "github.com/emer/emergent/v2/env.CurPrvString", LocalType: "CurPrvString", Doc: "if Table has a Group column, this is contents of that", Directives: gti.Directives{}, Tag: ""}},
		{"NameCol", &gti.Field{Name: "NameCol", Type: "string", LocalType: "string", Doc: "name of the Name column -- defaults to 'Name'", Directives: gti.Directives{}, Tag: ""}},
		{"GroupCol", &gti.Field{Name: "GroupCol", Type: "string", LocalType: "string", Doc: "name of the Group column -- defaults to 'Group'", Directives: gti.Directives{}, Tag: ""}},
		{"FreqCol", &gti.Field{Name: "FreqCol", Type: "string", LocalType: "string", Doc: "name of the Freq column -- defaults to 'Freq'", Directives: gti.Directives{}, Tag: ""}},
	}),
	Embeds:  ordmap.Make([]ordmap.KeyVal[string, *gti.Field]{}),
	Methods: ordmap.Make([]ordmap.KeyVal[string, *gti.Method]{}),
})

var _ = gti.AddType(&gti.Type{
	Name:       "github.com/emer/emergent/v2/env.MPIFixedTable",
	ShortName:  "env.MPIFixedTable",
	IDName:     "mpi-fixed-table",
	Doc:        "MPIFixedTable is an MPI-enabled version of the FixedTable, which is\na basic Env that manages patterns from an etable.Table, with\neither sequential or permuted random ordering, and uses standard Trial / Epoch\nTimeScale counters to record progress and iterations through the table.\nIt also records the outer loop of Run as provided by the model.\nIt uses an IdxView indexed view of the Table, so a single shared table\ncan be used across different environments, with each having its own unique view.\nThe MPI version distributes trials across MPI procs, in the Order list.\nIt is ESSENTIAL that the number of trials (rows) in Table is\nevenly divisible by number of MPI procs!\nIf all nodes start with the same seed, it should remain synchronized.",
	Directives: gti.Directives{},
	Fields: ordmap.Make([]ordmap.KeyVal[string, *gti.Field]{
		{"Nm", &gti.Field{Name: "Nm", Type: "string", LocalType: "string", Doc: "name of this environment", Directives: gti.Directives{}, Tag: ""}},
		{"Dsc", &gti.Field{Name: "Dsc", Type: "string", LocalType: "string", Doc: "description of this environment", Directives: gti.Directives{}, Tag: ""}},
		{"Table", &gti.Field{Name: "Table", Type: "*goki.dev/etable/v2/etable.IdxView", LocalType: "*etable.IdxView", Doc: "this is an indexed view of the table with the set of patterns to output -- the indexes are used for the *sequential* view so you can easily sort / split / filter the patterns to be presented using this view -- we then add the random permuted Order on top of those if !sequential", Directives: gti.Directives{}, Tag: ""}},
		{"Sequential", &gti.Field{Name: "Sequential", Type: "bool", LocalType: "bool", Doc: "present items from the table in sequential order (i.e., according to the indexed view on the Table)?  otherwise permuted random order", Directives: gti.Directives{}, Tag: ""}},
		{"Order", &gti.Field{Name: "Order", Type: "[]int", LocalType: "[]int", Doc: "permuted order of items to present if not sequential -- updated every time through the list", Directives: gti.Directives{}, Tag: ""}},
		{"Run", &gti.Field{Name: "Run", Type: "github.com/emer/emergent/v2/env.Ctr", LocalType: "Ctr", Doc: "current run of model as provided during Init", Directives: gti.Directives{}, Tag: "view:\"inline\""}},
		{"Epoch", &gti.Field{Name: "Epoch", Type: "github.com/emer/emergent/v2/env.Ctr", LocalType: "Ctr", Doc: "number of times through entire set of patterns", Directives: gti.Directives{}, Tag: "view:\"inline\""}},
		{"Trial", &gti.Field{Name: "Trial", Type: "github.com/emer/emergent/v2/env.Ctr", LocalType: "Ctr", Doc: "current ordinal item in Table -- if Sequential then = row number in table, otherwise is index in Order list that then gives row number in Table", Directives: gti.Directives{}, Tag: "view:\"inline\""}},
		{"TrialName", &gti.Field{Name: "TrialName", Type: "github.com/emer/emergent/v2/env.CurPrvString", LocalType: "CurPrvString", Doc: "if Table has a Name column, this is the contents of that", Directives: gti.Directives{}, Tag: ""}},
		{"GroupName", &gti.Field{Name: "GroupName", Type: "github.com/emer/emergent/v2/env.CurPrvString", LocalType: "CurPrvString", Doc: "if Table has a Group column, this is contents of that", Directives: gti.Directives{}, Tag: ""}},
		{"NameCol", &gti.Field{Name: "NameCol", Type: "string", LocalType: "string", Doc: "name of the Name column -- defaults to 'Name'", Directives: gti.Directives{}, Tag: ""}},
		{"GroupCol", &gti.Field{Name: "GroupCol", Type: "string", LocalType: "string", Doc: "name of the Group column -- defaults to 'Group'", Directives: gti.Directives{}, Tag: ""}},
		{"TrialSt", &gti.Field{Name: "TrialSt", Type: "int", LocalType: "int", Doc: "for MPI, trial we start each epoch on, as index into Order", Directives: gti.Directives{}, Tag: ""}},
		{"TrialEd", &gti.Field{Name: "TrialEd", Type: "int", LocalType: "int", Doc: "for MPI, trial number we end each epoch before (i.e., when ctr gets to Ed, restarts)", Directives: gti.Directives{}, Tag: ""}},
	}),
	Embeds:  ordmap.Make([]ordmap.KeyVal[string, *gti.Field]{}),
	Methods: ordmap.Make([]ordmap.KeyVal[string, *gti.Method]{}),
})

var _ = gti.AddType(&gti.Type{
	Name:       "github.com/emer/emergent/v2/env.TimeScales",
	ShortName:  "env.TimeScales",
	IDName:     "time-scales",
	Doc:        "TimeScales are the different time scales associated with overall simulation running, and\ncan be used to parameterize the updating and control flow of simulations at different scales.\nThe definitions become increasingly subjective imprecise as the time scales increase.\nEnvironments can implement updating along different such time scales as appropriate.\nThis list is designed to standardize terminology across simulations and\nestablish a common conceptual framework for time -- it can easily be extended in specific\nsimulations to add needed additional levels, although using one of the existing standard\nvalues is recommended wherever possible.",
	Directives: gti.Directives{},

	Methods: ordmap.Make([]ordmap.KeyVal[string, *gti.Method]{}),
})