# FlashFlags Demo Examples

This directory contains working examples demonstrating FlashFlags capabilities.

## Available Demos

| Demo | Description | Features Demonstrated |
|------|-------------|----------------------|
| **[advanced-syntax/](advanced-syntax/)** | POSIX/GNU flag syntax compatibility | `-f=value` assignment, `-abc` combined flags |
| **[basic/](basic/)** | Complete web server configuration | Groups, validation, dependencies, help system |
| **[config/](config/)** | Configuration file integration | JSON config files, auto-discovery |
| **[env/](env/)** | Environment variable integration | Env var lookup, prefixes, custom names |
| **[help/](help/)** | Help system showcase | Custom help, groups, descriptions |
| **[required/](required/)** | Required flags and dependencies | Required flags, dependencies, validation errors |

## Quick Start

### Try the Advanced Syntax Demo

```bash
cd advanced-syntax
go run main.go
```

This will show examples of POSIX/GNU compatible syntax including `-f=value` and `-abc` combined flags.

### Run the Basic Demo

```bash
cd basic
go run main.go --help
```

This will show the comprehensive help system with grouped options.

### Test with Real Arguments

```bash
cd basic
go run main.go \
  --host 0.0.0.0 \
  --port 8443 \
  --ssl \
  --cert /etc/ssl/server.crt \
  --key /etc/ssl/server.key \
  --db-user admin \
  --db-password secret123 \
  --log-level debug \
  --verbose
```

### Try Configuration Files

```bash
cd config
# The demo will automatically look for config files
go run main.go
```

### Test Environment Variables

```bash
cd env
export MYAPP_HOST=0.0.0.0
export MYAPP_PORT=3000
go run main.go
```

## Demo Structure

Each demo is self-contained with its own `go.mod` file, so you can run them independently or copy them to your own projects.

## Learning Path

1. **Start with `advanced-syntax/`** - New POSIX/GNU syntax features
2. **Try `basic/`** - Shows all major features
3. **Explore `help/`** - Understanding the help system
4. **Test `config/`** - Configuration file integration
5. **Check `env/`** - Environment variable support
6. **Validate with `required/`** - Validation and constraints

## Building and Running

Each demo can be built and run independently:

```bash
# Navigate to any demo directory
cd basic/

# Run directly
go run main.go [flags]

# Or build first
go build -o demo main.go
./demo [flags]
```

## Integration Tips

These demos show patterns you can use in your own applications:

- **Flag organization** with groups for better help output
- **Validation** for ensuring correct input values
- **Dependencies** between flags
- **Configuration hierarchy** (CLI > env vars > config files > defaults)
- **Error handling** and user-friendly messages

---

flash-flags • an AGILira library