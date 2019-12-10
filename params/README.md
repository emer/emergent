[GoDoc](https://godoc.org/github.com/emer/emergent/params)

See [Wiki Params](https://github.com/emer/emergent/wiki/Params) page for detailed docs.

Package `params` provides general-purpose parameter management functionality for organizing multiple sets of parameters efficiently, and basic IO for saving / loading from JSON files and generating Go code to embed into applications, and a basic GUI for viewing and editing.

The main overall unit that is generally operated upon at run-time is the `params.Set`, which is a collection of `params.Sheet`'s (akin to CSS style sheets) that constitute a coherent set of parameters.

A good strategy is to have a "Base" Set that has all the best parameters so far, and then other sets can modify specific params relative to that one. Order of application is critical, as subsequent params applications overwrite earlier ones, and the typical order is:

* `Defaults()` method called that establishes the hard-coded default parameters.
* Then apply "Base" `params.Set` for any changes relative to those.
* Then optionally apply one or more additional `params.Set`'s with current experimental parameters.

Critically, all of this is entirely up to the particular model program(s) to determine and control -- this package just provides the basic data structures for holding all of the parameters, and the IO / and Apply infrastructure.

Within a params.Set, multiple different params.Sheet's can be organized, with each CSS-style sheet achieving a relatively complete parameter styling of a given element of the overal model, e.g., "Network", "Sim", "Env". Or Network could be further broken down into "Learn" vs. "Act" etc, or according to different brain areas ("Hippo", "PFC", "BG", etc). Again, this is entirely at the discretion of the modeler and must be performed under explict program control, especially because order is so critical.

Each `params.Sheet` consists of a collection of params.Sel elements which actually finally contain the parameters.  The `Sel` field specifies a CSS-style selector determining over what scope the parameters should be applied:

* `Type` (no prefix) = name of a type -- anything having this type name will get these params.

* `.Class` = anything with a given class label (each object can have multiple Class labels and thus receive multiple parameter settings, but again, order matters!)

* `#Name` = a specific named object.

The order of application within a given Sheet is also critical -- typically put the most general Type params first, then .Class, then the most specific #Name cases, to achieve within a given Sheet the same logic of establishing Base params for all types and then more specific overrides for special cases (e.g., an overall learning rate that appplies across all projections, but maybe a faster or slower one for a .Class or specific #Name'd projection).

There is a params.Styler interface with methods that any Go type can implement to provide these different labels.  The emer.Network, .Layer, and .Prjn interfaces each implement this interface.

Otherwise, the Apply method will just directly apply params to a given struct type if it does not implement the Styler interface.

Parameter values are limited to float64 values *only*.  These can be specified using "enum" style const integer values, and can be applied to any numeric type (they will be automatically converted), but internally this is the only parameter value type, which greatly simplifies the overall interface, and handles the vast majority of use-cases (especially because named options are just integers and can be set as such).

Finally, there are methods to show where params.Set's set the same parameter differently, and to compare with the default settings on a given object type using go struct field tags of the form def:"val1[,val2...]".

