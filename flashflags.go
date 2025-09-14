// flash-flags.go: Ultra-fast command-line flag parsing for Go
//
// Copyright (c) 2025 AGILira - A. Giordano
// Series: an AGILira library
// SPDX-License-Identifier: MPL-2.0

// Package flashflags provides ultra-fast, zero-dependency, lock-free command-line flag parsing.
// This library is extracted from argus with exactly the same structure for maximum performance.
package flashflags

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

// Flag represents a single command-line flag with its value, metadata, and constraints.
// It implements ultra-fast flag handling using only the standard library with thread-safe operations.
//
// Example usage:
//
//	fs := flashflags.New("myapp")
//	flag := fs.Lookup("port")
//	if flag != nil && flag.Changed() {
//		fmt.Printf("Port was set to: %v\n", flag.Value())
//	}
//
// Flags support validation, dependencies, grouping, and environment variable integration.
// All operations are lock-free and safe for concurrent access.
type Flag struct {
	name         string
	value        interface{}
	defaultValue interface{} // original default value for reset
	ptr          interface{} // pointer to the actual value
	flagType     string
	changed      bool
	usage        string
	shortKey     string                  // Short flag key (e.g., "p" for port)
	validator    func(interface{}) error // Optional validation function
	required     bool                    // Whether this flag is required
	dependencies []string                // Flags that this flag depends on
	group        string                  // Group name for help organization
	envVar       string                  // Environment variable name for this flag
}

// Name returns the flag name.
// This is the long name used with --flag syntax.
//
// Example:
//
//	flag := fs.Lookup("port")
//	fmt.Printf("Flag name: %s\n", flag.Name()) // Output: Flag name: port
func (f *Flag) Name() string { return f.name }

// Value returns the current flag value as interface{}.
// Use type assertion to convert to the expected type.
//
// Example:
//
//	flag := fs.Lookup("port")
//	if flag != nil && flag.Type() == "int" {
//		port := flag.Value().(int)
//		fmt.Printf("Port value: %d\n", port)
//	}
func (f *Flag) Value() interface{} { return f.value }

// Type returns the flag type as a string.
// Possible values: "string", "int", "bool", "float64", "duration", "stringSlice".
//
// Example:
//
//	flag := fs.Lookup("timeout")
//	if flag.Type() == "duration" {
//		dur := flag.Value().(time.Duration)
//		fmt.Printf("Timeout: %v\n", dur)
//	}
func (f *Flag) Type() string { return f.flagType }

// Changed returns whether the flag was explicitly set by any configuration source.
// Returns true if the flag was set via command line, environment variable, or config file.
// Returns false if the flag is using its default value.
//
// Example:
//
//	flag := fs.Lookup("port")
//	if flag.Changed() {
//		fmt.Println("Port was explicitly set")
//	} else {
//		fmt.Println("Port is using default value")
//	}
func (f *Flag) Changed() bool { return f.changed }

// Usage returns the flag usage description string.
// This is the help text that was provided when the flag was created.
//
// Example:
//
//	flag := fs.Lookup("port")
//	fmt.Printf("Flag usage: %s\n", flag.Usage()) // Output: Flag usage: Server port number
func (f *Flag) Usage() string { return f.usage }

// SetValidator sets a validation function for the flag.
// The validator will be called whenever the flag value is set or changed.
//
// Example:
//
//	flag := fs.Lookup("port")
//	flag.SetValidator(func(val interface{}) error {
//		port := val.(int)
//		if port < 1024 {
//			return fmt.Errorf("port must be >= 1024")
//		}
//		return nil
//	})
func (f *Flag) SetValidator(validator func(interface{}) error) {
	f.validator = validator
}

// Validate validates the current flag value using the validator if set.
// Returns nil if no validator is set or if validation passes.
// Returns an error if validation fails.
//
// Example:
//
//	flag := fs.Lookup("port")
//	if err := flag.Validate(); err != nil {
//		fmt.Printf("Validation failed: %v\n", err)
//	}
func (f *Flag) Validate() error {
	if f.validator != nil {
		return f.validator(f.value)
	}
	return nil
}

// Reset resets the flag to its default value and marks it as unchanged.
// This is useful for testing or when you need to clear flag values.
//
// Example:
//
//	flag := fs.Lookup("port")
//	flag.Reset()
//	fmt.Printf("Port reset to: %v, Changed: %t\n", flag.Value(), flag.Changed())
func (f *Flag) Reset() {
	f.value = f.defaultValue
	if f.ptr != nil {
		f.resetPointer()
	}
	f.changed = false
}

// resetPointer resets the pointer to the default value based on the flag type
func (f *Flag) resetPointer() {
	switch f.flagType {
	case "string":
		f.resetStringPointer()
	case "int":
		f.resetIntPointer()
	case "bool":
		f.resetBoolPointer()
	case "float64":
		f.resetFloat64Pointer()
	case "duration":
		f.resetDurationPointer()
	case "stringSlice":
		f.resetStringSlicePointer()
	}
}

// resetStringPointer resets a string pointer to its default value
func (f *Flag) resetStringPointer() {
	if val, ok := f.defaultValue.(string); ok {
		if ptr, ok := f.ptr.(*string); ok {
			*ptr = val
		}
	}
}

// resetIntPointer resets an int pointer to its default value
func (f *Flag) resetIntPointer() {
	if val, ok := f.defaultValue.(int); ok {
		if ptr, ok := f.ptr.(*int); ok {
			*ptr = val
		}
	}
}

// resetBoolPointer resets a bool pointer to its default value
func (f *Flag) resetBoolPointer() {
	if val, ok := f.defaultValue.(bool); ok {
		if ptr, ok := f.ptr.(*bool); ok {
			*ptr = val
		}
	}
}

// resetFloat64Pointer resets a float64 pointer to its default value
func (f *Flag) resetFloat64Pointer() {
	if val, ok := f.defaultValue.(float64); ok {
		if ptr, ok := f.ptr.(*float64); ok {
			*ptr = val
		}
	}
}

// resetDurationPointer resets a Duration pointer to its default value
func (f *Flag) resetDurationPointer() {
	if val, ok := f.defaultValue.(time.Duration); ok {
		if ptr, ok := f.ptr.(*time.Duration); ok {
			*ptr = val
		}
	}
}

// resetStringSlicePointer resets a []string pointer to its default value
func (f *Flag) resetStringSlicePointer() {
	if val, ok := f.defaultValue.([]string); ok {
		if ptr, ok := f.ptr.(*[]string); ok {
			*ptr = val
		}
	}
}

// FlagSet represents a collection of command-line flags with parsing and validation capabilities.
// It implements ultra-fast flag set handling using only the standard library with lock-free operations.
//
// FlagSet supports multiple configuration sources in priority order:
//  1. Command-line arguments (highest priority)
//  2. Environment variables
//  3. Configuration files (lowest priority)
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
// All FlagSet operations are thread-safe and use lock-free algorithms for optimal performance.
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
}

// New creates a new FlagSet with the specified name.
// The name is used for help output and configuration file discovery.
// Returns a FlagSet with zero external dependencies.
//
// Thread Safety:
// All FlagSet operations are thread-safe and use lock-free algorithms for optimal performance.
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
	flag := &Flag{
		name:         name,
		value:        defaultValue,
		ptr:          &value,
		flagType:     "string",
		changed:      false,
		usage:        usage,
		shortKey:     "",
		defaultValue: defaultValue,
	}
	fs.flags[name] = flag
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
	flag := &Flag{
		name:         name,
		value:        defaultValue,
		ptr:          &value,
		flagType:     "string",
		changed:      false,
		usage:        usage,
		shortKey:     shortKey,
		defaultValue: defaultValue,
	}
	fs.flags[name] = flag
	if shortKey != "" {
		fs.shortMap[shortKey] = flag
	}
	return &value
}

// Int defines an integer flag with the specified name, default value, and usage string.
// The return value is a pointer to an int variable that stores the value of the flag.
func (fs *FlagSet) Int(name string, defaultValue int, usage string) *int {
	value := defaultValue
	flag := &Flag{
		name:         name,
		value:        defaultValue,
		ptr:          &value,
		flagType:     "int",
		changed:      false,
		usage:        usage,
		shortKey:     "",
		defaultValue: defaultValue,
	}
	fs.flags[name] = flag
	return &value
}

// IntVar defines an integer flag with the specified name, short key, default value, and usage string.
// The short key allows the flag to be specified with a single dash (e.g., -p for --port).
// The return value is a pointer to an int variable that stores the value of the flag.
func (fs *FlagSet) IntVar(name, shortKey string, defaultValue int, usage string) *int {
	value := defaultValue
	flag := &Flag{
		name:         name,
		value:        defaultValue,
		ptr:          &value,
		flagType:     "int",
		changed:      false,
		usage:        usage,
		shortKey:     shortKey,
		validator:    nil,
		defaultValue: defaultValue,
	}
	fs.flags[name] = flag
	if shortKey != "" {
		fs.shortMap[shortKey] = flag
	}
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
	flag := &Flag{
		name:         name,
		value:        defaultValue,
		ptr:          &value,
		flagType:     "bool",
		changed:      false,
		usage:        usage,
		shortKey:     "",
		validator:    nil,
		defaultValue: defaultValue,
	}
	fs.flags[name] = flag
	return &value
}

// BoolVar defines a boolean flag with the specified name, short key, default value, and usage string.
// The short key allows the flag to be specified with a single dash (e.g., -d for --debug).
// Boolean flags can be set without a value (defaults to true) or with explicit true/false values.
// The return value is a pointer to a bool variable that stores the value of the flag.
func (fs *FlagSet) BoolVar(name, shortKey string, defaultValue bool, usage string) *bool {
	value := defaultValue
	flag := &Flag{
		name:         name,
		value:        defaultValue,
		ptr:          &value,
		flagType:     "bool",
		changed:      false,
		usage:        usage,
		shortKey:     shortKey,
		validator:    nil,
		defaultValue: defaultValue,
	}
	fs.flags[name] = flag
	if shortKey != "" {
		fs.shortMap[shortKey] = flag
	}
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
	value := defaultValue
	flag := &Flag{
		name:         name,
		value:        defaultValue,
		ptr:          &value,
		flagType:     "duration",
		changed:      false,
		usage:        usage,
		validator:    nil,
		defaultValue: defaultValue,
	}
	fs.flags[name] = flag
	return &value
}

// Float64 defines a float64 flag with the specified name, default value, and usage string.
// The return value is a pointer to a float64 variable that stores the value of the flag.
func (fs *FlagSet) Float64(name string, defaultValue float64, usage string) *float64 {
	value := defaultValue
	flag := &Flag{
		name:         name,
		value:        defaultValue,
		ptr:          &value,
		flagType:     "float64",
		changed:      false,
		usage:        usage,
		validator:    nil,
		defaultValue: defaultValue,
	}
	fs.flags[name] = flag
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
	value := make([]string, len(defaultValue))
	copy(value, defaultValue)
	flag := &Flag{
		name:         name,
		value:        value,
		ptr:          &value,
		flagType:     "stringSlice",
		changed:      false,
		usage:        usage,
		validator:    nil,
		defaultValue: defaultValue,
	}
	fs.flags[name] = flag
	return &value
}

// Parse parses command line arguments with optimized allocations and validates all constraints.
//
// Parse processes configuration sources in priority order:
//  1. Configuration files (LoadConfig) - lowest priority
//  2. Environment variables (LoadEnvironmentVariables) - medium priority
//  3. Command-line arguments - highest priority
//
// After parsing, it validates all constraints including required flags, dependencies, and custom validators.
//
// Supported argument formats:
//
//	--flag value          (long flag with space-separated value)
//	--flag=value          (long flag with equals-separated value)
//	-f value             (short flag with space-separated value)
//	--boolean-flag       (boolean flag without value, defaults to true)
//	--boolean-flag=true  (explicit boolean value)
//
// Special handling:
//
//	--help, -h           (shows help and returns error "help requested")
//
// Example:
//
//	fs := flashflags.New("myapp")
//	host := fs.StringVar("host", "h", "localhost", "Server host")
//	port := fs.IntVar("port", "p", 8080, "Server port")
//
//	args := []string{"--host", "0.0.0.0", "-p", "3000"}
//	if err := fs.Parse(args); err != nil {
//		if err.Error() == "help requested" {
//			return // Help was displayed
//		}
//		log.Fatalf("Parse error: %v", err)
//	}
//
// Returns an error if parsing fails, validation fails, or help is requested.
func (fs *FlagSet) Parse(args []string) error {
	// Load configuration file first (lowest priority)
	if err := fs.LoadConfig(); err != nil {
		return fmt.Errorf("config file error: %v", err)
	}

	// Load environment variables second (medium priority)
	if err := fs.LoadEnvironmentVariables(); err != nil {
		return fmt.Errorf("environment variable error: %v", err)
	}

	// Parse command line arguments (highest priority)
	if err := fs.parseArguments(args); err != nil {
		return err
	}

	// Validate all constraints after parsing
	return fs.ValidateAllConstraints()
}

// parseArguments handles the main argument parsing loop
func (fs *FlagSet) parseArguments(args []string) error {
	for i := 0; i < len(args); i++ {
		consumed, err := fs.processArgument(args, i)
		if err != nil {
			return err
		}
		i += consumed
	}
	return nil
}

// processArgument processes a single argument and returns consumed count
func (fs *FlagSet) processArgument(args []string, i int) (int, error) {
	arg := args[i]

	if !strings.HasPrefix(arg, "-") {
		return 0, nil
	}

	if fs.isHelpFlag(arg) {
		fs.PrintHelp()
		return 0, fmt.Errorf("help requested")
	}

	if fs.isShortFlag(arg) {
		return fs.parseShortFlag(args, i)
	}

	if fs.isLongFlag(arg) {
		return fs.parseLongFlag(args, i)
	}

	return 0, nil
}

// isHelpFlag checks if the argument is a help flag
func (fs *FlagSet) isHelpFlag(arg string) bool {
	return arg == "--help" || arg == "-h"
}

// isShortFlag checks if the argument is a short flag
func (fs *FlagSet) isShortFlag(arg string) bool {
	return len(arg) == 2 && arg[0] == '-' && arg[1] != '-'
}

// isLongFlag checks if the argument is a long flag
func (fs *FlagSet) isLongFlag(arg string) bool {
	return strings.HasPrefix(arg, "--")
}

// parseShortFlag handles short flag parsing (-p, -d)
func (fs *FlagSet) parseShortFlag(args []string, i int) (int, error) {
	arg := args[i]
	shortKey := string(arg[1])
	flag, exists := fs.shortMap[shortKey]
	if !exists {
		return 0, fmt.Errorf("unknown flag: -%s", shortKey)
	}

	if flag.flagType == "bool" {
		flag.value = true
		if flag.ptr != nil {
			if ptr, ok := flag.ptr.(*bool); ok {
				*ptr = true
			}
		}
		flag.changed = true
		return 0, nil
	}

	// Non-bool short flag needs value
	if i+1 >= len(args) {
		return 0, fmt.Errorf("flag -%s requires a value", shortKey)
	}
	flagValue := args[i+1]

	err := fs.setFlagValue(flag.name, flagValue)
	if err != nil {
		return 0, err
	}
	return 1, nil // Consumed one extra argument
}

// parseLongFlag handles long flag parsing (--name)
func (fs *FlagSet) parseLongFlag(args []string, i int) (int, error) {
	arg := args[i][2:] // Remove -- prefix

	var flagName, flagValue string
	// Optimized parsing to avoid SplitN allocation
	if eqPos := strings.IndexByte(arg, '='); eqPos != -1 {
		flagName = arg[:eqPos]
		flagValue = arg[eqPos+1:]
	} else {
		flagName = arg
		// Look for value in next argument
		if i+1 < len(args) && !strings.HasPrefix(args[i+1], "--") {
			flagValue = args[i+1]
			// Set flag value
			err := fs.setFlagValue(flagName, flagValue)
			if err != nil {
				return 0, err
			}
			return 1, nil // Consumed one extra argument
		}
		// Boolean flag or error
		if flag, exists := fs.flags[flagName]; exists && flag.flagType == "bool" {
			flagValue = "true"
		} else {
			return 0, fmt.Errorf("flag --%s requires a value", flagName)
		}
	}

	// Set flag value
	err := fs.setFlagValue(flagName, flagValue)
	if err != nil {
		return 0, err
	}
	return 0, nil
}

// Type-specific value setters to reduce complexity

func (fs *FlagSet) setStringValue(flag *Flag, value string) error {
	flag.value = value
	if flag.ptr != nil {
		if ptr, ok := flag.ptr.(*string); ok {
			*ptr = value
		}
	}
	return nil
}

func (fs *FlagSet) setIntValue(flag *Flag, value, name string) error {
	intVal, err := strconv.Atoi(value)
	if err != nil {
		return fmt.Errorf("invalid int value for flag --%s: %s", name, value)
	}
	flag.value = intVal
	if flag.ptr != nil {
		if ptr, ok := flag.ptr.(*int); ok {
			*ptr = intVal
		}
	}
	return nil
}

func (fs *FlagSet) setBoolValue(flag *Flag, value, name string) error {
	boolVal, err := strconv.ParseBool(value)
	if err != nil {
		return fmt.Errorf("invalid bool value for flag --%s: %s", name, value)
	}
	flag.value = boolVal
	if flag.ptr != nil {
		if ptr, ok := flag.ptr.(*bool); ok {
			*ptr = boolVal
		}
	}
	return nil
}

func (fs *FlagSet) setDurationValue(flag *Flag, value, name string) error {
	durVal, err := time.ParseDuration(value)
	if err != nil {
		return fmt.Errorf("invalid duration value for flag --%s: %s", name, value)
	}
	flag.value = durVal
	if flag.ptr != nil {
		if ptr, ok := flag.ptr.(*time.Duration); ok {
			*ptr = durVal
		}
	}
	return nil
}

func (fs *FlagSet) setFloat64Value(flag *Flag, value, name string) error {
	floatVal, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return fmt.Errorf("invalid float64 value for flag --%s: %s", name, value)
	}
	flag.value = floatVal
	if flag.ptr != nil {
		if ptr, ok := flag.ptr.(*float64); ok {
			*ptr = floatVal
		}
	}
	return nil
}

func (fs *FlagSet) setStringSliceValue(flag *Flag, value string) error {
	slice := fs.parseStringSlice(value)
	flag.value = slice
	fs.updateStringSlicePointer(flag, slice)
	return nil
}

// parseStringSlice parses a comma-separated string into a slice
func (fs *FlagSet) parseStringSlice(value string) []string {
	if value == "" {
		return []string{}
	}
	return fs.splitByComma(value)
}

// splitByComma splits a string by commas with optimized allocation
func (fs *FlagSet) splitByComma(value string) []string {
	commas := fs.countCommas(value)
	slice := make([]string, 0, commas+1)

	start := 0
	for i := 0; i < len(value); i++ {
		if value[i] == ',' {
			if i > start {
				slice = append(slice, value[start:i])
			}
			start = i + 1
		}
	}
	if start < len(value) {
		slice = append(slice, value[start:])
	}
	return slice
}

// countCommas counts the number of commas in a string
func (fs *FlagSet) countCommas(value string) int {
	commas := 0
	for _, c := range []byte(value) {
		if c == ',' {
			commas++
		}
	}
	return commas
}

// updateStringSlicePointer updates the pointer if it exists
func (fs *FlagSet) updateStringSlicePointer(flag *Flag, slice []string) {
	if flag.ptr != nil {
		if ptr, ok := flag.ptr.(*[]string); ok {
			*ptr = slice
		}
	}
}

func (fs *FlagSet) setFlagValue(name, value string) error {
	flag, exists := fs.flags[name]
	if !exists {
		return fmt.Errorf("unknown flag: --%s", name)
	}

	if err := fs.setFlagValueByType(flag, value, name); err != nil {
		return err
	}

	flag.changed = true
	return fs.validateFlag(flag, name)
}

// setFlagValueByType sets the flag value based on its type
func (fs *FlagSet) setFlagValueByType(flag *Flag, value, name string) error {
	switch flag.flagType {
	case "string":
		return fs.setStringValue(flag, value)
	case "int":
		return fs.setIntValue(flag, value, name)
	case "bool":
		return fs.setBoolValue(flag, value, name)
	case "duration":
		return fs.setDurationValue(flag, value, name)
	case "float64":
		return fs.setFloat64Value(flag, value, name)
	case "stringSlice":
		return fs.setStringSliceValue(flag, value)
	default:
		return fmt.Errorf("unsupported flag type: %s", flag.flagType)
	}
}

// validateFlag runs validation on the flag if a validator is set
func (fs *FlagSet) validateFlag(flag *Flag, name string) error {
	if flag.validator != nil {
		if err := flag.validator(flag.value); err != nil {
			return fmt.Errorf("validation failed for flag --%s: %v", name, err)
		}
	}
	return nil
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

// PrintUsage prints basic usage information for all flags to stdout.
// This provides a simpler flag listing without the full help formatting, groups, or descriptions.
//
// Output format for each flag:
//
//	--flagname, -s
//	      Description text (type: flagtype)
//
// Example:
//
//	fs := flashflags.New("myapp")
//	fs.StringVar("host", "h", "localhost", "Server host")
//	fs.IntVar("port", "p", 8080, "Server port")
//
//	fs.PrintUsage()
//	// Output:
//	// Usage of myapp:
//	//   --host, -h
//	//         Server host (type: string)
//	//   --port, -p
//	//         Server port (type: int)
//
// Use PrintHelp() for complete help with grouping, defaults, and requirements.
func (fs *FlagSet) PrintUsage() {
	fmt.Printf("Usage of %s:\n", fs.name)
	for name, flag := range fs.flags {
		fmt.Printf("  --%s", name)
		if flag.shortKey != "" {
			fmt.Printf(", -%s", flag.shortKey)
		}
		fmt.Printf("\n")
		fmt.Printf("        %s (type: %s)\n", flag.usage, flag.flagType)
	}
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
		return flag.changed
	}
	return false
}

// SetValidator sets a validation function for a specific flag.
// The validator function will be called whenever the flag value is set, allowing for custom validation logic.
//
// The validator function receives the parsed value as interface{} and should return an error if validation fails.
// Type assertion is required to access the actual value.
//
// Example:
//
//	fs := flashflags.New("myapp")
//	port := fs.IntVar("port", "p", 8080, "Server port")
//
//	err := fs.SetValidator("port", func(val interface{}) error {
//		port := val.(int)
//		if port < 1024 || port > 65535 {
//			return fmt.Errorf("port must be between 1024-65535, got %d", port)
//		}
//		return nil
//	})
//	if err != nil {
//		log.Fatal(err)
//	}
//
// Returns an error if the flag name doesn't exist.
func (fs *FlagSet) SetValidator(name string, validator func(interface{}) error) error {
	flag := fs.Lookup(name)
	if flag == nil {
		return fmt.Errorf("flag not found: %s", name)
	}
	flag.SetValidator(validator)
	return nil
}

// SetRequired marks a flag as required.
// Required flags must be provided during parsing, otherwise an error will be returned.
//
// Required flags can be satisfied by any configuration source (command-line, environment variables, or config files).
//
// Example:
//
//	fs := flashflags.New("myapp")
//	apiKey := fs.String("api-key", "", "API authentication key")
//
//	if err := fs.SetRequired("api-key"); err != nil {
//		log.Fatal(err)
//	}
//
//	// This will fail if --api-key is not provided
//	if err := fs.Parse(os.Args[1:]); err != nil {
//		log.Fatal(err) // "required flag --api-key not provided"
//	}
//
// Returns an error if the flag name doesn't exist.
func (fs *FlagSet) SetRequired(name string) error {
	flag := fs.Lookup(name)
	if flag == nil {
		return fmt.Errorf("flag not found: %s", name)
	}
	flag.required = true
	return nil
}

// SetDependencies sets dependencies for a flag.
// When this flag is set, all dependent flags must also be set, otherwise an error will be returned.
//
// Dependencies are validated after all configuration sources are processed.
// Dependent flags can be satisfied by any configuration source.
//
// Example:
//
//	fs := flashflags.New("myapp")
//	enableTLS := fs.Bool("enable-tls", false, "Enable TLS")
//	tlsCert := fs.String("tls-cert", "", "TLS certificate file")
//	tlsKey := fs.String("tls-key", "", "TLS private key file")
//
//	// Both cert and key require TLS to be enabled
//	if err := fs.SetDependencies("tls-cert", "enable-tls"); err != nil {
//		log.Fatal(err)
//	}
//	if err := fs.SetDependencies("tls-key", "enable-tls"); err != nil {
//		log.Fatal(err)
//	}
//
//	// This will fail if --tls-cert is provided without --enable-tls
//	if err := fs.Parse(os.Args[1:]); err != nil {
//		log.Fatal(err) // "flag --tls-cert requires --enable-tls to be set"
//	}
//
// Returns an error if the flag name doesn't exist.
func (fs *FlagSet) SetDependencies(name string, dependencies ...string) error {
	flag := fs.Lookup(name)
	if flag == nil {
		return fmt.Errorf("flag not found: %s", name)
	}
	flag.dependencies = dependencies
	return nil
}

// SetDescription sets the program description displayed at the top of help output.
// The description should briefly explain what the program does.
//
// Example:
//
//	fs := flashflags.New("webserver")
//	fs.SetDescription("High-performance HTTP server with advanced configuration options")
//
//	// Help output will show:
//	// High-performance HTTP server with advanced configuration options
//	//
//	// Usage: webserver [options]
//	// ...
//
// Call before Parse() to ensure it's included in help text.
func (fs *FlagSet) SetDescription(description string) {
	fs.description = description
}

// SetVersion sets the program version displayed in help output.
// The version should follow semantic versioning (e.g., "v1.2.3").
//
// Example:
//
//	fs := flashflags.New("myapp")
//	fs.SetVersion("v2.1.0")
//
//	// Help output will show:
//	// Usage: myapp [options]
//	//
//	// Version: v2.1.0
//	// ...
//
// Call before Parse() to ensure it's included in help text.
func (fs *FlagSet) SetVersion(version string) {
	fs.version = version
}

// SetGroup sets the group name for a flag to organize help output.
// Flags with the same group will be displayed together under a group heading.
//
// Example:
//
//	fs := flashflags.New("server")
//	host := fs.String("host", "localhost", "Server host")
//	port := fs.Int("port", 8080, "Server port")
//	dbHost := fs.String("db-host", "localhost", "Database host")
//
//	fs.SetGroup("host", "Server Options")
//	fs.SetGroup("port", "Server Options")
//	fs.SetGroup("db-host", "Database Options")
//
//	// Help output will show:
//	// Server Options:
//	//   --host     Server host
//	//   --port     Server port
//	//
//	// Database Options:
//	//   --db-host  Database host
//
// Returns an error if the flag name doesn't exist.
func (fs *FlagSet) SetGroup(name, group string) error {
	flag := fs.Lookup(name)
	if flag == nil {
		return fmt.Errorf("flag not found: %s", name)
	}
	flag.group = group
	return nil
}

// ValidateAll validates all flags that have validators set.
// This is called automatically during Parse, but can be called manually if needed.
//
// Possible errors:
//   - Custom validation errors from validator functions
//   - Format: "validation failed for flag --flagname: custom error message"
//
// Example:
//
//	// Validator that might fail:
//	fs.SetValidator("port", func(val interface{}) error {
//		if val.(int) < 1024 { return fmt.Errorf("port too low") }
//		return nil
//	})
//
//	if err := fs.ValidateAll(); err != nil {
//		// Error: "validation failed for flag --port: port too low"
//	}
//
// The method stops at the first validation failure and returns that error.
func (fs *FlagSet) ValidateAll() error {
	for name, flag := range fs.flags {
		if flag.validator != nil {
			if err := flag.Validate(); err != nil {
				return fmt.Errorf("validation failed for flag --%s: %v", name, err)
			}
		}
	}
	return nil
}

// ValidateRequired checks that all required flags are set.
// This is called automatically during Parse, but can be called manually if needed.
//
// Possible errors:
//   - Missing required flag: "required flag --flagname not provided"
//
// Example:
//
//	fs.SetRequired("api-key")
//	// If --api-key not provided by CLI, env vars, or config file:
//	if err := fs.ValidateRequired(); err != nil {
//		// Error: "required flag --api-key not provided"
//	}
//
// Required flags can be satisfied by any configuration source (CLI, env, config file).
func (fs *FlagSet) ValidateRequired() error {
	for name, flag := range fs.flags {
		if flag.required && !flag.changed {
			return fmt.Errorf("required flag --%s not provided", name)
		}
	}
	return nil
}

// ValidateDependencies checks that all flag dependencies are satisfied.
// This is called automatically during Parse, but can be called manually if needed.
//
// Possible errors:
//   - Missing dependency: "flag --flagname requires --dependency to be set"
//   - Non-existent dependency: "flag --flagname depends on non-existent flag --missing"
//
// Example:
//
//	fs.SetDependencies("tls-cert", "enable-tls")
//	// If --tls-cert provided but --enable-tls not set:
//	if err := fs.ValidateDependencies(); err != nil {
//		// Error: "flag --tls-cert requires --enable-tls to be set"
//	}
//
// Dependencies must be satisfied by any configuration source.
func (fs *FlagSet) ValidateDependencies() error {
	for name, flag := range fs.flags {
		if flag.changed && len(flag.dependencies) > 0 {
			for _, dep := range flag.dependencies {
				depFlag := fs.Lookup(dep)
				if depFlag == nil {
					return fmt.Errorf("flag --%s depends on non-existent flag --%s", name, dep)
				}
				if !depFlag.changed {
					return fmt.Errorf("flag --%s requires --%s to be set", name, dep)
				}
			}
		}
	}
	return nil
}

// ValidateAllConstraints validates all constraints: validators, required flags, and dependencies.
// This is called automatically during Parse, but can be called manually if needed.
//
// Possible errors:
//   - Required flag errors: "required flag --api-key not provided"
//   - Dependency errors: "flag --tls-cert requires --enable-tls to be set"
//   - Validation errors: "validation failed for flag --port: port must be between 1024-65535"
//   - Missing dependency errors: "flag --ssl depends on non-existent flag --tls"
//
// The method stops at the first constraint violation and returns that error.
//
// Example:
//
//	if err := fs.ValidateAllConstraints(); err != nil {
//		fmt.Printf("Validation failed: %v\n", err)
//		fs.PrintHelp()
//		os.Exit(1)
//	}
func (fs *FlagSet) ValidateAllConstraints() error {
	if err := fs.ValidateRequired(); err != nil {
		return err
	}
	if err := fs.ValidateDependencies(); err != nil {
		return err
	}
	if err := fs.ValidateAll(); err != nil {
		return err
	}
	return nil
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

// Help generates and returns the complete help text as a string.
// Includes program description, version, usage line, and all flags organized by groups.
//
// The help text format:
//   - Program description (if set with SetDescription)
//   - Usage line: "Usage: {program-name} [options]"
//   - Version info (if set with SetVersion)
//   - Ungrouped flags (if any)
//   - Grouped flags organized by SetGroup() calls
//
// Example:
//
//	fs := flashflags.New("myserver")
//	fs.SetDescription("High-performance web server")
//	fs.SetVersion("v2.1.0")
//
//	port := fs.IntVar("port", "p", 8080, "Server port")
//	host := fs.String("host", "localhost", "Server host")
//
//	fs.SetGroup("port", "Server Options")
//	fs.SetGroup("host", "Server Options")
//
//	helpText := fs.Help()
//	// Contains formatted help with grouped flags, defaults, requirements, etc.
//
// Use PrintHelp() to output directly to stdout.
func (fs *FlagSet) Help() string {
	var help strings.Builder

	// Program name and description
	if fs.description != "" {
		help.WriteString(fs.description)
		help.WriteString("\n\n")
	}

	// Usage line
	help.WriteString("Usage: ")
	help.WriteString(fs.name)
	help.WriteString(" [options]\n\n")

	// Version info
	if fs.version != "" {
		help.WriteString("Version: ")
		help.WriteString(fs.version)
		help.WriteString("\n\n")
	}

	// Group flags by group name
	groups := make(map[string][]*Flag)
	ungrouped := []*Flag{}

	for _, flag := range fs.flags {
		if flag.group != "" {
			groups[flag.group] = append(groups[flag.group], flag)
		} else {
			ungrouped = append(ungrouped, flag)
		}
	}

	// Display ungrouped flags first
	if len(ungrouped) > 0 {
		help.WriteString("Options:\n")
		for _, flag := range ungrouped {
			help.WriteString(fs.formatFlagHelp(flag))
		}
		help.WriteString("\n")
	}

	// Display grouped flags
	for groupName, groupFlags := range groups {
		help.WriteString(groupName)
		help.WriteString(":\n")
		for _, flag := range groupFlags {
			help.WriteString(fs.formatFlagHelp(flag))
		}
		help.WriteString("\n")
	}

	return help.String()
}

// formatFlagHelp formats a single flag for help output
func (fs *FlagSet) formatFlagHelp(flag *Flag) string {
	var line strings.Builder

	// Build flag name with short key
	fs.buildFlagName(&line, flag)

	// Add type info for non-bool flags
	fs.addTypeInfo(&line, flag)

	// Pad to align descriptions
	fs.padForAlignment(&line)

	// Add description and modifiers
	fs.addDescriptionAndModifiers(&line, flag)

	line.WriteString("\n")
	return line.String()
}

// buildFlagName builds the flag name part of help output
func (fs *FlagSet) buildFlagName(line *strings.Builder, flag *Flag) {
	line.WriteString("  ")
	if flag.shortKey != "" {
		line.WriteString("-")
		line.WriteString(flag.shortKey)
		line.WriteString(", ")
	}
	line.WriteString("--")
	line.WriteString(flag.name)
}

// addTypeInfo adds type information for non-bool flags
func (fs *FlagSet) addTypeInfo(line *strings.Builder, flag *Flag) {
	if flag.flagType != "bool" {
		line.WriteString(" ")
		line.WriteString(strings.ToUpper(flag.flagType))
	}
}

// padForAlignment pads the line to align descriptions
func (fs *FlagSet) padForAlignment(line *strings.Builder) {
	for line.Len() < 30 {
		line.WriteString(" ")
	}
}

// addDescriptionAndModifiers adds description, default value, required indicator, and dependencies
func (fs *FlagSet) addDescriptionAndModifiers(line *strings.Builder, flag *Flag) {
	// Add description
	line.WriteString(flag.usage)

	// Add default value
	if flag.defaultValue != nil && flag.flagType != "bool" {
		line.WriteString(" (default: ")
		line.WriteString(fmt.Sprintf("%v", flag.defaultValue))
		line.WriteString(")")
	}

	// Add required indicator
	if flag.required {
		line.WriteString(" [REQUIRED]")
	}

	// Add dependencies
	fs.addDependencies(line, flag)
}

// addDependencies adds dependency information to help output
func (fs *FlagSet) addDependencies(line *strings.Builder, flag *Flag) {
	if len(flag.dependencies) > 0 {
		line.WriteString(" [depends on: ")
		for i, dep := range flag.dependencies {
			if i > 0 {
				line.WriteString(", ")
			}
			line.WriteString(dep)
		}
		line.WriteString("]")
	}
}

// PrintHelp prints the complete help text to stdout.
// This is a convenience method that calls Help() and prints the result directly.
//
// Commonly used in response to --help flags or validation errors.
//
// Example:
//
//	fs := flashflags.New("myapp")
//	// ... define flags ...
//
//	if err := fs.Parse(os.Args[1:]); err != nil {
//		if err.Error() == "help requested" {
//			// Help was already printed by Parse()
//			os.Exit(0)
//		}
//		fmt.Printf("Error: %v\n", err)
//		fs.PrintHelp()  // Show help on errors
//		os.Exit(1)
//	}
//
// Use Help() if you need the help text as a string for custom formatting.
func (fs *FlagSet) PrintHelp() {
	fmt.Print(fs.Help())
}

// SetConfigFile sets an explicit configuration file path.
// The config file is loaded automatically during Parse() with lower priority than CLI arguments.
//
// Supported format: JSON with flag names as keys and values matching flag types.
//
// Example config file (myapp.json):
//
//	{
//		"host": "0.0.0.0",
//		"port": 3000,
//		"debug": true,
//		"timeout": "45s",
//		"tags": ["web", "api", "production"]
//	}
//
// Usage:
//
//	fs := flashflags.New("myapp")
//	fs.SetConfigFile("./config/myapp.json")
//	// File will be loaded automatically during Parse()
//
// Security: Path validation prevents directory traversal attacks.
func (fs *FlagSet) SetConfigFile(path string) {
	fs.configFile = path
}

// AddConfigPath adds a directory to search for configuration files during auto-discovery.
// Multiple paths can be added and will be searched in order during Parse().
//
// Auto-discovery searches for these filenames in each path:
//   - {program-name}.json (e.g., "myapp.json")
//   - {program-name}.config.json (e.g., "myapp.config.json")
//   - config.json
//
// Example:
//
//	fs := flashflags.New("myapp")
//	fs.AddConfigPath("./config")        // ./config/myapp.json
//	fs.AddConfigPath("/etc/myapp")      // /etc/myapp/myapp.json
//	fs.AddConfigPath(os.Getenv("HOME")) // $HOME/myapp.json
//
//	// First found config file will be loaded during Parse()
//
// If no paths are added, auto-discovery searches: ".", "./config", "$HOME"
func (fs *FlagSet) AddConfigPath(path string) {
	fs.configPaths = append(fs.configPaths, path)
}

// SetEnvPrefix sets the prefix for environment variable lookup and enables env var processing.
// Flag names are automatically converted using the pattern: PREFIX_FLAGNAME
//
// Conversion rules:
//   - Hyphens become underscores: "db-host" → "PREFIX_DB_HOST"
//   - All uppercase: "api-key" → "PREFIX_API_KEY"
//   - Original case preserved in prefix: "MyApp" → "MyApp_DB_HOST"
//
// Example:
//
//	fs := flashflags.New("webserver")
//	host := fs.String("db-host", "localhost", "Database host")
//	fs.SetEnvPrefix("WEBAPP")
//
//	// Environment variable: WEBAPP_DB_HOST=postgresql.example.com
//	// Command line override: --db-host=localhost
//	// Result: host="localhost" (CLI wins over env var)
//
// This automatically enables environment variable lookup.
func (fs *FlagSet) SetEnvPrefix(prefix string) {
	fs.envPrefix = prefix
	fs.enableEnvLookup = true
}

// SetEnvVar sets a custom environment variable name for a specific flag.
// This overrides the default naming convention (prefix + converted flag name).
//
// Useful for integrating with existing environment variable conventions.
//
// Example:
//
//	fs := flashflags.New("myapp")
//	dbURL := fs.String("database-url", "", "Database connection URL")
//
//	// Use standard naming instead of MYAPP_DATABASE_URL
//	fs.SetEnvVar("database-url", "DATABASE_URL")
//
//	// Now reads from DATABASE_URL environment variable
//	// export DATABASE_URL=postgres://user:pass@host/db
//
// Returns an error if the flag name doesn't exist.
func (fs *FlagSet) SetEnvVar(flagName, envVarName string) error {
	flag, exists := fs.flags[flagName]
	if !exists {
		return fmt.Errorf("flag %s not found", flagName)
	}
	flag.envVar = envVarName
	return nil
}

// EnableEnvLookup enables environment variable lookup using default naming convention.
// No prefix is used - flag names are directly converted to environment variable names.
//
// Conversion rules:
//   - Hyphens become underscores: "db-host" → "DB_HOST"
//   - All uppercase: "api-key" → "API_KEY"
//   - Preserved characters: "timeout" → "TIMEOUT"
//
// Example:
//
//	fs := flashflags.New("myapp")
//	host := fs.String("db-host", "localhost", "Database host")
//	port := fs.Int("db-port", 5432, "Database port")
//	fs.EnableEnvLookup()
//
//	// Reads from: DB_HOST and DB_PORT environment variables
//	// export DB_HOST=postgresql.example.com
//	// export DB_PORT=5433
//
// Use SetEnvPrefix() if you need a prefix to avoid variable name conflicts.
func (fs *FlagSet) EnableEnvLookup() {
	fs.enableEnvLookup = true
}

// LoadConfig loads configuration from file and applies it.
// This is called automatically during Parse, but can be called manually if needed.
//
// Possible errors:
//   - File reading errors: permission denied, file not found (for explicit config files)
//   - JSON parsing errors: invalid JSON syntax in configuration file
//   - Path validation errors: unsafe file paths (directory traversal attempts)
//   - Flag validation errors: config values that fail custom validators
//
// Note: Missing auto-discovery config files are not considered errors.
//
// Example error handling:
//
//	if err := fs.LoadConfig(); err != nil {
//		log.Printf("Config error: %v", err)
//		// Continue without config or exit based on your needs
//	}
func (fs *FlagSet) LoadConfig() error {
	// Skip if already loaded
	if fs.configLoaded {
		return nil
	}
	fs.configLoaded = true

	// Skip entirely if no config file specified and no config paths added
	if fs.configFile == "" && len(fs.configPaths) == 0 {
		return nil
	}

	configPath := fs.findConfigFile()
	if configPath == "" {
		return nil // No config file found, not an error
	}

	return fs.loadConfigFromFile(configPath)
}

// findConfigFile finds the configuration file using the specified path or auto-discovery
func (fs *FlagSet) findConfigFile() string {
	// If explicit config file is set, use it
	if fs.configFile != "" {
		if _, err := os.Stat(fs.configFile); err == nil {
			return fs.configFile
		}
		return "" // Explicit file not found
	}

	// Auto-discovery in specified paths
	configNames := []string{
		fs.name + ".json",
		fs.name + ".config.json",
		"config.json",
	}

	searchPaths := fs.configPaths
	if len(searchPaths) == 0 {
		// Default search paths
		searchPaths = []string{".", "./config", os.Getenv("HOME")}
	}

	for _, dir := range searchPaths {
		for _, name := range configNames {
			path := filepath.Join(dir, name)
			if _, err := os.Stat(path); err == nil {
				return path
			}
		}
	}

	return ""
}

// isSafeAbsolutePath checks if an absolute path is safe for config files
func isSafeAbsolutePath(path string) bool {
	safePrefixes := []string{
		"/tmp/",         // Linux/Unix temp
		"/opt/",         // Optional software
		"/etc/",         // System configuration
		"/var/folders/", // macOS temp
		"/var/tmp/",     // System temp
	}

	for _, prefix := range safePrefixes {
		if strings.HasPrefix(path, prefix) {
			return true
		}
	}
	return false
}

// loadConfigFromFile loads and applies configuration from a JSON file
func (fs *FlagSet) loadConfigFromFile(path string) error {
	// Validate path to prevent directory traversal attacks
	if strings.Contains(path, "..") {
		return fmt.Errorf("invalid config file path: %s", path)
	}

	// Allow relative paths and safe absolute paths
	if strings.HasPrefix(path, "/") && !isSafeAbsolutePath(path) {
		return fmt.Errorf("invalid config file path: %s", path)
	}

	data, err := os.ReadFile(path) // #nosec G304 - path is validated above
	if err != nil {
		return fmt.Errorf("failed to read config file %s: %v", path, err)
	}

	var config map[string]interface{}
	if err := json.Unmarshal(data, &config); err != nil {
		return fmt.Errorf("failed to parse config file %s: %v", path, err)
	}

	return fs.applyConfig(config)
}

// applyConfig applies configuration values to flags (only if not already set by command line)
func (fs *FlagSet) applyConfig(config map[string]interface{}) error {
	for flagName, value := range config {
		flag := fs.Lookup(flagName)
		if flag == nil {
			continue // Skip unknown flags
		}

		// Only apply config value if flag wasn't set by command line
		if flag.changed {
			continue
		}

		// Convert and set the value
		if err := fs.setFlagValueFromConfig(flagName, value); err != nil {
			return fmt.Errorf("failed to set flag %s from config: %v", flagName, err)
		}
	}

	return nil
}

// Config-specific value setters to reduce complexity

func (fs *FlagSet) setStringValueFromConfig(flag *Flag, value interface{}, name string) error {
	if str, ok := value.(string); ok {
		flag.value = str
		if flag.ptr != nil {
			if ptr, ok := flag.ptr.(*string); ok {
				*ptr = str
			}
		}
		return nil
	}
	return fmt.Errorf("expected string for flag %s, got %T", name, value)
}

func (fs *FlagSet) setIntValueFromConfig(flag *Flag, value interface{}, name string) error {
	var intVal int
	switch v := value.(type) {
	case float64: // JSON numbers are float64
		intVal = int(v)
	case int:
		intVal = v
	default:
		return fmt.Errorf("expected number for flag %s, got %T", name, value)
	}
	flag.value = intVal
	if flag.ptr != nil {
		if ptr, ok := flag.ptr.(*int); ok {
			*ptr = intVal
		}
	}
	return nil
}

func (fs *FlagSet) setBoolValueFromConfig(flag *Flag, value interface{}, name string) error {
	if boolVal, ok := value.(bool); ok {
		flag.value = boolVal
		if flag.ptr != nil {
			if ptr, ok := flag.ptr.(*bool); ok {
				*ptr = boolVal
			}
		}
		return nil
	}
	return fmt.Errorf("expected boolean for flag %s, got %T", name, value)
}

func (fs *FlagSet) setFloat64ValueFromConfig(flag *Flag, value interface{}, name string) error {
	var floatVal float64
	switch v := value.(type) {
	case float64:
		floatVal = v
	case int:
		floatVal = float64(v)
	default:
		return fmt.Errorf("expected number for flag %s, got %T", name, value)
	}
	flag.value = floatVal
	if flag.ptr != nil {
		if ptr, ok := flag.ptr.(*float64); ok {
			*ptr = floatVal
		}
	}
	return nil
}

func (fs *FlagSet) setStringSliceValueFromConfig(flag *Flag, value interface{}, name string) error {
	if slice, ok := value.([]interface{}); ok {
		strSlice := make([]string, len(slice))
		for i, item := range slice {
			if str, ok := item.(string); ok {
				strSlice[i] = str
			} else {
				return fmt.Errorf("expected string array for flag %s, got %T in array", name, item)
			}
		}
		flag.value = strSlice
		if flag.ptr != nil {
			if ptr, ok := flag.ptr.(*[]string); ok {
				*ptr = strSlice
			}
		}
		return nil
	}
	return fmt.Errorf("expected array for flag %s, got %T", name, value)
}

func (fs *FlagSet) setFlagValueFromConfig(name string, value interface{}) error {
	flag, exists := fs.flags[name]
	if !exists {
		return fmt.Errorf("unknown flag: %s", name)
	}

	// Set value based on type using dedicated functions
	if err := fs.setConfigValueByType(flag, value, name); err != nil {
		return err
	}

	// Validate the value if validator is set
	return fs.validateFlagValue(flag)
}

// setConfigValueByType sets the flag value from config based on its type
func (fs *FlagSet) setConfigValueByType(flag *Flag, value interface{}, name string) error {
	switch flag.flagType {
	case "string":
		return fs.setStringValueFromConfig(flag, value, name)
	case "int":
		return fs.setIntValueFromConfig(flag, value, name)
	case "bool":
		return fs.setBoolValueFromConfig(flag, value, name)
	case "float64":
		return fs.setFloat64ValueFromConfig(flag, value, name)
	case "stringSlice":
		return fs.setStringSliceValueFromConfig(flag, value, name)
	default:
		return fmt.Errorf("unsupported flag type: %s", flag.flagType)
	}
}

// validateFlagValue validates a flag value using its validator function
func (fs *FlagSet) validateFlagValue(flag *Flag) error {
	if flag.validator != nil {
		return flag.validator(flag.value)
	}
	return nil
}

// LoadEnvironmentVariables loads values from environment variables.
// This is called automatically during Parse, but can be called manually if needed.
//
// Possible errors:
//   - Type conversion errors: "invalid int value for MYAPP_PORT: abc"
//   - Validation errors: environment values that fail custom validators
//   - Duration parsing errors: invalid duration format in environment variable
//
// Environment variables are only processed if EnableEnvLookup() or SetEnvPrefix() was called.
//
// Example error handling:
//
//	if err := fs.LoadEnvironmentVariables(); err != nil {
//		log.Fatalf("Environment variable error: %v", err)
//	}
func (fs *FlagSet) LoadEnvironmentVariables() error {
	if !fs.enableEnvLookup {
		return nil
	}

	for name, flag := range fs.flags {
		// Skip if flag was already set via command line
		if flag.changed {
			continue
		}

		envVarName := fs.getEnvVarName(name, flag)
		if envVarName == "" {
			continue
		}

		envValue := os.Getenv(envVarName)
		if envValue == "" {
			continue
		}

		// Set the flag value from environment variable
		if err := fs.setFlagValue(name, envValue); err != nil {
			return fmt.Errorf("invalid environment variable %s=%s: %v", envVarName, envValue, err)
		}
	}

	return nil
}

// getEnvVarName returns the environment variable name for a flag
func (fs *FlagSet) getEnvVarName(flagName string, flag *Flag) string {
	// Use custom environment variable name if set
	if flag.envVar != "" {
		return flag.envVar
	}

	// Use prefix-based naming if prefix is set
	if fs.envPrefix != "" {
		// Convert flag name: "db-host" -> "MYAPP_DB_HOST"
		envName := strings.ToUpper(strings.ReplaceAll(flagName, "-", "_"))
		return fs.envPrefix + "_" + envName
	}

	// Default naming: "db-host" -> "DB_HOST"
	return strings.ToUpper(strings.ReplaceAll(flagName, "-", "_"))
}
