# FlashFlags: Ultra-fast command-line flag parsing for Go
### an AGILira library

[![CI/CD Pipeline](https://github.com/agilira/flash-flags/actions/workflows/ci.yml/badge.svg)](https://github.com/agilira/flash-flags/actions/workflows/ci.yml)
[![Security](https://img.shields.io/badge/security-gosec%20verified-brightgreen.svg)](https://github.com/securego/gosec)
[![Go Report Card](https://goreportcard.com/badge/github.com/agilira/flash-flags)](https://goreportcard.com/report/github.com/agilira/flash-flags)
[![Coverage](https://img.shields.io/badge/coverage-92.7%25-brightgreen.svg)](.)

FlashFlags is an ultra-fast, zero-dependency, lock-free command-line flag parsing library for Go. Originally built for Argus, it provides great performance while maintaining simplicity and ease of use.

## Features

- **Ultra-Fast**: Zero-allocation parsing with optimized performance
- **Zero Dependencies**: Uses only Go standard library
- **Lock-Free**: Thread-safe operations without locks
- **Configuration Files**: JSON config file support with auto-discovery
- **Environment Variables**: Automatic environment variable integration
- **Validation**: Built-in validation system with custom validators
- **Help System**: Professional help output with grouping
- **Dependencies**: Flag dependency management
- **Type Safety**: Strong typing for all flag types
- **Short Flags**: Single-character flag support

## Quick Start

```go
package main

import (
    "fmt"
    "os"
    
    "github.com/agilira/flash-flags"
)

func main() {
    // Create flag set
    fs := flashflags.New("myapp")
    
    // Register flags
    host := fs.StringVar("host", "h", "localhost", "Server host")
    port := fs.IntVar("port", "p", 8080, "Server port")
    verbose := fs.BoolVar("verbose", "v", false, "Enable verbose logging")
    
    // Parse arguments
    if err := fs.Parse(os.Args[1:]); err != nil {
        if err.Error() == "help requested" {
            os.Exit(0) // Help was shown
        }
        fmt.Printf("Error: %v\n", err)
        os.Exit(1)
    }
    
    // Use flags
    fmt.Printf("Server starting on %s:%d (verbose: %t)\n", *host, *port, *verbose)
}
```

### Usage Examples

```bash
# Basic usage
./myapp --host 0.0.0.0 --port 3000 --verbose

# Short flags
./myapp -h 0.0.0.0 -p 3000 -v

# Mixed format
./myapp --host=192.168.1.1 -p 8080

# Help
./myapp --help
```

## Installation

```bash
go get github.com/agilira/flash-flags
```

## Why FlashFlags?

FlashFlags offers a clean, feature-rich API with solid performance and zero external dependencies. Perfect for applications that need advanced flag features without the complexity of large CLI frameworks.

## Performance

FlashFlags is designed for maximum performance, offering practically identical performance to Go's standard library with a lot of additional features:

| Operation | Time | Memory | Allocations |
|-----------|------|--------|-------------|
| Parse (5 flags) | 1,511 ns | 1,520 B | 25 allocs |
| Get String | 14.77 ns | 0 B | 0 allocs |
| Get Int | 7.54 ns | 0 B | 0 allocs |
| Get Bool | 8.17 ns | 0 B | 0 allocs |
| Get Duration | 8.74 ns | 0 B | 0 allocs |

**FlashFlags vs Go standard flag**: practically identical performance (~1.5μs) with zero significant additional overhead.

*Benchmarks run on AMD Ryzen 5 7520U, Go 1.23*

### Key Advantages

- **Fast Parsing**: Optimized for real-world flag parsing scenarios
- **Zero Dependencies**: Uses only Go standard library
- **Memory Efficient**: Reasonable allocation profile for complex parsing
- **Thread Safe**: Lock-free concurrent operations

## Documentation

- **[API Reference](docs/API.md)** - Complete API documentation
- **[Usage Guide](docs/USAGE.md)** - Comprehensive usage examples and patterns
- **[Demo Examples](demo/)** - Real-world examples and integrations

## Supported Flag Types

| Type | Go Type | Example | Description |
|------|---------|---------|-------------|
| `string` | `string` | `--name "John"` | Text values |
| `int` | `int` | `--port 8080` | Integer numbers |
| `bool` | `bool` | `--verbose` | Boolean flags |
| `float64` | `float64` | `--rate 0.75` | Floating point numbers |
| `duration` | `time.Duration` | `--timeout 30s` | Time durations |
| `stringSlice` | `[]string` | `--tags web,api` | Comma-separated lists |

## Configuration Priority

FlashFlags applies configuration in this priority order (higher numbers override lower):

1. **Default values** (lowest priority)
2. **Configuration file** values
3. **Environment variables**
4. **Command-line arguments** (highest priority)

### Configuration File Example

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

### Environment Variables

```bash
# With prefix
export MYAPP_HOST=localhost
export MYAPP_PORT=8080

# Custom names
export DATABASE_URL=postgres://...
```

## Validation & Constraints

```go
// Custom validation
fs.SetValidator("port", func(val interface{}) error {
    port := val.(int)
    if port < 1024 || port > 65535 {
        return fmt.Errorf("port must be between 1024 and 65535")
    }
    return nil
})

// Required flags
fs.SetRequired("api-key")

// Flag dependencies
fs.SetDependencies("tls-cert", "enable-tls")
```

## Real-World Example

```go
package main

import (
    "fmt"
    "log"
    "os"
    "time"
    
    "github.com/agilira/flash-flags"
)

func main() {
    fs := flashflags.New("webserver")
    fs.SetDescription("High-performance web server")
    fs.SetVersion("v1.0.0")
    
    // Server configuration
    host := fs.StringVar("host", "h", "localhost", "Server host")
    port := fs.IntVar("port", "p", 8080, "Server port")
    workers := fs.Int("workers", 4, "Number of worker threads")
    
    // TLS configuration  
    enableTLS := fs.Bool("enable-tls", false, "Enable TLS")
    tlsCert := fs.String("tls-cert", "", "TLS certificate file")
    tlsKey := fs.String("tls-key", "", "TLS private key file")
    
    // Performance tuning
    timeout := fs.Duration("timeout", 30*time.Second, "Request timeout")
    maxConns := fs.Int("max-connections", 1000, "Maximum connections")
    
    // Logging
    logLevel := fs.String("log-level", "info", "Log level (debug, info, warn, error)")
    logFile := fs.String("log-file", "", "Log file path (empty for stdout)")
    
    // Environment and config
    fs.SetEnvPrefix("WEBSERVER")
    fs.AddConfigPath("./config")
    fs.AddConfigPath("/etc/webserver")
    
    // Organize help output
    fs.SetGroup("host", "Server Options")
    fs.SetGroup("port", "Server Options")
    fs.SetGroup("workers", "Server Options")
    fs.SetGroup("enable-tls", "TLS Options")
    fs.SetGroup("tls-cert", "TLS Options")
    fs.SetGroup("tls-key", "TLS Options")
    
    // Validation
    fs.SetValidator("port", func(val interface{}) error {
        port := val.(int)
        if port < 1 || port > 65535 {
            return fmt.Errorf("port must be between 1 and 65535")
        }
        return nil
    })
    
    fs.SetValidator("log-level", func(val interface{}) error {
        level := val.(string)
        validLevels := []string{"debug", "info", "warn", "error"}
        for _, valid := range validLevels {
            if level == valid {
                return nil
            }
        }
        return fmt.Errorf("log-level must be one of: debug, info, warn, error")
    })
    
    // Dependencies
    fs.SetDependencies("tls-cert", "enable-tls")
    fs.SetDependencies("tls-key", "enable-tls")
    
    // Parse
    if err := fs.Parse(os.Args[1:]); err != nil {
        if err.Error() == "help requested" {
            os.Exit(0)
        }
        log.Fatalf("Error: %v", err)
    }
    
    // Use configuration
    fmt.Printf("Starting web server:\n")
    fmt.Printf("  Host: %s\n", *host)
    fmt.Printf("  Port: %d\n", *port)
    fmt.Printf("  Workers: %d\n", *workers)
    fmt.Printf("  TLS: %v\n", *enableTLS)
    fmt.Printf("  Timeout: %v\n", *timeout)
    fmt.Printf("  Max Connections: %d\n", *maxConns)
    fmt.Printf("  Log Level: %s\n", *logLevel)
    if *logFile != "" {
        fmt.Printf("  Log File: %s\n", *logFile)
    }
    
    // Start your server here...
}
```

## License

flash-flags is licensed under the [Mozilla Public License 2.0](./LICENSE.md).

---

flash-flags • an AGILira library
