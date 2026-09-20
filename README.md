# FlashFlags: Ultra-fast command-line flag parsing for Go
### an AGILira library

[![CI/CD Pipeline](https://github.com/agilira/flash-flags/actions/workflows/ci.yml/badge.svg)](https://github.com/agilira/flash-flags/actions/workflows/ci.yml)
[![CodeQL](https://github.com/agilira/flash-flags/actions/workflows/codeql.yml/badge.svg)](https://github.com/agilira/flash-flags/actions/workflows/codeql.yml)
[![Security](https://img.shields.io/badge/security-gosec-brightgreen.svg)](https://github.com/agilira/flash-flags/actions/workflows/ci.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/agilira/flash-flags?v=2)](https://goreportcard.com/report/github.com/agilira/flash-flags)
[![Coverage](https://img.shields.io/badge/coverage-94.2%25-brightgreen.svg)](https://github.com/agilira/flash-flags)
[![GoDoc](https://godoc.org/github.com/agilira/flash-flags?status.svg)](https://godoc.org/github.com/agilira/flash-flags)
[![Mentioned in Awesome Go](https://awesome.re/mentioned-badge.svg)](https://github.com/avelino/awesome-go)

FlashFlags is an ultra-fast, zero-dependency command-line flag parsing library for Go. Originally built for [Argus](https://github.com/agilira/argus), it provides great performance while maintaining simplicity and ease of use. FlashFlags serves as the core parsing engine for our CLI framework [Orpheus](https://github.com/agilira/orpheus).

## Live Demo

<div align="center">

See Flash-Flags in action - POSIX-compliant stdlib replacement with JSON config support:

<picture>
  <source media="(max-width: 768px)" srcset="https://asciinema.org/a/QJpvO4R70WE6kUE8.svg" width="100%">
  <source media="(max-width: 1024px)" srcset="https://asciinema.org/a/QJpvO4R70WE6kUE8.svg" width="90%">
  <img src="https://asciinema.org/a/QJpvO4R70WE6kUE8.svg" alt="Flash-Flags CLI Demo" style="max-width: 100%; height: auto;" width="800">
</picture>

*[Click to view interactive demo](https://asciinema.org/a/QJpvO4R70WE6kUE8)*

</div>

**[Features](#features) • [Quick Start](#quick-start) • [Performance](#performance) • [Demo](#demo) • [Flag Types](#supported-flag-types) • [Configuration](#configuration-priority) • [Examples](#real-world-example)**

## Features

- **Input Screening**: Rejects malformed values (null bytes, control characters, absurd length)
- **Ultra-Fast**: 85% of stdlib performance with comprehensive security validation
- **Zero Dependencies**: Can be use as drop-in stdlib replacement with security
- **Concurrent Reads**: Safe for concurrent reads once Parse() has returned -- see [Thread Safety](#thread-safety)
- **Configuration Files**: JSON config file support with auto-discovery
- **Environment Variables**: Automatic environment variable integration
- **Validation**: Built-in validation system with custom validators
- **Help System**: Professional help output with grouping
- **Dependencies**: Flag dependency management
- **Type Safety**: Strong typing for all flag types
- **POSIX/GNU Syntax**: Complete flag syntax support including combined short flags
- **Flexible Parsing**: Support for `-f=value` and `-abc` combined syntax

### Input Screening

Flag values are screened for input that is malformed as a string, whatever it is
later used for:

- **Length**: values above 10000 bytes are rejected
- **Null bytes**: `\x00` truncates the string in any C API it reaches
- **Control characters**: C0 controls except `\t`, `\n`, `\r` — terminal escape
  sequences, not data
- **Fast path**: values under 100 bytes made only of `[A-Za-z0-9-_.:]` skip the
  scan, because such a value cannot contain either

It stops there, deliberately. A flag parser does not know whether a value will
reach a shell, a SQL driver, an `fmt` verb or a file open, so it does not guess.
Escaping belongs at the point of use: run `os/exec` without a shell, parameterize
SQL, resolve and confine paths before opening them.

> **Changed in v1.1.9.** The screening previously also rejected values containing
> `/etc/`, `/proc/`, `/sys/`, `rm -rf`, `drop table`, `$(`, a backtick, the
> `%n %s %x %d %c %p` format verbs, `../` and Windows device names. That denylist
> stopped no attack — `a; rm -rf ~`, `deploy && restart` and `..%2f..%2fetc` all
> passed it untouched — while rejecting ordinary input such as
> `--config /etc/myapp.conf`. It has been removed. If you relied on it as a
> security control, it was not one.

**Overhead**: roughly 132ns per operation (17%).

## Thread Safety

FlashFlags holds no mutexes and performs no atomic operations, which is what keeps
a read down to a plain map lookup. The guarantee follows from the absence of
writes, not from synchronization:

- **Writes** — flag registration, `Set*` configuration, `Parse`, `LoadConfig`,
  `LoadEnvironmentVariables`, `Reset`, `ResetFlag` — must all happen on a single
  goroutine, normally the one running `main`.
- **Reads** — `Lookup`, `Value`, `Changed`, `Source`, the `Get*` accessors and the
  pointers returned at declaration — are safe from any number of goroutines once
  `Parse` has returned.

Calling `Reset` while another goroutine reads is a data race and `go test -race`
will report it. To re-parse at runtime, synchronize yourself or build a fresh
`FlagSet` and swap it behind a pointer.

## Compatibility and Support

FlashFlags requires Go 1.25.9 or later.

## Performance

FlashFlags delivers exceptional performance with security-hardened parsing:

```
AMD Ryzen 5 7520U 
BenchmarkFlashFlags-8      1,294,699    924 ns/op     945 B/op    11 allocs/op
BenchmarkStdFlag-8         1,527,176    792 ns/op     945 B/op    13 allocs/op  
BenchmarkPflag-8             785,904   1322 ns/op    1569 B/op    21 allocs/op  
BenchmarkGoFlags-8           147,394   7460 ns/op    5620 B/op    61 allocs/op  
BenchmarkKingpin-8           150,154   7567 ns/op    6504 B/op    97 allocs/op  
```

**Roughly 132ns overhead for input screening**

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

A higher-priority source always wins, whatever order the loaders run in: a value
found in the config file does not suppress the matching environment variable.
`fs.Source("port")` reports which layer supplied a value — `"cli"`, `"env"`,
`"config"` or `"default"` — which is the quickest way to debug precedence.

```go
fs.Parse(os.Args[1:])
fmt.Println(fs.GetInt("port"), "from", fs.Source("port")) // 3000 from env
```

Note that JSON has no duration type, so a duration flag in a config file accepts
either the string form (`"30s"`, `"1m30s"`) or a plain number of nanoseconds.

### Config file paths

The path must point at a regular file; a directory, a FIFO or a device is
refused. Nothing else about it is checked, because the path comes from your
program rather than from parsed arguments — `LoadConfig` runs before argument
parsing, so no `--config` value can reach it.

Symlinks are followed, because that is usually what you want: a Kubernetes
ConfigMap projects each key as a symlink, and dotfile managers link a config
into place. If your program reads configuration from a directory other local
users can write to, opt into refusing them:

```go
fs.SetConfigFile("/tmp/myapp.json")
fs.EnableStrictConfigPaths() // refuse a symlinked final component
```

Only the final component is examined, so a symlinked parent directory is
traversed normally — which is what keeps this usable on macOS, where `/tmp` is a
symlink to `/private/tmp`. On Unix the file is opened with `O_NOFOLLOW`, so the
kernel refuses the call and there is no window to race; on Windows the check
runs before the open, because Go exposes no portable equivalent there, making it
best-effort rather than a guarantee.

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
    port, ok := val.(int)
    if !ok {
        return fmt.Errorf("expected int, got %T", val)
    }
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

A validator receives the value boxed in an `interface{}`. Always use the comma-ok
form as shown above: a bare `val.(int)` panics on mismatch, and that panic
propagates out of `Parse` and terminates the program. The dynamic type matches the
flag's declared type — `int` for `Int`, `time.Duration` for `Duration`, `[]string`
for `StringSlice` — so a failed assertion means the validator was attached to the
wrong flag.

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
        port, ok := val.(int)
        if !ok {
            return fmt.Errorf("expected int, got %T", val)
        }
        if port < 1 || port > 65535 {
            return fmt.Errorf("port must be between 1 and 65535")
        }
        return nil
    })
    
    fs.SetValidator("log-level", func(val interface{}) error {
        level, ok := val.(string)
        if !ok {
            return fmt.Errorf("expected string, got %T", val)
        }
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
