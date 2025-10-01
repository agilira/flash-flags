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
BenchmarkFlashFlags-8            1259883               955.3 ns/op          1153 B/op         12 allocs/op
BenchmarkStdFlag-8               1518873               779.5 ns/op           945 B/op         13 allocs/op
BenchmarkPflag-8                  714727              1511 ns/op            1761 B/op         23 allocs/op
BenchmarkGoFlags-8                148956              7599 ns/op            5620 B/op         61 allocs/op
BenchmarkKingpin-8                152389              7351 ns/op            6504 B/op         97 allocs/op
```

## Performance Analysis

### Execution Speed (operations per second)
1. **Standard library flag**: 1,518,873 ops/sec
2. **Flash-flags**: 1,259,883 ops/sec (-17% vs stdlib)
3. **Pflag**: 714,727 ops/sec (-53% vs stdlib)
4. **Kingpin**: 152,389 ops/sec (-90% vs stdlib)
5. **Go-flags**: 148,956 ops/sec (-90% vs stdlib)

### Memory Allocation
1. **Standard library flag**: 945 B/op, 13 allocs/op
2. **Flash-flags**: 1,153 B/op, 12 allocs/op
3. **Pflag**: 1,761 B/op, 23 allocs/op
4. **Go-flags**: 5,620 B/op, 61 allocs/op
5. **Kingpin**: 6,504 B/op, 97 allocs/op

## Performance Comparison

### Relative Performance vs Standard Library
- Flash-flags: 83% performance, 122% memory usage
- Pflag: 47% performance, 186% memory usage
- Go-flags: 10% performance, 595% memory usage
- Kingpin: 10% performance, 688% memory usage

### Flash-flags vs Alternatives
- 1.8x faster than pflag
- 8.5x faster than go-flags
- 8.3x faster than kingpin
- 4.9x more memory efficient than go-flags
- 5.6x more memory efficient than kingpin

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
