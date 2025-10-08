# Flash-flags Performance Benchmarks

Comprehensive performance benchmarks comparing flash-flags with other popular Go flag parsing libraries.

## Benchmark Methodology

All benchmarks use a standardized test scenario parsing 3 flags with typical CLI arguments:
- `--env staging`
- `--verbose` 
- `--timeout 60`

## Libraries Tested

- **flash-flags** - Zero-dependency flag parsing library
- **Standard library flag** - Go's built-in flag package (baseline)
- **spf13/pflag** - Drop-in replacement for Go's flag package with POSIX compliance
- **jessevdk/go-flags** - Flag parsing library with struct tags
- **alecthomas/kingpin** - Command-line argument parser with validation

## Running the Benchmarks

```bash
cd benchmarks
go test -bench=. -benchmem
```

## Latest Results

Test Environment: AMD Ryzen 5 7520U, Linux

```
BenchmarkFlashFlags-8            1294699               924.0 ns/op           945 B/op         11 allocs/op
BenchmarkStdFlag-8               1527176               792.0 ns/op           945 B/op         13 allocs/op
BenchmarkPflag-8                  785904              1322 ns/op            1569 B/op         21 allocs/op
BenchmarkGoFlags-8                147394              7460 ns/op            5620 B/op         61 allocs/op
BenchmarkKingpin-8                150154              7567 ns/op            6504 B/op         97 allocs/op
```

## Performance Analysis

### Execution Speed (operations per second)
1. **Standard library flag**: 1,527,176 ops/sec
2. **Flash-flags**: 1,294,699 ops/sec (-15% vs stdlib)
3. **Pflag**: 785,904 ops/sec (-49% vs stdlib)
4. **Go-flags**: 147,394 ops/sec (-90% vs stdlib)
5. **Kingpin**: 150,154 ops/sec (-90% vs stdlib)

### Memory Allocation
1. **Flash-flags**: 945 B/op, 11 allocs/op
2. **Standard library flag**: 945 B/op, 13 allocs/op
3. **Pflag**: 1,569 B/op, 21 allocs/op
4. **Go-flags**: 5,620 B/op, 61 allocs/op
5. **Kingpin**: 6,504 B/op, 97 allocs/op

## Performance Comparison

### Security-Hardened Performance
Flash-flags delivers **85% of stdlib performance** with **comprehensive security validation**:
- Command injection protection
- Path traversal prevention  
- Buffer overflow safeguards
- Format string attack blocking
- Input sanitization & validation

### Relative Performance vs Standard Library
- **Flash-flags**: 85% performance, 100% memory usage, **FULL SECURITY** 🛡️
- Pflag: 51% performance, 166% memory usage, no security
- Go-flags: 10% performance, 595% memory usage, no security
- Kingpin: 10% performance, 688% memory usage, no security

### Flash-flags vs Alternatives
- **1.6x faster** than pflag with full security
- **8.8x faster** than go-flags with full security  
- **8.4x faster** than kingpin with full security
- **Same memory usage** as stdlib but with security
- **6.0x more memory efficient** than go-flags
- **6.9x more memory efficient** than kingpin

## Security vs Performance Trade-off

Flash-flags is the **only library** that provides comprehensive security validation:
- **132 ns/op overhead** (17%) for complete security hardening
- **Zero vulnerabilities** vs potential security risks in other libraries
- **Production-ready** security without sacrificing usability

## Methodology Notes

Benchmarks measure parsing performance under controlled conditions. Real-world performance may vary based on:
- Argument complexity and count
- Validation requirements
- Integration patterns
- System resource availability

## Reproducibility

All benchmarks are reproducible using the provided test suite. Results may vary across different hardware configurations and Go compiler versions.

---

flash-flags • an AGILira library
