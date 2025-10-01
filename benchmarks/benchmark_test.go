// benchmark_test.go: Performance benchmarks for flag parsing libraries
//
// Copyright (c) 2025 AGILira - A. Giordano
// Series: an AGILira library
// SPDX-License-Identifier: MPL-2.0

package benchmarks

import (
	"flag"
	"testing"

	flashflags "github.com/agilira/flash-flags"
	"github.com/alecthomas/kingpin/v2"
	"github.com/jessevdk/go-flags"
	"github.com/spf13/pflag"
)

// =============================================================================
// CORE COMPARISON BENCHMARKS
// =============================================================================
//
// These benchmarks test the same scenario across all libraries:
// - Parse 3 flags: string (--env), bool (--verbose), string (--timeout)
// - Arguments: --env staging --verbose --timeout 60
// - Simulate realistic CLI flag parsing workload

// BenchmarkFlashFlags tests our flash-flags library performance
func BenchmarkFlashFlags(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		fs := flashflags.New("benchmark")
		env := fs.StringVar("env", "e", "prod", "Environment")
		verbose := fs.BoolVar("verbose", "v", false, "Verbose output")
		timeout := fs.StringVar("timeout", "t", "30", "Timeout in seconds")

		args := []string{"--env", "staging", "--verbose", "--timeout", "60"}
		err := fs.Parse(args)
		if err != nil {
			b.Fatal(err)
		}

		// Simulate some work - access the parsed values
		_ = *env
		_ = *verbose
		_ = *timeout
	}
}

// BenchmarkStdFlag tests Go's standard library flag package (baseline)
func BenchmarkStdFlag(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Reset flag state for clean test
		fs := flag.NewFlagSet("benchmark", flag.ExitOnError)

		env := fs.String("env", "prod", "Environment")
		verbose := fs.Bool("verbose", false, "Verbose output")
		timeout := fs.String("timeout", "30", "Timeout in seconds")

		args := []string{"-env", "staging", "-verbose", "-timeout", "60"}
		err := fs.Parse(args)
		if err != nil {
			b.Fatal(err)
		}

		// Access parsed values
		_ = *env
		_ = *verbose
		_ = *timeout
	}
}

// BenchmarkPflag tests spf13/pflag (POSIX-compliant flag replacement)
func BenchmarkPflag(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		fs := pflag.NewFlagSet("benchmark", pflag.ExitOnError)
		env := fs.StringP("env", "e", "prod", "Environment")
		verbose := fs.BoolP("verbose", "v", false, "Verbose output")
		timeout := fs.StringP("timeout", "t", "30", "Timeout in seconds")

		args := []string{"--env", "staging", "--verbose", "--timeout", "60"}
		err := fs.Parse(args)
		if err != nil {
			b.Fatal(err)
		}

		// Access parsed values
		_ = *env
		_ = *verbose
		_ = *timeout
	}
}

// BenchmarkGoFlags tests jessevdk/go-flags (struct-based flag parsing)
func BenchmarkGoFlags(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var opts struct {
			Env     string `short:"e" long:"env" default:"prod" description:"Environment"`
			Verbose bool   `short:"v" long:"verbose" description:"Verbose output"`
			Timeout string `short:"t" long:"timeout" default:"30" description:"Timeout in seconds"`
		}

		args := []string{"--env", "staging", "--verbose", "--timeout", "60"}
		_, err := flags.ParseArgs(&opts, args)
		if err != nil {
			b.Fatal(err)
		}

		// Simulate some work
		_ = opts.Env
		_ = opts.Verbose
		_ = opts.Timeout
	}
}

// BenchmarkKingpin tests alecthomas/kingpin (command-line parser with validation)
func BenchmarkKingpin(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		app := kingpin.New("benchmark", "Benchmark application")
		env := app.Flag("env", "Environment").Short('e').Default("prod").String()
		verbose := app.Flag("verbose", "Verbose output").Short('v').Bool()
		timeout := app.Flag("timeout", "Timeout in seconds").Short('t').Default("30").String()

		args := []string{"--env", "staging", "--verbose", "--timeout", "60"}
		_, err := app.Parse(args)
		if err != nil {
			b.Fatal(err)
		}

		// Access parsed values
		_ = *env
		_ = *verbose
		_ = *timeout
	}
}
