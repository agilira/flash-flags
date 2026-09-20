# Flash-Flags API Reference
### an AGILira library

**Version**: v1.1.5  
**Go Version**: 1.25.9+  
**License**: MPL-2.0

This document provides comprehensive API documentation for Flash-Flags, covering all public types, functions, and interfaces.

---

## Table of Contents

1. [Core Types](#core-types)
   - [FlagSet](#flagset)
   - [Flag](#flag)
2. [Flag Definition Methods](#flag-definition-methods)
   - [String Flags](#string-flags)
   - [Integer Flags](#integer-flags)
   - [Boolean Flags](#boolean-flags)
   - [Float Flags](#float-flags)
   - [Duration Flags](#duration-flags)
   - [String Slice Flags](#string-slice-flags)
3. [Parsing & Processing](#parsing--processing)
   - [Parse](#parse)
   - [Arguments Access](#arguments-access)
4. [Configuration Sources](#configuration-sources)
   - [Configuration Files](#configuration-files)
   - [Environment Variables](#environment-variables)
5. [Validation & Constraints](#validation--constraints)
   - [Custom Validators](#custom-validators)
   - [Required Flags](#required-flags)
   - [Flag Dependencies](#flag-dependencies)
6. [Help & Documentation](#help--documentation)
7. [Flag Inspection](#flag-inspection)
8. [Utility Functions](#utility-functions)
9. [Interfaces](#interfaces)
10. [Input Screening](#input-screening)

---

## Core Types

### FlagSet

`FlagSet` represents a collection of command-line flags with parsing and validation capabilities.

#### Constructor

```go
func New(name string) *FlagSet
```

Creates a new FlagSet with the specified name. The name is used for help output and configuration file discovery.

**Parameters:**
- `name` (string): The program name

**Returns:**
- `*FlagSet`: A new flag set instance

**Example:**
```go
fs := flashflags.New("myapp")
```

**Thread Safety:**
- A `FlagSet` holds no mutexes and performs no atomic operations
- Multiple goroutines can safely read flag values concurrently **once `Parse()` has returned**
- Every mutating call — flag registration, `Set*`, `Parse`, `LoadConfig`, `LoadEnvironmentVariables`, `Reset`, `ResetFlag` — must happen on a single goroutine before the flag set is shared
- Calling `Reset()` or `ResetFlag()` while another goroutine reads is a data race and `go test -race` will report it

---

### Flag

`Flag` represents a single command-line flag with its value, metadata, and constraints.

#### Methods

##### Name
```go
func (f *Flag) Name() string
```

Returns the flag's long name (used with `--flag` syntax).

**Returns:**
- `string`: The flag name

**Example:**
```go
flag := fs.Lookup("port")
fmt.Printf("Flag name: %s\n", flag.Name()) // Output: Flag name: port
```

---

##### Value
```go
func (f *Flag) Value() interface{}
```

Returns the current flag value as `interface{}`. Use type assertion to convert to the expected type.

**Returns:**
- `interface{}`: The flag value

**Example:**
```go
flag := fs.Lookup("port")
if flag != nil && flag.Type() == "int" {
    if port, ok := flag.Value().(int); ok {
        fmt.Printf("Port value: %d\n", port)
    }
}
```

---

##### Type
```go
func (f *Flag) Type() string
```

Returns the flag type as a string.

**Returns:**
- `string`: One of `"string"`, `"int"`, `"bool"`, `"float64"`, `"duration"`, `"stringSlice"`

**Example:**
```go
flag := fs.Lookup("timeout")
if flag.Type() == "duration" {
    if dur, ok := flag.Value().(time.Duration); ok {
        fmt.Printf("Timeout: %v\n", dur)
    }
}
```

---

##### Changed
```go
func (f *Flag) Changed() bool
```

Returns whether the flag was explicitly set by any configuration source (command line, environment variable, or config file).

**Returns:**
- `bool`: `true` if explicitly set, `false` if using default value

**Example:**
```go
flag := fs.Lookup("port")
if flag.Changed() {
    fmt.Println("Port was explicitly set")
} else {
    fmt.Println("Port is using default value")
}
```

---

##### Source
```go
func (f *Flag) Source() string
```

Returns the configuration source that supplied the flag's current value.

Sources are resolved by precedence, so the returned name is the **highest-priority**
source that set the flag, not merely the last one consulted. A flag present in both
the config file and the environment reports `"env"`.

**Returns:**
- `string`: One of `"cli"`, `"env"`, `"config"`, `"default"`

**Example:**
```go
flag := fs.Lookup("port")
fmt.Printf("port=%v (from %s)\n", flag.Value(), flag.Source())
```

*Added in v1.1.9.*

---

##### Usage
```go
func (f *Flag) Usage() string
```

Returns the flag's usage description string.

**Returns:**
- `string`: The help text

**Example:**
```go
flag := fs.Lookup("port")
fmt.Printf("Flag usage: %s\n", flag.Usage())
```

---

##### ShortKey
```go
func (f *Flag) ShortKey() string
```

Returns the short flag key (single character for `-f` syntax). Returns empty string if no short key is defined.

**Returns:**
- `string`: The short key (e.g., `"p"`) or empty string

**Example:**
```go
flag := fs.Lookup("port")
if flag.ShortKey() != "" {
    fmt.Printf("Short key: -%s\n", flag.ShortKey())
}
```

---

##### SetValidator
```go
func (f *Flag) SetValidator(validator func(interface{}) error)
```

Sets a validation function for the flag. The validator will be called whenever the flag value is set or changed.

**Parameters:**
- `validator` (func(interface{}) error): Validation function

**Example:**
```go
flag := fs.Lookup("port")
flag.SetValidator(func(val interface{}) error {
    port, ok := val.(int)
    if !ok {
        return fmt.Errorf("expected int, got %T", val)
    }
    if port < 1024 {
        return fmt.Errorf("port must be >= 1024")
    }
    return nil
})
```

> Always use the comma-ok form of the type assertion. A bare `val.(int)` panics on
> mismatch, and the panic propagates out of `Parse()` and terminates the program.

---

##### Validate
```go
func (f *Flag) Validate() error
```

Validates the current flag value using the validator if set.

**Returns:**
- `error`: Validation error or `nil`

**Example:**
```go
flag := fs.Lookup("port")
if err := flag.Validate(); err != nil {
    fmt.Printf("Validation failed: %v\n", err)
}
```

---

##### Reset
```go
func (f *Flag) Reset()
```

Resets the flag to its default value and marks it as unchanged.

**Example:**
```go
flag := fs.Lookup("port")
flag.Reset()
fmt.Printf("Port reset to: %v, Changed: %t\n", flag.Value(), flag.Changed())
```

---

## Flag Definition Methods

### String Flags

#### String
```go
func (fs *FlagSet) String(name, defaultValue, usage string) *string
```

Defines a string flag with the specified name, default value, and usage string.

**Parameters:**
- `name` (string): Flag name
- `defaultValue` (string): Default value
- `usage` (string): Help text

**Returns:**
- `*string`: Pointer to the flag value

**Syntax:**
- `--name value`
- `--name=value`

**Example:**
```go
host := fs.String("host", "localhost", "Server host address")
fs.Parse(os.Args[1:])
fmt.Printf("Host: %s\n", *host)
```

---

#### StringVar
```go
func (fs *FlagSet) StringVar(name, shortKey string, defaultValue, usage string) *string
```

Defines a string flag with a short key (single character alias).

**Parameters:**
- `name` (string): Flag name
- `shortKey` (string): Short key (e.g., `"h"` for `-h`)
- `defaultValue` (string): Default value
- `usage` (string): Help text

**Returns:**
- `*string`: Pointer to the flag value

**Syntax:**
- `--name value`, `--name=value`
- `-shortKey value`, `-shortKey=value`

**Example:**
```go
host := fs.StringVar("host", "h", "localhost", "Server host address")
// Usage: --host 192.168.1.1 or -h 192.168.1.1
```

---

### Integer Flags

#### Int
```go
func (fs *FlagSet) Int(name string, defaultValue int, usage string) *int
```

Defines an integer flag.

**Parameters:**
- `name` (string): Flag name
- `defaultValue` (int): Default value
- `usage` (string): Help text

**Returns:**
- `*int`: Pointer to the flag value

**Example:**
```go
port := fs.Int("port", 8080, "Server port")
```

---

#### IntVar
```go
func (fs *FlagSet) IntVar(name, shortKey string, defaultValue int, usage string) *int
```

Defines an integer flag with a short key.

**Parameters:**
- `name` (string): Flag name
- `shortKey` (string): Short key
- `defaultValue` (int): Default value
- `usage` (string): Help text

**Returns:**
- `*int`: Pointer to the flag value

**Example:**
```go
port := fs.IntVar("port", "p", 8080, "Server port")
// Usage: --port 3000 or -p 3000
```

---

### Boolean Flags

#### Bool
```go
func (fs *FlagSet) Bool(name string, defaultValue bool, usage string) *bool
```

Defines a boolean flag. Boolean flags can be set without a value (defaults to true) or with explicit true/false values.

**Parameters:**
- `name` (string): Flag name
- `defaultValue` (bool): Default value
- `usage` (string): Help text

**Returns:**
- `*bool`: Pointer to the flag value

**Supported formats:**
- `--verbose` (sets to true)
- `--verbose=true` (explicit true)
- `--verbose=false` (explicit false)
- `--verbose true` (space-separated true)
- `--verbose false` (space-separated false)

**Example:**
```go
verbose := fs.Bool("verbose", false, "Enable verbose output")
debug := fs.Bool("debug", false, "Enable debug mode")

// Command line: --verbose --debug=false
fs.Parse(os.Args[1:])
fmt.Printf("Verbose: %t, Debug: %t\n", *verbose, *debug)
```

---

#### BoolVar
```go
func (fs *FlagSet) BoolVar(name, shortKey string, defaultValue bool, usage string) *bool
```

Defines a boolean flag with a short key.

**Parameters:**
- `name` (string): Flag name
- `shortKey` (string): Short key
- `defaultValue` (bool): Default value
- `usage` (string): Help text

**Returns:**
- `*bool`: Pointer to the flag value

**Example:**
```go
verbose := fs.BoolVar("verbose", "v", false, "Verbose output")
// Usage: --verbose or -v
```

---

### Float Flags

#### Float64
```go
func (fs *FlagSet) Float64(name string, defaultValue float64, usage string) *float64
```

Defines a float64 flag.

**Parameters:**
- `name` (string): Flag name
- `defaultValue` (float64): Default value
- `usage` (string): Help text

**Returns:**
- `*float64`: Pointer to the flag value

**Example:**
```go
rate := fs.Float64("rate", 1.0, "Processing rate")
// Usage: --rate 2.5 or --rate=3.14
```

---

### Duration Flags

#### Duration
```go
func (fs *FlagSet) Duration(name string, defaultValue time.Duration, usage string) *time.Duration
```

Defines a duration flag. Duration values are parsed using `time.ParseDuration` format.

**Parameters:**
- `name` (string): Flag name
- `defaultValue` (time.Duration): Default value
- `usage` (string): Help text

**Returns:**
- `*time.Duration`: Pointer to the flag value

**Supported formats:**
- `"5s"` (5 seconds)
- `"1m30s"` (1 minute 30 seconds)
- `"2h"` (2 hours)
- `"100ms"` (100 milliseconds)
- `"1h30m45s"` (1 hour 30 minutes 45 seconds)

**Example:**
```go
timeout := fs.Duration("timeout", 30*time.Second, "Request timeout duration")
// Usage: --timeout 5m or --timeout=1h30s
```

---

### String Slice Flags

#### StringSlice
```go
func (fs *FlagSet) StringSlice(name string, defaultValue []string, usage string) *[]string
```

Defines a string slice flag. Values are parsed as comma-separated.

**Parameters:**
- `name` (string): Flag name
- `defaultValue` ([]string): Default value
- `usage` (string): Help text

**Returns:**
- `*[]string`: Pointer to the flag value

**Input formats:**
- `"web,api,admin"` → `[]string{"web", "api", "admin"}`
- `"single"` → `[]string{"single"}`
- `""` → `[]string{}`

**Example:**
```go
tags := fs.StringSlice("tags", []string{"default"}, "Service tags")
// Usage: --tags web,api,production
fs.Parse(os.Args[1:])
for _, tag := range *tags {
    fmt.Printf("Tag: %s\n", tag)
}
```

---

## Parsing & Processing

### Parse

```go
func (fs *FlagSet) Parse(args []string) error
```

Parses command line arguments with optimized allocations and validates all constraints.

**Processing order (priority from lowest to highest):**
1. Defaults given at declaration (lowest priority)
2. Configuration files
3. Environment variables
4. Command-line arguments (highest priority)

A higher-priority source always wins, whatever order the loaders run in: a value
found in the config file does not suppress the matching environment variable.
Use [`Source`](#source-1) to find out which layer supplied a given value.

**Parameters:**
- `args` ([]string): Command line arguments (typically `os.Args[1:]`)

**Returns:**
- `error`: Parse error, validation error, or `nil` on success

**Supported argument formats:**

**Long flags:**
- `--flag value` (space-separated)
- `--flag=value` (equals-separated)
- `--boolean-flag` (boolean without value, defaults to true)
- `--boolean-flag=true` (explicit boolean)

**Short flags:**
- `-f value` (space-separated)
- `-f=value` (equals-separated)
- `-b` (boolean, defaults to true)
- `-b=false` (explicit boolean)

**Combined short flags:**
- `-abc` (equivalent to `-a -b -c`, all boolean)
- `-abc value` (last flag gets value)
- `-vdp 8080` (verbose + debug + port=8080)

**Special:**
- `--help`, `-h` (shows help and returns `"help requested"` error)
- `--` (end of flags marker)

**Example:**
```go
fs := flashflags.New("myapp")
host := fs.StringVar("host", "h", "localhost", "Server host")
port := fs.IntVar("port", "p", 8080, "Server port")

args := []string{"--host", "0.0.0.0", "-p", "3000"}
if err := fs.Parse(args); err != nil {
    if err.Error() == "help requested" {
        return // Help was displayed
    }
    log.Fatalf("Parse error: %v", err)
}
```

**Possible errors:**
- Unknown flag: `"unknown flag: --invalid"`
- Missing value: `"flag --port requires a value"`
- Invalid type: `"invalid int value for flag --port: abc"`
- Validation: `"validation failed for flag --port: port must be 1024-65535"`
- Required: `"required flag --api-key not provided"`
- Dependencies: `"flag --tls-cert requires --enable-tls to be set"`
- Help: `"help requested"` (special case)

---

### Arguments Access

#### Args
```go
func (fs *FlagSet) Args() []string
```

Returns the remaining non-flag arguments after parsing.

**Returns:**
- `[]string`: Non-flag arguments (returns a copy)

**Example:**
```go
fs := flashflags.New("myapp")
fs.String("host", "localhost", "Server host")
_ = fs.Parse([]string{"--host", "example.com", "file1", "file2"})
args := fs.Args()  // Returns ["file1", "file2"]
```

---

#### NArg
```go
func (fs *FlagSet) NArg() int
```

Returns the number of remaining non-flag arguments. Equivalent to `len(fs.Args())`.

**Returns:**
- `int`: Number of remaining arguments

**Example:**
```go
_ = fs.Parse([]string{"file1", "file2"})
count := fs.NArg()  // Returns 2
```

---

#### Arg
```go
func (fs *FlagSet) Arg(i int) string
```

Returns the i'th remaining argument. `Arg(0)` is the first remaining argument. Returns empty string if the index is out of bounds.

**Parameters:**
- `i` (int): Index (0-based)

**Returns:**
- `string`: The argument at index i, or empty string

**Example:**
```go
_ = fs.Parse([]string{"file1", "file2"})
first := fs.Arg(0)   // Returns "file1"
second := fs.Arg(1)  // Returns "file2"
third := fs.Arg(2)   // Returns ""
```

---

## Configuration Sources

### Configuration Files

#### SetConfigFile
```go
func (fs *FlagSet) SetConfigFile(path string)
```

Sets an explicit configuration file path. The config file is loaded automatically during `Parse()`. It is the lowest-priority source above the declared defaults, so both environment variables and CLI arguments override it.

**Parameters:**
- `path` (string): Path to configuration file

**Supported format:** JSON with flag names as keys

**Type mapping:** `string` → string, JSON number → `int` / `float64`, `true`/`false` → bool, array of strings → `[]string`. JSON has no duration type, so a duration flag accepts either the string form understood by `time.ParseDuration` (`"30s"`, `"1m30s"`) or a plain number of nanoseconds.

**Example config file (myapp.json):**
```json
{
  "host": "0.0.0.0",
  "port": 3000,
  "debug": true,
  "timeout": "45s",
  "tags": ["web", "api", "production"]
}
```

**Usage:**
```go
fs := flashflags.New("myapp")
fs.SetConfigFile("./config/myapp.json")
// File will be loaded automatically during Parse()
```

**Path handling:** The path must point at a regular file; a directory, a FIFO or a device is refused. Nothing else about the path is checked, because the path comes from your program, never from parsed arguments — `LoadConfig` runs before argument parsing, so no `--config` value can reach it.

> **Changed in v1.1.9.** A prefix allowlist (`/tmp/`, `/opt/`, `/etc/`, `/var/tmp/`, `/var/folders/`) previously rejected every other absolute path, including the user's own home directory — the location `AddConfigPath`'s example uses — and any path containing `..`, including a directory merely named `v1..2`. It has been removed. The regular-file check that replaced it closes a hang the allowlist permitted: a FIFO under `/tmp` passed validation, and reading one blocks `Parse` until a writer appears.

This is a robustness check, not a security boundary. If your program derives the config path from untrusted input, validate and confine it before calling `SetConfigFile`. Symlinks are followed; see [`EnableStrictConfigPaths`](#enablestrictconfigpaths) to refuse them.

---

#### AddConfigPath
```go
func (fs *FlagSet) AddConfigPath(path string)
```

Adds a directory to search for configuration files during auto-discovery.

**Parameters:**
- `path` (string): Directory path to search

**Auto-discovery searches for:**
- `{program-name}.json` (e.g., `"myapp.json"`)
- `{program-name}.config.json` (e.g., `"myapp.config.json"`)
- `config.json`

**No default search paths.** Discovery only visits directories added here, plus the file named by `SetConfigFile`. A `FlagSet` with neither loads no configuration, so a stray `config.json` in the working directory cannot alter a program that never opted into configuration files.

> **Changed in v1.1.9.** This section previously documented `"."`, `"./config"` and `$HOME` as defaults. `LoadConfig` returns early when no config file and no paths are set, so that list was unreachable — the coverage profile showed the branch at zero hits on the full suite. It has been removed rather than implemented: silently loading `./<name>.json` for programs that never asked is not a default worth having.

**Paths are not expanded.** The string is used as given, so `"$HOME/.myapp"` looks for a directory literally named `$HOME`. Resolve the home directory in your own code.

**Example:**
```go
fs := flashflags.New("myapp")
fs.AddConfigPath("./config")   // ./config/myapp.json
fs.AddConfigPath("/etc/myapp") // /etc/myapp/myapp.json

// os.UserHomeDir, not os.Getenv("HOME"): HOME is normally unset on Windows,
// where the home directory lives in USERPROFILE.
if home, err := os.UserHomeDir(); err == nil {
    fs.AddConfigPath(home)     // $HOME/myapp.json
}
```

Added paths are searched in the order they were added, each probed for every candidate filename before moving on.

---

#### EnableStrictConfigPaths
```go
func (fs *FlagSet) EnableStrictConfigPaths()
```

Makes `Parse` refuse a configuration file whose final path component is a symbolic link, for both an explicitly named file and one found by auto-discovery. **Off by default.**

**Why it is opt-in:** following a symlink is normally what a config file is expected to do. A Kubernetes ConfigMap mount projects every key as a symlink into a `..data` directory, and dotfile managers such as GNU Stow and chezmoi link a config into place from a repository. Refusing symlinks by default would break both, on the platform where they are most common.

**When to turn it on:** when the program reads configuration from a directory other local users can write to, where an attacker can plant a link before startup and have a privileged process read a file of their choosing. That is the classic `/tmp` symlink attack, and it is the only case this setting is for.

**Scope:** only the final path component is examined. A symlinked parent directory is traversed normally — deliberately, since `/tmp` on macOS is itself a symlink to `/private/tmp`.

**Guarantee by platform:**

| Platform | Mechanism | Strength |
|----------|-----------|----------|
| Unix | `O_NOFOLLOW` at open | The kernel refuses the call; check and open are one operation and cannot be raced |
| Windows | `Lstat` before open | Best-effort: Go exposes no portable `O_NOFOLLOW`, so a replacement racing between the two could be followed. Creating a symlink on Windows needs `SeCreateSymbolicLinkPrivilege` or Developer Mode |

**Example:**
```go
fs := flashflags.New("myapp")
fs.SetConfigFile("/tmp/myapp.json")
fs.EnableStrictConfigPaths()

if err := fs.Parse(os.Args[1:]); err != nil {
    // "config file /tmp/myapp.json is a symbolic link (strict config paths enabled)"
    log.Fatal(err)
}
```

*Added in v1.1.9.*

---

#### LoadConfig
```go
func (fs *FlagSet) LoadConfig() error
```

Loads configuration from file and applies it. This is called automatically during `Parse()`, but can be called manually if needed.

**Returns:**
- `error`: File reading, JSON parsing, or validation errors

**Possible errors:**
- File errors: permission denied, file not found
- JSON parsing: invalid JSON syntax
- Path validation: unsafe file paths
- Flag validation: config values that fail validators

**Note:** Missing auto-discovery config files are not considered errors.

**Example:**
```go
if err := fs.LoadConfig(); err != nil {
    log.Printf("Config error: %v", err)
}
```

---

### Environment Variables

#### SetEnvPrefix
```go
func (fs *FlagSet) SetEnvPrefix(prefix string)
```

Sets the prefix for environment variable lookup and enables env var processing.

**Parameters:**
- `prefix` (string): Prefix for environment variables

**Conversion rules:**
- Hyphens → underscores: `"db-host"` → `"PREFIX_DB_HOST"`
- All uppercase: `"api-key"` → `"PREFIX_API_KEY"`
- Prefix case preserved: `"MyApp"` → `"MyApp_DB_HOST"`

**Example:**
```go
fs := flashflags.New("webserver")
host := fs.String("db-host", "localhost", "Database host")
fs.SetEnvPrefix("WEBAPP")

// Environment variable: WEBAPP_DB_HOST=postgresql.example.com
// Command line override: --db-host=localhost
// Result: host="localhost" (CLI wins over env var)
```

---

#### EnableEnvLookup
```go
func (fs *FlagSet) EnableEnvLookup()
```

Enables environment variable lookup using default naming convention (no prefix).

**Conversion rules:**
- Hyphens → underscores: `"db-host"` → `"DB_HOST"`
- All uppercase: `"api-key"` → `"API_KEY"`

**Example:**
```go
fs := flashflags.New("myapp")
host := fs.String("db-host", "localhost", "Database host")
port := fs.Int("db-port", 5432, "Database port")
fs.EnableEnvLookup()

// Reads from: DB_HOST and DB_PORT environment variables
// export DB_HOST=postgresql.example.com
// export DB_PORT=5433
```

---

#### SetEnvVar
```go
func (fs *FlagSet) SetEnvVar(flagName, envVarName string) error
```

Sets a custom environment variable name for a specific flag.

**Parameters:**
- `flagName` (string): Flag name
- `envVarName` (string): Environment variable name

**Returns:**
- `error`: Error if flag doesn't exist

**Example:**
```go
fs := flashflags.New("myapp")
dbURL := fs.String("database-url", "", "Database connection URL")

// Use standard naming instead of MYAPP_DATABASE_URL
fs.SetEnvVar("database-url", "DATABASE_URL")

// Now reads from DATABASE_URL environment variable
// export DATABASE_URL=postgres://user:pass@host/db
```

---

#### LoadEnvironmentVariables
```go
func (fs *FlagSet) LoadEnvironmentVariables() error
```

Loads values from environment variables. Called automatically during `Parse()`, but can be called manually if needed.

**Returns:**
- `error`: Type conversion or validation errors

**Possible errors:**
- Type conversion: `"invalid int value for MYAPP_PORT: abc"`
- Validation: environment values that fail validators
- Duration parsing: invalid duration format

**Example:**
```go
if err := fs.LoadEnvironmentVariables(); err != nil {
    log.Fatalf("Environment variable error: %v", err)
}
```

---

## Validation & Constraints

### Custom Validators

#### SetValidator
```go
func (fs *FlagSet) SetValidator(name string, validator func(interface{}) error) error
```

Sets a validation function for a specific flag.

**Parameters:**
- `name` (string): Flag name
- `validator` (func(interface{}) error): Validation function

**Returns:**
- `error`: Error if flag doesn't exist

**Example:**
```go
fs := flashflags.New("myapp")
port := fs.IntVar("port", "p", 8080, "Server port")

err := fs.SetValidator("port", func(val interface{}) error {
    port, ok := val.(int)
    if !ok {
        return fmt.Errorf("expected int, got %T", val)
    }
    if port < 1024 || port > 65535 {
        return fmt.Errorf("port must be between 1024-65535, got %d", port)
    }
    return nil
})
```

---

#### ValidateAll
```go
func (fs *FlagSet) ValidateAll() error
```

Validates all flags that have validators set. Called automatically during `Parse()`, but can be called manually.

**Returns:**
- `error`: First validation error encountered

**Example:**
```go
if err := fs.ValidateAll(); err != nil {
    // Error: "validation failed for flag --port: port too low"
}
```

---

### Required Flags

#### SetRequired
```go
func (fs *FlagSet) SetRequired(name string) error
```

Marks a flag as required. Required flags must be provided during parsing.

**Parameters:**
- `name` (string): Flag name

**Returns:**
- `error`: Error if flag doesn't exist

**Example:**
```go
fs := flashflags.New("myapp")
apiKey := fs.String("api-key", "", "API authentication key")

if err := fs.SetRequired("api-key"); err != nil {
    log.Fatal(err)
}

// This will fail if --api-key is not provided
if err := fs.Parse(os.Args[1:]); err != nil {
    log.Fatal(err) // "required flag --api-key not provided"
}
```

---

#### ValidateRequired
```go
func (fs *FlagSet) ValidateRequired() error
```

Checks that all required flags are set. Called automatically during `Parse()`.

**Returns:**
- `error`: Error if required flag is missing

**Example:**
```go
fs.SetRequired("api-key")
if err := fs.ValidateRequired(); err != nil {
    // Error: "required flag --api-key not provided"
}
```

---

### Flag Dependencies

#### SetDependencies
```go
func (fs *FlagSet) SetDependencies(name string, dependencies ...string) error
```

Sets dependencies for a flag. When this flag is set, all dependent flags must also be set.

**Parameters:**
- `name` (string): Flag name
- `dependencies` (...string): Required dependency flag names

**Returns:**
- `error`: Error if flag doesn't exist

**Example:**
```go
fs := flashflags.New("myapp")
enableTLS := fs.Bool("enable-tls", false, "Enable TLS")
tlsCert := fs.String("tls-cert", "", "TLS certificate file")
tlsKey := fs.String("tls-key", "", "TLS private key file")

// Both cert and key require TLS to be enabled
if err := fs.SetDependencies("tls-cert", "enable-tls"); err != nil {
    log.Fatal(err)
}
if err := fs.SetDependencies("tls-key", "enable-tls"); err != nil {
    log.Fatal(err)
}

// This will fail if --tls-cert is provided without --enable-tls
if err := fs.Parse(os.Args[1:]); err != nil {
    log.Fatal(err) // "flag --tls-cert requires --enable-tls to be set"
}
```

---

#### ValidateDependencies
```go
func (fs *FlagSet) ValidateDependencies() error
```

Checks that all flag dependencies are satisfied. Called automatically during `Parse()`.

**Returns:**
- `error`: Dependency error

**Possible errors:**
- Missing dependency: `"flag --flagname requires --dependency to be set"`
- Non-existent dependency: `"flag --flagname depends on non-existent flag --missing"`

**Example:**
```go
fs.SetDependencies("tls-cert", "enable-tls")
if err := fs.ValidateDependencies(); err != nil {
    // Error: "flag --tls-cert requires --enable-tls to be set"
}
```

---

#### ValidateAllConstraints
```go
func (fs *FlagSet) ValidateAllConstraints() error
```

Validates all constraints: validators, required flags, and dependencies. Called automatically during `Parse()`.

**Returns:**
- `error`: First constraint violation encountered

**Example:**
```go
if err := fs.ValidateAllConstraints(); err != nil {
    fmt.Printf("Validation failed: %v\n", err)
    fs.PrintHelp()
    os.Exit(1)
}
```

---

## Help & Documentation

### SetDescription
```go
func (fs *FlagSet) SetDescription(description string)
```

Sets the program description displayed at the top of help output.

**Parameters:**
- `description` (string): Program description

**Example:**
```go
fs := flashflags.New("webserver")
fs.SetDescription("High-performance HTTP server with advanced configuration options")
```

---

### SetVersion
```go
func (fs *FlagSet) SetVersion(version string)
```

Sets the program version displayed in help output.

**Parameters:**
- `version` (string): Version string (e.g., `"v1.2.3"`)

**Example:**
```go
fs := flashflags.New("myapp")
fs.SetVersion("v2.1.0")
```

---

### SetGroup
```go
func (fs *FlagSet) SetGroup(name, group string) error
```

Sets the group name for a flag to organize help output.

**Parameters:**
- `name` (string): Flag name
- `group` (string): Group name

**Returns:**
- `error`: Error if flag doesn't exist

**Example:**
```go
fs := flashflags.New("server")
host := fs.String("host", "localhost", "Server host")
port := fs.Int("port", 8080, "Server port")
dbHost := fs.String("db-host", "localhost", "Database host")

fs.SetGroup("host", "Server Options")
fs.SetGroup("port", "Server Options")
fs.SetGroup("db-host", "Database Options")
```

---

### Help
```go
func (fs *FlagSet) Help() string
```

Generates and returns the complete help text as a string.

**Returns:**
- `string`: Formatted help text

**Example:**
```go
fs := flashflags.New("myserver")
fs.SetDescription("High-performance web server")
fs.SetVersion("v2.1.0")

port := fs.IntVar("port", "p", 8080, "Server port")
host := fs.String("host", "localhost", "Server host")

fs.SetGroup("port", "Server Options")
fs.SetGroup("host", "Server Options")

helpText := fs.Help()
// Contains formatted help with grouped flags, defaults, requirements, etc.
```

---

### PrintHelp
```go
func (fs *FlagSet) PrintHelp()
```

Prints the complete help text to stdout. Convenience method that calls `Help()` and prints the result.

**Example:**
```go
fs := flashflags.New("myapp")
// ... define flags ...

if err := fs.Parse(os.Args[1:]); err != nil {
    if err.Error() == "help requested" {
        // Help was already printed by Parse()
        os.Exit(0)
    }
    fmt.Printf("Error: %v\n", err)
    fs.PrintHelp()  // Show help on errors
    os.Exit(1)
}
```

---

### PrintUsage
```go
func (fs *FlagSet) PrintUsage()
```

Prints basic usage information for all flags to stdout. Simpler than `PrintHelp()`.

**Output format:**
```
Usage of myapp:
  --flagname, -s
        Description text (type: flagtype)
```

**Example:**
```go
fs := flashflags.New("myapp")
fs.StringVar("host", "h", "localhost", "Server host")
fs.IntVar("port", "p", 8080, "Server port")

fs.PrintUsage()
// Output:
// Usage of myapp:
//   --host, -h
//         Server host (type: string)
//   --port, -p
//         Server port (type: int)
```

---

## Flag Inspection

### Lookup
```go
func (fs *FlagSet) Lookup(name string) *Flag
```

Finds a flag by name and returns a pointer to the Flag, or `nil` if not found.

**Parameters:**
- `name` (string): Flag name

**Returns:**
- `*Flag`: Flag pointer or `nil`

**Example:**
```go
if flag := fs.Lookup("port"); flag != nil {
    fmt.Printf("Port flag exists, value: %v\n", flag.Value())
} else {
    fmt.Println("Port flag not found")
}
```

---

### VisitAll
```go
func (fs *FlagSet) VisitAll(fn func(*Flag))
```

Calls fn for each flag in the set. Order is not guaranteed.

**Parameters:**
- `fn` (func(*Flag)): Function to call for each flag

**Example:**
```go
fs.VisitAll(func(flag *Flag) {
    fmt.Printf("Flag: %s = %v (type: %s)\n", flag.Name(), flag.Value(), flag.Type())
})
```

---

### Changed
```go
func (fs *FlagSet) Changed(name string) bool
```

Returns whether the specified flag was set during parsing.

**Parameters:**
- `name` (string): Flag name

**Returns:**
- `bool`: `true` if flag was set

**Example:**
```go
if fs.Changed("debug") {
    fmt.Println("Debug mode was explicitly enabled")
}
```

---

### Source
```go
func (fs *FlagSet) Source(name string) string
```

Returns the configuration source that supplied the named flag's current value.
This is the quickest way to debug precedence: it tells you not just that a flag
changed, but which layer won.

**Parameters:**
- `name` (string): Flag name

**Returns:**
- `string`: One of `"cli"`, `"env"`, `"config"`, `"default"`. Unknown flags report `"default"`.

**Example:**
```go
fs.Parse(os.Args[1:])
fmt.Println(fs.GetInt("port"), "from", fs.Source("port")) // 3000 from env
```

*Added in v1.1.9.*

---

## Utility Functions

### GetString
```go
func (fs *FlagSet) GetString(name string) string
```

Gets a flag value as string, with automatic type conversion.

**Parameters:**
- `name` (string): Flag name

**Returns:**
- `string`: String value or empty string

**Example:**
```go
fmt.Println(fs.GetString("host"))  // "example.com"
fmt.Println(fs.GetString("port"))  // "3000" (converted from int)
```

---

### GetInt
```go
func (fs *FlagSet) GetInt(name string) int
```

Gets a flag value as int.

**Parameters:**
- `name` (string): Flag name

**Returns:**
- `int`: Integer value or 0

**Example:**
```go
fmt.Println(fs.GetInt("port"))     // 3000
```

---

### GetBool
```go
func (fs *FlagSet) GetBool(name string) bool
```

Gets a flag value as bool.

**Parameters:**
- `name` (string): Flag name

**Returns:**
- `bool`: Boolean value or false

**Example:**
```go
fmt.Println(fs.GetBool("verbose"))  // true
```

---

### GetDuration
```go
func (fs *FlagSet) GetDuration(name string) time.Duration
```

Gets a flag value as duration.

**Parameters:**
- `name` (string): Flag name

**Returns:**
- `time.Duration`: Duration value or 0

**Example:**
```go
fmt.Println(fs.GetDuration("timeout"))   // 45s
```

---

### GetFloat64
```go
func (fs *FlagSet) GetFloat64(name string) float64
```

Gets a flag value as float64.

**Parameters:**
- `name` (string): Flag name

**Returns:**
- `float64`: Float value or 0.0

**Example:**
```go
fmt.Println(fs.GetFloat64("rate"))       // 2.5
```

---

### GetStringSlice
```go
func (fs *FlagSet) GetStringSlice(name string) []string
```

Gets a flag value as string slice.

**Parameters:**
- `name` (string): Flag name

**Returns:**
- `[]string`: String slice or empty slice

**Example:**
```go
fmt.Println(fs.GetStringSlice("tags"))    // ["web", "api", "prod"]
```

---

### Reset
```go
func (fs *FlagSet) Reset()
```

Resets all flags to their default values and marks them as unchanged.

**Example:**
```go
fs := flashflags.New("test")
port := fs.Int("port", 8080, "Server port")

fs.Parse([]string{"--port", "3000"})
fmt.Println(*port)          // 3000
fmt.Println(fs.Changed("port")) // true

fs.Reset()
fmt.Println(*port)          // 8080 (default)
fmt.Println(fs.Changed("port")) // false
```

---

### ResetFlag
```go
func (fs *FlagSet) ResetFlag(name string) error
```

Resets a specific flag to its default value.

**Parameters:**
- `name` (string): Flag name

**Returns:**
- `error`: Error if flag doesn't exist

**Example:**
```go
fs.ResetFlag("port")  // Only reset port
fmt.Println(*port)    // 8080 (default)
```

---

## Interfaces

### ConfigFlag

Interface for configuration management integration.

```go
type ConfigFlag interface {
    Name() string
    Value() interface{}
    Type() string
    Changed() bool
    Usage() string
}
```

The `Flag` type implements this interface.

---

### ConfigFlagSet

Interface for flag set configuration integration.

```go
type ConfigFlagSet interface {
    VisitAll(func(ConfigFlag))
    Lookup(name string) ConfigFlag
}
```

---

### FlagSetAdapter

Wrapper for integrating with configuration management systems.

```go
type FlagSetAdapter struct {
    *FlagSet
}

func NewAdapter(fs *FlagSet) *FlagSetAdapter
```

**Example:**
```go
fs := flashflags.New("myapp")
adapter := flashflags.NewAdapter(fs)

// Use with configuration management systems
adapter.VisitAll(func(flag flashflags.ConfigFlag) {
    fmt.Printf("Flag: %s\n", flag.Name())
})
```

---

## Input Screening

Flag values are screened for input that is malformed as a string, whatever it is later used for.

### What is checked

- **Length**: values above 10000 bytes are rejected
- **Null bytes**: `\x00` is rejected — it truncates the string in any C API it reaches
- **Control characters**: C0 controls other than `\t`, `\n`, `\r` are rejected — terminal escape sequences, not data

### What is deliberately not checked

A flag parser does not know whether a value will reach a shell, a SQL driver, an `fmt` verb or a file open, so it does not guess. Shell metacharacters, SQL fragments, absolute or relative paths and format verbs all pass through unchanged. Escaping belongs at the point of use: run `os/exec` without a shell, parameterize SQL, resolve and confine paths before opening them.

> **Changed in v1.1.9.** The screening previously also rejected values containing `/etc/`, `/proc/`, `/sys/`, `rm -rf`, `drop table`, `$(`, a backtick, the `%n %s %x %d %c %p` format verbs, `../` and Windows device names. That denylist stopped no attack — `a; rm -rf ~`, `deploy && restart` and `..%2f..%2fetc` all passed it untouched — while rejecting ordinary input such as `--config /etc/myapp.conf`. It has been removed.

### Fast-Path Optimization

Values under 100 bytes made only of `a-z`, `A-Z`, `0-9`, `-`, `_`, `.`, `:` skip the scan entirely: such a value cannot contain a null byte or a control character.

### Overhead

Roughly 132ns per operation (17%).

### Example

```go
fs := flashflags.New("myapp")
cmd := fs.String("command", "", "Command to execute")

// Rejected — malformed as a string:
fs.Parse([]string{"--command", "bad\x00value"})   // null byte
fs.Parse([]string{"--command", "\x1b[31mred"})    // control character

// Accepted — the parser does not guess what these mean:
fs.Parse([]string{"--command", "rm -rf /tmp/build"})
fs.Parse([]string{"--command", "/etc/myapp.conf"})
fs.Parse([]string{"--command", "a; rm -rf ~"})
```

---

## Error Handling

Flash-flags returns descriptive errors for various scenarios:

| Error Type | Example |
|------------|---------|
| Unknown flag | `"unknown flag: --invalid"` |
| Missing value | `"flag --port requires a value"` |
| Invalid type | `"invalid int value for flag --port: abc"` |
| Validation | `"validation failed for flag --port: port must be 1024-65535"` |
| Required | `"required flag --api-key not provided"` |
| Dependencies | `"flag --tls-cert requires --enable-tls to be set"` |
| Type conversion | `"invalid int value for flag --port: abc"` |
| Config error | `"config file error: failed to read config.json"` |
| Help | `"help requested"` (special case) |
| Input screening | `"flag --name contains null byte at position 3"` |
| Over length | `"flag --data value too long: 15000 bytes (max: 10000)"` |

All errors include the flag name and specific details for debugging.

---

## Performance

### Benchmark Results

```
AMD Ryzen 5 7520U, Go 1.25.9+:
  Flash-flags (secure):      924 ns/op    (with full security validation)
  Go standard library flag:  792 ns/op    (baseline, no security)
  Spf13/pflag:             1,322 ns/op    (43% slower than flash-flags)
```

**Screening overhead:** roughly 132ns (17%)

### Internal Performance

```
BenchmarkGetters/GetString    136M    9.01 ns/op   0 B/op   0 allocs/op
BenchmarkGetters/GetInt       142M    8.35 ns/op   0 B/op   0 allocs/op
BenchmarkGetters/GetBool      135M    8.88 ns/op   0 B/op   0 allocs/op
BenchmarkGetters/GetDuration  134M    8.86 ns/op   0 B/op   0 allocs/op
```

### Key Characteristics

- 924ns with input screening enabled
- 43% faster than pflag
- Sub-nanosecond flag value access (8-9ns)
- Zero allocations for getter operations
- Unsynchronized concurrent reads once `Parse()` has returned (see Thread Safety)
- O(1) hash-based flag lookup

---

## Version History

**Current version:** v1.1.9

See `changelog/` directory for version history.

---

## License

Flash-flags is licensed under the [Mozilla Public License 2.0](./LICENSE.md).

---

## Links

- **Repository**: [github.com/agilira/flash-flags](https://github.com/agilira/flash-flags)
- **Documentation**: [pkg.go.dev](https://pkg.go.dev/github.com/agilira/flash-flags)
- **Issues**: [GitHub Issues](https://github.com/agilira/flash-flags/issues)

---

**flash-flags • an AGILira library**
