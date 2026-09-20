// Package flashflags provides ultra-fast, zero-dependency command-line flag parsing for Go.
//
// Flash-flags is designed for maximum performance with minimal memory allocations,
// input hardening, and compatibility with Go 1.25.9+.
// It provides a clean API similar to the standard library flag package, with
// additional features: configuration files, environment variables, validation,
// dependencies and grouped help.
//
// Key Features:
//
//   - Input hygiene screening -- see Input Hardening below
//   - Ultra-fast parsing (924ns/op)
//   - Zero external dependencies (only standard library)
//   - Safe for concurrent reads after Parse() -- see Thread Safety below
//   - Drop-in replacement for Go standard library flag package
//   - Support for configuration files (JSON)
//   - Environment variable integration
//   - Flag validation and constraints
//   - Grouped help output
//   - Comprehensive flag syntax support (POSIX/GNU-style)
//   - Short and long flag support with combined syntax
//   - Full support for remaining arguments (Args(), NArg(), Arg(i))
//   - Stdlib-compatible boolean flag behavior
//
// Input Hardening:
//
// Flag values are screened for input that is malformed as a string, whatever it
// is later used for:
//
//   - Length: values above 10000 bytes are rejected
//   - Null bytes: a value containing \x00 is rejected, because that byte
//     truncates the string in any C API it reaches
//   - Control characters: C0 controls other than tab, newline and carriage
//     return are rejected, since they are terminal escape sequences, not data
//   - Fast path: values under 100 bytes made only of [A-Za-z0-9-_.:] skip the
//     scan, because such a value cannot contain either
//
// The screening stops there, deliberately. A flag parser does not know whether
// a value will reach a shell, a SQL driver, an fmt verb or a file open, so it
// cannot decide which substrings are dangerous. Escaping belongs at the point
// of use: run os/exec without a shell, parameterize SQL, and resolve and
// confine paths before opening them.
//
// Until v1.1.9 the screening also rejected values containing "/etc/", "/proc/",
// "/sys/", "rm -rf", "drop table", "$(", a backtick, the %n %s %x %d %c %p
// format verbs, "../" and Windows device names. That denylist stopped no
// attack -- "a; rm -rf ~", "deploy && restart" and "..%2f..%2fetc" all passed
// it untouched -- while rejecting ordinary input such as
// "--config /etc/myapp.conf" or a --command argument containing
// "rm -rf /tmp/build". It has been removed. If you relied on it as a security
// control, it was not one; screen values where you know what they mean.
//
// Supported Flag Syntax:
//
// FlashFlags supports comprehensive POSIX/GNU-style flag syntax:
//
//	Long flags:
//	  --flag value          (space-separated)
//	  --flag=value          (equals-separated)
//	  --boolean-flag        (boolean without value)
//	  --boolean-flag=true   (explicit boolean value)
//
//	Short flags:
//	  -f value              (space-separated)
//	  -f=value              (equals-separated)
//	  -b                    (boolean short flag)
//	  -b=false              (explicit boolean value)
//
//	Combined short flags:
//	  -abc                  (equivalent to -a -b -c)
//	  -abc value            (with value for last flag)
//	  -vdp 8080             (verbose + debug + port=8080)
//
//	Special syntax:
//	  --help, -h            (shows help)
//	  --                    (end of flags marker)
//
// All boolean flags except the last in combined sequences (-abc) must be boolean.
// The last flag in a combined sequence can be any type and will consume the next argument as its value.
//
// Remaining Arguments:
//
// Flash-flags fully supports remaining non-flag arguments after parsing:
//
//	args := []string{"--host", "example.com", "file1.txt", "file2.txt"}
//	fs.Parse(args)                      // Parses flags, collects remaining args
//
//	remaining := fs.Args()              // Returns ["file1.txt", "file2.txt"]
//	count := fs.NArg()                  // Returns 2
//	first := fs.Arg(0)                  // Returns "file1.txt"
//
// The special "--" separator can be used to force all subsequent arguments to be treated as non-flags:
//
//	args := []string{"--debug", "--", "--not-a-flag", "file.txt"}
//	fs.Parse(args)                      // debug=true, remaining=["--not-a-flag", "file.txt"]
//
// Thread Safety:
//
// FlashFlags uses a sequential-write / concurrent-read model:
//   - Flag registration (String, Int, Bool, ...) and Parse() mutate internal maps
//     and MUST be called from a single goroutine (typically main/init)
//   - After Parse() completes, all flag value reads are safe for concurrent access
//     without any locks -- the underlying maps are never written again
//   - SetValidator() and other mutating methods must also be called before Parse()
//   - Reset() and ResetFlag() are mutating calls too. Calling either one while
//     another goroutine reads is a data race, and the race detector will report
//     it. An application that re-parses at runtime must synchronize on its own,
//     or build a fresh FlagSet and swap it behind a pointer
//
// There are no mutexes and no atomic operations in this package, which is what
// keeps a read down to a plain map lookup. The safety above comes from the
// absence of writes after Parse, not from synchronization.
//
// In short: register flags, call Parse(), then read freely from any goroutine.
//
// Drop-in Replacement for Standard Library flag Package:
//
// Flash-flags provides a complete drop-in replacement for Go's standard library flag package
// through the stdlib subpackage. Simply change your import and get all flash-flags benefits
// with zero code changes:
//
//	// OLD CODE
//	import "flag"
//
//	// NEW CODE (zero changes needed!)
//	import flag "github.com/agilira/flash-flags/stdlib"
//
// All stdlib flag APIs are supported:
//
//	package main
//
//	import (
//		"fmt"
//		flag "github.com/agilira/flash-flags/stdlib"  // Drop-in replacement
//	)
//
//	func main() {
//		// Exactly the same code as stdlib flag!
//		name := flag.String("name", "world", "Name to greet")
//		port := flag.Int("port", 8080, "Server port")
//		debug := flag.Bool("debug", false, "Debug mode")
//
//		flag.Parse()
//
//		fmt.Printf("Hello, %s! Server on port %d (debug: %v)\n", *name, *port, *debug)
//		fmt.Printf("Remaining args: %v\n", flag.Args())  // Full Args() support
//
//		// But you get all flash-flags benefits:
//		// - 1.5x faster parsing
//		// - Short flags: -n, -p, -d
//		// - Combined flags: -np 8080, -d
//		// - Environment variables: NAME=test ./app
//		// - Configuration files: JSON support
//		// - Better help output
//	}
//
// Migration benefits with zero code changes:
//   - Keep existing code unchanged
//   - Gain performance improvements immediately
//   - Access advanced features gradually as needed
//   - Full backward compatibility guaranteed
//
// Basic Usage:
//
//	package main
//
//	import (
//		"fmt"
//		"log"
//		"os"
//		"time"
//
//		"github.com/agilira/flash-flags"
//	)
//
//	func main() {
//		// Create a new flag set
//		fs := flashflags.New("myapp")
//		fs.SetDescription("Production-ready web server")
//		fs.SetVersion("2.1.0")
//
//		// Define flags (all supported types)
//		host := fs.StringVar("host", "h", "localhost", "Server host address")
//		port := fs.IntVar("port", "p", 8080, "Server port number")
//		debug := fs.BoolVar("debug", "d", false, "Enable debug mode")
//		timeout := fs.Duration("timeout", 30*time.Second, "Request timeout")
//		rate := fs.Float64("rate", 1.0, "Request rate limit")
//		tags := fs.StringSlice("tags", []string{}, "Service tags (comma-separated)")
//
//		// Configuration sources (priority: CLI > env > config > defaults)
//		fs.SetEnvPrefix("MYAPP")                    // MYAPP_HOST, MYAPP_PORT, etc.
//		fs.AddConfigPath("./config")                // Auto-discover config files
//		fs.SetConfigFile("./myapp.json")           // Explicit config file
//
//		// Validation and constraints
//		fs.SetValidator("port", func(val interface{}) error {
//			port, ok := val.(int)
//			if !ok {
//				return fmt.Errorf("expected int, got %T", val)
//			}
//			if port < 1024 || port > 65535 {
//				return fmt.Errorf("port must be 1024-65535, got %d", port)
//			}
//			return nil
//		})
//		fs.SetRequired("host")                      // Required flag
//
//		// Organized help output
//		fs.SetGroup("host", "Server Options")
//		fs.SetGroup("port", "Server Options")
//		fs.SetGroup("timeout", "Performance")
//		fs.SetGroup("rate", "Performance")
//		fs.SetGroup("debug", "Debugging")
//
//		// Parse all sources (config file → env vars → CLI args)
//		if err := fs.Parse(os.Args[1:]); err != nil {
//			if err.Error() == "help requested" {
//				os.Exit(0)                          // Help was shown
//			}
//			log.Fatalf("Parse error: %v", err)
//		}
//
//		// Use parsed values (safe to read concurrently now that Parse returned)
//		fmt.Printf("Server: %s:%d (debug=%t, timeout=%v, rate=%.1f)\n",
//			*host, *port, *debug, *timeout, *rate)
//		fmt.Printf("Tags: %v\n", *tags)
//
//		// Your application logic here...
//	}
//
// Configuration File Support:
//
// Flash-flags can load configuration from JSON files. A config file is the
// lowest-priority source above the declared defaults, so both environment
// variables and command-line arguments override it. Use FlagSet.Source to see
// which layer supplied a given value.
//
// JSON has no duration type, so a duration flag accepts either the string form
// understood by time.ParseDuration ("30s", "1m30s") or a plain number of
// nanoseconds.
//
//	fs := flashflags.New("myapp")
//	fs.SetConfigFile("./config.json")
//
//	// or use auto-discovery. Paths are used verbatim: "$HOME/.myapp" would
//	// look for a directory literally named "$HOME", so resolve the home
//	// directory yourself. os.UserHomeDir is the portable call -- HOME is
//	// normally unset on Windows, where the home lives in USERPROFILE.
//	fs.AddConfigPath("./config")
//	if home, err := os.UserHomeDir(); err == nil {
//		fs.AddConfigPath(filepath.Join(home, ".myapp"))
//	}
//
// A FlagSet with neither a config file nor an added path loads no
// configuration: there are no default search directories.
//
// Environment Variable Integration:
//
//	// Enable environment variable lookup
//	fs := flashflags.New("myapp")
//	fs.SetEnvPrefix("MYAPP")                    // MYAPP_HOST, MYAPP_PORT, etc.
//
//	// Or use default naming
//	fs.EnableEnvLookup()                        // HOST, PORT, DEBUG_MODE, etc.
//
//	// Custom environment variable names
//	fs.SetEnvVar("database-url", "DB_CONNECTION_STRING")
//
//	// Priority: CLI args > env vars > config file > defaults
//	// Example: MYAPP_PORT=3000 ./myapp --host=0.0.0.0
//	// Result: host=0.0.0.0 (CLI), port=3000 (env var)
//
// Validation and Constraints:
//
//	fs := flashflags.New("server")
//
//	// Define flags
//	port := fs.IntVar("port", "p", 8080, "Server port")
//	enableTLS := fs.Bool("enable-tls", false, "Enable TLS")
//	tlsCert := fs.String("tls-cert", "", "TLS certificate file")
//	apiKey := fs.String("api-key", "", "API authentication key")
//
//	// Custom validation with detailed error messages
//	fs.SetValidator("port", func(value interface{}) error {
//		port, ok := value.(int)
//		if !ok {
//			return fmt.Errorf("expected int, got %T", value)
//		}
//		if port < 1024 || port > 65535 {
//			return fmt.Errorf("port must be 1024-65535, got %d", port)
//		}
//		return nil
//	})
//
//	// Required flags (must be provided by any config source)
//	fs.SetRequired("api-key")
//
//	// Flag dependencies (cert requires TLS to be enabled)
//	fs.SetDependencies("tls-cert", "enable-tls")
//
//	// All constraints validated automatically during Parse()
//
// Performance and Benchmarks:
//
// Flash-flags delivers exceptional performance:
//
//	Benchmark Results (AMD Ryzen 5 7520U, Go 1.23+, v1.1.5):
//	  Flash-flags:               924 ns/op    (with input screening)
//	  Go standard library flag:  792 ns/op    (baseline, no screening)
//	  Spf13/pflag:             1,322 ns/op    (43% slower than flash-flags)
//	  Other libraries:       7,500+ ns/op    (8-10x slower)
//
//	Screening overhead: roughly 132ns (17%)
//
//	Internal performance metrics (zero allocations):
//	  BenchmarkGetters/GetString  136M    9.01 ns/op   0 B/op   0 allocs/op
//	  BenchmarkGetters/GetInt     142M    8.35 ns/op   0 B/op   0 allocs/op
//	  BenchmarkGetters/GetBool    135M    8.88 ns/op   0 B/op   0 allocs/op
//	  BenchmarkGetters/GetDuration 134M   8.86 ns/op   0 B/op   0 allocs/op
//
// Key performance characteristics:
//   - 924ns with input screening enabled
//   - 43% faster than pflag, with equivalent functionality
//   - Sub-nanosecond flag value access (8-9ns average)
//   - Zero allocations for all getter operations after parsing
//   - Concurrent-safe reads after Parse() (no locks needed at runtime)
//   - Hash-based O(1) flag lookup with minimal overhead
//   - Full support for remaining arguments with minimal overhead
//   - Fast-path optimization for simple alphanumeric inputs (bypasses heavy validation)
//
// Performance trade-offs:
//   - 17% slower than stdlib but gains: security, short flags, config files, env vars, validation
//   - 43% faster than pflag while providing more features and better security
//   - Optimal for production applications requiring security without sacrificing performance
//
// Compatibility and Requirements:
//
//   - Go 1.25.9 or later
//   - Zero external dependencies
//   - Full backward compatibility maintained
//   - Drop-in replacement for standard library flag package
//
// Error Handling:
//
// FlashFlags returns descriptive errors for various scenarios:
//
//   - Parse errors: "unknown flag: --invalid", "flag --port requires a value"
//   - Validation errors: "validation failed for flag --port: port must be between 1-65535"
//   - Required flag errors: "required flag --api-key not provided"
//   - Dependency errors: "flag --tls-cert requires --enable-tls to be set"
//   - Type conversion errors: "invalid int value for flag --port: abc"
//   - Configuration errors: "config file error: failed to read config.json"
//   - Help requests: "help requested" (special case, not a real error)
//   - Input screening errors: "flag --name contains null byte at position 3"
//   - Buffer overflow errors: "flag --data value too long: 15000 chars (max: 10000)"
//
// All errors include the flag name and specific details to help with debugging.
//
// Input Screening:
//
// See "Input Hardening" above for what is checked and what is deliberately not.
//
// Example:
//
//	fs := flashflags.New("myapp")
//	cmd := fs.String("command", "", "Command to execute")
//
//	// Rejected -- malformed as a string:
//	fs.Parse([]string{"--command", "bad\x00value"})  // null byte
//	fs.Parse([]string{"--command", "\x1b[31mred"})   // control character
//
//	// Accepted -- the parser does not guess what these mean:
//	fs.Parse([]string{"--command", "rm -rf /tmp/build"})
//	fs.Parse([]string{"--command", "/etc/myapp.conf"})
//	fs.Parse([]string{"--command", "a; rm -rf ~"})
//
// Version and Compatibility:
//
//   - Current version: v1.1.9
//   - Requires: Go 1.25.9 or later
//   - Changelog: See changelog/ for release notes
//   - Repository: github.com/agilira/flash-flags
//   - License: MPL-2.0 (Mozilla Public License 2.0)
//
// Copyright (c) 2025 AGILira
// Series: an AGILira library
// SPDX-License-Identifier: MPL-2.0
package flashflags
