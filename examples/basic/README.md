# Basic Example

A comprehensive demonstration of flash-flags features using a web server configuration scenario.

## What it demonstrates

- **Flag groups**: Server, Database, and Logging configurations
- **Multiple flag types**: String, Int, Bool, StringSlice
- **Short keys**: Quick access with single-letter flags (`-p`, `-H`, `-s`)
- **Validation**: Custom validators for ports and log levels
- **Dependencies**: SSL requires cert and key files
- **Required flags**: Database credentials when using database
- **Help system**: Organized help output with groups and descriptions

## Usage

```bash
# Build and run
go run main.go

# Show help
go run main.go --help

# Complete server configuration
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

# Using short flags
go run main.go -H 0.0.0.0 -p 8080 -v
```

## Key features shown

### Flag Groups
Flags are organized into logical groups for better help output:
- **Server Configuration**: host, port, SSL settings
- **Database Configuration**: connection parameters
- **Logging Configuration**: verbosity and log levels

### Validation
- Port numbers must be between 1-65535
- Log levels must be valid (debug, info, warn, error)
- SSL dependencies are enforced

### Error Handling
The example shows proper error handling and user-friendly validation messages.

## Build

```bash
go build -o basic
./basic --help
```

---

flash-flags • an AGILira library