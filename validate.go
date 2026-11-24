// validate.go: validators, required flags and dependencies
//
// Copyright (c) 2025 AGILira - A. Giordano
// Series: an AGILira library
// SPDX-License-Identifier: MPL-2.0

package flashflags

import (
	"fmt"
)

// SetValidator sets a validation function for a specific flag.
// The validator function will be called whenever the flag value is set, allowing for custom validation logic.
//
// The validator function receives the parsed value as interface{} and should return an error if validation fails.
// Type assertion is required to access the actual value.
//
// Example:
//
//	fs := flashflags.New("myapp")
//	port := fs.IntVar("port", "p", 8080, "Server port")
//
//	err := fs.SetValidator("port", func(val interface{}) error {
//		port, ok := val.(int)
//		if !ok {
//			return fmt.Errorf("expected int, got %T", val)
//		}
//		if port < 1024 || port > 65535 {
//			return fmt.Errorf("port must be between 1024-65535, got %d", port)
//		}
//		return nil
//	})
//	if err != nil {
//		log.Fatal(err)
//	}
//
// Use the comma-ok form of the type assertion as shown: a bare val.(int) panics
// on mismatch, and that panic propagates out of Parse.
//
// Returns an error if the flag name doesn't exist.
func (fs *FlagSet) SetValidator(name string, validator func(interface{}) error) error {
	flag := fs.Lookup(name)
	if flag == nil {
		return fmt.Errorf("flag not found: %s", name)
	}
	flag.SetValidator(validator)
	return nil
}

// SetRequired marks a flag as required.
// Required flags must be provided during parsing, otherwise an error will be returned.
//
// Required flags can be satisfied by any configuration source (command-line, environment variables, or config files).
//
// Example:
//
//	fs := flashflags.New("myapp")
//	apiKey := fs.String("api-key", "", "API authentication key")
//
//	if err := fs.SetRequired("api-key"); err != nil {
//		log.Fatal(err)
//	}
//
//	// This will fail if --api-key is not provided
//	if err := fs.Parse(os.Args[1:]); err != nil {
//		log.Fatal(err) // "required flag --api-key not provided"
//	}
//
// Returns an error if the flag name doesn't exist.
func (fs *FlagSet) SetRequired(name string) error {
	flag := fs.Lookup(name)
	if flag == nil {
		return fmt.Errorf("flag not found: %s", name)
	}
	flag.required = true
	return nil
}

// SetDependencies sets dependencies for a flag.
// When this flag is set, all dependent flags must also be set, otherwise an error will be returned.
//
// Dependencies are validated after all configuration sources are processed.
// Dependent flags can be satisfied by any configuration source.
//
// Example:
//
//	fs := flashflags.New("myapp")
//	enableTLS := fs.Bool("enable-tls", false, "Enable TLS")
//	tlsCert := fs.String("tls-cert", "", "TLS certificate file")
//	tlsKey := fs.String("tls-key", "", "TLS private key file")
//
//	// Both cert and key require TLS to be enabled
//	if err := fs.SetDependencies("tls-cert", "enable-tls"); err != nil {
//		log.Fatal(err)
//	}
//	if err := fs.SetDependencies("tls-key", "enable-tls"); err != nil {
//		log.Fatal(err)
//	}
//
//	// This will fail if --tls-cert is provided without --enable-tls
//	if err := fs.Parse(os.Args[1:]); err != nil {
//		log.Fatal(err) // "flag --tls-cert requires --enable-tls to be set"
//	}
//
// Returns an error if the flag name doesn't exist.
func (fs *FlagSet) SetDependencies(name string, dependencies ...string) error {
	flag := fs.Lookup(name)
	if flag == nil {
		return fmt.Errorf("flag not found: %s", name)
	}
	flag.dependencies = dependencies
	return nil
}

// ValidateAll validates all flags that have validators set.
// This is called automatically during Parse, but can be called manually if needed.
//
// Possible errors:
//   - Custom validation errors from validator functions
//   - Format: "validation failed for flag --flagname: custom error message"
//
// Example:
//
//	// Validator that might fail:
//	fs.SetValidator("port", func(val interface{}) error {
//		if val.(int) < 1024 { return fmt.Errorf("port too low") }
//		return nil
//	})
//
//	if err := fs.ValidateAll(); err != nil {
//		// Error: "validation failed for flag --port: port too low"
//	}
//
// The method stops at the first validation failure and returns that error.
func (fs *FlagSet) ValidateAll() error {
	for _, flag := range fs.order {
		name := flag.name
		if flag.validator != nil {
			if err := flag.Validate(); err != nil {
				return fmt.Errorf("validation failed for flag --%s: %v", name, err)
			}
		}
	}
	return nil
}

// ValidateRequired checks that all required flags are set.
// This is called automatically during Parse, but can be called manually if needed.
//
// Possible errors:
//   - Missing required flag: "required flag --flagname not provided"
//
// Example:
//
//	fs.SetRequired("api-key")
//	// If --api-key not provided by CLI, env vars, or config file:
//	if err := fs.ValidateRequired(); err != nil {
//		// Error: "required flag --api-key not provided"
//	}
//
// Required flags can be satisfied by any configuration source (CLI, env, config file).
func (fs *FlagSet) ValidateRequired() error {
	for _, flag := range fs.order {
		name := flag.name
		if flag.required && flag.source == sourceDefault {
			return fmt.Errorf("required flag --%s not provided", name)
		}
	}
	return nil
}

// ValidateDependencies checks that all flag dependencies are satisfied.
// This is called automatically during Parse, but can be called manually if needed.
//
// Possible errors:
//   - Missing dependency: "flag --flagname requires --dependency to be set"
//   - Non-existent dependency: "flag --flagname depends on non-existent flag --missing"
//
// Example:
//
//	fs.SetDependencies("tls-cert", "enable-tls")
//	// If --tls-cert provided but --enable-tls not set:
//	if err := fs.ValidateDependencies(); err != nil {
//		// Error: "flag --tls-cert requires --enable-tls to be set"
//	}
//
// Dependencies must be satisfied by any configuration source.
func (fs *FlagSet) ValidateDependencies() error {
	for _, flag := range fs.order {
		name := flag.name
		if flag.source != sourceDefault && len(flag.dependencies) > 0 {
			for _, dep := range flag.dependencies {
				depFlag := fs.Lookup(dep)
				if depFlag == nil {
					return fmt.Errorf("flag --%s depends on non-existent flag --%s", name, dep)
				}
				if depFlag.source == sourceDefault {
					return fmt.Errorf("flag --%s requires --%s to be set", name, dep)
				}
			}
		}
	}
	return nil
}

// ValidateAllConstraints validates all constraints: validators, required flags, and dependencies.
// This is called automatically during Parse, but can be called manually if needed.
//
// Possible errors:
//   - Required flag errors: "required flag --api-key not provided"
//   - Dependency errors: "flag --tls-cert requires --enable-tls to be set"
//   - Validation errors: "validation failed for flag --port: port must be between 1024-65535"
//   - Missing dependency errors: "flag --ssl depends on non-existent flag --tls"
//
// The method stops at the first constraint violation and returns that error.
//
// Example:
//
//	if err := fs.ValidateAllConstraints(); err != nil {
//		fmt.Printf("Validation failed: %v\n", err)
//		fs.PrintHelp()
//		os.Exit(1)
//	}
func (fs *FlagSet) ValidateAllConstraints() error {
	if err := fs.ValidateRequired(); err != nil {
		return err
	}
	if err := fs.ValidateDependencies(); err != nil {
		return err
	}
	if err := fs.ValidateAll(); err != nil {
		return err
	}
	return nil
}
