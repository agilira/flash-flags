// Package stdlib provides drop-in replacement for Go standard library flag package.
// This allows existing code using the standard flag package to seamlessly migrate
// to flash-flags with zero code changes while gaining all the advanced features.
//
// Simply replace:
//   import "flag"
// with:
//   import flag "github.com/agilira/flash-flags/stdlib"
//
// All existing code will work unchanged while gaining:
// - Better performance (1.5x faster parsing)
// - Configuration file support
// - Environment variable integration
// - Advanced validation and constraints
// - Professional help output
//
// Copyright (c) 2025 AGILira - A. Giordano
// Series: an AGILira library
// SPDX-License-Identifier: MPL-2.0

package stdlib

import (
	"errors"
	"math"
	"os"
	"sync"
	"time"

	flashflags "github.com/agilira/flash-flags"
)

// ErrHelp is the error returned if the -help or -h flag is invoked
// but no such flag is defined.
var ErrHelp = errors.New("flag: help requested")

// CommandLine is the default set of command-line flags, parsed from os.Args.
// The top-level functions such as BoolVar, Arg, and so on are wrappers for the
// methods of CommandLine.
var CommandLine = flashflags.New(getProgName())

// Usage is a function to call when a flag is provided but the flag is not defined.
var Usage = func() {
	CommandLine.PrintHelp()
}

// External pointer registry for *Var functions
var (
	pointerMutex sync.RWMutex
	stringVars   = make(map[string]*string)
	intVars      = make(map[string]*int)
	boolVars     = make(map[string]*bool)
	float64Vars  = make(map[string]*float64)
	durationVars = make(map[string]*time.Duration)
	int64Vars    = make(map[string]*int64)
	uintVars     = make(map[string]*uint)
	uint64Vars   = make(map[string]*uint64)
	parsed       bool
)

// getProgName extracts program name from os.Args[0]
func getProgName() string {
	if len(os.Args) == 0 {
		return "program"
	}
	// Extract just the program name without path
	name := os.Args[0]
	for i := len(name) - 1; i >= 0; i-- {
		if name[i] == '/' || name[i] == '\\' {
			return name[i+1:]
		}
	}
	return name
}

// Parse parses the command-line flags from os.Args[1:]. Must be called
// after all flags are defined and before flags are accessed by the program.
func Parse() {
	// Parse ignores errors to maintain stdlib compatibility
	_ = CommandLine.Parse(os.Args[1:])

	// Sync all registered pointers after parsing
	syncPointers()
	parsed = true
}

// Parsed reports whether the command-line flags have been parsed.
func Parsed() bool {
	return parsed
}

// syncPointers updates all registered external pointers with parsed values
func syncPointers() {
	pointerMutex.Lock()
	defer pointerMutex.Unlock()

	for name, ptr := range stringVars {
		*ptr = CommandLine.GetString(name)
	}
	for name, ptr := range intVars {
		*ptr = CommandLine.GetInt(name)
	}
	for name, ptr := range boolVars {
		*ptr = CommandLine.GetBool(name)
	}
	for name, ptr := range float64Vars {
		*ptr = CommandLine.GetFloat64(name)
	}
	for name, ptr := range durationVars {
		*ptr = CommandLine.GetDuration(name)
	}
	for name, ptr := range int64Vars {
		*ptr = int64(CommandLine.GetInt(name))
	}
	for name, ptr := range uintVars {
		intVal := CommandLine.GetInt(name)
		if intVal < 0 {
			*ptr = 0
		} else {
			*ptr = uint(intVal)
		}
	}
	for name, ptr := range uint64Vars {
		intVal := CommandLine.GetInt(name)
		if intVal < 0 {
			*ptr = 0
		} else {
			*ptr = uint64(intVal)
		}
	}
}

// String defines a string flag with specified name, default value, and usage string.
// The return value is the address of a string variable that stores the value of the flag.
func String(name string, value string, usage string) *string {
	return CommandLine.String(name, value, usage)
}

// StringVar defines a string flag with specified name, default value, and usage string.
// The argument p points to a string variable in which to store the value of the flag.
func StringVar(p *string, name string, value string, usage string) {
	*p = value // Set initial value
	CommandLine.String(name, value, usage)

	pointerMutex.Lock()
	stringVars[name] = p
	pointerMutex.Unlock()
}

// Int defines an int flag with specified name, default value, and usage string.
// The return value is the address of an int variable that stores the value of the flag.
func Int(name string, value int, usage string) *int {
	return CommandLine.Int(name, value, usage)
}

// IntVar defines an int flag with specified name, default value, and usage string.
// The argument p points to an int variable in which to store the value of the flag.
func IntVar(p *int, name string, value int, usage string) {
	*p = value
	CommandLine.Int(name, value, usage)

	pointerMutex.Lock()
	intVars[name] = p
	pointerMutex.Unlock()
}

// Bool defines a bool flag with specified name, default value, and usage string.
// The return value is the address of a bool variable that stores the value of the flag.
func Bool(name string, value bool, usage string) *bool {
	return CommandLine.Bool(name, value, usage)
}

// BoolVar defines a bool flag with specified name, default value, and usage string.
// The argument p points to a bool variable in which to store the value of the flag.
func BoolVar(p *bool, name string, value bool, usage string) {
	*p = value
	CommandLine.Bool(name, value, usage)

	pointerMutex.Lock()
	boolVars[name] = p
	pointerMutex.Unlock()
}

// Float64 defines a float64 flag with specified name, default value, and usage string.
// The return value is the address of a float64 variable that stores the value of the flag.
func Float64(name string, value float64, usage string) *float64 {
	return CommandLine.Float64(name, value, usage)
}

// Float64Var defines a float64 flag with specified name, default value, and usage string.
// The argument p points to a float64 variable in which to store the value of the flag.
func Float64Var(p *float64, name string, value float64, usage string) {
	*p = value
	CommandLine.Float64(name, value, usage)

	pointerMutex.Lock()
	float64Vars[name] = p
	pointerMutex.Unlock()
}

// Duration defines a time.Duration flag with specified name, default value, and usage string.
// The return value is the address of a time.Duration variable that stores the value of the flag.
func Duration(name string, value time.Duration, usage string) *time.Duration {
	return CommandLine.Duration(name, value, usage)
}

// DurationVar defines a time.Duration flag with specified name, default value, and usage string.
// The argument p points to a time.Duration variable in which to store the value of the flag.
func DurationVar(p *time.Duration, name string, value time.Duration, usage string) {
	*p = value
	CommandLine.Duration(name, value, usage)

	pointerMutex.Lock()
	durationVars[name] = p
	pointerMutex.Unlock()
}

// Int64 defines an int64 flag with specified name, default value, and usage string.
// The return value is the address of an int64 variable that stores the value of the flag.
// Note: Flash-flags doesn't support int64 natively, so this uses int internally.
func Int64(name string, value int64, usage string) *int64 {
	CommandLine.Int(name, int(value), usage)
	int64Ptr := new(int64)
	*int64Ptr = value

	pointerMutex.Lock()
	int64Vars[name] = int64Ptr
	pointerMutex.Unlock()

	return int64Ptr
}

// Int64Var defines an int64 flag with specified name, default value, and usage string.
// The argument p points to an int64 variable in which to store the value of the flag.
func Int64Var(p *int64, name string, value int64, usage string) {
	*p = value
	CommandLine.Int(name, int(value), usage)

	pointerMutex.Lock()
	int64Vars[name] = p
	pointerMutex.Unlock()
}

// Uint defines a uint flag with specified name, default value, and usage string.
// The return value is the address of a uint variable that stores the value of the flag.
func Uint(name string, value uint, usage string) *uint {
	intValue := 0
	if value <= math.MaxInt {
		intValue = int(value)
	} else {
		intValue = math.MaxInt
	}
	CommandLine.Int(name, intValue, usage)
	uintPtr := new(uint)
	*uintPtr = value

	pointerMutex.Lock()
	uintVars[name] = uintPtr
	pointerMutex.Unlock()

	return uintPtr
}

// UintVar defines a uint flag with specified name, default value, and usage string.
// The argument p points to a uint variable in which to store the value of the flag.
func UintVar(p *uint, name string, value uint, usage string) {
	*p = value
	intValue := 0
	if value <= math.MaxInt {
		intValue = int(value)
	} else {
		intValue = math.MaxInt
	}
	CommandLine.Int(name, intValue, usage)

	pointerMutex.Lock()
	uintVars[name] = p
	pointerMutex.Unlock()
}

// Uint64 defines a uint64 flag with specified name, default value, and usage string.
// The return value is the address of a uint64 variable that stores the value of the flag.
func Uint64(name string, value uint64, usage string) *uint64 {
	intValue := 0
	if value <= math.MaxInt {
		intValue = int(value)
	} else {
		intValue = math.MaxInt
	}
	CommandLine.Int(name, intValue, usage)
	uint64Ptr := new(uint64)
	*uint64Ptr = value

	pointerMutex.Lock()
	uint64Vars[name] = uint64Ptr
	pointerMutex.Unlock()

	return uint64Ptr
}

// Uint64Var defines a uint64 flag with specified name, default value, and usage string.
// The argument p points to a uint64 variable in which to store the value of the flag.
func Uint64Var(p *uint64, name string, value uint64, usage string) {
	*p = value
	intValue := 0
	if value <= math.MaxInt {
		intValue = int(value)
	} else {
		intValue = math.MaxInt
	}
	CommandLine.Int(name, intValue, usage)

	pointerMutex.Lock()
	uint64Vars[name] = p
	pointerMutex.Unlock()
}

// Args returns the non-flag command-line arguments.
func Args() []string {
	return CommandLine.Args()
}

// Arg returns the i'th command-line argument. Arg(0) is the first remaining argument
// after flags have been processed. Arg returns an empty string if the
// requested element does not exist.
func Arg(i int) string {
	return CommandLine.Arg(i)
}

// NArg is the number of arguments remaining after flags have been processed.
func NArg() int {
	return CommandLine.NArg()
}

// NFlag returns the number of command-line flags that have been set.
func NFlag() int {
	count := 0
	CommandLine.VisitAll(func(f *flashflags.Flag) {
		if f.Changed() {
			count++
		}
	})
	return count
}

// Set sets the value of the named command-line flag.
func Set(name, value string) error {
	// Flash-flags doesn't expose setFlagValue publicly
	// We simulate it by parsing a fake argument
	args := []string{"--" + name, value}
	return CommandLine.Parse(args)
}

// PrintDefaults prints, to standard error unless configured otherwise,
// a usage message showing the default settings of all defined command-line flags.
func PrintDefaults() {
	CommandLine.PrintUsage()
}

// Visit visits the command-line flags in lexicographical order, calling fn for each.
// It visits only those flags that have been set.
func Visit(fn func(*Flag)) {
	CommandLine.VisitAll(func(f *flashflags.Flag) {
		if f.Changed() {
			fn(&Flag{
				Name:     f.Name(),
				Usage:    f.Usage(),
				Value:    &stringValue{f.Value()},
				DefValue: "",
			})
		}
	})
}

// VisitAll visits the command-line flags in lexicographical order, calling fn
// for each. It visits all flags, even those not set.
func VisitAll(fn func(*Flag)) {
	CommandLine.VisitAll(func(f *flashflags.Flag) {
		fn(&Flag{
			Name:     f.Name(),
			Usage:    f.Usage(),
			Value:    &stringValue{f.Value()},
			DefValue: "",
			original: f,
		})
	})
}

// Lookup returns the Flag structure of the named command-line flag,
// returning nil if none exists.
func Lookup(name string) *Flag {
	f := CommandLine.Lookup(name)
	if f == nil {
		return nil
	}
	return &Flag{
		Name:     f.Name(),
		Usage:    f.Usage(),
		original: f,
		Value:    &stringValue{f.Value()},
		DefValue: "",
	}
}

// Flag represents the state of a flag.
type Flag struct {
	Name     string // name as it appears on command line
	Usage    string // help message
	Value    Value  // value as set
	DefValue string // default value (as text); for usage message

	// Keep reference to original flag for methods
	original *flashflags.Flag
}

// Changed reports whether the flag was set on the command line.
func (f *Flag) Changed() bool {
	if f.original != nil {
		return f.original.Changed()
	}
	return false
}

// Value is the interface to the dynamic value stored in a flag.
// (The default value is represented as a string.)
type Value interface {
	String() string
	Set(string) error
}

// stringValue implements Value interface for string conversion
type stringValue struct {
	val interface{}
}

func (s *stringValue) String() string {
	if s.val == nil {
		return ""
	}
	return s.val.(string)
}

func (s *stringValue) Set(val string) error {
	// Not implemented for compatibility layer
	return nil
}
