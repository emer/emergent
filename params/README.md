Docs: [GoDoc](https://pkg.go.dev/github.com/emer/emergent/params)

See [Wiki Params](https://github.com/emer/emergent/wiki/Params) page for detailed docs.

Package `params` applies parameters to struct fields using [css selectors](https://www.w3schools.com/cssref/css_selectors.php) to select which objects a given set of parameters applies to. The struct type must implement the `Styler` interface, with `StyleName() string` and `StyleClass() string` methods, which provide the name and class values against which the selectors test.

Parameters are set using a closure function that runs on matching objects, so that any type value can be set using full editor completion for accessing the struct fields, and any additional logic can be applied within the closure function, including parameter search functions.

Three levels of organization are supported:

* `Sheets` is a `map` of named `Sheet`s, typically with a "Base" Sheet that is applied first, and contains all the base-level defaults, and then different optional parameter `Sheet`s for different configurations or test cases being explored.

* `Sheet` is an ordered list (slice) of `Sel` elements, applied in order.  The ordering is critical for organizing parameters into broad defaults that apply more generally, which are put at the start, followed by progressively more specific parameters that override those defaults for specific cases as needed.

* `Sel` is an individual selector with an expression that matches on Name, Class or Type (Type can be left blank as the entire stack applies only to a specific type of object), and the `Set` function that sets the parameter values on matching objects.

TODO: replace with actual example from axon:

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

TODO


