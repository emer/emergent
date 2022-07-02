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

The TFPN matrix keeps a record of true/false positives (tp/fp) and true/false negatives (tn/fn) for each category/class. This table is used to calculate F1 scores either by class or across classes

A beginner’s guide on how to calculate Precision, Recall, F1-score for a multi-class classification problem can be found at https://towardsdatascience.com/confusion-matrix-for-your-multi-class-machine-learning-model-ff9aa3bf7826

API:
* `SumTFPN` to calculate the tp, fp, fn and tn scores for each class
* `ScoreClass` calculates the precision and recall scores that are needed for the F1 score
* `ScoreMatrix` uses the values calculated by ScoreClass to generate 3 different F1 scores for the entire matrix
    * `F1 Micro`
    * `F1 Macro`
    * `F1 Weighted`
