Docs: [GoDoc](https://pkg.go.dev/github.com/emer/emergent/ecmd)

Note: this is now deprecated in favor of the [econfig](../econfig) system, which provides a single common Config object for all configuration settings, with TOML config files and command-line arg support.

`ecmd.Args` provides maps for storing commandline arguments of basic types (bool, string, int, float64), along with associated defaults and descriptions, which then set the standard library `flags` for parsing command line arguments.

It has functions for populating standard emergent simulation args.


