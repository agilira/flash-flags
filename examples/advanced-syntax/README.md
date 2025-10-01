# Advanced Syntax Example

This example demonstrates the advanced flag parsing syntax features of flash-flags, including:

- **Short flag with equals assignment**: `-f=value`
- **Combined short flags**: `-abc` (equivalent to `-a -b -c`)

## Features Demonstrated

### 1. Short Flag Equals Assignment

You can assign values to short flags using the equals operator:

```bash
./advanced-syntax -n=myapp -p=9000
./advanced-syntax --name=myapp --port=9000
```

### 2. Combined Short Flags

Multiple boolean short flags can be combined into a single argument:

```bash
./advanced-syntax -vdf    # Equivalent to -v -d -f
./advanced-syntax -vq     # Equivalent to -v -q
```

### 3. Combined Flags with Values

The last flag in a combined sequence can take a value:

```bash
./advanced-syntax -vdn myapp    # Equivalent to -v -d -n myapp
./advanced-syntax -vdp 8090     # Equivalent to -v -d -p 8090
```

## Usage Examples

Build and run the example:

```bash
go build -o advanced-syntax
./advanced-syntax
```

### Standard Syntax
```bash
./advanced-syntax --verbose --name myapp --port 9000
./advanced-syntax -v -n myapp -p 9000
```

### Advanced Equals Syntax
```bash
./advanced-syntax -n=myapp -p=9000 --timeout=30s
./advanced-syntax --name=myapp --port=9000 --timeout=30s
```

### Combined Short Flags
```bash
./advanced-syntax -vdf
./advanced-syntax -vq -n=testapp
./advanced-syntax -vdn myapp
./advanced-syntax -vdp 8090
```

### Mixed Usage
```bash
./advanced-syntax -vd --name=server -p=3000 -c=/etc/config.json
./advanced-syntax -vqf --timeout=1m --config=/path/to/config
```

## Performance

flash-flags maintains superior performance while providing full POSIX/GNU syntax compatibility:

- **33% faster than pflags** in benchmarks
- **Zero-allocation parsing** for optimal performance
- **Lock-free architecture** for concurrent safety

## Implementation Notes

- Combined flags work with boolean flags and value flags
- Only the last flag in a combined sequence can accept a value
- Equals assignment works with both short (`-f=`) and long (`--flag=`) flags
- Full compatibility with POSIX and GNU flag parsing conventions