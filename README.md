# FlashFlags: Ultra-fast command-line flag parsing for Go
### an AGILira library

[![CI/CD Pipeline](https://github.com/agilira/flash-flags/actions/workflows/ci.yml/badge.svg)](https://github.com/agilira/flash-flags/actions/workflows/ci.yml)
[![CodeQL Security](https://github.com/agilira/flash-flags/actions/workflows/codeql.yml/badge.svg)](https://github.com/agilira/flash-flags/actions/workflows/codeql.yml)
[![Security](https://img.shields.io/badge/security-hardened-brightgreen.svg)](https://github.com/agilira/flash-flags/actions/workflows/codeql.yml)
[![Fuzz Testing](https://img.shields.io/badge/fuzz-tested-blue.svg)](https://github.com/agilira/flash-flags/blob/main/fuzz_test.go)
[![Go Report Card](https://goreportcard.com/badge/github.com/agilira/flash-flags?v=2)](https://goreportcard.com/report/github.com/agilira/flash-flags)
[![Coverage](https://codecov.io/gh/agilira/flash-flags/branch/main/graph/badge.svg)](https://codecov.io/gh/agilira/flash-flags)
[![GoDoc](https://godoc.org/github.com/agilira/flash-flags?status.svg)](https://godoc.org/github.com/agilira/flash-flags)

**[Features](#features) • [Quick Start](#quick-start) •  [Performance](#performance) • [Demo](#demo) • [Flag Types](#supported-flag-types) • [Configuration](#configuration-priority) • [Examples](#real-world-example)**

FlashFlags is an ultra-fast, zero-dependency, lock-free command-line flag parsing library for Go. Originally built for [Argus](https://github.com/agilira/argus), it provides great performance while maintaining simplicity and ease of use. FlashFlags serves as the core parsing engine for our CLI framework [Orpheus](https://github.com/agilira/orpheus).

## Features

- **Security-Hardened**: Built-in protection against injection attacks, path traversal, and buffer overflows
- **Ultra-Fast**: 85% of stdlib performance with comprehensive security validation
- **Zero Dependencies**: Can be use as drop-in stdlib replacement with security
- **Lock-Free**: Thread-safe operations without locks
- **Configuration Files**: JSON config file support with auto-discovery
- **Environment Variables**: Automatic environment variable integration
- **Validation**: Built-in validation system with custom validators
- **Help System**: Professional help output with grouping
- **Dependencies**: Flag dependency management
- **Type Safety**: Strong typing for all flag types
- **POSIX/GNU Syntax**: Complete flag syntax support including combined short flags
- **Flexible Parsing**: Support for `-f=value` and `-abc` combined syntax

### Security Features

Flash-flags is the **only** Go flag library with comprehensive security hardening:

- **Command Injection Protection**: Blocks `$(...)`, backticks, and shell metacharacters
- **Path Traversal Prevention**: Prevents `../` and `..\\` directory traversal attacks  
- **Buffer Overflow Safeguards**: 10KB input limits with fast-path optimization
- **Format String Attack Blocking**: Detects and blocks `%n`, `%s` format string exploits
- **Input Sanitization**: Removes null bytes and dangerous control characters
- **Windows Device Protection**: Blocks Windows reserved names (CON, PRN, AUX, etc.)

**Security overhead**: Only 132ns per operation (17%) for complete protection

## Compatibility and Support

FlashFlags is designed for Go 1.23+ environments and follows Long-Term Support guidelines to ensure consistent performance across production deployments.

## Performance

FlashFlags delivers exceptional performance with security-hardened parsing:

```
AMD Ryzen 5 7520U 
BenchmarkFlashFlags-8      1,294,699    924 ns/op     945 B/op    11 allocs/op  🛡️ SECURE
BenchmarkStdFlag-8         1,527,176    792 ns/op     945 B/op    13 allocs/op  
BenchmarkPflag-8             785,904   1322 ns/op    1569 B/op    21 allocs/op  
BenchmarkGoFlags-8           147,394   7460 ns/op    5620 B/op    61 allocs/op  
BenchmarkKingpin-8           150,154   7567 ns/op    6504 B/op    97 allocs/op  
```

**Flash-flags is 85% as fast as stdlib with FULL security hardening**  
**Only 132ns overhead for complete protection against injection attacks**

**Reproduce benchmarks**:
```bash
cd benchmarks && go test -bench=. -benchmem
```

## Quick Start

### Installation

```bash
go get github.com/agilira/flash-flags
```
### Basic Usage

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

# Short flags with space
./myapp -h 0.0.0.0 -p 3000 -v

# Short flags with equals (NEW!)
./myapp -h=192.168.1.1 -p=8080 -v=true

# Combined short flags (NEW!)
./myapp -hvp 3000              # -h -v -p 3000
./myapp -abc                   # -a -b -c (all boolean)

# Mixed formats
./myapp --host=192.168.1.1 -vp 8080 --debug=false

# Environment variables + CLI
MYAPP_HOST=api.example.com ./myapp -p=3000 --verbose

# Help
./myapp --help
```

### Flag Syntax

FlashFlags supports comprehensive POSIX/GNU-style flag syntax for maximum compatibility:

### Long Flags
```bash
--flag value          # Space-separated value
--flag=value          # Equals-separated value  
--boolean-flag        # Boolean without value (true)
--boolean-flag=false  # Explicit boolean value
```

### Short Flags
```bash
-f value              # Space-separated value
-f=value              # Equals-separated value (NEW!)
-b                    # Boolean short flag (true)
-b=false              # Explicit boolean value
```

### Combined Short Flags
```bash
-abc                  # Equivalent to -a -b -c (all boolean)
-abc value            # Last flag gets the value: -a -b -c value
-vdp 8080             # Verbose + debug + port: -v -d -p 8080
```

**Rules for combined flags:**
- All flags except the last must be boolean
- The last flag can be any type and consumes the next argument
- Example: `-vhp 3000` sets verbose=true, help=true, port=3000

### Drop-in Stdlib Replacement

Flash-flags includes a complete drop-in replacement for Go's standard `flag` package. Migrate with zero code changes:

```go
// Before - using stdlib
import "flag"

// After - using flash-flags
import "github.com/antonio-giordano/flash-flags/stdlib/flag"

// All your existing code works unchanged!
var name = flag.String("name", "default", "description")
var count = flag.Int("count", 42, "number of items")

func main() {
    flag.Parse()
    fmt.Printf("Name: %s, Count: %d\n", *name, *count)
    
    // Full remaining arguments support
    for i := 0; i < flag.NArg(); i++ {
        fmt.Printf("Arg[%d]: %s\n", i, flag.Arg(i))
    }
}
```

See the [stdlib example](examples/stdlib-drop-in/) for a complete working demonstration.

## Examples

- **[Examples](examples/)** - Real-world examples and integrations

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
