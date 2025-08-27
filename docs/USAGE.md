# FlashFlags Usage Guide

FlashFlags is an ultra-fast, zero-dependency command-line flag parsing library for Go. This guide covers practical usage examples and common patterns.

## Table of Contents

- [Quick Start](#quick-start)
- [Basic Flag Types](#basic-flag-types)
- [Short Flags](#short-flags)
- [Boolean Flags](#boolean-flags)
- [Configuration Files](#configuration-files)
- [Environment Variables](#environment-variables)
- [Validation and Constraints](#validation-and-constraints)
- [Help System](#help-system)
- [Advanced Usage](#advanced-usage)
- [Best Practices](#best-practices)

## Quick Start

### Basic Example

```go
package main

import (
    "fmt"
    "os"
    
    "github.com/agilira/flash-flags"
)

func main() {
    // Create a new flag set
    fs := flashflags.New("myapp")
    
    // Register flags
    host := fs.String("host", "localhost", "Server host")
    port := fs.Int("port", 8080, "Server port")
    verbose := fs.Bool("verbose", false, "Enable verbose logging")
    
    // Parse command line arguments
    if err := fs.Parse(os.Args[1:]); err != nil {
        fmt.Printf("Error: %v\n", err)
        os.Exit(1)
    }
    
    // Use the flag values
    fmt.Printf("Starting server on %s:%d\n", *host, *port)
    if *verbose {
        fmt.Println("Verbose logging enabled")
    }
}
```

### Usage

```bash
# Using default values
./myapp
# Output: Starting server on localhost:8080

# Setting values
./myapp --host 0.0.0.0 --port 3000 --verbose
# Output: Starting server on 0.0.0.0:3000
#         Verbose logging enabled

# Using equals syntax
./myapp --host=192.168.1.100 --port=9000
```

## Basic Flag Types

### String Flags

```go
fs := flashflags.New("example")

// Basic string flag
name := fs.String("name", "default", "Your name")

// String with validation
email := fs.String("email", "", "Email address")
fs.SetValidator("email", func(val interface{}) error {
    email := val.(string)
    if !strings.Contains(email, "@") {
        return fmt.Errorf("invalid email format")
    }
    return nil
})
```

### Integer Flags

```go
// Basic integer
workers := fs.Int("workers", 4, "Number of worker threads")

// Integer with validation (port range)
port := fs.Int("port", 8080, "Server port")
fs.SetValidator("port", func(val interface{}) error {
    port := val.(int)
    if port < 1 || port > 65535 {
        return fmt.Errorf("port must be between 1 and 65535")
    }
    return nil
})
```

### Float Flags

```go
rate := fs.Float64("rate", 0.5, "Processing rate (0.0-1.0)")
fs.SetValidator("rate", func(val interface{}) error {
    rate := val.(float64)
    if rate < 0.0 || rate > 1.0 {
        return fmt.Errorf("rate must be between 0.0 and 1.0")
    }
    return nil
})
```

### Duration Flags

```go
timeout := fs.Duration("timeout", 30*time.Second, "Request timeout")

// Usage examples:
// --timeout 5s
// --timeout 1m30s  
// --timeout 2h
```

### String Slice Flags

```go
tags := fs.StringSlice("tags", []string{}, "Comma-separated tags")

// Usage examples:
// --tags web,api,frontend
// --tags "tag1,tag with spaces,tag3"
```

## Short Flags

Short flags provide convenient single-character alternatives:

```go
fs := flashflags.New("example")

// Register flags with short keys
host := fs.StringVar("host", "h", "localhost", "Server host")
port := fs.IntVar("port", "p", 8080, "Server port") 
verbose := fs.BoolVar("verbose", "v", false, "Enable verbose logging")
debug := fs.BoolVar("debug", "d", false, "Enable debug mode")

// Usage examples:
// Long form: --host localhost --port 8080 --verbose
// Short form: -h localhost -p 8080 -v
// Mixed: --host localhost -p 8080 -v
```

## Boolean Flags

Boolean flags have special behavior and multiple usage patterns:

```go
verbose := fs.Bool("verbose", false, "Enable verbose logging")
debug := fs.BoolVar("debug", "d", false, "Enable debug mode")
```

### Usage Patterns

```bash
# Enable flags (all equivalent)
--verbose
--verbose=true
--verbose true
-d

# Disable flags
--verbose=false
--verbose false

# Multiple boolean flags
-vd  # NOT SUPPORTED - use separate flags
-v -d  # Correct way
```

## Configuration Files

FlashFlags supports JSON configuration files with automatic discovery:

### Basic Configuration

```go
fs := flashflags.New("myapp")

// Register flags
host := fs.String("host", "localhost", "Server host")
port := fs.Int("port", 8080, "Server port")
workers := fs.Int("workers", 4, "Number of workers")
enableTLS := fs.Bool("enable-tls", false, "Enable TLS")
tags := fs.StringSlice("tags", []string{}, "Service tags")

// Set program description
fs.SetDescription("My awesome application")
fs.SetVersion("1.0.0")

// Parse (will automatically load config if found)
if err := fs.Parse(os.Args[1:]); err != nil {
    fmt.Printf("Error: %v\n", err)
    os.Exit(1)
}
```

### Configuration File Example

Create `myapp.json` or `config.json`:

```json
{
  "host": "0.0.0.0",
  "port": 3000,
  "workers": 8,
  "enable-tls": true,
  "tags": ["web", "api", "production"],
  "timeout": "60s"
}
```

### Custom Configuration Paths

```go
// Set explicit config file
fs.SetConfigFile("/etc/myapp/config.json")

// Add search paths
fs.AddConfigPath("/etc/myapp")
fs.AddConfigPath("/usr/local/etc/myapp")
fs.AddConfigPath(".")
```

### Auto-discovery

FlashFlags automatically searches for configuration files in this order:

1. `{program-name}.json`
2. `{program-name}.config.json`  
3. `config.json`

In these directories:
1. Current directory (`.`)
2. `./config/`
3. User home directory

## Environment Variables

FlashFlags can automatically load values from environment variables:

### Enable Environment Variable Support

```go
fs := flashflags.New("myapp")

// Method 1: Enable with prefix
fs.SetEnvPrefix("MYAPP")

// Method 2: Enable with default naming
fs.EnableEnvLookup()

// Register flags
host := fs.String("host", "localhost", "Server host")
port := fs.Int("port", 8080, "Server port")
dbURL := fs.String("db-url", "", "Database URL")

// Custom environment variable name
fs.SetEnvVar("db-url", "DATABASE_URL")
```

### Environment Variable Naming

| Flag Name | Default Env Var | With Prefix "MYAPP" | Custom |
|-----------|----------------|-------------------|---------|
| `host` | `HOST` | `MYAPP_HOST` | Via `SetEnvVar()` |
| `db-url` | `DB_URL` | `MYAPP_DB_URL` | `DATABASE_URL` |
| `enable-tls` | `ENABLE_TLS` | `MYAPP_ENABLE_TLS` | - |

### Usage Example

```bash
# Set environment variables
export MYAPP_HOST=0.0.0.0
export MYAPP_PORT=9000
export DATABASE_URL=postgres://user:pass@localhost/db

# Run application (env vars will be used as defaults)
./myapp

# Override with command line (highest priority)
./myapp --port 3000  # Uses env vars for host and db-url, but port=3000
```

## Validation and Constraints

### Custom Validation

```go
fs := flashflags.New("example")

// Port validation
port := fs.Int("port", 8080, "Server port")
fs.SetValidator("port", func(val interface{}) error {
    port := val.(int)
    if port < 1024 || port > 65535 {
        return fmt.Errorf("port must be between 1024 and 65535")
    }
    return nil
})

// Email validation
email := fs.String("email", "", "Email address")
fs.SetValidator("email", func(val interface{}) error {
    email := val.(string)
    if email == "" {
        return nil // Allow empty (use SetRequired for mandatory)
    }
    if !strings.Contains(email, "@") || !strings.Contains(email, ".") {
        return fmt.Errorf("invalid email format")
    }
    return nil
})
```

### Required Flags

```go
// Mark flags as required
fs.SetRequired("email")
fs.SetRequired("api-key")

// Parse will fail if required flags are not provided
if err := fs.Parse(os.Args[1:]); err != nil {
    fmt.Printf("Error: %v\n", err)
    // Error: required flag --email not provided
    os.Exit(1)
}
```

### Flag Dependencies

```go
// TLS cert requires TLS to be enabled
fs.SetDependencies("tls-cert", "enable-tls")
fs.SetDependencies("tls-key", "enable-tls")

// Database password requires database URL
fs.SetDependencies("db-password", "db-url")
```

## Help System

### Basic Help

```go
fs := flashflags.New("myapp")
fs.SetDescription("A powerful web server application")
fs.SetVersion("v1.2.3")

// Register flags with groups
host := fs.StringVar("host", "h", "localhost", "Server host")
port := fs.IntVar("port", "p", 8080, "Server port")
workers := fs.Int("workers", 4, "Number of worker threads")

// Organize flags into groups
fs.SetGroup("host", "Server Options")
fs.SetGroup("port", "Server Options")
fs.SetGroup("workers", "Performance Options")

// Help is automatically available via --help or -h
```

### Generated Help Output

```
A powerful web server application

Usage: myapp [options]

Version: v1.2.3

Server Options:
  -h, --host STRING          Server host (default: localhost)
  -p, --port INT             Server port (default: 8080)

Performance Options:
  --workers INT              Number of worker threads (default: 4)

Options:
  --help                     Show this help message
```

### Custom Help

```go
// Generate help programmatically
helpText := fs.Help()
fmt.Print(helpText)

// Or print directly
fs.PrintHelp()
```

## Advanced Usage

### Programmatic Flag Access

```go
// Check if a flag was explicitly set
if fs.Changed("verbose") {
    fmt.Println("Verbose mode was explicitly enabled")
}

// Get flag information
if flag := fs.Lookup("port"); flag != nil {
    fmt.Printf("Port flag type: %s\n", flag.Type())
    fmt.Printf("Port flag value: %v\n", flag.Value())
    fmt.Printf("Port flag changed: %v\n", flag.Changed())
}

// Visit all flags
fs.VisitAll(func(flag *flashflags.Flag) {
    fmt.Printf("Flag %s = %v (type: %s)\n", 
        flag.Name(), flag.Value(), flag.Type())
})
```

### Dynamic Configuration

```go
// Reset flags to defaults
fs.Reset()

// Reset specific flag
fs.ResetFlag("port")

// Get values using type-safe methods
host := fs.GetString("host")
port := fs.GetInt("port")
verbose := fs.GetBool("verbose")
timeout := fs.GetDuration("timeout")
tags := fs.GetStringSlice("tags")
```

### Error Handling

```go
if err := fs.Parse(os.Args[1:]); err != nil {
    switch {
    case strings.Contains(err.Error(), "help requested"):
        // User requested help, exit gracefully
        os.Exit(0)
    case strings.Contains(err.Error(), "required flag"):
        fmt.Printf("Missing required flag: %v\n", err)
        fs.PrintHelp()
        os.Exit(1)
    case strings.Contains(err.Error(), "validation failed"):
        fmt.Printf("Invalid flag value: %v\n", err)
        os.Exit(1)
    default:
        fmt.Printf("Error parsing flags: %v\n", err)
        os.Exit(1)
    }
}
```

## Best Practices

### 1. Use Clear Flag Names

```go
// Good
fs.String("database-url", "", "PostgreSQL connection string")
fs.String("log-level", "info", "Logging level (debug, info, warn, error)")

// Avoid
fs.String("db", "", "Database")
fs.String("level", "info", "Level")
```

### 2. Provide Good Help Text

```go
// Good - explains what the flag does and acceptable values
fs.Int("workers", 4, "Number of worker goroutines (1-100)")
fs.String("format", "json", "Output format (json, yaml, text)")

// Good - includes examples
fs.String("listen", ":8080", "Listen address (e.g., ':8080', '127.0.0.1:3000')")
```

### 3. Use Validation for Critical Flags

```go
// Validate ranges
fs.SetValidator("workers", func(val interface{}) error {
    workers := val.(int)
    if workers < 1 || workers > 100 {
        return fmt.Errorf("workers must be between 1 and 100")
    }
    return nil
})

// Validate formats
fs.SetValidator("format", func(val interface{}) error {
    format := val.(string)
    validFormats := []string{"json", "yaml", "text"}
    for _, valid := range validFormats {
        if format == valid {
            return nil
        }
    }
    return fmt.Errorf("format must be one of: %s", strings.Join(validFormats, ", "))
})
```

### 4. Use Short Flags for Common Options

```go
// Common flags that benefit from short versions
verbose := fs.BoolVar("verbose", "v", false, "Enable verbose output")
help := fs.BoolVar("help", "h", false, "Show help")
version := fs.BoolVar("version", "V", false, "Show version")
config := fs.StringVar("config", "c", "", "Configuration file path")
```

### 5. Organize Flags with Groups

```go
// Server configuration
fs.SetGroup("host", "Server Options")
fs.SetGroup("port", "Server Options")
fs.SetGroup("tls-cert", "TLS Options")
fs.SetGroup("tls-key", "TLS Options")

// Database configuration  
fs.SetGroup("db-url", "Database Options")
fs.SetGroup("db-timeout", "Database Options")

// Logging configuration
fs.SetGroup("log-level", "Logging Options")
fs.SetGroup("log-file", "Logging Options")
```
---

flash-flags • an AGILira library