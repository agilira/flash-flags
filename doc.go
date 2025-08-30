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
//   - Lock-free design for concurrent access
//   - Support for configuration files (JSON)
//   - Environment variable integration
//   - Flag validation and constraints
//   - Grouped help output
//   - Short and long flag support
//
// Basic Usage:
//
//	package main
//
//	import (
//		"fmt"
//		"log"
//		"os"
//
//		"github.com/agilira/flash-flags"
//	)
//
//	func main() {
//		// Create a new flag set
//		fs := flashflags.New("myapp")
//		fs.SetDescription("My awesome application")
//		fs.SetVersion("1.0.0")
//
//		// Define flags
//		host := fs.StringVar("host", "h", "localhost", "Server host address")
//		port := fs.IntVar("port", "p", 8080, "Server port number")
//		debug := fs.BoolVar("debug", "d", false, "Enable debug mode")
//
//		// Parse command line arguments
//		if err := fs.Parse(os.Args[1:]); err != nil {
//			if err.Error() == "help requested" {
//				os.Exit(0)
//			}
//			log.Fatal(err)
//		}
//
//		fmt.Printf("Starting server on %s:%d (debug: %v)\n", *host, *port, *debug)
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
// Flags can be set via environment variables with automatic naming conversion:
//
//	fs := flashflags.New("myapp")
//	fs.SetEnvPrefix("MYAPP")  // MYAPP_HOST, MYAPP_PORT, etc.
//	// or use default naming (HOST, PORT, etc.)
//	fs.EnableEnvLookup()
//
// Flag Validation and Constraints:
//
// Flash-flags supports custom validation, required flags, and dependencies:
//
//	fs := flashflags.New("myapp")
//	port := fs.IntVar("port", "p", 8080, "Server port")
//
//	// Custom validation
//	fs.SetValidator("port", func(value interface{}) error {
//		port := value.(int)
//		if port < 1 || port > 65535 {
//			return fmt.Errorf("port must be between 1 and 65535")
//		}
//		return nil
//	})
//
//	// Required flags
//	fs.SetRequired("host")
//
//	// Dependencies
//	fs.SetDependencies("ssl-cert", "ssl-key")
//
// Performance Considerations:
//
// Flash-flags is optimized for performance:
//
//   - Minimal memory allocations during parsing
//   - Lock-free design for concurrent access
//   - Optimized string operations
//   - Efficient flag lookup using maps
//
// The library is extracted from the Argus configuration management system
// with the same high-performance characteristics.
//
// Copyright (c) 2025 AGILira
// Series: an AGILira library
// SPDX-License-Identifier: MPL-2.0
package flashflags
