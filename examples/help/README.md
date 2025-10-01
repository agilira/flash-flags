# Help System Example

Demonstrates flash-flags' advanced help system with groups, custom formatting, and comprehensive documentation using a web server application.

## What it demonstrates

- **Organized help output**: Flags grouped by functionality
- **Custom descriptions**: Detailed help text for complex applications
- **Usage examples**: Built-in usage examples and patterns
- **Version information**: Application versioning support
- **Professional formatting**: Clean, readable help output
- **Group management**: Logical organization of related flags

## Usage

```bash
# Show comprehensive help
go run main.go --help
go run main.go -h

# Show version information
go run main.go --version

# Run with example configuration
go run main.go \
  --host 0.0.0.0 \
  --port 8443 \
  --ssl \
  --cert /etc/ssl/server.crt \
  --key /etc/ssl/server.key \
  --db-host localhost \
  --db-user admin \
  --db-password secret \
  --log-level debug \
  --verbose
```

## Key features shown

### Flag Groups
Flags are organized into logical groups for better readability:

- **Server Configuration**: Host, port, SSL settings
- **Database Configuration**: Connection parameters and credentials  
- **Logging Configuration**: Log levels and output options
- **Monitoring**: Metrics and performance settings
- **Security**: Authentication and authorization

### Professional Help Output
The help system generates clean, professional documentation:
- Clear section headers
- Consistent formatting
- Type information for each flag
- Default values display
- Required flag indicators

### Usage Examples
Built-in examples show common usage patterns:
- Basic server setup
- SSL configuration
- Database connection
- Development vs production settings

### Custom Descriptions
Rich descriptions help users understand:
- What each flag does
- When to use specific options
- Valid value formats
- Dependencies between flags

## Help output structure

```
Usage: webserver [options]

Description:
    A fast web server with advanced configuration options.
    Supports SSL/TLS, database connections, and monitoring.

Server Configuration:
    --host HOST            Server bind address (default: localhost)
    --port PORT            Server port number (default: 8080)
    --ssl                  Enable SSL/TLS encryption

Database Configuration:
    --db-host HOST         Database server address
    --db-user USER         Database username  
    --db-password PASS     Database password

... (additional groups)

Examples:
    webserver --host 0.0.0.0 --port 8443 --ssl
    webserver --db-host localhost --db-user admin --log-level debug
```

## Build

```bash
go build -o help
./help --help
```