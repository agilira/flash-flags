# Configuration Files Example

Demonstrates how flash-flags integrates with JSON configuration files for flexible application setup.

## What it demonstrates

- **JSON configuration support**: Load flags from config files
- **Configuration hierarchy**: Command line overrides config files
- **Automatic config discovery**: Searches for config files in standard locations
- **Config file creation**: Dynamic generation of sample configurations
- **Mixed sources**: Combining CLI arguments with config files

## Usage

```bash
# Run with automatic config file detection
go run main.go

# Specify a custom config file
go run main.go --config myapp.json

# Override config values with CLI arguments
go run main.go --config myapp.json --host 0.0.0.0 --port 9000

# Show help
go run main.go --help
```

## Configuration format

The example uses JSON configuration files:

```json
{
  "host": "config.example.com",
  "port": 9090,
  "ssl": true,
  "workers": 4,
  "log-level": "debug",
  "timeout": 30.5,
  "features": ["auth", "metrics", "cache"]
}
```

## Key features shown

### Automatic Discovery
The application searches for config files in:
- Current directory: `./config.json`, `./app.json`
- User config: `~/.config/myapp/config.json`
- System config: `/etc/myapp/config.json`

### Priority Order
1. **Command line arguments** (highest priority)
2. **Configuration files**
3. **Default values** (lowest priority)

### Error Handling
- Graceful handling of missing config files
- JSON parsing error reporting
- Validation after config loading

## Build

```bash
go build -o config
./config
```

The example will create a sample `demo-config.json` file to demonstrate the functionality.