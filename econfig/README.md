Docs: [GoDoc](https://pkg.go.dev/github.com/emer/emergent/econfig)

`econfig` provides methods to set values on a `Config` struct through a (TOML) config file or command-line args (`flags` in Go terminology), with support for setting Network params and values on any other struct as well (e.g., an Env to be constructed later in a ConfigEnv method).

* Standard usage:
    + `cfg := &ss.Config
    + `cfg.Defaults()` -- sets hard-coded defaults -- user should define and call this method first.  It is better to use the `def:` field tag however because it then shows in `-h` or `--help` usage and in the [GoGi](https://github.com/goki/gi) GUI.
    + `econfig.Config(cfg, "config.toml")` -- sets config values according to the standard order, with given file name specifying the default config file name.

* Has support for nested `Include` paths, which are processed in the natural deepest-first order. The processed `Config` struct field will contain a list of all such files processed.  Config must implement the `IncludesPtr() *[]string` method which satisfies the `Includer` interface, and returns a pointer to an `Includes []string` field containing a list of config files to include.

* Order of setting in `econfig.Config`:
    + Apply any `def:` field tag default values.
    + Look for `--config`, `--cfg` arg, specifying config file(s) on the command line (comma separated if multiple, with no spaces).
    + Fall back on default config file name passed to `Config` function, if arg not found.
    + Read any `Include[s]` files in config file in deepest-first (natural) order, then the specified config file last -- includee overwrites included settings.
    + Process command-line args based on Config field names, with `.` separator for sub-fields.
        
* All field names, including arg names, are case-insensitive.  For args (flags) kebab-case (with either `-` or `_` delimiter) can be used.  For bool args, use "No" prefix in any form (e.g., "NoRunLog" or "no-run-log"). Instead of polluting the flags space with all the different options, custom args processing code is used.

* Is a replacement for `ecmd` and includes the helper methods for saving log files etc.

* A `map[string]any` type can be used for deferred raw params to be applied later (`Network`, `Env` etc).  Example: `Network = {'.PFCLayer:Layer.Inhib.Layer.Gi' = '2.4', '#VSPatchPrjn:Prjn.Learn.LRate' =  '0.01'}` where the key expression contains the [params](../params) selector : path to variable.

* Supports full set of `Open` (file), `OpenFS` (takes fs.FS arg, e.g., for embedded), `Read` (bytes) methods for loading config files.  The overall `Config()` version uses `OpenWithIncludes` which processes includes -- others are just for single files.  Also supports `Write` and `Save` methods for saving from current state.

* If needed, different config file encoding formats can be supported, with TOML being the default (currently only TOML).

# Special fields, supported types, and field tags

* A limited number of standard field types are supported, consistent with emer neural network usage:
    + `bool` and `[]bool`
    + `float32` and `[]float32`
    + `int` and `[]int`
    + `string` and `[]string`
    + [kit](https://github.com/goki/ki) registered "enum" `const` types, with names automatically parsed from string values (including | bit flags).  Must use the [goki stringer](https://github.com/goki/stringer) version to generate `FromString()` method, and register the type like this: `var KiT_TestEnum = kit.Enums.AddEnum(TestEnumN, kit.NotBitFlag, nil)` -- see [enum.go](enum.go) file for example.

* To enable include file processing, add a `Includes []string` field and a `func (cfg *Config) IncludesPtr() *[]string { return &cfg.Includes }` method.  The include file(s) are read first before the current one.  A stack of such includes is created and processed in the natural order encountered, so each includer is applied after the includees, recursively.  Note: use `--config` to specify the first config file read -- the `Includes` field is excluded from arg processing because it would be processed _after_ the point where include files are processed.

* `Field map[string]any` -- allows raw parsing of values that can be applied later.  Use this for `Network`, `Env` etc fields.

* Field tag `def:"value"`, used in the [GoGi](https://github.com/goki/gi) GUI, sets the initial default value and is shown for the `-h` or `--help` usage info.

# Standard Config

Here's a standard `Config` struct, corresponding to the `AddStd` args from `ecmd`, which can be used as a starting point.

```Go
// Config is a standard Sim config -- use as a starting point.
// don't forget to update defaults, delete unused fields, etc.
typeConfig struct {
	Includes     []string       `desc:"specify include files here, and after configuration, it contains list of include files added"`
	GUI          bool           `def:"true" desc:"open the GUI -- does not automatically run -- if false, then runs automatically and quits"`
	GPU          bool           `desc:"use the GPU for computation"`
	Debug        bool           `desc:"log debugging information"`
	Network      map[string]any `desc:"network parameters"`
	ParamSet     string         `desc:"ParamSet name to use -- must be valid name as listed in compiled-in params or loaded params"`
	ParamFile    string         `desc:"Name of the JSON file to input saved parameters from."`
	ParamDocFile string         `desc:"Name of the file to output all parameter data. If not empty string, program should write file(s) and then exit"`
	Tag          string         `desc:"extra tag to add to file names and logs saved from this run"`
	Note         string         `desc:"user note -- describe the run params etc -- like a git commit message for the run"`
	Run          int            `def:"0" desc:"starting run number -- determines the random seed -- runs counts from there -- can do all runs in parallel by launching separate jobs with each run, runs = 1"`
	Runs         int            `def:"10" desc:"total number of runs to do when running Train"`
	Epochs       int            `def:"100" desc:"total number of epochs per run"`
	NTrials      int            `def:"128" desc:"total number of trials per epoch.  Should be an even multiple of NData."`
	NData        int            `def:"16" desc:"number of data-parallel items to process in parallel per trial -- works (and is significantly faster) for both CPU and GPU.  Results in an effective mini-batch of learning."`
	SaveWts      bool           `desc:"if true, save final weights after each run"`
	EpochLog     bool           `def:"true" desc:"if true, save train epoch log to file, as .epc.tsv typically"`
	RunLog       bool           `def:"true" desc:"if true, save run log to file, as .run.tsv typically"`
	TrialLog     bool           `def:"true" desc:"if true, save train trial log to file, as .trl.tsv typically. May be large."`
	TestEpochLog bool           `def:"false" desc:"if true, save testing epoch log to file, as .tst_epc.tsv typically.  In general it is better to copy testing items over to the training epoch log and record there."`
	TestTrialLog bool           `def:"false" desc:"if true, save testing trial log to file, as .tst_trl.tsv typically. May be large."`
	NetData      bool           `desc:"if true, save network activation etc data from testing trials, for later viewing in netview"`
}
```    

# Key design considerations

* Can set config values from command-line args and/or config file (TOML being the preferred format) (or env vars)
    + current axon models only support args. obelisk models only support TOML.  conflicts happen.

* Sims use a Config struct with fields that represents the definitive value of all arg / config settings (vs a `map[string]interface{}`)
    + struct provides _compile time_ error checking -- very important and precludes map.
    + Add Config to Sim so it is visible in the GUI for easy visual debugging etc (current args map is organized by types -- makes it hard to see everything).

* Enable setting Network or Env params directly:
    + Use `Network.`, `Env.`, `TrainEnv.`, `TestEnv.` etc prefixes followed by standard `params` selectors (e.g., `Layer.Act.Gain`) or paths to fields in relevant env.  These can be added to Config as `map[string]any` and then applied during ConfigNet, ConfigEnv etc.

* TOML Go implementations are case insensitive (TOML spec says case sensitive..) -- makes sense to use standard Go CamelCase conventions as in every other Go struct.

