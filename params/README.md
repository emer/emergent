Docs: [GoDoc](https://pkg.go.dev/github.com/emer/emergent/params)

See [Wiki Params](https://github.com/emer/emergent/wiki/Params) page for detailed docs.

Package `params` applies parameters to struct fields using [css selectors](https://www.w3schools.com/cssref/css_selectors.php) to select which objects a given set of parameters applies to. The struct type must implement the `Styler` interface, with `StyleName() string` and `StyleClass() string` methods, which provide the name and class values against which the selectors test.

Parameters are set using a closure function that runs on matching objects, so that any type value can be set using full editor completion for accessing the struct fields, and any additional logic can be applied within the closure function, including parameter search functions.

Three levels of organization are supported:

* `Sheets` is a `map` of named `Sheet`s, typically with a "Base" Sheet that is applied first, and contains all the base-level defaults, and then different optional parameter `Sheet`s for different configurations or test cases being explored.

* `Sheet` is an ordered list (slice) of `Sel` elements, applied in order.  The ordering is critical for organizing parameters into broad defaults that apply more generally, which are put at the start, followed by progressively more specific parameters that override those defaults for specific cases as needed.

* `Sel` is an individual selector with an expression that matches on Name, Class or Type (Type can be left blank as the entire stack applies only to a specific type of object), and the `Set` function that sets the parameter values on matching objects.

```
var LayerParams = axon.LayerSheets{
	"Base": {
		{Sel: "Layer", Doc: "all defaults",
			Set: func(ly *axon.LayerParams) {
				ly.Inhib.Layer.Gi = 1.05                     // 1.05 > 1.1 for short-term; 1.1 better long-run stability
				ly.Inhib.Layer.FB = 0.5                      // 0.5 > 0.2 > 0.1 > 1.0 -- usu 1.0
				ly.Inhib.ActAvg.Nominal = 0.06               // 0.6 > 0.5
				ly.Acts.NMDA.MgC = 1.2                       // 1.2 > 1.4 here, still..
				ly.Learn.RLRate.SigmoidLinear.SetBool(false) // false > true here
			}},
	},
}
```

In summary, the overall logic is all about the order of application, going from broad defaults to more specific overrides, with the following overall ordering:
* A `Defaults()` method defined on the struct type, which establishes hard-coded default parameters.
* The "Base" `Sheet` applies default parameters for a specific simulation, relative to hard-coded defaults.
* Other `Sheet` cases defined in the map of `Sheets` can then optionally be applied with various experiments, parameter searches, or other specific cases.
* Order of `Sel`s within within a given Sheet is also critical, with the most general Type params first, then .Class, then the most specific #Name cases. For example, an overall learning rate that applies across all pathways with a Type sel, but then a slower one is needed for a for a .Class or specific #Name'd pathway.

## Selectors

The `Sel` field of the `Sel` specifies a CSS-style selector determining over what scope the parameters should be applied:

* `.Class` = anything with a given class label (each object can have multiple Class labels and thus receive multiple parameter settings, but again, order matters!)

* `#Name` = a specific named object.

* `Type` (no prefix) = name of a type -- because parameters only apply to a specific type of object, this can typically just be left blank.

There is a `params.Styler` interface with methods that any Go type can implement to provide these different labels.


## Parameter Searching

A parameter search runs a set of jobs, each with one parameter changed to one value, with everything else at the base configuration. It is a set of independent, one-at-a-time tweaks around a reasonable starting point, not an automated optimization: you look at the results, decide which dimensions show any opportunity for improvement, set new base values, and then run a new search, or particular combinations of the values that helped. Because each job is independent, all of them run in parallel, e.g., with the `ParamSearch` button of [simmer](https://github.com/emer/simmer).

### Search lists

A [Search] has a selector `Sel`, a `Set` function that sets a value on an object, and a `Vals` function that returns the values to try. [Searches] is a list of them: the jobs are all the values of the first one, then all the values of the second one, and so on, so the total number of jobs ([Searches.NumParams]) is the total number of values. Jobs are numbered from 1, as in the `-search-at` argument.

The values are typically:
* [Range]`(start, end, inc)`: discrete values in even steps, e.g., `Range(0.05, 0.3, 0.05)` is .05, .1, .15, .2, .25, .3 (exact decimal values, without float32 rounding error).
* [Tweak]`(v, Increment)`: one value on either side of the current value `v` at its last significant digit, e.g., .4 and .6 for .5; [Log] uses the 1, 2, 5, 10 series.
* [TweakPct]`(v, pct)`: the given proportion below and above `v`.
* An explicit list, e.g., `[]float32{0.7, 0.8}`.

[Baseline]`()` is a special element with one job that changes nothing, so that it runs the base configuration, with the same arguments as the other jobs, to compare them against. Put it first, so that it is job 1, labeled `Search 1: Baseline`.

The objects can be of two kinds:
* **Network parameters**, e.g., `axon.LayerSearches` and `axon.PathSearches` (`params.Searches[*axon.LayerParams]` etc.), which select the layers or pathways with `Sel` (`.Class`, `#Name`) as in a params `Sheet`. `axon.ApplyLayerSearch(net, search, index)` and `axon.ApplyPathSearch` apply a job to all the matching layers or pathways (egomlx has `ApplyLayerSearch` too).
* **Any other type, e.g., the sim's `Config`.** Then `Sel` is only the label of the parameter, and the `Set` function sets the value directly. [Searches.ApplyTo] applies a job to the object. This is the simplest way to search parameters that are set from the config, or that change how the network is built. Use [Float64] to set float64 fields to the exact decimal value.

For example, in the sim's `search.go`:

```go
var Search = params.Searches[*Config]{
	params.Baseline[*Config](),
	{Sel: "Params.Dropout", Set: func(cfg *Config, v float32) {
		cfg.Params.Dropout = params.Float64(v)
	}, Vals: func() []float32 {
		return params.Range(0.05, 0.5, 0.05)
	}},
	{Sel: "Params.WeightDecay", Set: func(cfg *Config, v float32) {
		cfg.Params.WeightDecay = params.Float64(v)
	}, Vals: func() []float32 {
		return params.Range(0.05, 0.3, 0.05)
	}},
}
```

The label of each job ([Search.JobString], [Searches.Labels]) is like `Search 4: Params.Dropout=0.2`.

### Sim setup

Add two fields to the sim's `ParamConfig`, which become the `-search-n` and `-search-at` command line arguments:

```go
	// SearchN prints the total number of parameter search jobs,
	// followed by the label of each job, then quits.
	SearchN bool

	// SearchAt runs the parameter search job at this index,
	// in the [1..N] range, so that a non-zero value means to search.
	SearchAt int
```

and handle them at the start of `RunNoGUI`:

```go
	if ss.Config.Params.SearchN {
		fmt.Println(Search.NumParams())
		for _, lbl := range Search.Labels() {
			fmt.Println(lbl)
		}
		return
	}
	if at := ss.Config.Params.SearchAt; at > 0 {
		lbl, err := Search.ApplyTo(ss.Config, at-1)
		if errors.Log(err) != nil {
			return
		}
		errors.Log(params.WriteJobLabel(lbl))
	}
	ss.Init()
```

Config parameters must be applied before they are used: here `Init` applies the config values to the network. Network parameter searches (`axon.ApplyLayerSearch` etc.) are applied after `Init` (which applies the params sheets) and before the runs start. [WriteJobLabel] saves the label in the `job.label` file of the job.

Check the jobs with `-nogui -search-n`, and run one with `-nogui -search-at 4` (plus short `-runs`, `-epochs` etc. arguments) to make sure that it applies the right value.

### Running a search in simmer

Build the sim locally (simmer runs the local executable with `-search-n` to get the number of jobs and their labels), and press `ParamSearch`. The `Args` apply to all the jobs, so they can set the base configuration of the search. Simmer launches one job per value, with `-search-at 1` to N, labeled with the job labels.

With a [Baseline] element in the search, job 1 runs the base configuration with the same `Args`. Without one, submit a regular job with the same `Args` (or use an existing one) to compare against.

### Workflow

1. Start from a reasonable base configuration, with defaults in the code (e.g., the sim's `config.go` and params sheets).
2. Search a few parameters around their base values, with a modest number of discrete values each.
3. Compare each job with the baseline on the key stats: look for dimensions where some values do better, and for the sensitive range of each parameter.
4. Set the new base values of the parameters that helped, and search again around them, or try combinations of them (e.g., with the `Args` of regular jobs, or a new search over one parameter with another one set in the `Args`).


