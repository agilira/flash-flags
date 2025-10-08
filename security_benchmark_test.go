// security_benchmark_test.go - Security-Focused Performance Benchmarks for Flash-Flags
//
// This file implements performance benchmarks specifically for security-critical functions
// to ensure that hardening measures don't significantly impact performance.
//
// BENCHMARKED FUNCTIONS:
// - Parse with various input sizes and complexity
// - parseStringSlice with different CSV inputs
// - LoadConfig with JSON parsing
// - LoadEnvironmentVariables with various env sizes
// - Validation functions with security checks
//
// PERFORMANCE TARGETS:
// - Parse: < 10μs for typical arguments (< 20 flags)
// - parseStringSlice: < 1μs for typical CSV (< 100 items)
// - LoadConfig: < 1ms for typical JSON (< 10KB)
// - Validation: < 100ns per flag
//
// These benchmarks help detect performance regressions introduced by security hardening.
//
// Copyright (c) 2025 AGILira - A. Giordano
// Series: an AGILira library
// SPDX-License-Identifier: MPL-2.0

package flashflags

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"
)

// =============================================================================
// CORE PARSING BENCHMARKS
// =============================================================================

// BenchmarkParse_Typical benchmarks typical flag parsing workload
func BenchmarkParse_Typical(b *testing.B) {
	args := []string{
		"--host", "example.com",
		"--port", "8080",
		"--timeout", "30s",
		"--debug",
		"--tags", "web,api,prod",
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		fs := New("benchmark")
		host := fs.StringVar("host", "h", "localhost", "Server host")
		port := fs.IntVar("port", "p", 8080, "Server port")
		timeout := fs.Duration("timeout", 30*time.Second, "Request timeout")
		debug := fs.Bool("debug", false, "Debug mode")
		tags := fs.StringSlice("tags", []string{}, "Service tags")

		err := fs.Parse(args)
		if err != nil {
			b.Fatal(err)
		}

		// Prevent optimization
		_ = *host
		_ = *port
		_ = *timeout
		_ = *debug
		_ = *tags
	}
}

// BenchmarkParse_ManyFlags benchmarks parsing with many flags
func BenchmarkParse_ManyFlags(b *testing.B) {
	// Generate many flags
	var args []string
	for i := 0; i < 20; i++ {
		args = append(args, "--flag"+strconv.Itoa(i), "value"+strconv.Itoa(i))
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		fs := New("benchmark")

		// Define many flags
		var ptrs []*string
		for j := 0; j < 20; j++ {
			ptr := fs.String("flag"+strconv.Itoa(j), "default", "Flag "+strconv.Itoa(j))
			ptrs = append(ptrs, ptr)
		}

		err := fs.Parse(args)
		if err != nil {
			b.Fatal(err)
		}

		// Prevent optimization
		for _, ptr := range ptrs {
			_ = *ptr
		}
	}
}

// BenchmarkParse_LongValues benchmarks parsing with long values (potential security risk)
func BenchmarkParse_LongValues(b *testing.B) {
	longValue := strings.Repeat("A", 10000)
	args := []string{
		"--host", longValue,
		"--config", longValue,
		"--description", longValue,
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		fs := New("benchmark")
		host := fs.String("host", "localhost", "Server host")
		config := fs.String("config", "", "Config path")
		desc := fs.String("description", "", "Description")

		err := fs.Parse(args)
		if err != nil {
			b.Fatal(err)
		}

		// Prevent optimization
		_ = *host
		_ = *config
		_ = *desc
	}
}

// BenchmarkParse_ComplexCombined benchmarks complex combined short flags
func BenchmarkParse_ComplexCombined(b *testing.B) {
	args := []string{
		"-vdqh", "example.com", // Combined: verbose, debug, quiet, host
		"-p", "8080", // Port
		"-abc", // All boolean flags
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		fs := New("benchmark")
		verbose := fs.BoolVar("verbose", "v", false, "Verbose")
		debug := fs.BoolVar("debug", "d", false, "Debug")
		quiet := fs.BoolVar("quiet", "q", false, "Quiet")
		host := fs.StringVar("host", "h", "localhost", "Host")
		port := fs.IntVar("port", "p", 8080, "Port")
		flagA := fs.BoolVar("flag-a", "a", false, "Flag A")
		flagB := fs.BoolVar("flag-b", "b", false, "Flag B")
		flagC := fs.BoolVar("flag-c", "c", false, "Flag C")

		err := fs.Parse(args)
		if err != nil {
			b.Fatal(err)
		}

		// Prevent optimization
		_ = *verbose
		_ = *debug
		_ = *quiet
		_ = *host
		_ = *port
		_ = *flagA
		_ = *flagB
		_ = *flagC
	}
}

// =============================================================================
// STRING SLICE PARSING BENCHMARKS
// =============================================================================

// BenchmarkParseStringSlice_Small benchmarks small CSV parsing
func BenchmarkParseStringSlice_Small(b *testing.B) {
	fs := &FlagSet{}
	csvValue := "web,api,database,cache,monitoring"

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		result := fs.parseStringSlice(csvValue)
		_ = result // Prevent optimization
	}
}

// BenchmarkParseStringSlice_Medium benchmarks medium CSV parsing
func BenchmarkParseStringSlice_Medium(b *testing.B) {
	fs := &FlagSet{}
	// Generate 100 items
	items := make([]string, 100)
	for i := range items {
		items[i] = "item" + string(rune(i%26+'a'))
	}
	csvValue := strings.Join(items, ",")

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		result := fs.parseStringSlice(csvValue)
		_ = result
	}
}

// BenchmarkParseStringSlice_Large benchmarks large CSV parsing (stress test)
func BenchmarkParseStringSlice_Large(b *testing.B) {
	fs := &FlagSet{}
	// Generate 10,000 items
	items := make([]string, 10000)
	for i := range items {
		items[i] = "service" + string(rune(i%1000))
	}
	csvValue := strings.Join(items, ",")

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		result := fs.parseStringSlice(csvValue)
		_ = result
	}
}

// BenchmarkParseStringSlice_EmptyAndEdges benchmarks edge cases
func BenchmarkParseStringSlice_EmptyAndEdges(b *testing.B) {
	fs := &FlagSet{}
	testCases := []string{
		"",     // Empty
		",",    // Single comma
		",,",   // Double comma
		",a,",  // Commas around
		"a,,b", // Double comma middle
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		csvValue := testCases[i%len(testCases)]
		result := fs.parseStringSlice(csvValue)
		_ = result
	}
}

// =============================================================================
// CONFIGURATION LOADING BENCHMARKS
// =============================================================================

// BenchmarkLoadConfig_Small benchmarks small JSON config loading
func BenchmarkLoadConfig_Small(b *testing.B) {
	tmpDir := b.TempDir()
	configPath := filepath.Join(tmpDir, "small_config.json")

	jsonContent := `{
		"host": "localhost",
		"port": 8080,
		"debug": true,
		"timeout": "30s",
		"tags": ["web", "api"]
	}`

	if err := os.WriteFile(configPath, []byte(jsonContent), 0644); err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		fs := New("benchmark")
		fs.String("host", "localhost", "Host")
		fs.Int("port", 8080, "Port")
		fs.Bool("debug", false, "Debug")
		fs.Duration("timeout", 30*time.Second, "Timeout")
		fs.StringSlice("tags", []string{}, "Tags")
		fs.SetConfigFile(configPath)

		err := fs.LoadConfig()
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkLoadConfig_Medium benchmarks medium JSON config loading
func BenchmarkLoadConfig_Medium(b *testing.B) {
	tmpDir := b.TempDir()
	configPath := filepath.Join(tmpDir, "medium_config.json")

	// Generate medium-sized JSON with nested objects and arrays
	jsonContent := `{
		"server": {
			"host": "example.com",
			"port": 8080,
			"ssl": {
				"enabled": true,
				"cert": "/path/to/cert.pem",
				"key": "/path/to/key.pem"
			}
		},
		"database": {
			"hosts": ["db1.example.com", "db2.example.com", "db3.example.com"],
			"port": 5432,
			"pool_size": 20,
			"timeout": "10s"
		},
		"features": ["auth", "cache", "monitoring", "logging", "metrics"],
		"environment": "production"
	}`

	if err := os.WriteFile(configPath, []byte(jsonContent), 0644); err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		fs := New("benchmark")
		fs.String("environment", "dev", "Environment")
		fs.StringSlice("features", []string{}, "Features")
		fs.SetConfigFile(configPath)

		err := fs.LoadConfig()
		if err != nil {
			b.Fatal(err)
		}
	}
}

// =============================================================================
// ENVIRONMENT VARIABLE BENCHMARKS
// =============================================================================

// BenchmarkLoadEnvironmentVariables_Few benchmarks few env vars
func BenchmarkLoadEnvironmentVariables_Few(b *testing.B) {
	// Set up environment variables
	testVars := map[string]string{
		"BENCH_HOST":    "example.com",
		"BENCH_PORT":    "8080",
		"BENCH_DEBUG":   "true",
		"BENCH_TIMEOUT": "30s",
	}

	// Backup and set test vars
	backup := make(map[string]string)
	for key, value := range testVars {
		backup[key] = os.Getenv(key)
		_ = os.Setenv(key, value)
	}
	defer func() {
		for key, value := range backup {
			if value != "" {
				_ = os.Setenv(key, value)
			} else {
				_ = os.Unsetenv(key)
			}
		}
	}()

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		fs := New("benchmark")
		fs.String("host", "localhost", "Host")
		fs.Int("port", 8080, "Port")
		fs.Bool("debug", false, "Debug")
		fs.Duration("timeout", 30*time.Second, "Timeout")
		fs.SetEnvPrefix("BENCH")

		err := fs.LoadEnvironmentVariables()
		if err != nil {
			b.Fatal(err)
		}
	}
}

// =============================================================================
// VALIDATION BENCHMARKS
// =============================================================================

// BenchmarkValidation_Simple benchmarks simple validation functions
func BenchmarkValidation_Simple(b *testing.B) {
	args := []string{"--port", "8080", "--host", "example.com"}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		fs := New("benchmark")
		host := fs.String("host", "localhost", "Host")
		port := fs.Int("port", 8080, "Port")

		// Add simple validators
		_ = fs.SetValidator("host", func(val interface{}) error {
			h := val.(string)
			if len(h) == 0 {
				return nil // Don't fail in benchmark
			}
			return nil
		})

		_ = fs.SetValidator("port", func(val interface{}) error {
			p := val.(int)
			if p <= 0 || p > 65535 {
				return nil // Don't fail in benchmark
			}
			return nil
		})

		err := fs.Parse(args)
		if err != nil {
			b.Fatal(err)
		}

		_ = *host
		_ = *port
	}
}

// BenchmarkValidation_Complex benchmarks complex validation with security checks
func BenchmarkValidation_Complex(b *testing.B) {
	args := []string{
		"--config", "/etc/myapp/config.json",
		"--host", "web-server-01.example.com",
		"--tags", "production,web,api,v2.1.0",
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		fs := New("benchmark")
		config := fs.String("config", "", "Config path")
		host := fs.String("host", "localhost", "Host")
		tags := fs.StringSlice("tags", []string{}, "Tags")

		// Add security-focused validators
		_ = fs.SetValidator("config", func(val interface{}) error {
			path := val.(string)
			// Simulate path validation (security check)
			if strings.Contains(path, "..") {
				return nil // Don't fail in benchmark
			}
			if strings.Contains(path, "\x00") {
				return nil // Don't fail in benchmark
			}
			return nil
		})

		_ = fs.SetValidator("host", func(val interface{}) error {
			h := val.(string)
			// Simulate hostname validation
			if len(h) > 253 {
				return nil // Don't fail in benchmark
			}
			if strings.ContainsAny(h, "\x00\r\n") {
				return nil // Don't fail in benchmark
			}
			return nil
		})

		_ = fs.SetValidator("tags", func(val interface{}) error {
			tagList := val.([]string)
			// Simulate tag validation
			if len(tagList) > 100 {
				return nil // Don't fail in benchmark
			}
			for _, tag := range tagList {
				if len(tag) > 50 {
					return nil // Don't fail in benchmark
				}
			}
			return nil
		})

		err := fs.Parse(args)
		if err != nil {
			b.Fatal(err)
		}

		_ = *config
		_ = *host
		_ = *tags
	}
}

// =============================================================================
// MEMORY ALLOCATION BENCHMARKS
// =============================================================================

// BenchmarkParse_ZeroAlloc benchmarks for zero-allocation parsing (ideal case)
func BenchmarkParse_ZeroAlloc(b *testing.B) {
	args := []string{"--host", "localhost", "--port", "8080"}

	fs := New("benchmark")
	host := fs.String("host", "localhost", "Host")
	port := fs.Int("port", 8080, "Port")

	// Parse once to set up state
	err := fs.Parse(args)
	if err != nil {
		b.Fatal(err)
	}

	// Reset and measure subsequent parses
	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		fs.Reset() // Reset to default values
		err := fs.Parse(args)
		if err != nil {
			b.Fatal(err)
		}

		_ = *host
		_ = *port
	}
}

// =============================================================================
// COMPARATIVE SECURITY BENCHMARKS
// =============================================================================

// BenchmarkSecurityOverhead_Minimal benchmarks minimal security checking overhead
func BenchmarkSecurityOverhead_Minimal(b *testing.B) {
	args := []string{"--input", "normal_value"}

	b.Run("NoValidation", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			fs := New("benchmark")
			input := fs.String("input", "", "Input value")

			err := fs.Parse(args)
			if err != nil {
				b.Fatal(err)
			}
			_ = *input
		}
	})

	b.Run("WithValidation", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			fs := New("benchmark")
			input := fs.String("input", "", "Input value")

			// Add minimal security validation
			_ = fs.SetValidator("input", func(val interface{}) error {
				s := val.(string)
				if len(s) > 10000 {
					return nil // Don't fail in benchmark
				}
				if strings.ContainsAny(s, "\x00\r\n") {
					return nil // Don't fail in benchmark
				}
				return nil
			})

			err := fs.Parse(args)
			if err != nil {
				b.Fatal(err)
			}
			_ = *input
		}
	})
}
