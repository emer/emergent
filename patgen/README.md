[GoDoc](https://godoc.org/github.com/emer/emergent/patgen)

Package `patgen` contains functions that generate patterns, typically based on various constrained-forms of random patterns, e.g., permuted binary random patterns.

# Pattern Generator

Pattern generator is capable of flexibly making patterns for models. To configure your own patterns, follow these four steps: 

1) make your vocabulary as a **pool name -- tensor** map; 
2) use your vocabulary to initialize a big pattern (e.g., TrainAB), in which you later use to store input & output patterns; 
3) mix different pools in the vocabulary into one pattern (e.g., A+B+context-->input pattern), which is stored in the big pattern; 
4) repeat 3) for the output pattern.

Example code could be found in `ConfigPats()` in [hip bench](https://github.com/emer/leabra/blob/master/examples/hip_bench/hip_bench.go).
