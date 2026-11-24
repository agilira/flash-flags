// flag.go: a single flag, its metadata and its reset
//
// Copyright (c) 2025 AGILira - A. Giordano
// Series: an AGILira library
// SPDX-License-Identifier: MPL-2.0

package flashflags

// Flag represents a single command-line flag with its value, metadata, and constraints.
// It implements ultra-fast flag handling using only the standard library.
// Safe for concurrent reads after Parse() completes.
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
// A Flag is safe for concurrent reads once Parse has returned; it is not safe to
// read while any goroutine mutates the owning FlagSet. See the package
// documentation on concurrency.
type Flag struct {
	name         string
	value        interface{}
	defaultValue interface{} // original default value for reset
	ptr          interface{} // pointer to the actual value
	kind         *flagKind   // per-type parsing, conversion and storage
	source       flagSource  // highest-priority source that supplied the current value
	usage        string
	shortKey     string                  // Short flag key (e.g., "p" for port)
	validator    func(interface{}) error // Optional validation function
	required     bool                    // Whether this flag is required
	dependencies []string                // Flags that this flag depends on
	group        string                  // Group name for help organization
	envVar       string                  // Environment variable name for this flag
	confineTo    string                  // Directory this path-valued flag is confined to, if any
	givenIn      uint64                  // FlagSet.parseGen of the command line that last set it
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
func (f *Flag) Type() string { return f.kind.name }

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
func (f *Flag) Changed() bool { return f.source != sourceDefault }

// Source returns the configuration source that supplied the flag's current
// value: "cli", "env", "config" or "default".
//
// Sources are resolved by precedence, so the returned name is the highest-priority
// source that set the flag, not merely the last one consulted. A flag present in
// both the config file and the environment reports "env".
//
// Example:
//
//	flag := fs.Lookup("port")
//	fmt.Printf("port=%v (from %s)\n", flag.Value(), flag.Source())
func (f *Flag) Source() string { return f.source.String() }

// Usage returns the flag usage description string.
// This is the help text that was provided when the flag was created.
//
// Example:
//
//	flag := fs.Lookup("port")
//	fmt.Printf("Flag usage: %s\n", flag.Usage()) // Output: Flag usage: Server port number
func (f *Flag) Usage() string { return f.usage }

// ShortKey returns the short flag key (single character for -f syntax).
// Returns empty string if no short key is defined for this flag.
//
// Example:
//
//	flag := fs.Lookup("port")
//	if flag.ShortKey() != "" {
//		fmt.Printf("Short key: -%s\n", flag.ShortKey()) // Output: Short key: -p
//	}
func (f *Flag) ShortKey() string { return f.shortKey }

// SetValidator sets a validation function for the flag.
// The validator will be called whenever the flag value is set or changed.
//
// The validator receives the value boxed in an interface{}. Always use the
// comma-ok form of the type assertion: a bare val.(int) panics, and a panic
// inside a validator propagates out of Parse and terminates the program. The
// dynamic type matches the flag's declared type -- int for Int, time.Duration
// for Duration, []string for StringSlice -- so a failed assertion means the
// validator was attached to the wrong flag.
//
// Example:
//
//	flag := fs.Lookup("port")
//	flag.SetValidator(func(val interface{}) error {
//		port, ok := val.(int)
//		if !ok {
//			return fmt.Errorf("expected int, got %T", val)
//		}
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
	f.kind.store(f.ptr, f.defaultValue)
	f.source = sourceDefault
}

// set stores v as the flag's value and writes it through the declared pointer.
func (f *Flag) set(v interface{}) {
	f.value = v
	f.kind.store(f.ptr, v)
}
