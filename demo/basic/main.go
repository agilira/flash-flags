// flash-flags/demo/basic: Ultra-fast command-line flag parsing for Go - basic example
//
// Copyright (c) 2025 AGILira
// Series: an AGILira library
// SPDX-License-Identifier: MPL-2.0

package main

import (
	"fmt"
	"os"

	flashflags "github.com/agilira/flash-flags"
)

func main() {
	fs := flashflags.New("webserver")
	fs.SetDescription("A fast web server with advanced configuration options")
	fs.SetVersion("2.1.0")

	// Server configuration group
	host := fs.StringVar("host", "H", "localhost", "Server host address")
	port := fs.IntVar("port", "p", 8080, "Server port")
	ssl := fs.BoolVar("ssl", "s", false, "Enable SSL/TLS")
	certFile := fs.String("cert", "", "SSL certificate file")
	keyFile := fs.String("key", "", "SSL private key file")

	// Database configuration group
	dbHost := fs.String("db-host", "localhost", "Database host")
	dbPort := fs.Int("db-port", 5432, "Database port")
	dbName := fs.String("db-name", "app", "Database name")
	dbUser := fs.String("db-user", "", "Database user")
	_ = fs.String("db-password", "", "Database password")

	// Logging configuration group
	logLevel := fs.String("log-level", "info", "Log level (debug, info, warn, error)")
	logFile := fs.String("log-file", "", "Log file path (empty for stdout)")
	verbose := fs.BoolVar("verbose", "v", false, "Enable verbose logging")

	// Set groups for better help organization
	fs.SetGroup("host", "Server Options")
	fs.SetGroup("port", "Server Options")
	fs.SetGroup("ssl", "Server Options")
	fs.SetGroup("cert", "SSL Options")
	fs.SetGroup("key", "SSL Options")

	fs.SetGroup("db-host", "Database Options")
	fs.SetGroup("db-port", "Database Options")
	fs.SetGroup("db-name", "Database Options")
	fs.SetGroup("db-user", "Database Options")
	fs.SetGroup("db-password", "Database Options")

	fs.SetGroup("log-level", "Logging Options")
	fs.SetGroup("log-file", "Logging Options")
	fs.SetGroup("verbose", "Logging Options")

	// Set required flags
	fs.SetRequired("db-user")
	fs.SetRequired("db-password")

	// Set dependencies
	fs.SetDependencies("cert", "ssl") // cert requires ssl
	fs.SetDependencies("key", "ssl")  // key requires ssl

	// Set validation for log level
	fs.SetValidator("log-level", func(val interface{}) error {
		level := val.(string)
		validLevels := []string{"debug", "info", "warn", "error"}
		for _, valid := range validLevels {
			if level == valid {
				return nil
			}
		}
		return fmt.Errorf("invalid log level '%s', must be one of: debug, info, warn, error", level)
	})

	// Set validation for port range
	fs.SetValidator("port", func(val interface{}) error {
		port := val.(int)
		if port < 1 || port > 65535 {
			return fmt.Errorf("port must be between 1 and 65535, got %d", port)
		}
		return nil
	})

	fmt.Println("=== Help System Demo ===")

	// Show help first
	fmt.Println("\n1. Generated Help Output:")
	fmt.Println("----------------------------------------")
	fs.PrintHelp()

	// Test parsing some arguments
	fmt.Println("\n2. Testing argument parsing:")
	fmt.Println("----------------------------------------")

	// Example valid configuration
	args := []string{
		"--host", "0.0.0.0",
		"-p", "8443",
		"--ssl",
		"--cert", "/etc/ssl/server.crt",
		"--key", "/etc/ssl/server.key",
		"--db-user", "admin",
		"--db-password", "secret123",
		"--log-level", "debug",
		"-v",
	}

	if err := fs.Parse(args); err != nil {
		fmt.Printf("❌ Parse error: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("✅ Configuration parsed successfully!\n")
	fmt.Printf("   Server: %s:%d (SSL: %t)\n", *host, *port, *ssl)
	fmt.Printf("   SSL Cert: %s, Key: %s\n", *certFile, *keyFile)
	fmt.Printf("   Database: %s@%s:%d/%s\n", *dbUser, *dbHost, *dbPort, *dbName)
	fmt.Printf("   Logging: level=%s, verbose=%t\n", *logLevel, *verbose)

	if *logFile != "" {
		fmt.Printf("   Log file: %s\n", *logFile)
	}
}
