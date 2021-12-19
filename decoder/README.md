[GoDoc](https://pkg.go.dev/github.com/emer/emergent/decoder)

The decoder package provides standalone decoders that can sample variables from `emer` network layers and provide a supervised one-layer categorical decoding of what is being represented in those layers.  This can provide an important point of reference relative to whatever the network itself is generating, and is especially useful for more self-organizing networks that may not have supervised training at all.

# SoftMax

The `SoftMax` decoder is the best choice for a 1-hot classification decoder, using the widely-used SoftMax function: https://en.wikipedia.org/wiki/Softmax_function

Here's the basic API:

* `InitLayer` to initialize with number of categories and layer(s) for input.

* `Decode` with variable name to record that variable from layers, and decode based on the current state info for that variable.  You can also access the full sorted list of category outputs in the Sorted field of the SoftMax object.

* `Train` *after* Decode with index of current ground-truth category value.

It is also possible to use the decoder without reference to emer Layers by just calling `Init`, `Forward`, `Sort`, and `Train`.

A learning rate of about 0.05 works well for large layers, and 0.1 can be used for smaller, less complex cases.

# Sigmoid

The `Sigmoid` decoder is the best choice for factorial, independent categories where any number of them might be active at a time.  It learns using the delta rule for each output unit.  Uses the same API as above, except Decode takes a full slice of target values for each category output, and the results are found in the `Units[i].Act` variable, which can be returned into a slice using the `Output` method.

# Vote

`TopVoteInt` takes a slice of ints representing votes for which category index was selected (or anything really), and returns the one with the most votes, choosing at random for any ties at the top, along with the number of votes for it.

`TopVoteString` is the same as TopVoteInt but with string-valued votes.

