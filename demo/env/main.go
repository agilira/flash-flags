// flash-flags/demo/env: Ultra-fast command-line flag parsing for Go - env example
//
// Copyright (c) 2025 AGILira - A. Giordano
// Series: an AGILira library
// SPDX-License-Identifier: MPL-2.0

package main

import (
	"fmt"
	"os"

	flashflags "github.com/agilira/flash-flags"
)

func main() {
	fmt.Println("=== Environment Variables Demo ===")

	// Create flag set with environment support
	fs := flashflags.New("envdemo")
	fs.SetDescription("Demo application showing environment variable support")
	fs.SetVersion("1.0.0")

	// Define flags
	host := fs.String("host", "localhost", "Server host")
	port := fs.Int("port", 8080, "Server port")
	debug := fs.Bool("debug", false, "Debug mode")
	maxConns := fs.Int("max-connections", 100, "Maximum connections")

	// Demo 1: Using environment prefix
	fmt.Println("\n1. Environment Variables with Prefix:")
	fmt.Println("----------------------------------------")
	fmt.Println("Setting environment variables:")
	fmt.Println("  export ENVDEMO_HOST=env.example.com")
	fmt.Println("  export ENVDEMO_PORT=9090")
	fmt.Println("  export ENVDEMO_DEBUG=true")

	// Set environment variables for demo
	if err := os.Setenv("ENVDEMO_HOST", "env.example.com"); err != nil {
		fmt.Printf("Error setting ENVDEMO_HOST: %v\n", err)
		return
	}
	if err := os.Setenv("ENVDEMO_PORT", "9090"); err != nil {
		fmt.Printf("Error setting ENVDEMO_PORT: %v\n", err)
		return
	}
	if err := os.Setenv("ENVDEMO_DEBUG", "true"); err != nil {
		fmt.Printf("Error setting ENVDEMO_DEBUG: %v\n", err)
		return
	}

	// Enable environment lookup with prefix
	fs.SetEnvPrefix("ENVDEMO")

	err := fs.Parse([]string{})
	if err != nil {
		fmt.Printf("Parse error: %v\n", err)
		return
	}

	fmt.Printf("Values loaded from environment:\n")
	fmt.Printf("  host: %s\n", *host)
	fmt.Printf("  port: %d\n", *port)
	fmt.Printf("  debug: %t\n", *debug)
	fmt.Printf("  max-connections: %d (default, no env var set)\n", *maxConns)

	// Demo 2: Command line override
	fmt.Println("\n2. Command Line Override:")
	fmt.Println("----------------------------------------")
	fmt.Println("Command: --host override.example.com --port 3000")

	fs.Reset() // Reset to test override

	err = fs.Parse([]string{"--host", "override.example.com", "--port", "3000"})
	if err != nil {
		fmt.Printf("Parse error: %v\n", err)
		return
	}

	fmt.Printf("Values after command line override:\n")
	fmt.Printf("  host: %s (from command line)\n", *host)
	fmt.Printf("  port: %d (from command line)\n", *port)
	fmt.Printf("  debug: %t (from environment - not overridden)\n", *debug)

	// Demo 3: Custom environment variable names
	fmt.Println("\n3. Custom Environment Variable Names:")
	fmt.Println("----------------------------------------")

	// Create new flag set for custom env vars demo
	fs2 := flashflags.New("envdemo2")
	fs2.EnableEnvLookup()

	database := fs2.String("database-url", "localhost:5432", "Database URL")

	// Set custom environment variable name
	if err := fs2.SetEnvVar("database-url", "DATABASE_CONNECTION_STRING"); err != nil {
		fmt.Printf("Error setting custom env var: %v\n", err)
		return
	}

	fmt.Println("Setting custom environment variable:")
	fmt.Println("  export DATABASE_CONNECTION_STRING=postgres://user:pass@db.example.com:5432/mydb")

	if err := os.Setenv("DATABASE_CONNECTION_STRING", "postgres://user:pass@db.example.com:5432/mydb"); err != nil {
		fmt.Printf("Error setting DATABASE_CONNECTION_STRING: %v\n", err)
		return
	}

	err = fs2.Parse([]string{})
	if err != nil {
		fmt.Printf("Parse error: %v\n", err)
		return
	}

	fmt.Printf("Value from custom env var:\n")
	fmt.Printf("  database-url: %s\n", *database)

	// Demo 4: Default naming convention
	fmt.Println("\n4. Default Naming Convention:")
	fmt.Println("----------------------------------------")

	fs3 := flashflags.New("envdemo3")
	fs3.EnableEnvLookup() // No prefix - uses default naming

	dbHost := fs3.String("db-host", "localhost", "Database host")
	logLevel := fs3.String("log-level", "info", "Log level")

	fmt.Println("Setting environment variables with default naming:")
	fmt.Println("  export DB_HOST=db.example.com")
	fmt.Println("  export LOG_LEVEL=debug")

	if err := os.Setenv("DB_HOST", "db.example.com"); err != nil {
		fmt.Printf("Error setting DB_HOST: %v\n", err)
		return
	}
	if err := os.Setenv("LOG_LEVEL", "debug"); err != nil {
		fmt.Printf("Error setting LOG_LEVEL: %v\n", err)
		return
	}

	err = fs3.Parse([]string{})
	if err != nil {
		fmt.Printf("Parse error: %v\n", err)
		return
	}

	fmt.Printf("Values from default naming:\n")
	fmt.Printf("  db-host: %s (from DB_HOST)\n", *dbHost)
	fmt.Printf("  log-level: %s (from LOG_LEVEL)\n", *logLevel)

	// Demo 5: Priority order
	fmt.Println("\n5. Priority Order Demonstration:")
	fmt.Println("----------------------------------------")
	fmt.Println("Priority (highest to lowest):")
	fmt.Println("  1. Command line arguments")
	fmt.Println("  2. Environment variables")
	fmt.Println("  3. Configuration files")
	fmt.Println("  4. Default values")

	// Clean up environment variables
	os.Unsetenv("ENVDEMO_HOST")
	os.Unsetenv("ENVDEMO_PORT")
	os.Unsetenv("ENVDEMO_DEBUG")
	os.Unsetenv("DATABASE_CONNECTION_STRING")
	os.Unsetenv("DB_HOST")
	os.Unsetenv("LOG_LEVEL")

	fmt.Println("\n=== Demo Complete ===")
}
