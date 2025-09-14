// flash-flags/demo/required: Ultra-fast command-line flag parsing for Go - required flags example
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
	fs := createFlagSet()
	flags := registerFlags(fs)
	setupRequiredFlags(fs)
	setupDependencies(fs)
	setupGroups(fs)

	parseArguments(fs)
	displayConfiguration(flags)
}

// createFlagSet creates and configures the basic flag set
func createFlagSet() *flashflags.FlagSet {
	fs := flashflags.New("required-demo")
	fs.SetDescription("Demo application showing required flags functionality")
	fs.SetVersion("v1.0.0")
	return fs
}

// flagPointers holds all the flag pointers for easy access
type flagPointers struct {
	apiKey     *string
	host       *string
	port       *int
	enableTLS  *bool
	tlsCert    *string
	tlsKey     *string
	dbURL      *string
	dbPassword *string
}

// registerFlags registers all application flags
func registerFlags(fs *flashflags.FlagSet) *flagPointers {
	return &flagPointers{
		apiKey:     fs.String("api-key", "", "API key for authentication (required)"),
		host:       fs.String("host", "localhost", "Server host (optional)"),
		port:       fs.Int("port", 8080, "Server port (optional)"),
		enableTLS:  fs.Bool("enable-tls", false, "Enable TLS encryption (optional)"),
		tlsCert:    fs.String("tls-cert", "", "TLS certificate file (required if TLS enabled)"),
		tlsKey:     fs.String("tls-key", "", "TLS private key file (required if TLS enabled)"),
		dbURL:      fs.String("db-url", "", "Database connection URL (required)"),
		dbPassword: fs.String("db-password", "", "Database password (required if db-url provided)"),
	}
}

// setupRequiredFlags marks the required flags
func setupRequiredFlags(fs *flashflags.FlagSet) {
	requiredFlags := []string{"api-key", "db-url"}
	for _, flag := range requiredFlags {
		if err := fs.SetRequired(flag); err != nil {
			fmt.Printf("Error setting %s as required: %v\n", flag, err)
			os.Exit(1)
		}
	}
}

// setupDependencies configures flag dependencies
func setupDependencies(fs *flashflags.FlagSet) {
	dependencies := map[string]string{
		"tls-cert":    "enable-tls",
		"tls-key":     "enable-tls",
		"db-password": "db-url",
	}

	for flag, dependency := range dependencies {
		if err := fs.SetDependencies(flag, dependency); err != nil {
			fmt.Printf("Error setting %s dependencies: %v\n", flag, err)
			os.Exit(1)
		}
	}
}

// setupGroups organizes flags into groups for help output
func setupGroups(fs *flashflags.FlagSet) {
	groups := map[string]string{
		"api-key":     "Authentication",
		"host":        "Server Options",
		"port":        "Server Options",
		"enable-tls":  "TLS Options",
		"tls-cert":    "TLS Options",
		"tls-key":     "TLS Options",
		"db-url":      "Database Options",
		"db-password": "Database Options",
	}

	for flag, group := range groups {
		fs.SetGroup(flag, group) // #nosec G104
	}
}

// parseArguments parses command line arguments with error handling
func parseArguments(fs *flashflags.FlagSet) {
	if err := fs.Parse(os.Args[1:]); err != nil {
		if err.Error() == "help requested" {
			os.Exit(0) // Help was shown
		}

		fmt.Printf("❌ Configuration Error: %v\n\n", err)
		fmt.Println("💡 Run with --help to see all available options")
		os.Exit(1)
	}
}

// displayConfiguration shows the final configuration
func displayConfiguration(flags *flagPointers) {
	fmt.Println("✅ All required flags provided successfully!")
	fmt.Println()
	fmt.Println("🚀 Starting application with configuration:")

	displayServerConfig(flags)
	displayTLSConfig(flags)
	displayDatabaseConfig(flags)

	fmt.Println()
	fmt.Println("🎯 Application would start here with the provided configuration!")
}

// displayServerConfig shows server-related configuration
func displayServerConfig(flags *flagPointers) {
	fmt.Printf("   API Key: %s\n", maskSecret(*flags.apiKey))
	fmt.Printf("   Server: %s:%d\n", *flags.host, *flags.port)
}

// displayTLSConfig shows TLS-related configuration
func displayTLSConfig(flags *flagPointers) {
	if *flags.enableTLS {
		fmt.Printf("   TLS Enabled: ✅\n")
		fmt.Printf("   TLS Cert: %s\n", *flags.tlsCert)
		fmt.Printf("   TLS Key: %s\n", *flags.tlsKey)
	} else {
		fmt.Printf("   TLS Enabled: ❌\n")
	}
}

// displayDatabaseConfig shows database-related configuration
func displayDatabaseConfig(flags *flagPointers) {
	fmt.Printf("   Database URL: %s\n", maskConnectionString(*flags.dbURL))
	if *flags.dbPassword != "" {
		fmt.Printf("   Database Password: %s\n", maskSecret(*flags.dbPassword))
	}
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
