# Required Flags Demo

This demo shows how to use FlashFlags' required flags functionality, including flag dependencies.

## Features Demonstrated

- **Required Flags**: Flags that must be provided by the user
- **Flag Dependencies**: Flags that are required only when other flags are set
- **Error Handling**: Proper validation and user-friendly error messages
- **Help Integration**: Professional help output with grouped flags

## Running the Demo

### Show Help
```bash
go run main.go --help
```

### Missing Required Flags (will fail)
```bash
go run main.go --host localhost --port 3000
# Error: required flag --api-key not provided
```

### Provide Required Flags (will succeed)
```bash
go run main.go --api-key "your-secret-key" --db-url "postgres://user@localhost/db"
```

### TLS Example with Dependencies
```bash
# This will fail because TLS is enabled but cert/key are missing
go run main.go --api-key "key123" --db-url "postgres://user@localhost/db" --enable-tls

# This will succeed - TLS enabled with required cert and key
go run main.go \
  --api-key "your-secret-key" \
  --db-url "postgres://user:pass@localhost/mydb" \
  --db-password "dbpass123" \
  --enable-tls \
  --tls-cert "/path/to/cert.pem" \
  --tls-key "/path/to/key.pem" \
  --host "0.0.0.0" \
  --port 8443
```

## Code Structure

### Required Flags
The demo marks these flags as absolutely required:
- `--api-key`: Authentication token
- `--db-url`: Database connection string

### Conditional Requirements (Dependencies)
These flags are required only when their dependencies are set:
- `--tls-cert` and `--tls-key`: Required when `--enable-tls` is true
- `--db-password`: Required when `--db-url` is provided

### Implementation Details

```go
// Mark a flag as required
fs.SetRequired("api-key")

// Set up dependencies
fs.SetDependencies("tls-cert", "enable-tls")
fs.SetDependencies("tls-key", "enable-tls")
```

## Error Messages

FlashFlags provides clear error messages for validation failures:

```
Configuration Error: required flag --api-key not provided

Run with --help to see all available options
```

## Security Features

The demo includes security best practices:
- **Secret Masking**: API keys and passwords are masked in output
- **Connection String Protection**: Database URLs are partially hidden
- **Clear Separation**: Required vs optional flags are clearly documented

## Real-World Usage

This pattern is common in production applications where:
- API keys are mandatory for external service access
- Database connections require proper configuration
- TLS setup needs both certificate and key files
- Different deployment environments have different requirements

The required flags feature ensures your application fails fast with clear error messages rather than running with incomplete configuration.

---

flash-flags • an AGILira library
