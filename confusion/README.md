Docs: [GoDoc](https://pkg.go.dev/github.com/emer/emergent/confusion)

Confusion implements a confusion matrix: records output responses for discrete categories / classes.

* Rows (outer dimension) are for each class as the ground truth, correct answer.

* Columns (inner dimension) are the response generated for each ground-truth class.

The main result is in the Prob field, computed from the Sum and N values added incrementally.

Main API:

* `InitFromLabels` to initialize with list of class labels and display font size.
* `Incr` on each trial with network's response index and correct target index.
* `Probs` when done, to compute probabilities from accumulated data.
* `SaveCSV` / `OpenCSV` for saving / loading data (for nogui usage).

