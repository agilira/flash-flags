// flash-flags.go: Ultra-fast command-line flag parsing for Go
//
// Copyright (c) 2025 AGILira
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

// Flag implements ultra-fast flag using only standard library
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

// Name returns the flag name
func (f *Flag) Name() string { return f.name }

// Value returns the flag value
func (f *Flag) Value() interface{} { return f.value }

// Type returns the flag type
func (f *Flag) Type() string { return f.flagType }

// Changed returns whether the flag was set
func (f *Flag) Changed() bool { return f.changed }

// Usage returns the flag usage string
func (f *Flag) Usage() string { return f.usage }

// SetValidator sets a validation function for the flag
func (f *Flag) SetValidator(validator func(interface{}) error) {
	f.validator = validator
}

// Validate validates the current flag value using the validator if set
func (f *Flag) Validate() error {
	if f.validator != nil {
		return f.validator(f.value)
	}
	return nil
}

// Reset resets the flag to its default value
func (f *Flag) Reset() {
	f.value = f.defaultValue
	if f.ptr != nil {
		switch f.flagType {
		case "string":
			*f.ptr.(*string) = f.defaultValue.(string)
		case "int":
			*f.ptr.(*int) = f.defaultValue.(int)
		case "bool":
			*f.ptr.(*bool) = f.defaultValue.(bool)
		case "float64":
			*f.ptr.(*float64) = f.defaultValue.(float64)
		case "duration":
			*f.ptr.(*time.Duration) = f.defaultValue.(time.Duration)
		case "stringSlice":
			*f.ptr.(*[]string) = f.defaultValue.([]string)
		}
	}
	f.changed = false
}

// FlagSet implements ultra-fast flag set using only standard library
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

// New creates a new flag set with zero external dependencies
func New(name string) *FlagSet {
	return &FlagSet{
		flags:    make(map[string]*Flag),
		shortMap: make(map[string]*Flag),
		name:     name,
	}
}

// String adds a string flag
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

// StringVar adds a string flag with optional short key
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

// Int adds an integer flag
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

// IntVar adds an integer flag with optional short key
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

// Bool adds a boolean flag
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

// BoolVar adds a boolean flag with optional short key
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

// Duration adds a duration flag
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

// Float64 adds a float64 flag
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

// StringSlice adds a string slice flag
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

// Parse parses command line arguments with optimized allocations
func (fs *FlagSet) Parse(args []string) error {
	// Load configuration file first (lowest priority)
	if err := fs.LoadConfig(); err != nil {
		return fmt.Errorf("config file error: %v", err)
	}

	// Load environment variables second (medium priority)
	if err := fs.LoadEnvironmentVariables(); err != nil {
		return fmt.Errorf("environment variable error: %v", err)
	}

	for i := 0; i < len(args); i++ {
		arg := args[i]

		if !strings.HasPrefix(arg, "-") {
			continue
		}

		// Handle help flags
		if arg == "--help" || arg == "-h" {
			fs.PrintHelp()
			return fmt.Errorf("help requested")
		}

		// Handle short flags (-p, -d)
		if len(arg) == 2 && arg[0] == '-' && arg[1] != '-' {
			shortKey := string(arg[1])
			flag, exists := fs.shortMap[shortKey]
			if !exists {
				return fmt.Errorf("unknown flag: -%s", shortKey)
			}

			if flag.flagType == "bool" {
				flag.value = true
				if flag.ptr != nil {
					*flag.ptr.(*bool) = true
				}
				flag.changed = true
				continue
			}

			// Non-bool short flag needs value
			if i+1 >= len(args) {
				return fmt.Errorf("flag -%s requires a value", shortKey)
			}
			flagValue := args[i+1]
			i++ // Skip next argument

			err := fs.setFlagValue(flag.name, flagValue)
			if err != nil {
				return err
			}
			continue
		}

		// Handle long flags (--name)
		if !strings.HasPrefix(arg, "--") {
			continue
		}

		// Remove -- prefix without allocation
		arg = arg[2:]

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
				i++ // Skip next argument
			} else {
				// Boolean flag or error
				if flag, exists := fs.flags[flagName]; exists && flag.flagType == "bool" {
					flagValue = "true"
				} else {
					return fmt.Errorf("flag --%s requires a value", flagName)
				}
			}
		}

		// Set flag value
		err := fs.setFlagValue(flagName, flagValue)
		if err != nil {
			return err
		}
	}

	// Validate all constraints after parsing
	return fs.ValidateAllConstraints()
}

// setFlagValue sets a flag value with type conversion and minimal allocations
func (fs *FlagSet) setFlagValue(name, value string) error {
	flag, exists := fs.flags[name]
	if !exists {
		return fmt.Errorf("unknown flag: --%s", name)
	}

	switch flag.flagType {
	case "string":
		flag.value = value
		if flag.ptr != nil {
			*flag.ptr.(*string) = value
		}
		flag.changed = true

	case "int":
		intVal, err := strconv.Atoi(value)
		if err != nil {
			return fmt.Errorf("invalid int value for flag --%s: %s", name, value)
		}
		flag.value = intVal
		if flag.ptr != nil {
			*flag.ptr.(*int) = intVal
		}
		flag.changed = true

	case "bool":
		boolVal, err := strconv.ParseBool(value)
		if err != nil {
			return fmt.Errorf("invalid bool value for flag --%s: %s", name, value)
		}
		flag.value = boolVal
		if flag.ptr != nil {
			*flag.ptr.(*bool) = boolVal
		}
		flag.changed = true

	case "duration":
		durVal, err := time.ParseDuration(value)
		if err != nil {
			return fmt.Errorf("invalid duration value for flag --%s: %s", name, value)
		}
		flag.value = durVal
		if flag.ptr != nil {
			*flag.ptr.(*time.Duration) = durVal
		}
		flag.changed = true

	case "float64":
		floatVal, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return fmt.Errorf("invalid float64 value for flag --%s: %s", name, value)
		}
		flag.value = floatVal
		if flag.ptr != nil {
			*flag.ptr.(*float64) = floatVal
		}
		flag.changed = true

	case "stringSlice":
		// Optimized string slice parsing with minimal allocations
		if value == "" {
			slice := []string{}
			flag.value = slice
			if flag.ptr != nil {
				*flag.ptr.(*[]string) = slice
			}
		} else {
			// Manual parsing to avoid strings.Count allocation
			commas := 0
			for _, c := range []byte(value) {
				if c == ',' {
					commas++
				}
			}
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
			flag.value = slice
			if flag.ptr != nil {
				*flag.ptr.(*[]string) = slice
			}
		}
		flag.changed = true

	default:
		return fmt.Errorf("unsupported flag type: %s", flag.flagType)
	}

	// Run validation if validator is set
	if flag.validator != nil {
		if err := flag.validator(flag.value); err != nil {
			return fmt.Errorf("validation failed for flag --%s: %v", name, err)
		}
	}

	return nil
}

// VisitAll calls fn for each flag in the set
func (fs *FlagSet) VisitAll(fn func(*Flag)) {
	for _, flag := range fs.flags {
		fn(flag)
	}
}

// Lookup finds a flag by name
func (fs *FlagSet) Lookup(name string) *Flag {
	flag, exists := fs.flags[name]
	if !exists {
		return nil
	}
	return flag
}

// PrintUsage prints usage information for all flags
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

// Changed returns whether the flag was set
func (fs *FlagSet) Changed(name string) bool {
	if flag := fs.Lookup(name); flag != nil {
		return flag.changed
	}
	return false
}

// SetValidator sets a validation function for a specific flag
func (fs *FlagSet) SetValidator(name string, validator func(interface{}) error) error {
	flag := fs.Lookup(name)
	if flag == nil {
		return fmt.Errorf("flag not found: %s", name)
	}
	flag.SetValidator(validator)
	return nil
}

// SetRequired marks a flag as required
func (fs *FlagSet) SetRequired(name string) error {
	flag := fs.Lookup(name)
	if flag == nil {
		return fmt.Errorf("flag not found: %s", name)
	}
	flag.required = true
	return nil
}

// SetDependencies sets dependencies for a flag (this flag requires the dependent flags to be set)
func (fs *FlagSet) SetDependencies(name string, dependencies ...string) error {
	flag := fs.Lookup(name)
	if flag == nil {
		return fmt.Errorf("flag not found: %s", name)
	}
	flag.dependencies = dependencies
	return nil
}

// SetDescription sets the program description for help output
func (fs *FlagSet) SetDescription(description string) {
	fs.description = description
}

// SetVersion sets the program version for help output
func (fs *FlagSet) SetVersion(version string) {
	fs.version = version
}

// SetGroup sets the group for a flag (for help organization)
func (fs *FlagSet) SetGroup(name, group string) error {
	flag := fs.Lookup(name)
	if flag == nil {
		return fmt.Errorf("flag not found: %s", name)
	}
	flag.group = group
	return nil
}

// ValidateAll validates all flags that have validators set
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

// ValidateRequired checks that all required flags are set
func (fs *FlagSet) ValidateRequired() error {
	for name, flag := range fs.flags {
		if flag.required && !flag.changed {
			return fmt.Errorf("required flag --%s not provided", name)
		}
	}
	return nil
}

// ValidateDependencies checks that all flag dependencies are satisfied
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

// ValidateAllConstraints validates all constraints: validators, required flags, and dependencies
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

// Reset resets all flags to their default values
func (fs *FlagSet) Reset() {
	for _, flag := range fs.flags {
		flag.Reset()
	}
}

// ResetFlag resets a specific flag to its default value
func (fs *FlagSet) ResetFlag(name string) error {
	if flag, exists := fs.flags[name]; exists {
		flag.Reset()
		return nil
	}
	return fmt.Errorf("flag --%s not found", name)
}

// GetString gets a flag value as string
func (fs *FlagSet) GetString(name string) string {
	if flag, exists := fs.flags[name]; exists {
		if str, ok := flag.value.(string); ok {
			return str
		}
		return fmt.Sprintf("%v", flag.value)
	}
	return ""
}

// GetInt gets a flag value as int
func (fs *FlagSet) GetInt(name string) int {
	if flag, exists := fs.flags[name]; exists {
		if intVal, ok := flag.value.(int); ok {
			return intVal
		}
	}
	return 0
}

// GetBool gets a flag value as bool
func (fs *FlagSet) GetBool(name string) bool {
	if flag, exists := fs.flags[name]; exists {
		if boolVal, ok := flag.value.(bool); ok {
			return boolVal
		}
	}
	return false
}

// GetDuration gets a flag value as duration
func (fs *FlagSet) GetDuration(name string) time.Duration {
	if flag, exists := fs.flags[name]; exists {
		if durVal, ok := flag.value.(time.Duration); ok {
			return durVal
		}
	}
	return 0
}

// GetFloat64 gets a flag value as float64
func (fs *FlagSet) GetFloat64(name string) float64 {
	if flag, exists := fs.flags[name]; exists {
		if floatVal, ok := flag.value.(float64); ok {
			return floatVal
		}
	}
	return 0.0
}

// GetStringSlice gets a flag value as string slice
func (fs *FlagSet) GetStringSlice(name string) []string {
	if flag, exists := fs.flags[name]; exists {
		if slice, ok := flag.value.([]string); ok {
			return slice
		}
	}
	return []string{}
}

// Help generates and returns the help text
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
	line.WriteString("  ")
	if flag.shortKey != "" {
		line.WriteString("-")
		line.WriteString(flag.shortKey)
		line.WriteString(", ")
	}
	line.WriteString("--")
	line.WriteString(flag.name)

	// Add type info for non-bool flags
	if flag.flagType != "bool" {
		line.WriteString(" ")
		line.WriteString(strings.ToUpper(flag.flagType))
	}

	// Pad to align descriptions
	for line.Len() < 30 {
		line.WriteString(" ")
	}

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

	line.WriteString("\n")
	return line.String()
}

// PrintHelp prints the help text to stdout
func (fs *FlagSet) PrintHelp() {
	fmt.Print(fs.Help())
}

// SetConfigFile sets the configuration file path
func (fs *FlagSet) SetConfigFile(path string) {
	fs.configFile = path
}

// AddConfigPath adds a path to search for configuration files
func (fs *FlagSet) AddConfigPath(path string) {
	fs.configPaths = append(fs.configPaths, path)
}

// SetEnvPrefix sets the prefix for environment variable lookup
// For example, if prefix is "MYAPP", flag "host" will look for "MYAPP_HOST"
func (fs *FlagSet) SetEnvPrefix(prefix string) {
	fs.envPrefix = prefix
	fs.enableEnvLookup = true
}

// SetEnvVar sets a custom environment variable name for a specific flag
func (fs *FlagSet) SetEnvVar(flagName, envVarName string) error {
	flag, exists := fs.flags[flagName]
	if !exists {
		return fmt.Errorf("flag %s not found", flagName)
	}
	flag.envVar = envVarName
	return nil
}

// EnableEnvLookup enables environment variable lookup with default naming
// Flag names are converted to uppercase with dashes replaced by underscores
func (fs *FlagSet) EnableEnvLookup() {
	fs.enableEnvLookup = true
}

// LoadConfig loads configuration from file and applies it
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

// loadConfigFromFile loads and applies configuration from a JSON file
func (fs *FlagSet) loadConfigFromFile(path string) error {
	// Validate path to prevent directory traversal attacks
	if strings.Contains(path, "..") {
		return fmt.Errorf("invalid config file path: %s", path)
	}

	// Allow relative paths, /tmp/ (for tests), /opt/, /etc/
	if strings.HasPrefix(path, "/") &&
		!strings.HasPrefix(path, "/tmp/") &&
		!strings.HasPrefix(path, "/opt/") &&
		!strings.HasPrefix(path, "/etc/") {
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

// setFlagValueFromConfig sets a flag value from config data with type conversion
func (fs *FlagSet) setFlagValueFromConfig(name string, value interface{}) error {
	flag, exists := fs.flags[name]
	if !exists {
		return fmt.Errorf("unknown flag: %s", name)
	}

	switch flag.flagType {
	case "string":
		if str, ok := value.(string); ok {
			flag.value = str
			if flag.ptr != nil {
				*flag.ptr.(*string) = str
			}
		} else {
			return fmt.Errorf("expected string for flag %s, got %T", name, value)
		}

	case "int":
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
			*flag.ptr.(*int) = intVal
		}

	case "bool":
		if boolVal, ok := value.(bool); ok {
			flag.value = boolVal
			if flag.ptr != nil {
				*flag.ptr.(*bool) = boolVal
			}
		} else {
			return fmt.Errorf("expected boolean for flag %s, got %T", name, value)
		}

	case "float64":
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
			*flag.ptr.(*float64) = floatVal
		}

	case "stringSlice":
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
				*flag.ptr.(*[]string) = strSlice
			}
		} else {
			return fmt.Errorf("expected array for flag %s, got %T", name, value)
		}

	default:
		return fmt.Errorf("unsupported flag type: %s", flag.flagType)
	}

	// Validate the value if validator is set
	if flag.validator != nil {
		if err := flag.validator(flag.value); err != nil {
			return err
		}
	}

	return nil
}

// LoadEnvironmentVariables loads values from environment variables
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
