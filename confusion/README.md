Confusion implements a confusion matrix: records output responses for discrete categories / classes.

* Rows (outer dimension) are for each class as the ground truth, correct answer.

* Columns (inner dimension) are the response generated for each ground-truth class.

The main result is in the Prob field, computed from the Sum and N values added incrementally.


