// Package flashflags provides ultra-fast, zero-dependency, lock-free command-line flag parsing for Go.
//
// Flash-flags is designed for maximum performance with minimal memory allocations.
// It provides a clean API similar to the standard library flag package but with
// significant performance improvements and additional features.
//
// Key Features:
//
//   - Ultra-fast parsing with optimized memory allocations
//   - Zero external dependencies (only standard library)
//   - Lock-free design for concurrent access (thread-safe)
//   - Support for configuration files (JSON)
//   - Environment variable integration
//   - Flag validation and constraints
//   - Grouped help output
//   - Comprehensive flag syntax support (POSIX/GNU-style)
//   - Short and long flag support with combined syntax
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
// Thread Safety:
//
// FlashFlags is designed to be thread-safe with lock-free operations:
//   - All flag reading operations are safe for concurrent access
//   - Parse() should be called only once from a single goroutine
//   - Flag registration should be done before calling Parse()
//   - After Parse() completes, all flag values can be read concurrently
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
// Flash-flags delivers exceptional performance with minimal overhead:
//
//	Benchmark Results (AMD Ryzen 5 7520U, Go 1.23):
//	  BenchmarkParse-8           742,216   1,471 ns/op   1,520 B/op   25 allocs/op
//	  BenchmarkGetters/GetString 121M      8.58 ns/op        0 B/op    0 allocs/op
//	  BenchmarkGetters/GetInt    150M      7.56 ns/op        0 B/op    0 allocs/op
//	  BenchmarkGetters/GetBool   147M      8.34 ns/op        0 B/op    0 allocs/op
//	  BenchmarkGetters/GetDuration 141M    8.13 ns/op        0 B/op    0 allocs/op
//
// Key performance characteristics:
//   - ~1.47μs total parse time (5 flags including validation and config loading)
//   - Sub-nanosecond flag value access (7-8ns average)
//   - Zero allocations for all getter operations after parsing
//   - Lock-free concurrent reads (thread-safe)
//   - Hash-based O(1) flag lookup with minimal overhead
//
// Memory efficiency:
//   - 1,520 bytes total allocation for complete parsing cycle
//   - 25 allocations total (setup phase only, no runtime allocations)
//   - Constant memory footprint regardless of flag access frequency
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
//
// All errors include the flag name and specific details to help with debugging.
//
// Copyright (c) 2025 AGILira
// Series: an AGILira library
// SPDX-License-Identifier: MPL-2.0
package flashflags
