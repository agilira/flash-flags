// flash-flags/demo/config: Ultra-fast command-line flag parsing for Go - config example
//
// Copyright (c) 2025 AGILira - A. Giordano
// Series: an AGILira library
// SPDX-License-Identifier: MPL-2.0

package main

import (
	"fmt"
	"log"
	"os"

	flashflags "github.com/agilira/flash-flags"
)

func main() {
	fmt.Println("=== Configuration Files Demo ===")

	// Create a sample config file
	configContent := `{
		"host": "config.example.com",
		"port": 9090,
		"ssl": true,
		"workers": 4,
		"log-level": "debug",
		"timeout": 30.5,
		"features": ["auth", "metrics", "cache"]
	}`

	configFile := "demo-config.json"
	if err := os.WriteFile(configFile, []byte(configContent), 0600); err != nil { // More restrictive permissions
		log.Fatalf("Failed to create config file: %v", err)
	}
	defer os.Remove(configFile) // Clean up

	fmt.Println("\n1. Created sample config file:")
	fmt.Println("----------------------------------------")
	fmt.Println(configContent)

	// Setup flag set
	fs := flashflags.New("myapp")
	fs.SetDescription("Demo application with configuration file support")
	fs.SetVersion("1.2.0")

	// Define flags with defaults
	host := fs.StringVar("host", "H", "localhost", "Server host")
	port := fs.IntVar("port", "p", 8080, "Server port")
	ssl := fs.BoolVar("ssl", "s", false, "Enable SSL")
	workers := fs.Int("workers", 2, "Number of workers")
	logLevel := fs.String("log-level", "info", "Log level")
	timeout := fs.Float64("timeout", 10.0, "Timeout in seconds")
	features := fs.StringSlice("features", []string{}, "Enabled features")

	// Set config file
	fs.SetConfigFile(configFile)

	// Test 1: Parse with config only
	fmt.Println("\n2. Values after loading config (no command line args):")
	fmt.Println("----------------------------------------")

	if err := fs.Parse([]string{}); err != nil {
		log.Fatalf("Parse failed: %v", err)
	}

	fmt.Printf("host: %s (from config)\n", *host)
	fmt.Printf("port: %d (from config)\n", *port)
	fmt.Printf("ssl: %t (from config)\n", *ssl)
	fmt.Printf("workers: %d (from config)\n", *workers)
	fmt.Printf("log-level: %s (from config)\n", *logLevel)
	fmt.Printf("timeout: %.1f (from config)\n", *timeout)
	fmt.Printf("features: %v (from config)\n", *features)

	// Test 2: Reset and parse with command line override
	fmt.Println("\n3. Values after command line override:")
	fmt.Println("----------------------------------------")

	fs.Reset() // Reset to test override behavior

	args := []string{
		"--host", "cmdline.example.com",
		"-p", "3000",
		"--workers", "8",
	}

	if err := fs.Parse(args); err != nil {
		log.Fatalf("Parse with override failed: %v", err)
	}

	fmt.Printf("host: %s (from command line - overrides config)\n", *host)
	fmt.Printf("port: %d (from command line - overrides config)\n", *port)
	fmt.Printf("ssl: %t (from config - not overridden)\n", *ssl)
	fmt.Printf("workers: %d (from command line - overrides config)\n", *workers)
	fmt.Printf("log-level: %s (from config - not overridden)\n", *logLevel)
	fmt.Printf("timeout: %.1f (from config - not overridden)\n", *timeout)
	fmt.Printf("features: %v (from config - not overridden)\n", *features)

	// Test 3: Show priority order
	fmt.Println("\n4. Priority demonstration:")
	fmt.Println("----------------------------------------")
	fmt.Println("Value priority order (highest to lowest):")
	fmt.Println("1. Command line arguments")
	fmt.Println("2. Configuration file")
	fmt.Println("3. Default values")
	fmt.Println("")
	fmt.Println("In this demo:")
	fmt.Println("- host, port, workers came from command line (highest priority)")
	fmt.Println("- ssl, log-level, timeout, features came from config file")

	// Test 4: Show config file auto-discovery
	fmt.Println("\n5. Config file auto-discovery:")
	fmt.Println("----------------------------------------")
	fmt.Println("The library searches for config files in this order:")
	fmt.Println("- Explicit file set with SetConfigFile()")
	fmt.Println("- Auto-discovery: {appname}.json, {appname}.config.json, config.json")
	fmt.Println("- Search paths: current dir, ./config, $HOME (or custom paths)")

	fmt.Println("\n=== Demo Complete ===")
}
