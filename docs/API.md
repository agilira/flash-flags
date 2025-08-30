# FlashFlags API Reference

FlashFlags is an ultra-fast, zero-dependency, lock-free command-line flag parsing library for Go.

## Table of Contents

- [Installation](#installation)
- [Core Types](#core-types)
- [FlagSet Methods](#flagset-methods)
- [Flag Types](#flag-types)
- [Configuration](#configuration)
- [Validation](#validation)
- [Environment Variables](#environment-variables)
- [Advanced Features](#advanced-features)

## Installation

```bash
go get github.com/agilira/flash-flags
```

## Core Types

### FlagSet

The main structure for managing command-line flags.

```go
type FlagSet struct {
    // Private fields
}
```

### Flag

Represents a single command-line flag.

```go
type Flag struct {
    // Private fields
}
```

## FlagSet Methods

### Constructor

#### `New(name string) *FlagSet`

Creates a new flag set with the specified name.

**Parameters:**
- `name`: The name of the program/command

**Returns:**
- `*FlagSet`: A new FlagSet instance

**Example:**
```go
fs := flashflags.New("myapp")
```

### Flag Registration

#### `String(name, defaultValue, usage string) *string`

Registers a string flag and returns a pointer to its value.

**Parameters:**
- `name`: Flag name (used with `--name`)
- `defaultValue`: Default value if flag is not provided
- `usage`: Help text description

**Returns:**
- `*string`: Pointer to the flag value

#### `StringVar(name, shortKey string, defaultValue, usage string) *string`

Registers a string flag with optional short key.

**Parameters:**
- `name`: Flag name (used with `--name`)
- `shortKey`: Single character for short flag (used with `-x`)
- `defaultValue`: Default value if flag is not provided
- `usage`: Help text description

**Returns:**
- `*string`: Pointer to the flag value

#### `Int(name string, defaultValue int, usage string) *int`

Registers an integer flag.

#### `IntVar(name, shortKey string, defaultValue int, usage string) *int`

Registers an integer flag with optional short key.

#### `Bool(name string, defaultValue bool, usage string) *bool`

Registers a boolean flag.

#### `BoolVar(name, shortKey string, defaultValue bool, usage string) *bool`

Registers a boolean flag with optional short key.

#### `Duration(name string, defaultValue time.Duration, usage string) *time.Duration`

Registers a duration flag (accepts values like "1h30m", "5s", etc.).

#### `Float64(name string, defaultValue float64, usage string) *float64`

Registers a float64 flag.

#### `StringSlice(name string, defaultValue []string, usage string) *[]string`

Registers a string slice flag (comma-separated values).

### Parsing

#### `Parse(args []string) error`

Parses command-line arguments in the following priority order:
1. Command-line arguments (highest priority)
2. Environment variables (medium priority)  
3. Configuration file (lowest priority)

**Parameters:**
- `args`: Command-line arguments (typically `os.Args[1:]`)

**Returns:**
- `error`: Parsing error or nil on success

### Value Retrieval

#### `GetString(name string) string`

Gets a flag value as string.

#### `GetInt(name string) int`

Gets a flag value as integer.

#### `GetBool(name string) bool`

Gets a flag value as boolean.

#### `GetDuration(name string) time.Duration`

Gets a flag value as duration.

#### `GetFloat64(name string) float64`

Gets a flag value as float64.

#### `GetStringSlice(name string) []string`

Gets a flag value as string slice.

### Flag Information

#### `Lookup(name string) *Flag`

Finds a flag by name.

**Returns:**
- `*Flag`: The flag or nil if not found

#### `Changed(name string) bool`

Returns whether a flag was explicitly set.

#### `VisitAll(fn func(*Flag))`

Calls the provided function for each registered flag.

### Help System

#### `SetDescription(description string)`

Sets the program description for help output.

#### `SetVersion(version string)`

Sets the program version for help output.

#### `Help() string`

Returns the formatted help text.

#### `PrintHelp()`

Prints help text to stdout.

#### `PrintUsage()`

Prints basic usage information.

### Configuration Loading

#### `LoadConfig() error`

Loads configuration from file and applies it. This is called automatically during Parse, but can be called manually if needed.

#### `LoadEnvironmentVariables() error`

Loads values from environment variables. This is called automatically during Parse, but can be called manually if needed.

## Flag Types

### Supported Types

| Type | Go Type | Example Values |
|------|---------|----------------|
| `string` | `string` | `"hello"`, `"world"` |
| `int` | `int` | `42`, `-10` |
| `bool` | `bool` | `true`, `false` |
| `float64` | `float64` | `3.14`, `-2.5` |
| `duration` | `time.Duration` | `"1h30m"`, `"5s"`, `"100ms"` |
| `stringSlice` | `[]string` | `"a,b,c"` → `["a", "b", "c"]` |

### Boolean Flags

Boolean flags can be used in several ways:

```bash
# All equivalent to setting flag to true
--verbose
--verbose=true
--verbose true

# Setting to false
--verbose=false
--verbose false
```

## Configuration

### Configuration Files

#### `SetConfigFile(path string)`

Sets an explicit configuration file path.

#### `AddConfigPath(path string)`

Adds a directory to search for configuration files.

#### Auto-discovery

If no explicit config file is set, FlashFlags searches for:
- `{program-name}.json`
- `{program-name}.config.json`
- `config.json`

In these directories (in order):
- Current directory (`.`)
- `./config/`
- User home directory

#### JSON Format

Configuration files use JSON format:

```json
{
  "host": "localhost",
  "port": 8080,
  "verbose": true,
  "tags": ["web", "api"],
  "timeout": "30s"
}
```

## Validation

### Flag Validation

#### `SetValidator(name string, validator func(interface{}) error) error`

Sets a custom validation function for a flag.

**Example:**
```go
err := fs.SetValidator("port", func(val interface{}) error {
    port := val.(int)
    if port < 1 || port > 65535 {
        return fmt.Errorf("port must be between 1 and 65535")
    }
    return nil
})
```

### Required Flags

#### `SetRequired(name string) error`

Marks a flag as required.

### Dependencies

#### `SetDependencies(name string, dependencies ...string) error`

Sets flag dependencies (this flag requires other flags to be set).

### Validation Methods

#### `ValidateAll() error`

Validates all flags that have validators.

#### `ValidateRequired() error`

Checks that all required flags are set.

#### `ValidateDependencies() error`

Validates flag dependencies.

#### `ValidateAllConstraints() error`

Validates all constraints (validators, required flags, dependencies).

## Environment Variables

### Enable Environment Variable Lookup

#### `EnableEnvLookup()`

Enables automatic environment variable lookup using default naming convention.

#### `SetEnvPrefix(prefix string)`

Sets a prefix for environment variable names.

**Example:**
```go
fs.SetEnvPrefix("MYAPP")
// Flag "db-host" will look for "MYAPP_DB_HOST"
```

#### `SetEnvVar(flagName, envVarName string) error`

Sets a custom environment variable name for a specific flag.

### Naming Convention

- Default: `DB_HOST` for flag `db-host`
- With prefix: `MYAPP_DB_HOST` for flag `db-host` with prefix `MYAPP`
- Custom: Use `SetEnvVar()` for custom names

## Configuration Integration

### Interfaces

#### `ConfigFlag` Interface

Represents a flag interface for configuration management integration. This interface allows flash-flags to integrate seamlessly with configuration management systems like Argus, Viper, or custom solutions.

```go
type ConfigFlag interface {
    Name() string
    Value() interface{}
    Type() string
    Changed() bool
    Usage() string
}
```

#### `ConfigFlagSet` Interface

Represents a collection of flags for configuration integration. This interface provides the standard contract expected by configuration managers.

```go
type ConfigFlagSet interface {
    VisitAll(func(ConfigFlag))
    Lookup(name string) ConfigFlag
}
```

### Adapter

#### `FlagSetAdapter`

Wraps a FlagSet to implement ConfigFlagSet interface. This allows seamless integration with configuration management systems.

#### `NewAdapter(fs *FlagSet) *FlagSetAdapter`

Creates a new FlagSetAdapter for configuration integration. This allows a FlagSet to be used with configuration management systems that expect the ConfigFlagSet interface.

## Advanced Features

### Grouping

#### `SetGroup(name, group string) error`

Organizes flags into groups for better help output.

### Reset

#### `Reset()`

Resets all flags to their default values.

#### `ResetFlag(name string) error`

Resets a specific flag to its default value.

### Priority Order

Values are applied in this order (later values override earlier ones):

1. Default values
2. Configuration file values
3. Environment variable values
4. Command-line argument values

## Flag Methods

The `Flag` type provides these methods:

#### `Name() string`

Returns the flag name.

#### `Value() interface{}`

Returns the current flag value.

#### `Type() string`

Returns the flag type.

#### `Changed() bool`

Returns whether the flag was explicitly set.

#### `Usage() string`

Returns the flag usage string.

#### `SetValidator(validator func(interface{}) error)`

Sets a validation function for the flag. The validator function will be called whenever the flag value is set.

#### `Validate() error`

Validates the flag using its validator (if set).

#### `Reset()`

Resets the flag to its default value.

## Error Handling

FlashFlags returns errors for:

- Unknown flags
- Invalid flag values
- Missing required flags
- Validation failures
- Dependency violations
- Configuration file errors

All errors include descriptive messages for easy debugging.

## Performance Features

- **Zero-allocation parsing** for optimal performance
- **Lock-free operations** for concurrent safety
- **Minimal memory footprint**
- **Fast flag lookup** using hash maps
- **Optimized string operations** to avoid unnecessary allocations

---

flash-flags • an AGILira library