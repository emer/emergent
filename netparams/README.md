Docs: [GoDoc](https://pkg.go.dev/github.com/emer/emergent/netparams)

See [Wiki Params](https://github.com/emer/emergent/wiki/Params) page for detailed docs.

Package `netparams` provides general-purpose parameter management functionality for organizing multiple sets of parameters efficiently, and basic IO for saving / loading from JSON files and generating Go code to embed into applications, and a basic GUI for viewing and editing.

`netparams` is a simplification of `params` focused only on setting `Network` params -- eliminating the `params.Set` and `params.Sheets` levels that support setting params on arbitrary non-Network objects.  It is better to use [econfig](../econfig) to configure standard struct Config objects outside of the Network -- the styling flexibility is really only needed for Network level params.

The `netparams.Sets` contains one or more `params.Sheet`s (akin to CSS style sheets) that constitute a coherent set of parameters.  Here's the structure:

```
Sets {
    "Base": {
        Sel: "Layer" {
            Params: {
                "Layer.Inhib.Layer.Gi": "1.1",
                ...
            }
        },
        Sel: ".Back" {
            Params: {
                "Path.PathScale.Rel": "0.2",
                ...
            }
        }
    },
}
```


The default supported organization is to have a "Base" Set that has all the best parameters so far, and then other sets can modify specific params relative to that one. Order of application is critical, as subsequent params applications overwrite earlier ones, and the typical order is:

* `Defaults()` method called that establishes the hard-coded default parameters.
* Then apply "Base" `params.Sheet` for any changes relative to those.
* Then optionally apply one or more additional `params.Sheet`s with current experimental parameters or for other special use-cases.

Critically, all of this is entirely up to the particular model program(s) to determine and control -- this package just provides the basic data structures for holding all of the parameters, and the IO / and Apply infrastructure.

Each `params.Sheet` consists of a collection of params.Sel elements which finally contain the parameters.  The `Sel` field specifies a CSS-style selector determining over what scope the parameters should be applied:

* `Type` (no prefix) = name of a type -- anything having this type name will get these params.

* `.Class` = anything with a given class label (each object can have multiple Class labels and thus receive multiple parameter settings, but again, order matters!)

* `#Name` = a specific named object.

The order of application within a given Sheet is also critical -- typically put the most general Type params first, then `.Class`, then the most specific `#Name` cases, to achieve within a given Sheet the same logic of establishing Base params for all types and then more specific overrides for special cases (e.g., an overall learning rate that appplies across all pathways, but maybe a faster or slower one for a .Class or specific #Name'd pathway).

There is a params.Styler interface with methods that any Go type can implement to provide these different labels.  The emer.Network, .Layer, and .Path interfaces each implement this interface.

Parameter values are stored as strings, which can represent any value.

Finally, there are methods to show where params.Sheet sets the same parameter differently, and to compare with the default settings on a given object type using go struct field tags of the form def:"val1[,val2...]".

# Providing direct access to specific params

The best way to provide the user direct access to specific parameter values through the Params mechanisms is to put the relevant params in the `Sim` object, where they will be editable fields, and then call `SetFloat` or `SetString` as appropriate with the path to the parameter in question, followed by a call to apply the params.

The current value can be obtained by the `ParamVal` methods.


