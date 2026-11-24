// parse.go: command-line parsing and the value setters it drives
//
// Copyright (c) 2025 AGILira - A. Giordano
// Series: an AGILira library
// SPDX-License-Identifier: MPL-2.0

package flashflags

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

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
//	-f value              (short flag with space-separated value)
//	-f=value              (short flag with equals-separated value)
//	-abc                  (combined short flags, all must be boolean except last)
//	-abc value            (combined short flags with value for last non-boolean flag)
//	--boolean-flag        (boolean flag without value, defaults to true)
//	--boolean-flag=true   (explicit boolean value)
//	-b                    (boolean short flag, defaults to true)
//	-b=false              (explicit boolean short flag value)
//
// Special handling:
//
//	--help, -h            (shows help and returns error "help requested")
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
	// Reset args slice for new parsing
	fs.args = nil
	fs.repeated = map[string]bool{}

	for i := 0; i < len(args); i++ {
		arg := args[i]

		// Handle special "--" separator (everything after is non-flag)
		if arg == "--" {
			// Collect all remaining args as non-flag arguments
			if i+1 < len(args) {
				fs.args = append(fs.args, args[i+1:]...)
			}
			return nil
		}

		// Check if this is a flag (starts with -). A lone "-" is not: by
		// POSIX convention it is an operand, usually meaning standard input.
		//
		// WHY: it starts with '-' but is neither a short nor a long flag, so
		// processArgument consumed nothing and it vanished silently from
		// both the flags and Args().
		if !strings.HasPrefix(arg, "-") || arg == "-" {
			// Non-flag argument - collect it
			fs.args = append(fs.args, arg)
			continue
		}

		// Process the flag
		consumed, err := fs.processArgument(args, i)
		if err != nil {
			return err
		}

		// Move past consumed arguments
		i += consumed
	}
	return nil
}

// processArgument processes a single argument and returns consumed count
// Assumes the argument is a flag (starts with -)
func (fs *FlagSet) processArgument(args []string, i int) (int, error) {
	arg := args[i]

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

// isHelpFlag reports whether the argument should be handled as the built-in
// help flag.
//
// WHY it consults the registered flags: --help and -h are a convenience the
// parser offers, not names it reserves. Intercepting them unconditionally made
// a flag registered under either spelling unreachable -- StringVar("host", "h",
// ...) could never be set with -h, because Parse printed the help text and
// returned "help requested" before the short key was ever looked up. A caller
// that registers one of these names means it, so the built-in yields and the
// argument is parsed like any other.
//
// The two spellings are judged independently: claiming --help as a flag name
// leaves -h to the built-in, and claiming the short key "h" leaves --help to it.
func (fs *FlagSet) isHelpFlag(arg string) bool {
	switch arg {
	case "--help":
		_, claimed := fs.flags["help"]
		return !claimed
	case "-h":
		_, claimed := fs.shortMap["h"]
		return !claimed
	default:
		return false
	}
}

// isShortFlag checks if the argument is a short flag (includes -f, -f=value, -abc)
func (fs *FlagSet) isShortFlag(arg string) bool {
	return len(arg) >= 2 && arg[0] == '-' && arg[1] != '-'
}

// isLongFlag checks if the argument is a long flag
func (fs *FlagSet) isLongFlag(arg string) bool {
	return strings.HasPrefix(arg, "--")
}

// parseShortFlag handles short flag parsing (-p, -d, -p=value, -abc)
func (fs *FlagSet) parseShortFlag(args []string, i int) (int, error) {
	arg := args[i]

	// Check for combined shorthand flags or single flag with equals
	if len(arg) > 2 {
		return fs.parseComplexShortFlag(args, i)
	}

	// Simple single short flag: -p
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
		flag.source = sourceCLI
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

// parseComplexShortFlag handles complex short flag patterns: -f=value and -abc (combined flags)
func (fs *FlagSet) parseComplexShortFlag(args []string, i int) (int, error) {
	arg := args[i][1:] // Remove initial '-'

	// Check for equals sign: -f=value
	if eqPos := strings.IndexByte(arg, '='); eqPos != -1 {
		return fs.parseShortFlagWithEquals(arg, eqPos)
	}

	// Combined shorthand flags: -abc
	return fs.parseCombinedShortFlags(args, i, arg)
}

// parseShortFlagWithEquals handles -f=value syntax with optimized parsing
func (fs *FlagSet) parseShortFlagWithEquals(arg string, eqPos int) (int, error) {
	if eqPos != 1 {
		// Invalid format: should be -f=value, not -ab=value
		return 0, fmt.Errorf("invalid short flag format with equals: -%s", arg)
	}

	shortKey := string(arg[0])
	flagValue := arg[eqPos+1:]

	flag, exists := fs.shortMap[shortKey]
	if !exists {
		return 0, fmt.Errorf("unknown flag: -%s", shortKey)
	}

	// For boolean flags, -f=true or -f=false
	if flag.flagType == "bool" {
		err := fs.setFlagValue(flag.name, flagValue)
		if err != nil {
			return 0, err
		}
		return 0, nil
	}

	// For non-boolean flags, -f=value
	err := fs.setFlagValue(flag.name, flagValue)
	if err != nil {
		return 0, err
	}
	return 0, nil
}

// parseCombinedShortFlags handles -abc syntax (GNU-style combined short flags)
// Performance-optimized with single-pass parsing and minimal allocations
func (fs *FlagSet) parseCombinedShortFlags(args []string, i int, flagChars string) (int, error) {
	consumed := 0

	// Process each character as a separate flag
	for pos, char := range []byte(flagChars) {
		shortKey := string(char)
		flag, exists := fs.shortMap[shortKey]
		if !exists {
			return 0, fmt.Errorf("unknown flag in combined sequence: -%s", shortKey)
		}

		// All flags except the last must be boolean for combined syntax
		isLastFlag := pos == len(flagChars)-1

		if flag.flagType == "bool" {
			// Set boolean flag to true
			flag.value = true
			if flag.ptr != nil {
				if ptr, ok := flag.ptr.(*bool); ok {
					*ptr = true
				}
			}
			flag.source = sourceCLI
		} else {
			// Non-boolean flag must be the last in the sequence
			if !isLastFlag {
				return 0, fmt.Errorf("non-boolean flag -%s must be last in combined sequence -%s", shortKey, flagChars)
			}

			// Get value from next argument
			if i+1 >= len(args) {
				return 0, fmt.Errorf("flag -%s in combined sequence requires a value", shortKey)
			}

			flagValue := args[i+1]
			err := fs.setFlagValue(flag.name, flagValue)
			if err != nil {
				return 0, err
			}
			consumed = 1 // Consumed the next argument for value
		}
	}

	return consumed, nil
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
		// Check if this is a boolean flag first
		if flag, exists := fs.flags[flagName]; exists && flag.flagType == "bool" {
			// Boolean flag without explicit value = true
			flagValue = "true"
		} else {
			// Non-boolean flag: take value from the next argument.
			// WHY: we do NOT check for a '-' prefix on the next arg because
			// legitimate values can start with '-' (e.g., negative numbers
			// like "-5", grep patterns like "-foo"). This matches the
			// behavior of parseShortFlag which already works correctly.
			if i+1 < len(args) {
				flagValue = args[i+1]
				err := fs.setFlagValue(flagName, flagValue)
				if err != nil {
					return 0, err
				}
				return 1, nil // Consumed one extra argument
			}
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
	// WHY: a slice flag repeated on the command line accumulates
	// (--tag a --tag b means [a b]), as the NOTE on splitByComma promises.
	// Only a repetition within the command line appends: a value from the
	// command line still replaces one from the default, env or config.
	if fs.repeated[flag.name] {
		if current, ok := flag.value.([]string); ok {
			slice = append(append([]string(nil), current...), slice...)
		}
	}

	// WHY no per-element screening here: the caller screened the whole value
	// before it was split, and input hygiene is a per-character property, so no
	// element can fail a check the joined string already passed. The config
	// file path does need it, because a JSON array never passes through this
	// function -- see screenConfigValue.

	// Additional validation for slice size (DoS protection)
	if len(slice) > 10000 {
		return fmt.Errorf("string slice too large: %d items (max: 10000)", len(slice))
	}

	flag.value = slice
	fs.updateStringSlicePointer(flag, slice)
	if fs.repeated != nil {
		fs.repeated[flag.name] = true
	}
	return nil
}

// parseStringSlice parses a comma-separated string into a slice
func (fs *FlagSet) parseStringSlice(value string) []string {
	if value == "" {
		return []string{}
	}
	return fs.splitByComma(value)
}

// splitByComma splits a string by commas with optimized allocation.
// NOTE: backslash-escaped commas (\,) are NOT supported -- values containing
// literal commas must be passed as separate --flag invocations or quoted at the
// shell level.  Escape support is deferred to a future minor release to avoid
// a breaking change in the split semantics.
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

// validateInputHygiene rejects flag values that are malformed as strings,
// whatever they are later used for: a value longer than maxValueLength, or one
// containing a null byte or a C0 control character other than tab, newline and
// carriage return.
//
// WHY it stops there: a flag parser does not know whether a value will reach a
// shell, a SQL driver, an fmt verb or a file open, so it cannot decide which
// substrings are dangerous. Until v1.1.9 this function also rejected values
// containing "/etc/", "rm -rf", "drop table", "$(", a backtick, format verbs
// and Windows device names. That denylist blocked no attack -- "a; rm -rf ~",
// "deploy && restart" and "..%2f..%2fetc" all passed it -- while rejecting
// ordinary input such as "--config /etc/myapp.conf". Escaping belongs at the
// point of use: exec without a shell, parameterize SQL, resolve and confine
// paths. See screening_test.go, which pins both halves of this contract.
func (fs *FlagSet) validateInputHygiene(name, value string) error {
	// Fast path: length check first (most common case)
	valueLen := len(value)
	if valueLen > maxValueLength {
		return fmt.Errorf("flag --%s value too long: %d bytes (max: %d)", name, valueLen, maxValueLength)
	}

	if valueLen == 0 {
		return nil
	}

	// Fast path: a simple alphanumeric value cannot contain a null byte or a
	// control character, so the full scan would find nothing.
	if valueLen < 100 && isSimpleAlphanumeric(value) {
		return nil
	}

	return fs.validateInputHygieneSlow(name, value)
}

// maxValueLength is the largest accepted flag value, in bytes. It is compared
// against len(value), which counts bytes rather than runes, so a value of
// multi-byte characters reaches the cap sooner than its character count
// suggests. The error message names the unit for that reason.
const maxValueLength = 10000

// isSimpleAlphanumeric checks if a string contains only safe characters
func isSimpleAlphanumeric(s string) bool {
	for _, r := range s {
		if !((r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') ||
			(r >= '0' && r <= '9') || r == '-' || r == '_' || r == '.' || r == ':') {
			return false
		}
	}
	return true
}

// validateInputHygieneSlow scans a value that did not qualify for the fast path.
// It is a single pass looking for bytes that cannot appear in a well-formed
// value: a null byte, which truncates the string in any C API it reaches, and a
// C0 control character other than tab, newline or carriage return, which is
// almost always a terminal escape sequence rather than data.
func (fs *FlagSet) validateInputHygieneSlow(name, value string) error {
	for i, r := range value {
		if r == '\x00' {
			return fmt.Errorf("flag --%s contains null byte at position %d", name, i)
		}
		if r < 32 && r != '\t' && r != '\n' && r != '\r' {
			return fmt.Errorf("flag --%s contains control character: \\x%02x", name, r)
		}
	}
	return nil
}

// setFlagValue sets a flag from a command-line argument. It is a thin wrapper
// over setFlagValueFrom that records the command-line source.
func (fs *FlagSet) setFlagValue(name, value string) error {
	return fs.setFlagValueFrom(name, value, sourceCLI)
}

// setFlagValueFrom parses value into the named flag and records src as its
// origin, provided src outranks the source that currently owns the value.
//
// A write from a lower-priority source is discarded and reported as success:
// being outranked is the normal outcome of layered configuration, not a
// failure. The value is not parsed or validated in that case, so a leftover
// environment variable cannot fail a parse whose result it would never reach.
func (fs *FlagSet) setFlagValueFrom(name, value string, src flagSource) error {
	flag, exists := fs.flags[name]
	if !exists {
		return fmt.Errorf("unknown flag: --%s", name)
	}

	if !flag.canSet(src) {
		return nil
	}

	// Apply security validation before processing the value (optimized path)
	if len(value) > 0 && (len(value) > 100 || !isSimpleAlphanumeric(value)) {
		if err := fs.validateInputHygiene(name, value); err != nil {
			return err
		}
	}

	if err := fs.setFlagValueByType(flag, value, name); err != nil {
		return err
	}

	flag.source = src
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
	if err := fs.checkConfinement(flag, name); err != nil {
		return err
	}
	if flag.validator != nil {
		if err := flag.validator(flag.value); err != nil {
			return fmt.Errorf("validation failed for flag --%s: %v", name, err)
		}
	}
	return nil
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

// Args returns the remaining non-flag arguments after parsing.
// These are the arguments that were not consumed by any flag.
// Returns empty slice if no arguments remain or Parse() has not been called.
//
// Example:
//
//	fs := flashflags.New("myapp")
//	fs.String("host", "localhost", "Server host")
//	_ = fs.Parse([]string{"--host", "example.com", "file1", "file2"})
//	args := fs.Args()  // Returns ["file1", "file2"]
func (fs *FlagSet) Args() []string {
	if fs.args == nil {
		return []string{}
	}
	// Return a copy to prevent external modification
	result := make([]string, len(fs.args))
	copy(result, fs.args)
	return result
}

// NArg returns the number of remaining non-flag arguments after parsing.
// Equivalent to len(fs.Args()).
//
// Example:
//
//	fs := flashflags.New("myapp")
//	_ = fs.Parse([]string{"file1", "file2"})
//	count := fs.NArg()  // Returns 2
func (fs *FlagSet) NArg() int {
	return len(fs.args)
}

// Arg returns the i'th remaining argument. Arg(0) is the first remaining argument
// after flags have been processed. Arg returns an empty string if the
// requested element does not exist.
//
// Example:
//
//	fs := flashflags.New("myapp")
//	_ = fs.Parse([]string{"file1", "file2"})
//	first := fs.Arg(0)   // Returns "file1"
//	second := fs.Arg(1)  // Returns "file2"
//	third := fs.Arg(2)   // Returns ""
func (fs *FlagSet) Arg(i int) string {
	if i < 0 || i >= len(fs.args) {
		return ""
	}
	return fs.args[i]
}
