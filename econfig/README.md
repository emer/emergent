Docs: [GoDoc](https://pkg.go.dev/github.com/emer/emergent/econfig)

`econfig` provides methods to set values on a `Config` struct through a (TOML) config file or command-line args (`flags` in Go terminology), with support for setting Network params and values on any other struct as well (e.g., an Env to be constructed later in a ConfigEnv method).

* Standard usage:
    + `cfg := &ss.Config`
    + `cfg.Defaults()` -- sets hard-coded defaults -- user should define and call this method first.
    + It is better to use the `default:` field tag however because it then shows in `-h` or `--help` usage and in the [Cogent Core](https://cogentcore.org/core) GUI.  See [Default Tags](#def_default_tags) for how to specify def values for more complex types.
    + `econfig.Config(cfg, "config.toml")` -- sets config values according to the standard order, with given file name specifying the default config file name.

* Has support for nested `Include` paths, which are processed in the natural deepest-first order. The processed `Config` struct field will contain a list of all such files processed.  There are two options for include file support:
    + `Includes []string` for multiple includes in one config: Config implements the `IncludesPtr() *[]string` method which satisfies the `Includeser` interface, and returns a pointer to the `Includes` field containing a list of config files to include.
    + `Include string` for single include in one config: Config implements the `IncludePtr() *string` method which satisfies the `Includer` interface, and returns a pointer to the Include field.
    + The default `IncludePaths` includes current dir (`.`) and `configs` directory, which is recommended location to store different configs.    
    
* Order of setting in `econfig.Config`:
    + Apply any `def:` field tag default values.
    + Look for `--config`, `--cfg` arg, specifying config file(s) on the command line (comma separated if multiple, with no spaces).
    + Fall back on default config file name passed to `Config` function, if arg not found.
    + Read any `Include[s]` files in config file in deepest-first (natural) order, then the specified config file last -- includee overwrites included settings.
    + Process command-line args based on Config field names, with `.` separator for sub-fields.
        
* All field name references in toml files and command-line args are case-insensitive.  For args (flags) kebab-case (with either `-` or `_` delimiter) can be used.  For bool args, use "No" prefix in any form (e.g., "NoRunLog" or "no-run-log"). Instead of polluting the flags space with all the different options, custom args processing code is used.

* Args in sub-structs are automatically available with just the field name and also nested within the name of the parent struct field -- for example, `-Run.NEpochs` and just `-NEpochs` (or `-nepochs` lowercase).  Use `nest:"+"` to force a field to only be available in its nested form, in case of conflict of names without nesting (which are logged).

* Is a replacement for `ecmd` and includes the helper methods for saving log files etc.

* A `map[string]any` type can be used for deferred raw params to be applied later (`Network`, `Env` etc).  Example: `Network = {'.PFCLayer:Layer.Inhib.Layer.Gi' = '2.4', '#VSPatchPrjn:Prjn.Learn.LRate' =  '0.01'}` where the key expression contains the [params](../params) selector : path to variable.

* Supports full set of `Open` (file), `OpenFS` (takes fs.FS arg, e.g., for embedded), `Read` (bytes) methods for loading config files.  The overall `Config()` version uses `OpenWithIncludes` which processes includes -- others are just for single files.  Also supports `Write` and `Save` methods for saving from current state.

* If needed, different config file encoding formats can be supported, with TOML being the default (currently only TOML).

# Special fields, supported types, and field tags

* To enable include file processing, add a `Includes []string` field and a `func (cfg *Config) IncludesPtr() *[]string { return &cfg.Includes }` method.  The include file(s) are read first before the current one.  A stack of such includes is created and processed in the natural order encountered, so each includer is applied after the includees, recursively.  Note: use `--config` to specify the first config file read -- the `Includes` field is excluded from arg processing because it would be processed _after_ the point where include files are processed.

* `Field map[string]any` -- allows raw parsing of values that can be applied later.  Use this for `Network`, `Env` etc fields.  Here are some examples:

```toml
[Env.Env]
	"NItems" = 10
	"NTrials" = 10
```

```toml
[Params.Network]
  "#Output:Layer.Inhib.Layer.Gi" = 0.7
  "Prjn:Prjn.Learn.LRate.Base" = 0.05
```

* Field tag `default:"value"`, used in the [Cogent Core](https://cogentcore.org/core) GUI, sets the initial default value and is shown for the `-h` or `--help` usage info.

* [kit](https://cogentcore.org/core/ki) registered "enum" `const` types, with names automatically parsed from string values (including bit flags).  Must use the [goki stringer](https://github.com/goki/stringer) version to generate `FromString()` method, and register the type like this: `var KitTestEnum = kit.Enums.AddEnum(TestEnumN, kit.NotBitFlag, nil)` -- see [enum.go](enum.go) file for example.

# `default` Default Tags

The [Cogent Core](https://cogentcore.org/core) GUI processes `default:"value"` struct tags to highlight values that are not at their defaults.  econfig uses these same tags to auto-initialize fields as well, ensuring that the tag and the actual initial value are the same.  The value for strings or numbers is just the string representation.  For more complex types, here ar some examples:

* `struct`: specify using standard Go literal expression as a string, with single-quotes `'` used instead of double-quotes around strings, such as the name of the fields:
    + `evec.Vector2i`: `default:"{'X':10,'Y':10}"`

* `slice`: comma-separated list of values in square braces -- use `'` for internal string boundaries:
    + `[]float32`: `default:"[1, 2.14, 3.14]"`
    + `[]string`: `default:"{'A', 'bbb bbb', 'c c c'}"`

* `map`: comma-separated list of key:value in curly braces -- use `'` for internal string boundaries:
    + `map[string]float32`: `default:"{'key1': 1, 'key2': 2.14, 'key3': 3.14]"`

# Standard Config Example

Here's the `Config` struct from [axon/examples/ra25](https://github.com/emer/axon), which can provide a useful starting point.  It uses Params, Run and Log sub-structs to better organize things.  For sims with extensive Env config, that should be added as a separate sub-struct as well.  The `view:"add-fields"` struct tag shows all of the fields in one big dialog in the GUI -- if you want separate ones, omit that.

```Go
// ParamConfig has config parameters related to sim params
type ParamConfig struct {

	// network parameters
	Network map[string]any

	// size of hidden layer -- can use emer.LaySize for 4D layers
	Hidden1Size evec.Vector2i `default:"{'X':10,'Y':10}" nest:"+"`

	// size of hidden layer -- can use emer.LaySize for 4D layers
	Hidden2Size evec.Vector2i `default:"{'X':10,'Y':10}" nest:"+"`

	// Extra Param Sheet name(s) to use (space separated if multiple) -- must be valid name as listed in compiled-in params or loaded params
	Sheet string

	// extra tag to add to file names and logs saved from this run
	Tag string

	// user note -- describe the run params etc -- like a git commit message for the run
	Note string

	// Name of the JSON file to input saved parameters from.
	File string `nest:"+"`

	// Save a snapshot of all current param and config settings in a directory named params_<datestamp> (or _good if Good is true), then quit -- useful for comparing to later changes and seeing multiple views of current params
	SaveAll bool `nest:"+"`

	// for SaveAll, save to params_good for a known good params state.  This can be done prior to making a new release after all tests are passing -- add results to git to provide a full diff record of all params over time.
	Good bool `nest:"+"`
}

// RunConfig has config parameters related to running the sim
type RunConfig struct {

	// use the GPU for computation -- generally faster even for small models if NData ~16
	GPU bool `default:"true"`

	// number of data-parallel items to process in parallel per trial -- works (and is significantly faster) for both CPU and GPU.  Results in an effective mini-batch of learning.
	NData int `default:"16" min:"1"`

	// number of parallel threads for CPU computation -- 0 = use default
	NThreads int `default:"0"`

	// starting run number -- determines the random seed -- runs counts from there -- can do all runs in parallel by launching separate jobs with each run, runs = 1
	Run int `default:"0"`

	// total number of runs to do when running Train
	NRuns int `default:"5" min:"1"`

	// total number of epochs per run
	NEpochs int `default:"100"`

	// stop run after this number of perfect, zero-error epochs
	NZero int `default:"2"`

	// total number of trials per epoch.  Should be an even multiple of NData.
	NTrials int `default:"32"`

	// how often to run through all the test patterns, in terms of training epochs -- can use 0 or -1 for no testing
	TestInterval int `default:"5"`

	// how frequently (in epochs) to compute PCA on hidden representations to measure variance?
	PCAInterval int `default:"5"`

	// if non-empty, is the name of weights file to load at start of first run -- for testing
	StartWts string
}

// LogConfig has config parameters related to logging data
type LogConfig struct {

	// if true, save final weights after each run
	SaveWts bool

	// if true, save train epoch log to file, as .epc.tsv typically
	Epoch bool `default:"true" nest:"+"`

	// if true, save run log to file, as .run.tsv typically
	Run bool `default:"true" nest:"+"`

	// if true, save train trial log to file, as .trl.tsv typically. May be large.
	Trial bool `default:"false" nest:"+"`

	// if true, save testing epoch log to file, as .tst_epc.tsv typically.  In general it is better to copy testing items over to the training epoch log and record there.
	TestEpoch bool `default:"false" nest:"+"`

	// if true, save testing trial log to file, as .tst_trl.tsv typically. May be large.
	TestTrial bool `default:"false" nest:"+"`

	// if true, save network activation etc data from testing trials, for later viewing in netview
	NetData bool
}

// Config is a standard Sim config -- use as a starting point.
type Config struct {

	// specify include files here, and after configuration, it contains list of include files added
	Includes []string

	// open the GUI -- does not automatically run -- if false, then runs automatically and quits
	GUI bool `default:"true"`

	// log debugging information
	Debug bool

	// parameter related configuration options
	Params ParamConfig `view:"add-fields"`

	// sim running related configuration options
	Run RunConfig `view:"add-fields"`

	// data logging related configuration options
	Log LogConfig `view:"add-fields"`
}

func (cfg *Config) IncludesPtr() *[]string { return &cfg.Includes }

```    

# Key design considerations

* Can set config values from command-line args and/or config file (TOML being the preferred format) (or env vars)
    + current axon models only support args. obelisk models only support TOML.  conflicts happen.

* Sims use a Config struct with fields that represents the definitive value of all arg / config settings (vs a `map[string]any`)
    + struct provides _compile time_ error checking (and IDE completion) -- very important and precludes map.
    + Add Config to Sim so it is visible in the GUI for easy visual debugging etc (current args map is organized by types -- makes it hard to see everything).

* Enable setting Network or Env params directly:
    + Use `Network.`, `Env.`, `TrainEnv.`, `TestEnv.` etc prefixes followed by standard `params` selectors (e.g., `Layer.Act.Gain`) or paths to fields in relevant env.  These can be added to Config as `map[string]any` and then applied during ConfigNet, ConfigEnv etc.

* TOML Go implementations are case insensitive (TOML spec says case sensitive..) -- makes sense to use standard Go CamelCase conventions as in every other Go struct.


