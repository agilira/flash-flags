# Environment Variables Example

Shows how flash-flags automatically integrates with environment variables for configuration management.

## What it demonstrates

- **Environment variable integration**: Automatic env var lookup
- **Custom env var prefixes**: Organize variables with prefixes
- **Custom env var names**: Override default naming conventions  
- **Priority hierarchy**: CLI > Environment > Config > Defaults
- **Multiple data types**: String, Int, Bool, Duration support

## Usage

```bash
# Run without environment variables
go run main.go

# Set environment variables and run
export ENVDEMO_HOST=0.0.0.0
export ENVDEMO_PORT=3000
export ENVDEMO_DEBUG=true
go run main.go

# Override env vars with CLI arguments
ENVDEMO_PORT=8080 go run main.go --port 9000

# Show help
go run main.go --help
```

## Environment variable naming

### Default naming convention
- Flag: `--host` → Env var: `ENVDEMO_HOST`
- Flag: `--port` → Env var: `ENVDEMO_PORT`
- Flag: `--debug` → Env var: `ENVDEMO_DEBUG`
- Flag: `--max-connections` → Env var: `ENVDEMO_MAX_CONNECTIONS`

### Custom naming
The example demonstrates environment variables with the app prefix `ENVDEMO_`.

## Key features shown

### Automatic Integration
flash-flags automatically looks up environment variables for each flag using standard naming conventions.

### Type Support
Environment variables work with all flag types:
- **String**: Direct string values
- **Int**: Parsed as integers with validation
- **Bool**: Accepts `true`/`false`, `1`/`0`, `yes`/`no`
- **Duration**: Parsed as Go duration strings (`30s`, `5m`, etc.)

### Priority Order
1. **Command line arguments** (highest priority)
2. **Environment variables**
3. **Configuration files**
4. **Default values** (lowest priority)

## Example environment setup

```bash
# Server configuration
export ENVDEMO_HOST=0.0.0.0
export ENVDEMO_PORT=8080
export ENVDEMO_DEBUG=true
export ENVDEMO_MAX_CONNECTIONS=200

# Run the application
go run main.go
```

## Build

```bash
go build -o env
./env
```

---

flash-flags • an AGILira library