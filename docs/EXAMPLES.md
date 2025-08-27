# FlashFlags Demo Examples

The `demo/` directory contains practical examples demonstrating how to use FlashFlags in real-world scenarios.

## Available Examples

### Basic Examples

- **[basic/](../demo/basic/)** - Simple flag parsing example
- **[help/](../demo/help/)** - Help system demonstration
- **[config/](../demo/config/)** - Configuration file integration
- **[env/](../demo/env/)** - Environment variable integration
- **[required/](../demo/required/)** - Required flags and validation

## Running Examples

```bash
# Navigate to demo directory
cd demo

# Run basic example
go run basic/main.go --help
go run basic/main.go --name "John" --port 8080 --verbose

# Run help system demo
go run help/main.go --help

# Run with config file
go run config/main.go

# Run with environment variables
go run env/main.go

# Run required flags demo
go run required/main.go --help
```

## Demo Structure

Each demo is self-contained in its own directory:

```
demo/
├── main.go          # Package declaration
├── go.mod           # Module file
├── basic/           # Basic usage examples
├── config/          # Configuration file examples
├── env/             # Environment variable examples
├── help/            # Help system examples
└── required/        # Required flags and validation
```

## Building Demos

```bash
# Navigate to a specific demo
cd demo/basic
go build -o basic main.go

# Or run directly
go run main.go --help
```
---

flash-flags • an AGILira library