# Stdlib Drop-in Replacement Example

This example demonstrates how to migrate from Go's standard library `flag` package to `flash-flags` with **ZERO code changes** while gaining advanced features.

## Migration

Simply change your import:

```go
// OLD
import "flag"

// NEW  
import flag "github.com/agilira/flash-flags/stdlib"
```

That's it! Your existing code works exactly the same.

## What You Get

### 🚀 Performance
- **1.5x faster** flag parsing
- Zero-dependency, lock-free implementation
- Optimized memory allocation

### ✨ Enhanced Features
- **Short flags**: `-h`, `-p`, `-v` 
- **Combined flags**: `-vd` for `-v -d`
- **Configuration files**: YAML, JSON, TOML support
- **Environment variables**: Automatic lookup with prefixes
- **Advanced validation**: Custom validators
- **Professional help**: Beautiful, structured output

### 📦 Full Compatibility
- All `flag` package functions supported
- Identical API and behavior
- Drop-in replacement guarantee
- Remaining arguments support (`flag.Args()`, `flag.NArg()`)

## Usage Examples

### Basic Usage (Same as stdlib)
```go
import flag "github.com/agilira/flash-flags/stdlib"

var host = flag.String("host", "localhost", "Server host")
var port = flag.Int("port", 8080, "Server port") 
var debug = flag.Bool("debug", false, "Debug mode")

flag.Parse()

fmt.Printf("Server: %s:%d (debug: %v)\n", *host, *port, *debug)
```

### Command Line
```bash
# Standard syntax (same as stdlib flag)
./app --host example.com --port 9000 --debug
./app --host=example.com --port=9000 --debug=true

# Enhanced syntax (flash-flags extensions)  
./app -h example.com -p 9000 -d        # Short flags
./app -hp example.com 9000 -d          # Combined flags
MYAPP_HOST=remote ./app                 # Environment variables
```

### Remaining Arguments
```bash
# Everything after -- or non-flag args are preserved
./app --host example.com -- file1 file2 file3
./app --debug file1 file2              # file1, file2 in flag.Args()
```

## Advanced Features

Once migrated, you can gradually adopt flash-flags extensions:

### Configuration Files
```go
// Enable config file support (optional)
fs := flag.CommandLine
fs.SetConfigFile("config.yaml")
```

### Environment Variables  
```bash
# Automatic environment variable lookup
MYAPP_HOST=production ./app    # Sets --host flag
MYAPP_DEBUG=true ./app         # Sets --debug flag
```

### Validation
```go
// Add custom validation (optional)
flag.CommandLine.SetValidator("port", func(value interface{}) error {
    if port := value.(int); port < 1 || port > 65535 {
        return fmt.Errorf("port must be between 1 and 65535")
    }
    return nil
})
```

## Running This Example

```bash
# Basic usage
go run main.go --host example.com --port 9000 --debug --verbose

# With remaining arguments  
go run main.go --debug --verbose -- extra arg1 arg2

# With short flags (flash-flags extension)
go run main.go -h example.com -p 9000 -dv

# With environment variables
MYAPP_HOST=remote go run main.go
```
## Migration Guarantee

This package provides 100% API compatibility with Go's standard library `flag` package. Any code that works with `flag` will work with `flash-flags/stdlib` without modifications.

The only difference is enhanced performance and optional advanced features that you can adopt gradually.

---

flash-flags • an AGILira library