// flash-flags.go: Ultra-fast command-line flag parsing for Go
//
// Copyright (c) 2025 AGILira - A. Giordano
// Series: an AGILira library
// SPDX-License-Identifier: MPL-2.0

// Package flashflags provides ultra-fast, zero-dependency command-line flag parsing.
// This library is extracted from argus with exactly the same structure for maximum performance.
//
// # Concurrency
//
// A FlagSet is not internally synchronized: it holds no mutexes and performs no
// atomic operations, which is what keeps reads down to a plain map lookup.
// The guarantee it offers is the one a flag set actually needs:
//
//   - Reads are safe from any number of goroutines once Parse has returned,
//     because nothing mutates the flag set afterwards. This covers Lookup, Value,
//     Changed, Source, the Get* accessors and the pointers returned at
//     declaration time.
//   - Mutating calls are not safe to run concurrently with anything. These are
//     Parse, LoadConfig, LoadEnvironmentVariables, Reset, ResetFlag and every
//     Set* configuration method. Declare, configure and parse from a single
//     goroutine -- normally the one running main -- before sharing the FlagSet.
//
// Calling Reset while another goroutine reads is a data race and will be
// reported by the race detector. If an application needs to re-parse at runtime,
// it must provide its own synchronization, or build a fresh FlagSet and swap it
// behind a pointer.
package flashflags

import (
	"fmt"
	"time"
)

// FlagSet represents a collection of command-line flags with parsing and validation capabilities.
// It implements ultra-fast flag set handling using only the standard library.
//
// FlagSet supports multiple configuration sources in priority order:
//  1. Command-line arguments (highest priority)
//  2. Environment variables
//  3. Configuration files
//  4. Defaults given at declaration (lowest priority)
//
// A higher-priority source always wins, whatever order the loaders run in: a
// value found in the config file does not suppress the matching environment
// variable. Use Source to find out which layer supplied a given value.
//
// Example usage:
//
//	fs := flashflags.New("myapp")
//	host := fs.StringVar("host", "h", "localhost", "Server host address")
//	port := fs.IntVar("port", "p", 8080, "Server port number")
//
//	if err := fs.Parse(os.Args[1:]); err != nil {
//		log.Fatal(err)
//	}
//
//	fmt.Printf("Server starting on %s:%d\n", *host, *port)
//
// A FlagSet is safe for concurrent reads once Parse has returned, and is not
// internally synchronized for writes. See the package documentation on
// concurrency.
type FlagSet struct {
	flags           map[string]*Flag // Long flag name -> Flag
	shortMap        map[string]*Flag // Short flag key -> Flag
	name            string
	description     string   // Program description for help
	version         string   // Program version for help
	configFile      string   // Configuration file path
	configPaths     []string // Auto-discovery paths for config files
	configLoaded    bool     // Whether config has been loaded
	envPrefix       string   // Prefix for environment variables (e.g., "MYAPP")
	enableEnvLookup bool     // Whether to lookup environment variables
	strictPaths     bool     // Whether to refuse a symlinked configuration file
	args            []string // Remaining non-flag arguments after parsing
	// repeated holds the slice flags already given in the current Parse, so
	// a repetition appends while a new Parse starts over.
	repeated map[string]bool
}

// New creates a new FlagSet with the specified name.
// The name is used for help output and configuration file discovery.
// Returns a FlagSet with zero external dependencies.
//
// Thread Safety:
// A FlagSet is safe for concurrent reads once Parse has returned, and is not
// internally synchronized for writes. See the package documentation on
// concurrency.
// Multiple goroutines can safely read flag values concurrently after parsing is complete.
// However, Parse() should only be called once from a single goroutine.
//
// Example:
//
//	fs := flashflags.New("myapp")
//	// Safe to use fs concurrently after Parse() completes
func New(name string) *FlagSet {
	return &FlagSet{
		flags:    make(map[string]*Flag),
		shortMap: make(map[string]*Flag),
		name:     name,
	}
}

// define registers a flag. value is its starting value, defaultValue what
// Reset restores, and ptr the pointer handed back to the caller. They differ
// only for a slice, whose starting value is a copy of the caller's default.
// An empty shortKey registers the long form only.
func (fs *FlagSet) define(name, shortKey, usage string, kind *flagKind, value, defaultValue, ptr interface{}) {
	flag := &Flag{
		name:         name,
		value:        value,
		defaultValue: defaultValue,
		ptr:          ptr,
		kind:         kind,
		source:       sourceDefault,
		usage:        usage,
		shortKey:     shortKey,
	}
	fs.flags[name] = flag
	if shortKey != "" {
		fs.shortMap[shortKey] = flag
	}
}

// String defines a string flag with the specified name, default value, and usage string.
// The return value is a pointer to a string variable that stores the value of the flag.
//
// The flag can be set using: --name value or --name=value
//
// Example:
//
//	fs := flashflags.New("myapp")
//	host := fs.String("host", "localhost", "Server host address")
//
//	// Command line: --host 192.168.1.1 or --host=192.168.1.1
//	fs.Parse(os.Args[1:])
//	fmt.Printf("Host: %s\n", *host)
//
// The pointer value is updated immediately when the flag is parsed.
func (fs *FlagSet) String(name, defaultValue, usage string) *string {
	value := defaultValue
	fs.define(name, "", usage, stringKind, defaultValue, defaultValue, &value)
	return &value
}

// StringVar defines a string flag with the specified name, short key, default value, and usage string.
// The short key allows the flag to be specified with a single dash (e.g., -h for --host).
// The return value is a pointer to a string variable that stores the value of the flag.
//
// The flag can be set using: --name value, --name=value, -shortKey value
//
// Example:
//
//	fs := flashflags.New("myapp")
//	host := fs.StringVar("host", "h", "localhost", "Server host address")
//
//	// Command line options:
//	// --host 192.168.1.1
//	// --host=192.168.1.1
//	// -h 192.168.1.1
//	fs.Parse(os.Args[1:])
//	fmt.Printf("Host: %s\n", *host)
//
// If shortKey is empty string, only the long form (--name) is available.
func (fs *FlagSet) StringVar(name, shortKey string, defaultValue, usage string) *string {
	value := defaultValue
	fs.define(name, shortKey, usage, stringKind, defaultValue, defaultValue, &value)
	return &value
}

// Int defines an integer flag with the specified name, default value, and usage string.
// The return value is a pointer to an int variable that stores the value of the flag.
func (fs *FlagSet) Int(name string, defaultValue int, usage string) *int {
	value := defaultValue
	fs.define(name, "", usage, intKind, defaultValue, defaultValue, &value)
	return &value
}

// IntVar defines an integer flag with the specified name, short key, default value, and usage string.
// The short key allows the flag to be specified with a single dash (e.g., -p for --port).
// The return value is a pointer to an int variable that stores the value of the flag.
func (fs *FlagSet) IntVar(name, shortKey string, defaultValue int, usage string) *int {
	value := defaultValue
	fs.define(name, shortKey, usage, intKind, defaultValue, defaultValue, &value)
	return &value
}

// Bool defines a boolean flag with the specified name, default value, and usage string.
// Boolean flags can be set without a value (defaults to true) or with explicit true/false values.
// The return value is a pointer to a bool variable that stores the value of the flag.
//
// Boolean flags support multiple formats:
//
//	--verbose           (sets to true)
//	--verbose=true      (explicit true)
//	--verbose=false     (explicit false)
//	--verbose true      (space-separated true)
//	--verbose false     (space-separated false)
//
// Example:
//
//	fs := flashflags.New("myapp")
//	verbose := fs.Bool("verbose", false, "Enable verbose output")
//	debug := fs.Bool("debug", false, "Enable debug mode")
//
//	// Command line: --verbose --debug=false
//	fs.Parse(os.Args[1:])
//	fmt.Printf("Verbose: %t, Debug: %t\n", *verbose, *debug) // Output: Verbose: true, Debug: false
func (fs *FlagSet) Bool(name string, defaultValue bool, usage string) *bool {
	value := defaultValue
	fs.define(name, "", usage, boolKind, defaultValue, defaultValue, &value)
	return &value
}

// BoolVar defines a boolean flag with the specified name, short key, default value, and usage string.
// The short key allows the flag to be specified with a single dash (e.g., -d for --debug).
// Boolean flags can be set without a value (defaults to true) or with explicit true/false values.
// The return value is a pointer to a bool variable that stores the value of the flag.
func (fs *FlagSet) BoolVar(name, shortKey string, defaultValue bool, usage string) *bool {
	value := defaultValue
	fs.define(name, shortKey, usage, boolKind, defaultValue, defaultValue, &value)
	return &value
}

// Duration defines a duration flag with the specified name, default value, and usage string.
// Duration values are parsed using time.ParseDuration format (e.g., "5s", "1m30s", "2h").
// The return value is a pointer to a time.Duration variable that stores the value of the flag.
//
// Supported duration formats:
//
//	"5s"        (5 seconds)
//	"1m30s"     (1 minute 30 seconds)
//	"2h"        (2 hours)
//	"100ms"     (100 milliseconds)
//	"1h30m45s"  (1 hour 30 minutes 45 seconds)
//
// Example:
//
//	fs := flashflags.New("myapp")
//	timeout := fs.Duration("timeout", 30*time.Second, "Request timeout duration")
//
//	// Command line: --timeout 5m or --timeout=1h30s
//	fs.Parse(os.Args[1:])
//	fmt.Printf("Timeout: %v\n", *timeout)
//
// Returns an error during parsing if the duration format is invalid.
func (fs *FlagSet) Duration(name string, defaultValue time.Duration, usage string) *time.Duration {
	return fs.DurationVar(name, "", defaultValue, usage)
}

// DurationVar defines a duration flag with a short key, the same way StringVar
// does for strings. The flag can be set as --name value, --name=value or
// -shortKey value. An empty shortKey registers the long form only.
func (fs *FlagSet) DurationVar(name, shortKey string, defaultValue time.Duration, usage string) *time.Duration {
	value := defaultValue
	fs.define(name, shortKey, usage, durationKind, defaultValue, defaultValue, &value)
	return &value
}

// Float64 defines a float64 flag with the specified name, default value, and usage string.
// The return value is a pointer to a float64 variable that stores the value of the flag.
func (fs *FlagSet) Float64(name string, defaultValue float64, usage string) *float64 {
	return fs.Float64Var(name, "", defaultValue, usage)
}

// Float64Var defines a float64 flag with a short key, the same way StringVar
// does for strings. The flag can be set as --name value, --name=value or
// -shortKey value. An empty shortKey registers the long form only.
func (fs *FlagSet) Float64Var(name, shortKey string, defaultValue float64, usage string) *float64 {
	value := defaultValue
	fs.define(name, shortKey, usage, float64Kind, defaultValue, defaultValue, &value)
	return &value
}

// StringSlice defines a string slice flag with the specified name, default value, and usage string.
// String slice values are parsed as comma-separated values (e.g., "a,b,c").
// The return value is a pointer to a []string variable that stores the value of the flag.
//
// Input formats:
//
//	"web,api,admin"     →  []string{"web", "api", "admin"}
//	"single"           →  []string{"single"}
//	""                 →  []string{} (empty slice)
//
// Example:
//
//	fs := flashflags.New("myapp")
//	tags := fs.StringSlice("tags", []string{"default"}, "Service tags")
//
//	// Command line: --tags web,api,production
//	fs.Parse(os.Args[1:])
//	for _, tag := range *tags {
//		fmt.Printf("Tag: %s\n", tag)
//	}
//	// Output: Tag: web, Tag: api, Tag: production
//
// Spaces around commas are not trimmed. Use "a, b, c" carefully as it will include spaces.
func (fs *FlagSet) StringSlice(name string, defaultValue []string, usage string) *[]string {
	return fs.StringSliceVar(name, "", defaultValue, usage)
}

// StringSliceVar defines a string slice flag with a short key, the same way
// StringVar does for strings. The flag can be set as --name a,b, --name=a,b or
// -shortKey a,b. An empty shortKey registers the long form only.
//
// As with StringSlice, the default is copied, so the caller's slice is never
// aliased by the value the parser writes into.
func (fs *FlagSet) StringSliceVar(name, shortKey string, defaultValue []string, usage string) *[]string {
	value := make([]string, len(defaultValue))
	copy(value, defaultValue)
	fs.define(name, shortKey, usage, stringSliceKind, value, defaultValue, &value)
	return &value
}

// VisitAll calls fn for each flag in the set.
// This is useful for iterating over all flags, for example to generate help text or validate all flags.
// The order of iteration is not guaranteed.
//
// Example:
//
//	fs.VisitAll(func(flag *Flag) {
//		fmt.Printf("Flag: %s = %v (type: %s)\n", flag.Name(), flag.Value(), flag.Type())
//	})
func (fs *FlagSet) VisitAll(fn func(*Flag)) {
	for _, flag := range fs.flags {
		fn(flag)
	}
}

// Lookup finds a flag by name and returns a pointer to the Flag, or nil if not found.
// This is useful for accessing flag metadata, checking if a flag exists, or getting flag values dynamically.
//
// Example:
//
//	if flag := fs.Lookup("port"); flag != nil {
//		fmt.Printf("Port flag exists, value: %v\n", flag.Value())
//	} else {
//		fmt.Println("Port flag not found")
//	}
func (fs *FlagSet) Lookup(name string) *Flag {
	flag, exists := fs.flags[name]
	if !exists {
		return nil
	}
	return flag
}

// Changed returns whether the specified flag was set during parsing.
// This is useful for conditional logic based on whether a flag was explicitly provided
// via command line, environment variable, or configuration file.
//
// Example:
//
//	if fs.Changed("debug") {
//		fmt.Println("Debug mode was explicitly enabled")
//	}
func (fs *FlagSet) Changed(name string) bool {
	if flag := fs.Lookup(name); flag != nil {
		return flag.Changed()
	}
	return false
}

// Source returns the configuration source that supplied the named flag's current
// value: "cli", "env", "config" or "default". Unknown flags report "default".
//
// This is the recommended way to debug precedence: it tells you not just that a
// flag changed, but which layer won.
//
// Example:
//
//	if fs.Source("port") == "config" {
//		log.Println("port came from the config file")
//	}
func (fs *FlagSet) Source(name string) string {
	if flag := fs.Lookup(name); flag != nil {
		return flag.Source()
	}
	return sourceDefault.String()
}

// Reset resets all flags to their default values and marks them as unchanged.
// This is useful for testing scenarios where you need to clear flag state.
//
// Example:
//
//	fs := flashflags.New("test")
//	port := fs.Int("port", 8080, "Server port")
//
//	fs.Parse([]string{"--port", "3000"})
//	fmt.Println(*port)          // 3000
//	fmt.Println(fs.Changed("port")) // true
//
//	fs.Reset()
//	fmt.Println(*port)          // 8080 (default)
//	fmt.Println(fs.Changed("port")) // false
//
// After Reset(), all flags return to their initial state as if Parse() was never called.
func (fs *FlagSet) Reset() {
	for _, flag := range fs.flags {
		flag.Reset()
	}
}

// ResetFlag resets a specific flag to its default value and marks it as unchanged.
// This is useful for testing or when you need to clear a specific flag's state.
//
// Example:
//
//	fs := flashflags.New("test")
//	port := fs.Int("port", 8080, "Server port")
//	host := fs.String("host", "localhost", "Server host")
//
//	fs.Parse([]string{"--port", "3000", "--host", "example.com"})
//
//	fs.ResetFlag("port")  // Only reset port
//	fmt.Println(*port)    // 8080 (default)
//	fmt.Println(*host)    // "example.com" (unchanged)
//
// Returns an error if the flag name doesn't exist.
func (fs *FlagSet) ResetFlag(name string) error {
	if flag, exists := fs.flags[name]; exists {
		flag.Reset()
		return nil
	}
	return fmt.Errorf("flag --%s not found", name)
}

// GetString gets a flag value as string, with automatic type conversion.
// Returns the string value of the flag, or an empty string if the flag is not found.
//
// Type conversion rules:
//   - String flags: returned as-is
//   - Other types: converted using fmt.Sprintf("%v", value)
//
// Example:
//
//	fs := flashflags.New("myapp")
//	host := fs.String("host", "localhost", "Server host")
//	port := fs.Int("port", 8080, "Server port")
//
//	fs.Parse([]string{"--host", "example.com", "--port", "3000"})
//
//	fmt.Println(fs.GetString("host"))  // "example.com"
//	fmt.Println(fs.GetString("port"))  // "3000" (converted from int)
//	fmt.Println(fs.GetString("missing")) // "" (not found)
//
// This method is safe for concurrent access after Parse() completes.
func (fs *FlagSet) GetString(name string) string {
	if flag, exists := fs.flags[name]; exists {
		if str, ok := flag.value.(string); ok {
			return str
		}
		return fmt.Sprintf("%v", flag.value)
	}
	return ""
}

// GetInt gets a flag value as int.
// Returns the int value of the flag, or 0 if the flag is not found or not an int type.
//
// Only returns non-zero values for flags defined using Int() or IntVar().
// Other flag types will return 0 (use type-specific getters or Lookup() for other types).
//
// Example:
//
//	fs := flashflags.New("myapp")
//	port := fs.IntVar("port", "p", 8080, "Server port")
//	timeout := fs.Duration("timeout", 30*time.Second, "Timeout")
//
//	fs.Parse([]string{"-p", "3000", "--timeout", "45s"})
//
//	fmt.Println(fs.GetInt("port"))     // 3000
//	fmt.Println(fs.GetInt("timeout"))  // 0 (not an int flag)
//	fmt.Println(fs.GetInt("missing"))  // 0 (not found)
//
// This method is safe for concurrent access after Parse() completes.
func (fs *FlagSet) GetInt(name string) int {
	if flag, exists := fs.flags[name]; exists {
		if intVal, ok := flag.value.(int); ok {
			return intVal
		}
	}
	return 0
}

// GetBool gets a flag value as bool.
// Returns the bool value of the flag, or false if the flag is not found or not a bool type.
//
// Only returns true for flags defined using Bool() or BoolVar() that were set to true.
// Other flag types will return false.
//
// Example:
//
//	fs := flashflags.New("myapp")
//	verbose := fs.BoolVar("verbose", "v", false, "Verbose output")
//	debug := fs.Bool("debug", false, "Debug mode")
//
//	fs.Parse([]string{"--verbose", "--debug=false"})
//
//	fmt.Println(fs.GetBool("verbose"))  // true
//	fmt.Println(fs.GetBool("debug"))    // false
//	fmt.Println(fs.GetBool("missing"))  // false (not found)
//
// This method is safe for concurrent access after Parse() completes.
func (fs *FlagSet) GetBool(name string) bool {
	if flag, exists := fs.flags[name]; exists {
		if boolVal, ok := flag.value.(bool); ok {
			return boolVal
		}
	}
	return false
}

// GetDuration gets a flag value as duration.
// Returns the time.Duration value of the flag, or 0 if the flag is not found or not a duration type.
//
// Only returns non-zero values for flags defined using Duration().
// Duration flags accept Go duration format: "1h30m", "45s", "100ms", etc.
//
// Example:
//
//	fs := flashflags.New("myapp")
//	timeout := fs.Duration("timeout", 30*time.Second, "Request timeout")
//	interval := fs.Duration("interval", 5*time.Minute, "Check interval")
//
//	fs.Parse([]string{"--timeout", "45s", "--interval=2m30s"})
//
//	fmt.Println(fs.GetDuration("timeout"))   // 45s
//	fmt.Println(fs.GetDuration("interval"))  // 2m30s
//	fmt.Println(fs.GetDuration("missing"))   // 0s (not found)
//
// This method is safe for concurrent access after Parse() completes.
func (fs *FlagSet) GetDuration(name string) time.Duration {
	if flag, exists := fs.flags[name]; exists {
		if durVal, ok := flag.value.(time.Duration); ok {
			return durVal
		}
	}
	return 0
}

// GetFloat64 gets a flag value as float64.
// Returns the float64 value of the flag, or 0.0 if the flag is not found or not a float64 type.
//
// Only returns non-zero values for flags defined using Float64().
// Accepts standard float formats: "3.14", "-2.5", "1e6", etc.
//
// Example:
//
//	fs := flashflags.New("myapp")
//	rate := fs.Float64("rate", 1.0, "Processing rate")
//	threshold := fs.Float64("threshold", 0.5, "Error threshold")
//
//	fs.Parse([]string{"--rate", "2.5", "--threshold=0.95"})
//
//	fmt.Println(fs.GetFloat64("rate"))       // 2.5
//	fmt.Println(fs.GetFloat64("threshold"))  // 0.95
//	fmt.Println(fs.GetFloat64("missing"))    // 0.0 (not found)
//
// This method is safe for concurrent access after Parse() completes.
func (fs *FlagSet) GetFloat64(name string) float64 {
	if flag, exists := fs.flags[name]; exists {
		if floatVal, ok := flag.value.(float64); ok {
			return floatVal
		}
	}
	return 0.0
}

// GetStringSlice gets a flag value as string slice.
// Returns the []string value of the flag, or an empty slice if the flag is not found or not a string slice type.
//
// Only returns non-empty values for flags defined using StringSlice().
// String slices are parsed from comma-separated values.
//
// Example:
//
//	fs := flashflags.New("myapp")
//	tags := fs.StringSlice("tags", []string{}, "Service tags")
//	hosts := fs.StringSlice("hosts", []string{"localhost"}, "Host list")
//
//	fs.Parse([]string{"--tags", "web,api,prod", "--hosts=srv1,srv2"})
//
//	fmt.Println(fs.GetStringSlice("tags"))    // ["web", "api", "prod"]
//	fmt.Println(fs.GetStringSlice("hosts"))   // ["srv1", "srv2"]
//	fmt.Println(fs.GetStringSlice("missing")) // []
//
// This method is safe for concurrent access after Parse() completes.
func (fs *FlagSet) GetStringSlice(name string) []string {
	if flag, exists := fs.flags[name]; exists {
		if slice, ok := flag.value.([]string); ok {
			return slice
		}
	}
	return []string{}
}
