looper implements a fully generic looping control system with extensible functionality at each level of the loop.

TODO: add names to funcs and support replacement, add after, add before?

stack_test.go output shows the logic of the looping functions:

```
Run Start: 0
Run Pre: 0
	Epoch Start: 0
	Epoch Pre: 0
		Trial Start: 0
		Trial Pre: 0
		Trial Post: 1
		Trial Pre: 1
		Trial Post: 2
		Trial Pre: 2
		Trial Post: 3
		Trial Stop: 3
		Trial End: 3
	Epoch Post: 1
	Epoch Pre: 1
		Trial Start: 0
		Trial Pre: 0
		Trial Post: 1
		Trial Pre: 1
		Trial Post: 2
		Trial Pre: 2
		Trial Post: 3
		Trial Stop: 3
		Trial End: 3
	Epoch Post: 2
	Epoch Pre: 2
		Trial Start: 0
		Trial Pre: 0
		Trial Post: 1
		Trial Pre: 1
		Trial Post: 2
		Trial Pre: 2
		Trial Post: 3
		Trial Stop: 3
		Trial End: 3
	Epoch Post: 3
	Epoch Stop: 3
	Epoch End: 3
Run Post: 1
Run Pre: 1
	Epoch Start: 0
	Epoch Pre: 0
		Trial Start: 0
		Trial Pre: 0
		Trial Post: 1
		Trial Pre: 1
		Trial Post: 2
		Trial Pre: 2
		Trial Post: 3
		Trial Stop: 3
		Trial End: 3
	Epoch Post: 1
	Epoch Pre: 1
		Trial Start: 0
		Trial Pre: 0
		Trial Post: 1
		Trial Pre: 1
		Trial Post: 2
		Trial Pre: 2
		Trial Post: 3
		Trial Stop: 3
		Trial End: 3
	Epoch Post: 2
	Epoch Pre: 2
		Trial Start: 0
		Trial Pre: 0
		Trial Post: 1
		Trial Pre: 1
		Trial Post: 2
		Trial Pre: 2
		Trial Post: 3
		Trial Stop: 3
		Trial End: 3
	Epoch Post: 3
	Epoch Stop: 3
	Epoch End: 3
Run Post: 2
Run Stop: 2
Run End: 2
```
