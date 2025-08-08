Docs: [GoDoc](https://pkg.go.dev/github.com/emer/emergent/actrf)

Package actrf provides activation-based receptive field computation, otherwise known as *reverse correlation* or *spike-triggered averaging*.  It simply computes the activation weighted average of other *source* patterns of activation -- i.e., sum(act * src) / sum(src) which then shows you the patterns of source activity for which a given unit was active.

The RF's are computed and stored in 4D tensors, where the outer 2D are the 2D pathway of the activation tensor (e.g., the activations of units in a layer), and the inner 2D are the 2D pathway of the source tensor.

This results in a nice standard RF plot that can be visualized in a tensor grid view.

There is a standard ActRF which is cumulative over a user-defined interval and a RunningAvg version which is computed online and continuously updated but is more susceptible to sampling bias (i.e., more sampled areas are more active in general), and a recency bias.

See [objrec CCN sim](https://github.com/compcogneuro/sims/blob/main/ch6/objrec) for example usage.

