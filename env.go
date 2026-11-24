// env.go: environment variable lookup
//
// Copyright (c) 2025 AGILira - A. Giordano
// Series: an AGILira library
// SPDX-License-Identifier: MPL-2.0

package flashflags

import (
	"fmt"
	"os"
	"strings"
)

// SetEnvPrefix sets the prefix for environment variable lookup and enables env var processing.
// Flag names are automatically converted using the pattern: PREFIX_FLAGNAME
//
// Conversion rules:
//   - Hyphens become underscores: "db-host" → "PREFIX_DB_HOST"
//   - All uppercase: "api-key" → "PREFIX_API_KEY"
//   - Original case preserved in prefix: "MyApp" → "MyApp_DB_HOST"
//
// Example:
//
//	fs := flashflags.New("webserver")
//	host := fs.String("db-host", "localhost", "Database host")
//	fs.SetEnvPrefix("WEBAPP")
//
//	// Environment variable: WEBAPP_DB_HOST=postgresql.example.com
//	// Command line override: --db-host=localhost
//	// Result: host="localhost" (CLI wins over env var)
//
// This automatically enables environment variable lookup.
func (fs *FlagSet) SetEnvPrefix(prefix string) {
	fs.envPrefix = prefix
	fs.enableEnvLookup = true
}

// SetEnvVar sets a custom environment variable name for a specific flag.
// This overrides the default naming convention (prefix + converted flag name).
//
// Useful for integrating with existing environment variable conventions.
//
// Example:
//
//	fs := flashflags.New("myapp")
//	dbURL := fs.String("database-url", "", "Database connection URL")
//
//	// Use standard naming instead of MYAPP_DATABASE_URL
//	fs.SetEnvVar("database-url", "DATABASE_URL")
//
//	// Now reads from DATABASE_URL environment variable
//	// export DATABASE_URL=postgres://user:pass@host/db
//
// Returns an error if the flag name doesn't exist.
func (fs *FlagSet) SetEnvVar(flagName, envVarName string) error {
	flag, exists := fs.flags[flagName]
	if !exists {
		return fmt.Errorf("flag %s not found", flagName)
	}
	flag.envVar = envVarName
	return nil
}

// EnableEnvLookup enables environment variable lookup using default naming convention.
// No prefix is used - flag names are directly converted to environment variable names.
//
// Conversion rules:
//   - Hyphens become underscores: "db-host" → "DB_HOST"
//   - All uppercase: "api-key" → "API_KEY"
//   - Preserved characters: "timeout" → "TIMEOUT"
//
// Example:
//
//	fs := flashflags.New("myapp")
//	host := fs.String("db-host", "localhost", "Database host")
//	port := fs.Int("db-port", 5432, "Database port")
//	fs.EnableEnvLookup()
//
//	// Reads from: DB_HOST and DB_PORT environment variables
//	// export DB_HOST=postgresql.example.com
//	// export DB_PORT=5433
//
// Use SetEnvPrefix() if you need a prefix to avoid variable name conflicts.
func (fs *FlagSet) EnableEnvLookup() {
	fs.enableEnvLookup = true
}

// LoadEnvironmentVariables loads values from environment variables.
// This is called automatically during Parse, but can be called manually if needed.
//
// Possible errors:
//   - Type conversion errors: "invalid int value for MYAPP_PORT: abc"
//   - Validation errors: environment values that fail custom validators
//   - Duration parsing errors: invalid duration format in environment variable
//
// Environment variables are only processed if EnableEnvLookup() or SetEnvPrefix() was called.
//
// Example error handling:
//
//	if err := fs.LoadEnvironmentVariables(); err != nil {
//		log.Fatalf("Environment variable error: %v", err)
//	}
func (fs *FlagSet) LoadEnvironmentVariables() error {
	if !fs.enableEnvLookup {
		return nil
	}

	for _, flag := range fs.order {
		name := flag.name
		// setFlagValueFrom enforces this too; skipping early avoids the env
		// lookup for a value that would be discarded anyway. A value from the
		// config file ranks lower and is overridden here.
		if !flag.canSet(sourceEnv) {
			continue
		}

		envVarName := fs.getEnvVarName(name, flag)
		if envVarName == "" {
			continue
		}

		envValue := os.Getenv(envVarName)
		if envValue == "" {
			continue
		}

		// Set the flag value from environment variable
		if err := fs.setFlagValueFrom(name, envValue, sourceEnv); err != nil {
			return fmt.Errorf("invalid environment variable %s=%s: %v", envVarName, envValue, err)
		}
	}

	return nil
}

// getEnvVarName returns the environment variable name for a flag
func (fs *FlagSet) getEnvVarName(flagName string, flag *Flag) string {
	// Use custom environment variable name if set
	if flag.envVar != "" {
		return flag.envVar
	}

	// Use prefix-based naming if prefix is set
	if fs.envPrefix != "" {
		// Convert flag name: "db-host" -> "MYAPP_DB_HOST"
		envName := strings.ToUpper(strings.ReplaceAll(flagName, "-", "_"))
		return fs.envPrefix + "_" + envName
	}

	// Default naming: "db-host" -> "DB_HOST"
	return strings.ToUpper(strings.ReplaceAll(flagName, "-", "_"))
}
