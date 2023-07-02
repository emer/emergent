Docs: [GoDoc](https://pkg.go.dev/github.com/emer/emergent/econfig)

econfig provides methods to set values on a `Config` struct through a (TOML) config file or command-line args (`flags` in Go terminology), with support for setting Network params and values on any other struct as well (e.g., an Env to be constructed later in a ConfigEnv method).

* Standard usage:
    + `cfg := &ss.Config
    + `cfg.Defaults()` -- sets hard-coded defaults -- user should define and call this method first.  It is better to use the `def:` field tag however because it then shows in `-h` or `--help` usage and in the [GoGi](https://github.com/goki/gi) GUI.
    + `econfig.Config(cfg, "config.toml")` -- sets config values according to the standard order, with given file name specifying the default config file name.

* Standard order in `econfig.Config`:
    + Apply any `def:` field tag default values.
    + Look for `--config`, `--cfg`, or `-c` arg, specifying config file(s) on the command line (comma separated if multiple, with no spaces).
    + Fall back on default config file name passed to `Config` function, if arg not found.
    + Read any `Include[s]` files in config file in deepest-first (natural) order, then the specified config file last.
    + Process command-line args based on Config field names, with `.` separator for sub-fields (see field tags for shorthand and aliases)
        
* Args are processed using the POSIX standard naming conventions (as in [pflag](https://github.com/spf13/pflag) ) instead of as in the Go `flags` package, supporting shorthand args (see tags below) and requiring `--` double-dash for long names.  Arg names are case-insensitive and kebab-case (with either `-` or `_` delimiter) can be used.  Instead of polluting the flags space with all the different options, custom args processing code is used.

* Is a replacement for `ecmd` and includes the helper methods for saving log files etc.

* Has support for nested `Include` paths, which are processed in the natural deepest-first order (see below). The processed `Config` struct field will contain a list of all such files processed.

* A `map[string]any` type can be used for deferred raw params to be applied later (Network, Env etc) (see below).

* Is case insensitive for field names -- use Go CamelCase for consistency but any naming scheme is supported.

* Supports full set of `Open` (file), `OpenFS` (takes fs.FS arg, e.g., for embedded), `Read` (bytes) methods for loading config files.  Only the overall `Config()` version processes includes -- others are just for single files.

* If needed, different config file encoding formats can be supported, with TOML being the default (currently only TOML).

# Special fields, supported types, and field tags

* A limited number of standard field types are supported, consistent with emer neural network usage:
    + `float32` and `[]float32`
    + `int` and `[]int`
    + `string` and `[]string`
    + [kit](https://github.com/goki/ki) registered "enum" `const` types, with names automatically parsed from string values (including | bit flags).  Must use the [goki stringer](https://github.com/goki/stringer) version to generate `FromString()` method, and register the type like this: `var KiT_GlobalVars = kit.Enums.AddEnum(GlobalVarsN, kit.NotBitFlag, nil)`

* `Include string` or `Include []string` or `Includes []string` -- given file path is read first before the current one.  A stack of such includes is created and processed in the natural order encountered, so each includer is applied after the includees, recursively.  Note that, due to the order of processing, a command-line arg of `Include` is not valid and is flagged as an error -- use `--config` instead.

* `Field map[string]any` -- allows raw parsing of values that can be applied later.  Use this for `Network`, `Env` etc fields.

* Field tag `def:"value"`, used in the [GoGi](https://github.com/goki/gi) GUI, sets the initial default value and is shown for the `-h` or `--help` usage info.

* Field tag `short:"s"` specifies a shorthand name for the arg (using the ) (e.g., `-s` in this example) (multiple can be specified).  Note that `-c` is reserved for the config file name.

* Field tag `alias:"other-name"` specifies an alias alternative long-format name for field (multiple can be specified).

# Standard Config

Here's a standard `Config` struct, corresponding to the `AddStd` args from `ecmd`, which can be used as a starting point.

```Go
type Config struct {
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


