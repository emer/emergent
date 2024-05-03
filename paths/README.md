Docs: [GoDoc](https://pkg.go.dev/github.com/emer/emergent/paths)

See [Wiki Params](https://github.com/emer/emergent/wiki/Paths) page for detailed docs.

Package `paths` is a separate package for defining `Pattern`s of pathways between layers.  This is done using a fully independent structure that *only* knows about the shapes of the two layers, and it returns a fully general bitmap representation of the pattern of connectivity between them.

The algorithm-specific code then uses these patterns to do all the nitty-gritty of connecting up neurons.

This makes the path code *much* simpler.  This should be the *last* time any of those pathway patterns need to be written (having re-written this code too many times in the C++ version as the details of memory allocations changed).

A Pattern maintains nothing about a specific pathway -- it only has the parameters that are applied in creating a new pattern of connectivity, so it can be shared among any number of pathways that need the same connectivity parameters.

All Patttern types have a New<Name> where <Name> is the type name, that creates a new instance of given pattern initialized with default values.

Individual Pattern types may have a Defaults() method to initialize default values, but it is not mandatory.

# Topographic Weights

Some paths (e.g., Circle, PoolTile) support the generation of topographic weight patterns that can be used to set initial weights, or per-synapse scaling factors.  The `Pattern` interface does not define any standard for how this done, as there are various possible approaches.  Circle defines a method with a standard signature that can be called for each point in the pattern, while PoolTile has a lot more overhead per point and is thus more efficient to generate the whole set of weights to tensor, which can then be used.

It is recommended to have some kind of positive flag(s) for enabling the use of TopoWts -- the standard weight initialization methods, e.g., leabra.Network.InitWts, can then automatically do the correct thing for each type of standard path -- custom ones outside of this standard set would need custom code..

