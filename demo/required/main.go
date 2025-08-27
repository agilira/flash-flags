// flash-flags/demo/required: Ultra-fast command-line flag parsing for Go - required flags example
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
	// Create flag set
	fs := flashflags.New("required-demo")
	fs.SetDescription("Demo application showing required flags functionality")
	fs.SetVersion("v1.0.0")

	// Register flags - some required, some optional
	apiKey := fs.String("api-key", "", "API key for authentication (required)")
	host := fs.String("host", "localhost", "Server host (optional)")
	port := fs.Int("port", 8080, "Server port (optional)")
	enableTLS := fs.Bool("enable-tls", false, "Enable TLS encryption (optional)")

	// TLS related flags (required only if TLS is enabled)
	tlsCert := fs.String("tls-cert", "", "TLS certificate file (required if TLS enabled)")
	tlsKey := fs.String("tls-key", "", "TLS private key file (required if TLS enabled)")

	// Database configuration
	dbURL := fs.String("db-url", "", "Database connection URL (required)")
	dbPassword := fs.String("db-password", "", "Database password (required if db-url provided)")

	// Mark required flags
	if err := fs.SetRequired("api-key"); err != nil {
		fmt.Printf("Error setting api-key as required: %v\n", err)
		os.Exit(1)
	}

	if err := fs.SetRequired("db-url"); err != nil {
		fmt.Printf("Error setting db-url as required: %v\n", err)
		os.Exit(1)
	}

	// Set up dependencies - these flags are required only if their dependencies are set
	if err := fs.SetDependencies("tls-cert", "enable-tls"); err != nil {
		fmt.Printf("Error setting tls-cert dependencies: %v\n", err)
		os.Exit(1)
	}

	if err := fs.SetDependencies("tls-key", "enable-tls"); err != nil {
		fmt.Printf("Error setting tls-key dependencies: %v\n", err)
		os.Exit(1)
	}

	if err := fs.SetDependencies("db-password", "db-url"); err != nil {
		fmt.Printf("Error setting db-password dependencies: %v\n", err)
		os.Exit(1)
	}

	// Organize flags in groups for better help output
	fs.SetGroup("api-key", "Authentication")       // #nosec G104
	fs.SetGroup("host", "Server Options")          // #nosec G104
	fs.SetGroup("port", "Server Options")          // #nosec G104
	fs.SetGroup("enable-tls", "TLS Options")       // #nosec G104
	fs.SetGroup("tls-cert", "TLS Options")         // #nosec G104
	fs.SetGroup("tls-key", "TLS Options")          // #nosec G104
	fs.SetGroup("db-url", "Database Options")      // #nosec G104
	fs.SetGroup("db-password", "Database Options") // #nosec G104

	// Parse command line arguments
	if err := fs.Parse(os.Args[1:]); err != nil {
		if err.Error() == "help requested" {
			os.Exit(0) // Help was shown
		}

		// Show specific error messages for different types of validation failures
		fmt.Printf("❌ Configuration Error: %v\n\n", err)

		// Show help after error for user convenience
		fmt.Println("💡 Run with --help to see all available options")
		os.Exit(1)
	}

	// If we get here, all required flags and dependencies are satisfied
	fmt.Println("✅ All required flags provided successfully!")
	fmt.Println()

	// Display the configuration
	fmt.Println("🚀 Starting application with configuration:")
	fmt.Printf("   API Key: %s\n", maskSecret(*apiKey))
	fmt.Printf("   Server: %s:%d\n", *host, *port)

	if *enableTLS {
		fmt.Printf("   TLS Enabled: ✅\n")
		fmt.Printf("   TLS Cert: %s\n", *tlsCert)
		fmt.Printf("   TLS Key: %s\n", *tlsKey)
	} else {
		fmt.Printf("   TLS Enabled: ❌\n")
	}

	fmt.Printf("   Database URL: %s\n", maskConnectionString(*dbURL))
	if *dbPassword != "" {
		fmt.Printf("   Database Password: %s\n", maskSecret(*dbPassword))
	}

	fmt.Println()
	fmt.Println("🎯 Application would start here with the provided configuration!")
}

// Helper function to mask sensitive information in output
func maskSecret(secret string) string {
	if len(secret) <= 8 {
		return "****"
	}
	return secret[:4] + "****" + secret[len(secret)-4:]
}

// Helper function to mask connection strings
func maskConnectionString(url string) string {
	if url == "" {
		return ""
	}
	// Simple masking - in real apps you'd want more sophisticated URL parsing
	if len(url) <= 20 {
		return "****"
	}
	return url[:10] + "****" + url[len(url)-6:]
}
