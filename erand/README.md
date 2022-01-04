Docs: [GoDoc](https://pkg.go.dev/github.com/emer/emergent/erand)

Package erand provides randomization functionality built on top of standard math/rand
random number generation functions.  Includes:
*  RndParams: specifies parameters for random number generation according to various distributions
   used e.g., for initializing random weights and generating random noise in neurons
*  Permute*: basic convenience methods calling rand.Shuffle on e.g., []int slice



