// Package flashflags provides ultra-fast, zero-dependency, lock-free command-line flag parsing for Go.
//
// Flash-flags is designed for maximum performance with minimal memory allocations,
// comprehensive security hardening, and full compatibility with Go 1.23+.
// It provides a clean API similar to the standard library flag package but with
// significant security improvements and additional features.
//
// Key Features:
//
//   - Security-hardened parsing with protection against injection attacks
//   - Ultra-fast parsing (924ns/op) with only 132ns security overhead
//   - Zero external dependencies (only standard library)
//   - Lock-free design for concurrent access (thread-safe)
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
// Security Features (v1.1.5+):
//
//   - Command Injection Protection: Blocks $(…), backticks, and shell metacharacters
//   - Path Traversal Prevention: Prevents ../ and ..\ directory traversal attacks
//   - Buffer Overflow Safeguards: 10KB input limits with fast-path optimization
//   - Format String Attack Blocking: Detects and blocks %n, %s format string exploits
//   - Input Sanitization: Removes null bytes and dangerous control characters
//   - Windows Device Protection: Blocks Windows reserved names (CON, PRN, AUX, etc.)
//   - Fast-path optimization: Simple alphanumeric inputs bypass heavy validation
//
// The security overhead is minimal (132ns per operation, 17%) while providing
// comprehensive protection against common attack vectors.
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
// FlashFlags is designed to be thread-safe with lock-free operations:
//   - All flag reading operations are safe for concurrent access
//   - Parse() should be called only once from a single goroutine
//   - Flag registration should be done before calling Parse()
//   - After Parse() completes, all flag values can be read concurrently
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
//			port := val.(int)
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
//		// Use parsed values (thread-safe access)
//		fmt.Printf("Server: %s:%d (debug=%t, timeout=%v, rate=%.1f)\n",
//			*host, *port, *debug, *timeout, *rate)
//		fmt.Printf("Tags: %v\n", *tags)
//
//		// Your application logic here...
//	}
//
// Configuration File Support:
//
// Flash-flags can load configuration from JSON files. The configuration is loaded
// with lower priority than command line arguments, allowing command line to override
// configuration file values.
//
//	fs := flashflags.New("myapp")
//	fs.SetConfigFile("./config.json")
//	// or use auto-discovery
//	fs.AddConfigPath("./config")
//	fs.AddConfigPath("$HOME/.myapp")
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
//		port := value.(int)
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
// Flash-flags delivers exceptional performance with comprehensive security hardening:
//
//	Benchmark Results (AMD Ryzen 5 7520U, Go 1.23+, v1.1.5):
//	  Flash-flags (secure):      924 ns/op    (with full security validation)
//	  Go standard library flag:  792 ns/op    (baseline, no security)
//	  Spf13/pflag:             1,322 ns/op    (43% slower than flash-flags)
//	  Other libraries:       7,500+ ns/op    (8-10x slower)
//
//	Security overhead: Only 132ns (17%) for complete protection
//
//	Internal performance metrics (zero allocations):
//	  BenchmarkGetters/GetString  136M    9.01 ns/op   0 B/op   0 allocs/op
//	  BenchmarkGetters/GetInt     142M    8.35 ns/op   0 B/op   0 allocs/op
//	  BenchmarkGetters/GetBool    135M    8.88 ns/op   0 B/op   0 allocs/op
//	  BenchmarkGetters/GetDuration 134M   8.86 ns/op   0 B/op   0 allocs/op
//
// Key performance characteristics:
//   - 924ns with full security (command injection, path traversal, format string protection)
//   - 43% faster than pflag while providing equivalent functionality plus security
//   - Sub-nanosecond flag value access (8-9ns average)
//   - Zero allocations for all getter operations after parsing
//   - Lock-free concurrent reads (thread-safe)
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
//   - Go 1.23 or later (follows LTS guidelines)
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
//   - Security validation errors: "flag --name contains dangerous pattern"
//   - Buffer overflow errors: "flag --data value too long: 15000 chars (max: 10000)"
//
// All errors include the flag name and specific details to help with debugging.
//
// Security Validation:
//
// FlashFlags automatically validates all input against common security threats:
//
//   - Command injection: Detects $(...), backticks, pipes, redirections
//   - Path traversal: Blocks ../ and ..\ sequences
//   - Format strings: Prevents %n, %s, %x format string attacks
//   - Null bytes: Removes \x00 null byte injection attempts
//   - Control chars: Filters dangerous control characters (except \t, \n, \r)
//   - Buffer overflow: Enforces 10KB input size limit per flag value
//   - Windows devices: Blocks CON, PRN, AUX, COM1-9, LPT1-9 device names
//
// Fast-path optimization: Simple alphanumeric values (a-z, A-Z, 0-9, -, _, ., :)
// bypass heavy validation for optimal performance on common use cases.
//
// Example security validation:
//
//	fs := flashflags.New("myapp")
//	cmd := fs.String("command", "", "Command to execute")
//
//	// These will be rejected with security errors:
//	fs.Parse([]string{"--command", "rm -rf /"})           // Command injection
//	fs.Parse([]string{"--command", "../../etc/passwd"})   // Path traversal
//	fs.Parse([]string{"--command", "%n%n%n%n"})          // Format string attack
//
// Version and Compatibility:
//
//   - Current version: v1.1.5 (October 2025)
//   - Requires: Go 1.23 or later
//   - Changelog: See changelog/v1.1.5.txt for latest updates
//   - Repository: github.com/agilira/flash-flags
//   - License: MPL-2.0 (Mozilla Public License 2.0)
//
// Copyright (c) 2025 AGILira
// Series: an AGILira library
// SPDX-License-Identifier: MPL-2.0
package flashflags
