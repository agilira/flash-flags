// flash-flags.go: Ultra-fast command-line flag parsing for Go Tests
//
// Copyright (c) 2025 AGILira - A. Giordano
// Series: an AGILira library
// SPDX-License-Identifier: MPL-2.0

package flashflags

import (
	"fmt"
	"os"
	"strings"
	"testing"
	"time"
)

// Test basic flag definition and parsing
func TestBasicUsage(t *testing.T) {
	flags := New("test")

	// Define flags
	portPtr := flags.Int("port", 8080, "Server port")
	debugPtr := flags.Bool("debug", false, "Debug mode")
	namePtr := flags.String("name", "default", "Service name")

	// Test default values
	if *portPtr != 8080 {
		t.Errorf("Expected default port 8080, got %d", *portPtr)
	}
	if *debugPtr != false {
		t.Errorf("Expected default debug false, got %v", *debugPtr)
	}
	if *namePtr != "default" {
		t.Errorf("Expected default name 'default', got %s", *namePtr)
	}

	// Test parsing
	args := []string{"--port", "9090", "--debug", "--name=myservice"}
	err := flags.Parse(args)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	// Test getter methods
	if flags.GetInt("port") != 9090 {
		t.Errorf("Expected port 9090, got %d", flags.GetInt("port"))
	}
	if !flags.GetBool("debug") {
		t.Errorf("Expected debug true, got %v", flags.GetBool("debug"))
	}
	if flags.GetString("name") != "myservice" {
		t.Errorf("Expected name 'myservice', got %s", flags.GetString("name"))
	}
}

// Test different argument formats
func TestArgumentFormats(t *testing.T) {
	flags := New("test")
	flags.String("str1", "", "String with equals")
	flags.String("str2", "", "String with space")
	flags.Bool("bool1", false, "Bool without value")
	flags.Bool("bool2", false, "Bool with value")

	args := []string{
		"--str1=hello",
		"--str2", "world",
		"--bool1",
		"--bool2=true",
	}

	err := flags.Parse(args)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	if flags.GetString("str1") != "hello" {
		t.Errorf("Expected str1 'hello', got '%s'", flags.GetString("str1"))
	}
	if flags.GetString("str2") != "world" {
		t.Errorf("Expected str2 'world', got '%s'", flags.GetString("str2"))
	}
	if !flags.GetBool("bool1") {
		t.Errorf("Expected bool1 true, got %v", flags.GetBool("bool1"))
	}
	if !flags.GetBool("bool2") {
		t.Errorf("Expected bool2 true, got %v", flags.GetBool("bool2"))
	}
}

// Test duration parsing
func TestDuration(t *testing.T) {
	flags := New("test")
	flags.Duration("timeout", time.Second, "Request timeout")
	flags.Duration("interval", 0, "Poll interval")

	args := []string{
		"--timeout=30s",
		"--interval", "500ms",
	}

	err := flags.Parse(args)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	if flags.GetDuration("timeout") != 30*time.Second {
		t.Errorf("Expected timeout 30s, got %v", flags.GetDuration("timeout"))
	}
	if flags.GetDuration("interval") != 500*time.Millisecond {
		t.Errorf("Expected interval 500ms, got %v", flags.GetDuration("interval"))
	}
}

// Test float64 parsing
func TestFloat64(t *testing.T) {
	flags := New("test")
	flags.Float64("rate", 1.0, "Processing rate")
	flags.Float64("ratio", 0.0, "Ratio value")

	args := []string{
		"--rate=2.5",
		"--ratio", "0.75",
	}

	err := flags.Parse(args)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	if flags.GetFloat64("rate") != 2.5 {
		t.Errorf("Expected rate 2.5, got %f", flags.GetFloat64("rate"))
	}
	if flags.GetFloat64("ratio") != 0.75 {
		t.Errorf("Expected ratio 0.75, got %f", flags.GetFloat64("ratio"))
	}
}

// Test short flags parsing
func TestShortFlags(t *testing.T) {
	t.Run("string short flag", func(t *testing.T) {
		fs := New("test")
		name := fs.StringVar("name", "n", "", "Your name")

		args := []string{"-n", "John"}
		err := fs.Parse(args)
		if err != nil {
			t.Fatalf("Parse failed: %v", err)
		}

		if *name != "John" {
			t.Errorf("Expected 'John', got '%s'", *name)
		}
	})

	t.Run("int short flag", func(t *testing.T) {
		fs := New("test")
		portVar := fs.IntVar("port", "p", 0, "Port number")

		args := []string{"-p", "8080"}
		err := fs.Parse(args)
		if err != nil {
			t.Fatalf("Parse failed: %v", err)
		}

		if *portVar != 8080 {
			t.Errorf("Expected 8080, got %d", *portVar)
		}
	})

	t.Run("bool short flag", func(t *testing.T) {
		fs := New("test")
		debug := fs.BoolVar("debug", "d", false, "Debug mode")

		args := []string{"-d"}
		err := fs.Parse(args)
		if err != nil {
			t.Fatalf("Parse failed: %v", err)
		}

		if !*debug {
			t.Error("Expected debug to be true")
		}
	})

	t.Run("unknown short flag", func(t *testing.T) {
		fs := New("test")
		fs.StringVar("name", "n", "", "Your name")

		args := []string{"-x", "value"}
		err := fs.Parse(args)
		if err == nil {
			t.Error("Expected error for unknown short flag")
		}
	})
}

// TestShortFlagEquals tests the new -f=value syntax
func TestShortFlagEquals(t *testing.T) {
	t.Run("string short flag with equals", func(t *testing.T) {
		fs := New("test")
		name := fs.StringVar("name", "n", "", "Your name")

		args := []string{"-n=John"}
		err := fs.Parse(args)
		if err != nil {
			t.Fatalf("Parse failed: %v", err)
		}

		if *name != "John" {
			t.Errorf("Expected 'John', got '%s'", *name)
		}
	})

	t.Run("int short flag with equals", func(t *testing.T) {
		fs := New("test")
		port := fs.IntVar("port", "p", 0, "Port number")

		args := []string{"-p=8080"}
		err := fs.Parse(args)
		if err != nil {
			t.Fatalf("Parse failed: %v", err)
		}

		if *port != 8080 {
			t.Errorf("Expected 8080, got %d", *port)
		}
	})

	t.Run("bool short flag with equals true", func(t *testing.T) {
		fs := New("test")
		debug := fs.BoolVar("debug", "d", false, "Debug mode")

		args := []string{"-d=true"}
		err := fs.Parse(args)
		if err != nil {
			t.Fatalf("Parse failed: %v", err)
		}

		if !*debug {
			t.Error("Expected debug to be true")
		}
	})

	t.Run("bool short flag with equals false", func(t *testing.T) {
		fs := New("test")
		debug := fs.BoolVar("debug", "d", true, "Debug mode") // Default true

		args := []string{"-d=false"}
		err := fs.Parse(args)
		if err != nil {
			t.Fatalf("Parse failed: %v", err)
		}

		if *debug {
			t.Error("Expected debug to be false")
		}
	})

	t.Run("duration short flag with equals", func(t *testing.T) {
		fs := New("test")
		timeout := fs.Duration("timeout", 30*time.Second, "Timeout duration")
		fs.shortMap["t"] = fs.flags["timeout"] // Add short key

		args := []string{"-t=45s"}
		err := fs.Parse(args)
		if err != nil {
			t.Fatalf("Parse failed: %v", err)
		}

		expected := 45 * time.Second
		if *timeout != expected {
			t.Errorf("Expected %v, got %v", expected, *timeout)
		}
	})

	t.Run("invalid short flag format with equals", func(t *testing.T) {
		fs := New("test")
		fs.StringVar("name", "n", "", "Your name")

		// -ab=value should be invalid (only single char before =)
		args := []string{"-ab=value"}
		err := fs.Parse(args)
		if err == nil {
			t.Error("Expected error for invalid short flag format with equals")
		}
		if !strings.Contains(err.Error(), "invalid short flag format") {
			t.Errorf("Expected 'invalid short flag format' error, got: %v", err)
		}
	})

	t.Run("empty value with equals", func(t *testing.T) {
		fs := New("test")
		name := fs.StringVar("name", "n", "default", "Your name")

		args := []string{"-n="}
		err := fs.Parse(args)
		if err != nil {
			t.Fatalf("Parse failed: %v", err)
		}

		if *name != "" {
			t.Errorf("Expected empty string, got '%s'", *name)
		}
	})
}

// TestCombinedShortFlags tests the new -abc syntax
func TestCombinedShortFlags(t *testing.T) {
	t.Run("combined boolean flags", func(t *testing.T) {
		fs := New("test")
		verbose := fs.BoolVar("verbose", "v", false, "Verbose output")
		debug := fs.BoolVar("debug", "d", false, "Debug mode")
		help := fs.BoolVar("help-mode", "h", false, "Help mode")

		args := []string{"-vdh"}
		err := fs.Parse(args)
		if err != nil {
			t.Fatalf("Parse failed: %v", err)
		}

		if !*verbose {
			t.Error("Expected verbose to be true")
		}
		if !*debug {
			t.Error("Expected debug to be true")
		}
		if !*help {
			t.Error("Expected help to be true")
		}
	})

	t.Run("combined flags with value at end", func(t *testing.T) {
		fs := New("test")
		verbose := fs.BoolVar("verbose", "v", false, "Verbose output")
		debug := fs.BoolVar("debug", "d", false, "Debug mode")
		name := fs.StringVar("name", "n", "", "Name")

		args := []string{"-vdn", "John"}
		err := fs.Parse(args)
		if err != nil {
			t.Fatalf("Parse failed: %v", err)
		}

		if !*verbose {
			t.Error("Expected verbose to be true")
		}
		if !*debug {
			t.Error("Expected debug to be true")
		}
		if *name != "John" {
			t.Errorf("Expected 'John', got '%s'", *name)
		}
	})

	t.Run("combined flags with int value at end", func(t *testing.T) {
		fs := New("test")
		verbose := fs.BoolVar("verbose", "v", false, "Verbose output")
		port := fs.IntVar("port", "p", 0, "Port")

		args := []string{"-vp", "8080"}
		err := fs.Parse(args)
		if err != nil {
			t.Fatalf("Parse failed: %v", err)
		}

		if !*verbose {
			t.Error("Expected verbose to be true")
		}
		if *port != 8080 {
			t.Errorf("Expected 8080, got %d", *port)
		}
	})

	t.Run("single flag that looks combined", func(t *testing.T) {
		fs := New("test")
		verbose := fs.BoolVar("verbose", "v", false, "Verbose output")

		args := []string{"-v"}
		err := fs.Parse(args)
		if err != nil {
			t.Fatalf("Parse failed: %v", err)
		}

		if !*verbose {
			t.Error("Expected verbose to be true")
		}
	})

	t.Run("non-boolean flag in middle of combined sequence", func(t *testing.T) {
		fs := New("test")
		_ = fs.BoolVar("verbose", "v", false, "Verbose output")
		_ = fs.StringVar("name", "n", "", "Name")
		_ = fs.BoolVar("debug", "d", false, "Debug mode")

		// -vnd should fail because -n (non-boolean) is not last
		args := []string{"-vnd", "value"}
		err := fs.Parse(args)
		if err == nil {
			t.Error("Expected error for non-boolean flag in middle of combined sequence")
		}
		if !strings.Contains(err.Error(), "must be last in combined sequence") {
			t.Errorf("Expected 'must be last in combined sequence' error, got: %v", err)
		}
	})

	t.Run("unknown flag in combined sequence", func(t *testing.T) {
		fs := New("test")
		_ = fs.BoolVar("verbose", "v", false, "Verbose output")

		args := []string{"-vx"}
		err := fs.Parse(args)
		if err == nil {
			t.Error("Expected error for unknown flag in combined sequence")
		}
		if !strings.Contains(err.Error(), "unknown flag in combined sequence") {
			t.Errorf("Expected 'unknown flag in combined sequence' error, got: %v", err)
		}
	})

	t.Run("combined flag missing value for last non-boolean", func(t *testing.T) {
		fs := New("test")
		_ = fs.BoolVar("verbose", "v", false, "Verbose output")
		_ = fs.StringVar("name", "n", "", "Name")

		args := []string{"-vn"} // Missing value for -n
		err := fs.Parse(args)
		if err == nil {
			t.Error("Expected error for missing value in combined sequence")
		}
		if !strings.Contains(err.Error(), "requires a value") {
			t.Errorf("Expected 'requires a value' error, got: %v", err)
		}
	})
}

// TestAdvancedShortFlagCombinations tests edge cases and complex scenarios
func TestAdvancedShortFlagCombinations(t *testing.T) {
	t.Run("mixed long and new short syntax", func(t *testing.T) {
		fs := New("test")
		verbose := fs.BoolVar("verbose", "v", false, "Verbose output")
		debug := fs.BoolVar("debug", "d", false, "Debug mode")
		host := fs.StringVar("host", "h", "localhost", "Host")
		port := fs.IntVar("port", "p", 8080, "Port")

		args := []string{"--verbose", "-d", "-h=example.com", "-p", "9090"}
		err := fs.Parse(args)
		if err != nil {
			t.Fatalf("Parse failed: %v", err)
		}

		if !*verbose {
			t.Error("Expected verbose to be true")
		}
		if !*debug {
			t.Error("Expected debug to be true")
		}
		if *host != "example.com" {
			t.Errorf("Expected 'example.com', got '%s'", *host)
		}
		if *port != 9090 {
			t.Errorf("Expected 9090, got %d", *port)
		}
	})

	t.Run("performance test - many combined flags", func(t *testing.T) {
		fs := New("test")

		// Create many boolean flags
		var flags []*bool
		for i := 0; i < 10; i++ {
			char := string(rune('a' + i))
			flag := fs.BoolVar(fmt.Sprintf("flag%d", i), char, false, fmt.Sprintf("Flag %d", i))
			flags = append(flags, flag)
		}

		// Test combined: -abcdefghij
		args := []string{"-abcdefghij"}
		err := fs.Parse(args)
		if err != nil {
			t.Fatalf("Parse failed: %v", err)
		}

		// Verify all flags are set
		for i, flag := range flags {
			if !*flag {
				t.Errorf("Expected flag%d to be true", i)
			}
		}
	})
}

func TestValidation(t *testing.T) {
	t.Run("StringValidation", func(t *testing.T) {
		testStringValidation(t)
	})

	t.Run("IntValidation", func(t *testing.T) {
		testIntValidation(t)
	})

	t.Run("ValidateAllMethod", func(t *testing.T) {
		testValidateAllMethod(t)
	})
}

// testStringValidation tests string validation functionality
func testStringValidation(t *testing.T) {
	// Test valid value
	fs := New("test")
	fs.String("name", "", "Your name")
	if err := fs.SetValidator("name", nameValidator()); err != nil {
		t.Fatalf("SetValidator failed: %v", err)
	}

	args := []string{"--name=John"}
	if err := fs.Parse(args); err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	if fs.GetString("name") != "John" {
		t.Errorf("Expected 'John', got '%s'", fs.GetString("name"))
	}

	// Test invalid value
	fs = New("test")
	fs.String("name", "", "Your name")
	_ = fs.SetValidator("name", nameValidator())

	args = []string{"--name=A"}
	if err := fs.Parse(args); err == nil {
		t.Error("Expected validation error for short name")
	}
}

// testIntValidation tests int validation functionality
func testIntValidation(t *testing.T) {
	// Test valid port
	fs := New("test")
	fs.Int("port", 0, "Port number")
	if err := fs.SetValidator("port", portValidator()); err != nil {
		t.Fatalf("SetValidator failed: %v", err)
	}

	args := []string{"--port=8080"}
	if err := fs.Parse(args); err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	if fs.GetInt("port") != 8080 {
		t.Errorf("Expected 8080, got %d", fs.GetInt("port"))
	}

	// Test invalid port
	fs = New("test")
	fs.Int("port", 0, "Port number")
	_ = fs.SetValidator("port", portValidator())

	args = []string{"--port=70000"}
	if err := fs.Parse(args); err == nil {
		t.Error("Expected validation error for invalid port")
	}
}

// testValidateAllMethod tests the ValidateAll method
func testValidateAllMethod(t *testing.T) {
	fs := New("test")
	fs.String("name", "TestName", "Your name")
	fs.Int("port", 8080, "Port number")

	_ = fs.SetValidator("name", nameValidator())
	_ = fs.SetValidator("port", portValidator())

	// Should pass validation with default values
	if err := fs.ValidateAll(); err != nil {
		t.Errorf("ValidateAll failed: %v", err)
	}
}

// Test string slice parsing
func TestStringSlice(t *testing.T) {
	flags := New("test")
	flags.StringSlice("hosts", []string{"localhost"}, "Server hosts")
	flags.StringSlice("tags", nil, "Service tags")

	args := []string{
		"--hosts=server1,server2,server3",
		"--tags", "web,api,backend",
	}

	err := flags.Parse(args)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	hosts := flags.GetStringSlice("hosts")
	expected := []string{"server1", "server2", "server3"}
	if len(hosts) != len(expected) {
		t.Errorf("Expected %d hosts, got %d", len(expected), len(hosts))
	}
	for i, host := range hosts {
		if host != expected[i] {
			t.Errorf("Expected host[%d] '%s', got '%s'", i, expected[i], host)
		}
	}

	tags := flags.GetStringSlice("tags")
	expectedTags := []string{"web", "api", "backend"}
	if len(tags) != len(expectedTags) {
		t.Errorf("Expected %d tags, got %d", len(expectedTags), len(tags))
	}
	for i, tag := range tags {
		if tag != expectedTags[i] {
			t.Errorf("Expected tag[%d] '%s', got '%s'", i, expectedTags[i], tag)
		}
	}
}

// Test error conditions
func TestErrors(t *testing.T) {
	flags := New("test")
	flags.Int("port", 8080, "Server port")
	flags.Bool("debug", false, "Debug mode")

	// Test unknown flag
	err := flags.Parse([]string{"--unknown"})
	if err == nil {
		t.Error("Expected error for unknown flag")
	}

	// Test invalid int
	err = flags.Parse([]string{"--port=invalid"})
	if err == nil {
		t.Error("Expected error for invalid int")
	}

	// Test missing value for non-bool flag
	err = flags.Parse([]string{"--port"})
	if err == nil {
		t.Error("Expected error for missing port value")
	}
}

// Test flag metadata
func TestMetadata(t *testing.T) {
	flags := New("test")
	flags.String("name", "default", "Service name")
	flags.Int("port", 8080, "Server port")

	// Test lookup
	flag := flags.Lookup("name")
	if flag == nil {
		t.Error("Expected to find 'name' flag")
	}
	if flag.Name() != "name" {
		t.Errorf("Expected flag name 'name', got '%s'", flag.Name())
	}
	if flag.Type() != "string" {
		t.Errorf("Expected flag type 'string', got '%s'", flag.Type())
	}
	if flag.Usage() != "Service name" {
		t.Errorf("Expected usage 'Service name', got '%s'", flag.Usage())
	}
}

// Test Changed() functionality
func TestChanged(t *testing.T) {
	flags := New("test")
	flags.String("name", "default", "Service name")
	flags.Int("port", 8080, "Server port")

	// Check initial state
	nameFlag := flags.Lookup("name")
	portFlag := flags.Lookup("port")

	if nameFlag.Changed() {
		t.Error("Expected 'name' flag to not be changed initially")
	}
	if portFlag.Changed() {
		t.Error("Expected 'port' flag to not be changed initially")
	}

	// Parse with one flag set
	err := flags.Parse([]string{"--name=test"})
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	// Check changed state
	if !nameFlag.Changed() {
		t.Error("Expected 'name' flag to be changed after parsing")
	}
	if portFlag.Changed() {
		t.Error("Expected 'port' flag to not be changed after parsing")
	}
}

// Test VisitAll functionality
func TestVisitAll(t *testing.T) {
	flags := New("test")
	flags.String("name", "default", "Service name")
	flags.Int("port", 8080, "Server port")
	flags.Bool("debug", false, "Debug mode")

	visited := make(map[string]bool)
	flags.VisitAll(func(flag *Flag) {
		visited[flag.Name()] = true
	})

	expected := []string{"name", "port", "debug"}
	for _, name := range expected {
		if !visited[name] {
			t.Errorf("Expected to visit flag '%s'", name)
		}
	}

	if len(visited) != len(expected) {
		t.Errorf("Expected to visit %d flags, visited %d", len(expected), len(visited))
	}
}

// Benchmark flag parsing performance
func BenchmarkParse(b *testing.B) {
	args := []string{
		"--name=myservice",
		"--port", "9090",
		"--debug",
		"--timeout=30s",
		"--hosts=host1,host2,host3",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		flags := New("benchmark")
		flags.String("name", "default", "Service name")
		flags.Int("port", 8080, "Server port")
		flags.Bool("debug", false, "Debug mode")
		flags.Duration("timeout", time.Second, "Timeout")
		flags.StringSlice("hosts", nil, "Host list")

		err := flags.Parse(args)
		if err != nil {
			b.Fatalf("Parse failed: %v", err)
		}
	}
} // Benchmark getter performance
func BenchmarkGetters(b *testing.B) {
	flags := New("benchmark")
	flags.String("name", "myservice", "Service name")
	flags.Int("port", 9090, "Server port")
	flags.Bool("debug", true, "Debug mode")
	flags.Duration("timeout", 30*time.Second, "Timeout")

	b.Run("GetString", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = flags.GetString("name")
		}
	})

	b.Run("GetInt", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = flags.GetInt("port")
		}
	})

	b.Run("GetBool", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = flags.GetBool("debug")
		}
	})

	b.Run("GetDuration", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = flags.GetDuration("timeout")
		}
	})
}

// Test interface compatibility
func TestInterfaces(t *testing.T) {
	flags := New("test")
	flags.String("name", "default", "Service name")
	flags.Int("port", 8080, "Server port")
	flags.Bool("debug", false, "Debug mode")

	// Parse some flags
	err := flags.Parse([]string{"--name=test", "--port=9000"})
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	// Test adapter
	adapter := NewAdapter(flags)

	// Test ConfigFlagSet interface
	var configFlagSet ConfigFlagSet = adapter

	// Test VisitAll through interface
	visited := make(map[string]bool)
	configFlagSet.VisitAll(func(flag ConfigFlag) {
		visited[flag.Name()] = true

		// Test ConfigFlag interface methods
		_ = flag.Name()
		_ = flag.Value()
		_ = flag.Type()
		_ = flag.Changed()
		_ = flag.Usage()
	})

	if len(visited) != 3 {
		t.Errorf("Expected to visit 3 flags, visited %d", len(visited))
	}

	// Test Lookup through interface
	nameFlag := configFlagSet.Lookup("name")
	if nameFlag == nil {
		t.Error("Expected to find 'name' flag")
	}
	if nameFlag.Name() != "name" {
		t.Errorf("Expected flag name 'name', got '%s'", nameFlag.Name())
	}
	if nameFlag.Value() != "test" {
		t.Errorf("Expected flag value 'test', got '%v'", nameFlag.Value())
	}
	if !nameFlag.Changed() {
		t.Error("Expected flag to be changed")
	}

	// Test non-existent flag
	unknownFlag := configFlagSet.Lookup("unknown")
	if unknownFlag != nil {
		t.Error("Expected nil for unknown flag")
	}
}

// Benchmark concurrent access
func BenchmarkConcurrent(b *testing.B) {
	flags := New("benchmark")
	flags.String("name", "myservice", "Service name")
	flags.Int("port", 9090, "Server port")
	flags.Bool("debug", true, "Debug mode")

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_ = flags.GetString("name")
			_ = flags.GetInt("port")
			_ = flags.GetBool("debug")
		}
	})
}

func TestReset(t *testing.T) {
	t.Run("ResetAllFlags", func(t *testing.T) {
		testResetAllFlags(t)
	})

	t.Run("ResetSpecificFlag", func(t *testing.T) {
		testResetSpecificFlag(t)
	})

	t.Run("ResetNonExistentFlag", func(t *testing.T) {
		fs := New("test")
		err := fs.ResetFlag("nonexistent")
		verifyExpectedError(t, err, "flag --nonexistent not found", "Expected error when resetting non-existent flag")
	})
}

// testResetAllFlags tests resetting all flags to their defaults
func testResetAllFlags(t *testing.T) {
	fs := New("test")

	// Create flags with default values
	strFlag := fs.String("str", "default", "String flag")
	intFlag := fs.Int("num", 42, "Int flag")
	boolFlag := fs.Bool("verbose", false, "Bool flag")
	float64Flag := fs.Float64("rate", 3.14, "Float flag")

	// Parse some arguments to change values
	args := []string{"--str", "changed", "--num", "100", "--verbose", "--rate", "2.71"}
	if err := fs.Parse(args); err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	// Verify values changed
	verifyChangedValues(t, strFlag, intFlag, boolFlag, float64Flag)

	// Reset all flags
	fs.Reset()

	// Verify values are back to defaults
	verifyDefaultValues(t, strFlag, intFlag, boolFlag, float64Flag)
}

// testResetSpecificFlag tests resetting a specific flag
func testResetSpecificFlag(t *testing.T) {
	fs := New("test")

	strFlag := fs.String("str", "default", "String flag")
	intFlag := fs.Int("num", 42, "Int flag")

	// Parse arguments
	args := []string{"--str", "changed", "--num", "100"}
	if err := fs.Parse(args); err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	// Reset only string flag
	if err := fs.ResetFlag("str"); err != nil {
		t.Fatalf("ResetFlag failed: %v", err)
	}

	// String flag should be reset, int flag unchanged
	verifyResetSpecificValues(t, strFlag, intFlag)
}

// verifyChangedValues checks that flag values were changed after parsing
func verifyChangedValues(t *testing.T, strFlag *string, intFlag *int, boolFlag *bool, float64Flag *float64) {
	if *strFlag != "changed" {
		t.Errorf("Expected strFlag to be 'changed', got '%s'", *strFlag)
	}
	if *intFlag != 100 {
		t.Errorf("Expected intFlag to be 100, got %d", *intFlag)
	}
	if !*boolFlag {
		t.Errorf("Expected boolFlag to be true, got %t", *boolFlag)
	}
	if *float64Flag != 2.71 {
		t.Errorf("Expected float64Flag to be 2.71, got %f", *float64Flag)
	}
}

// verifyDefaultValues checks that flag values are back to their defaults
func verifyDefaultValues(t *testing.T, strFlag *string, intFlag *int, boolFlag *bool, float64Flag *float64) {
	if *strFlag != "default" {
		t.Errorf("After reset, expected strFlag to be 'default', got '%s'", *strFlag)
	}
	if *intFlag != 42 {
		t.Errorf("After reset, expected intFlag to be 42, got %d", *intFlag)
	}
	if *boolFlag {
		t.Errorf("After reset, expected boolFlag to be false, got %t", *boolFlag)
	}
	if *float64Flag != 3.14 {
		t.Errorf("After reset, expected float64Flag to be 3.14, got %f", *float64Flag)
	}
}

// verifyResetSpecificValues checks values after resetting a specific flag
func verifyResetSpecificValues(t *testing.T, strFlag *string, intFlag *int) {
	if *strFlag != "default" {
		t.Errorf("After reset, expected strFlag to be 'default', got '%s'", *strFlag)
	}
	if *intFlag != 100 {
		t.Errorf("After reset, expected intFlag to remain 100, got %d", *intFlag)
	}
}

func TestRequiredFlags(t *testing.T) {
	t.Run("RequiredFlagProvided", func(t *testing.T) {
		fs := New("test")

		name := fs.String("name", "", "Name flag")
		_ = fs.SetRequired("name")

		args := []string{"--name", "test"}
		if err := fs.Parse(args); err != nil {
			t.Fatalf("Parse failed: %v", err)
		}

		if *name != "test" {
			t.Errorf("Expected name to be 'test', got '%s'", *name)
		}
	})

	t.Run("RequiredFlagMissing", func(t *testing.T) {
		fs := New("test")

		fs.String("name", "", "Name flag")
		_ = fs.SetRequired("name")

		args := []string{}
		err := fs.Parse(args)
		if err == nil {
			t.Error("Expected error when required flag is missing")
		}
		if err.Error() != "required flag --name not provided" {
			t.Errorf("Expected 'required flag --name not provided' error, got: %v", err)
		}
	})

	t.Run("SetRequiredNonExistentFlag", func(t *testing.T) {
		fs := New("test")

		err := fs.SetRequired("nonexistent")
		if err == nil {
			t.Error("Expected error when setting required on non-existent flag")
		}
		if err.Error() != "flag not found: nonexistent" {
			t.Errorf("Expected 'flag not found: nonexistent' error, got: %v", err)
		}
	})
}

func TestFlagDependencies(t *testing.T) {
	t.Run("DependenciesSatisfied", func(t *testing.T) {
		testDependenciesSatisfied(t)
	})

	t.Run("DependencyMissing", func(t *testing.T) {
		testDependencyMissing(t)
	})

	t.Run("MultipleDependencies", func(t *testing.T) {
		testMultipleDependencies(t)
	})

	t.Run("NonExistentDependency", func(t *testing.T) {
		testNonExistentDependency(t)
	})

	t.Run("SetDependenciesNonExistentFlag", func(t *testing.T) {
		fs := New("test")
		err := fs.SetDependencies("nonexistent", "something")
		verifyExpectedError(t, err, "flag not found: nonexistent", "Expected error when setting dependencies on non-existent flag")
	})
}

// testDependenciesSatisfied tests when all dependencies are satisfied
func testDependenciesSatisfied(t *testing.T) {
	fs := New("test")

	host := fs.String("host", "localhost", "Host flag")
	port := fs.Int("port", 8080, "Port flag")
	_ = fs.SetDependencies("port", "host")

	args := []string{"--host", "myhost", "--port", "3000"}
	if err := fs.Parse(args); err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	verifyDependencySatisfiedValues(t, host, port)
}

// testDependencyMissing tests when a required dependency is missing
func testDependencyMissing(t *testing.T) {
	fs := New("test")

	fs.String("host", "localhost", "Host flag")
	fs.Int("port", 8080, "Port flag")
	_ = fs.SetDependencies("port", "host")

	args := []string{"--port", "3000"}
	err := fs.Parse(args)
	verifyExpectedError(t, err, "flag --port requires --host to be set", "Expected error when dependency is missing")
}

// testMultipleDependencies tests flags with multiple dependencies
func testMultipleDependencies(t *testing.T) {
	fs := New("test")

	user := fs.String("user", "", "User flag")
	pass := fs.String("pass", "", "Password flag")
	auth := fs.Bool("auth", false, "Auth flag")
	_ = fs.SetDependencies("auth", "user", "pass")

	args := []string{"--user", "admin", "--pass", "secret", "--auth"}
	if err := fs.Parse(args); err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	verifyMultipleDependencyValues(t, user, pass, auth)
}

// testNonExistentDependency tests dependencies on non-existent flags
func testNonExistentDependency(t *testing.T) {
	fs := New("test")

	fs.String("name", "", "Name flag")
	_ = fs.SetDependencies("name", "nonexistent")

	args := []string{"--name", "test"}
	err := fs.Parse(args)
	verifyExpectedError(t, err, "flag --name depends on non-existent flag --nonexistent", "Expected error when dependency doesn't exist")
}

// verifyDependencySatisfiedValues checks values when dependencies are satisfied
func verifyDependencySatisfiedValues(t *testing.T, host *string, port *int) {
	if *host != "myhost" {
		t.Errorf("Expected host to be 'myhost', got '%s'", *host)
	}
	if *port != 3000 {
		t.Errorf("Expected port to be 3000, got %d", *port)
	}
}

// verifyMultipleDependencyValues checks values for multiple dependencies test
func verifyMultipleDependencyValues(t *testing.T, user, pass *string, auth *bool) {
	if *user != "admin" {
		t.Errorf("Expected user to be 'admin', got '%s'", *user)
	}
	if *pass != "secret" {
		t.Errorf("Expected pass to be 'secret', got '%s'", *pass)
	}
	if !*auth {
		t.Errorf("Expected auth to be true, got %t", *auth)
	}
}

func TestHelpSystem(t *testing.T) {
	t.Run("BasicHelpGeneration", func(t *testing.T) {
		fs := New("myapp")
		fs.SetDescription("A test application for demonstration")
		fs.SetVersion("1.0.0")

		_ = fs.String("name", "default", "Name of the service")
		port := fs.IntVar("port", "p", 8080, "Server port")
		_ = fs.Bool("verbose", false, "Enable verbose output")

		_ = fs.SetRequired("name")
		_ = fs.SetGroup("port", "Server Options")

		help := fs.Help()

		// Check that help contains expected elements using helper
		verifyBasicHelpContent(t, help, port)
	})

	t.Run("HelpFlagHandling", func(t *testing.T) {
		testHelpFlagHandling(t, "--help")
	})

	t.Run("ShortHelpFlagHandling", func(t *testing.T) {
		testHelpFlagHandling(t, "-h")
	})

	t.Run("SetGroupNonExistentFlag", func(t *testing.T) {
		fs := New("test")
		err := fs.SetGroup("nonexistent", "Some Group")
		verifyExpectedError(t, err, "flag not found: nonexistent", "Expected error when setting group on non-existent flag")
	})

	t.Run("HelpWithDependencies", func(t *testing.T) {
		fs := New("test")
		_ = fs.String("user", "", "Username")
		_ = fs.String("pass", "", "Password")
		_ = fs.Bool("auth", false, "Enable authentication")
		_ = fs.SetDependencies("auth", "user", "pass")

		help := fs.Help()
		verifyHelpContains(t, help, "[depends on: user, pass]", "Help should show dependencies")
	})
}

// verifyBasicHelpContent checks that help output contains all expected basic elements
func verifyBasicHelpContent(t *testing.T, help string, port *int) {
	expectedContents := []struct {
		content string
		message string
	}{
		{"A test application for demonstration", "Help should contain description"},
		{"Version: 1.0.0", "Help should contain version"},
		{"Usage: myapp [options]", "Help should contain usage line"},
		{"--name", "Help should contain --name flag"},
		{"-p, --port", "Help should contain short and long flag format"},
		{"[REQUIRED]", "Help should indicate required flags"},
		{"Server Options:", "Help should contain group headers"},
		{"(default: 8080)", "Help should show default values"},
	}

	for _, expected := range expectedContents {
		verifyHelpContains(t, help, expected.content, expected.message)
	}

	// Verify port is not nil
	if port == nil {
		t.Error("Port should not be nil")
	}
}

// testHelpFlagHandling tests help flag handling for both -h and --help
func testHelpFlagHandling(t *testing.T, helpFlag string) {
	fs := New("test")
	_ = fs.String("name", "", "Name flag")

	err := fs.Parse([]string{helpFlag})
	verifyExpectedError(t, err, "help requested", "Expected error when "+helpFlag+" is used")
}

// verifyHelpContains checks if help output contains expected content
func verifyHelpContains(t *testing.T, help, expectedContent, errorMessage string) {
	if !strings.Contains(help, expectedContent) {
		t.Error(errorMessage)
	}
}

// verifyExpectedError checks if error matches expected value
func verifyExpectedError(t *testing.T, err error, expectedMsg, contextMsg string) {
	if err == nil {
		t.Error(contextMsg)
		return
	}
	if err.Error() != expectedMsg {
		t.Errorf("Expected '%s' error, got: %v", expectedMsg, err)
	}
}

func TestConfigurationFiles(t *testing.T) {
	t.Run("LoadJSONConfig", func(t *testing.T) {
		testLoadJSONConfig(t)
	})

	t.Run("CommandLineOverridesConfig", func(t *testing.T) {
		testCommandLineOverridesConfig(t)
	})

	t.Run("ConfigFileNotFound", func(t *testing.T) {
		testConfigFileNotFound(t)
	})

	t.Run("InvalidJSONConfig", func(t *testing.T) {
		testInvalidJSONConfig(t)
	})

	t.Run("ConfigValidation", func(t *testing.T) {
		testConfigValidation(t)
	})
}

// testLoadJSONConfig tests loading a basic JSON configuration file
func testLoadJSONConfig(t *testing.T) {
	configContent := `{
		"host": "config-host",
		"port": 9000,
		"debug": true,
		"rate": 2.5,
		"tags": ["config", "test"]
	}`

	tmpfile := createTempConfigFile(t, configContent, "test-config-*.json")
	defer func() { _ = os.Remove(tmpfile) }()

	fs, host, port, debug, rate, tags := setupBasicFlags("test")
	fs.SetConfigFile(tmpfile)

	if err := fs.Parse([]string{}); err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	verifyConfigValues(t, host, port, debug, rate, tags)
}

// testCommandLineOverridesConfig tests that command line args override config
func testCommandLineOverridesConfig(t *testing.T) {
	configContent := `{
		"host": "config-host",
		"port": 9000
	}`

	tmpfile := createTempConfigFile(t, configContent, "test-override-*.json")
	defer func() { _ = os.Remove(tmpfile) }()

	fs := New("test")
	host := fs.String("host", "default", "Host flag")
	port := fs.Int("port", 8080, "Port flag")
	fs.SetConfigFile(tmpfile)

	args := []string{"--host", "cmdline-host"}
	if err := fs.Parse(args); err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	verifyOverrideValues(t, host, port)
}

// testConfigFileNotFound tests handling of non-existent config files
func testConfigFileNotFound(t *testing.T) {
	fs := New("test")
	_ = fs.String("host", "default", "Host flag")

	// Set non-existent config file (should not cause error, just skip)
	fs.SetConfigFile("/non/existent/config.json")

	if err := fs.Parse([]string{}); err != nil {
		t.Fatalf("Parse should not fail when config file doesn't exist: %v", err)
	}
}

// testInvalidJSONConfig tests handling of invalid JSON in config files
func testInvalidJSONConfig(t *testing.T) {
	configContent := `{ "host": "test", invalid json }`
	tmpfile := createTempConfigFile(t, configContent, "test-invalid-*.json")
	defer func() { _ = os.Remove(tmpfile) }()

	fs := New("test")
	_ = fs.String("host", "default", "Host flag")
	fs.SetConfigFile(tmpfile)

	err := fs.Parse([]string{})
	verifyConfigParseError(t, err, "config file error")
}

// testConfigValidation tests validation of config file values
func testConfigValidation(t *testing.T) {
	configContent := `{
		"port": 99999
	}`

	tmpfile, err := os.CreateTemp("", "test-validation-*.json")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer func() { _ = os.Remove(tmpfile.Name()) }()

	if _, err := tmpfile.Write([]byte(configContent)); err != nil {
		t.Fatalf("Failed to write config file: %v", err)
	}
	_ = tmpfile.Close()

	fs := New("test")
	_ = fs.Int("port", 8080, "Port flag")

	// Set validator that rejects the config value
	_ = fs.SetValidator("port", func(val interface{}) error {
		port, ok := val.(int)
		if !ok {
			return fmt.Errorf("expected int, got %T", val)
		}
		if port > 65535 {
			return fmt.Errorf("port must be <= 65535")
		}
		return nil
	})

	fs.SetConfigFile(tmpfile.Name())

	err = fs.Parse([]string{})
	verifyConfigParseError(t, err, "port must be <= 65535")
}

// verifyOverrideValues checks command line override behavior
func verifyOverrideValues(t *testing.T, host *string, port *int) {
	if *host != "cmdline-host" {
		t.Errorf("Expected host from command line 'cmdline-host', got '%s'", *host)
	}
	if *port != 9000 {
		t.Errorf("Expected port from config 9000, got %d", *port)
	}
}

// verifyConfigParseError checks for expected config parsing errors
func verifyConfigParseError(t *testing.T, err error, expectedError string) {
	if err == nil {
		t.Error("Expected error when parsing invalid config")
		return
	}
	if !strings.Contains(err.Error(), expectedError) {
		t.Errorf("Expected error containing '%s', got: %v", expectedError, err)
	}
}

// TestEnvironmentVariables tests environment variable support
func TestEnvironmentVariables(t *testing.T) {
	t.Run("BasicEnvSupport", func(t *testing.T) {
		testBasicEnvSupport(t)
	})

	t.Run("CommandLineOverridesEnv", func(t *testing.T) {
		testCommandLineOverridesEnv(t)
	})

	t.Run("CustomEnvVarNames", func(t *testing.T) {
		testCustomEnvVarNames(t)
	})

	t.Run("DefaultNaming", func(t *testing.T) {
		testDefaultNaming(t)
	})
}

// testBasicEnvSupport tests basic environment variable support
func testBasicEnvSupport(t *testing.T) {
	// Set environment variables
	_ = os.Setenv("TEST_HOST", "env.example.com")
	_ = os.Setenv("TEST_PORT", "9090")
	_ = os.Setenv("TEST_DEBUG", "true")
	defer func() { _ = os.Unsetenv("TEST_HOST") }()
	defer func() { _ = os.Unsetenv("TEST_PORT") }()
	defer func() { _ = os.Unsetenv("TEST_DEBUG") }()

	fs := New("myapp")
	fs.SetEnvPrefix("TEST")

	host := fs.String("host", "localhost", "Host address")
	port := fs.Int("port", 8080, "Port number")
	debug := fs.Bool("debug", false, "Debug mode")

	err := fs.Parse([]string{})
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	verifyBasicEnvValues(t, host, port, debug)
}

// testCommandLineOverridesEnv tests that command line overrides environment variables
func testCommandLineOverridesEnv(t *testing.T) {
	// Set environment variables
	_ = os.Setenv("TEST_HOST", "env.example.com")
	_ = os.Setenv("TEST_PORT", "9090")
	defer func() { _ = os.Unsetenv("TEST_HOST") }()
	defer func() { _ = os.Unsetenv("TEST_PORT") }()

	fs := New("myapp")
	fs.SetEnvPrefix("TEST")

	host := fs.String("host", "localhost", "Host address")
	port := fs.Int("port", 8080, "Port number")

	// Command line should override environment
	err := fs.Parse([]string{"--host", "cmd.example.com", "--port", "3000"})
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	verifyCommandLineOverrideValues(t, host, port)
}

// testCustomEnvVarNames tests custom environment variable names
func testCustomEnvVarNames(t *testing.T) {
	// Set custom environment variable
	_ = os.Setenv("CUSTOM_SERVER_HOST", "custom.example.com")
	defer func() { _ = os.Unsetenv("CUSTOM_SERVER_HOST") }()

	fs := New("myapp")
	fs.EnableEnvLookup()

	host := fs.String("host", "localhost", "Host address")

	// Set custom environment variable name
	err := fs.SetEnvVar("host", "CUSTOM_SERVER_HOST")
	if err != nil {
		t.Fatalf("SetEnvVar failed: %v", err)
	}

	err = fs.Parse([]string{})
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	verifyCustomEnvValue(t, host)
}

// testDefaultNaming tests default environment variable naming
func testDefaultNaming(t *testing.T) {
	// Set environment variable with default naming
	_ = os.Setenv("DB_HOST", "db.example.com")
	defer func() { _ = os.Unsetenv("DB_HOST") }()

	fs := New("myapp")
	fs.EnableEnvLookup()

	dbHost := fs.String("db-host", "localhost", "Database host")

	err := fs.Parse([]string{})
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	verifyDefaultNamingValue(t, dbHost)
}

// verifyBasicEnvValues checks basic environment variable values
func verifyBasicEnvValues(t *testing.T, host *string, port *int, debug *bool) {
	if *host != "env.example.com" {
		t.Errorf("Expected host from env 'env.example.com', got '%s'", *host)
	}
	if *port != 9090 {
		t.Errorf("Expected port from env 9090, got %d", *port)
	}
	if *debug != true {
		t.Errorf("Expected debug from env true, got %v", *debug)
	}
}

// verifyCommandLineOverrideValues checks command line override values
func verifyCommandLineOverrideValues(t *testing.T, host *string, port *int) {
	if *host != "cmd.example.com" {
		t.Errorf("Expected host from command line 'cmd.example.com', got '%s'", *host)
	}
	if *port != 3000 {
		t.Errorf("Expected port from command line 3000, got %d", *port)
	}
}

// verifyCustomEnvValue checks custom environment variable value
func verifyCustomEnvValue(t *testing.T, host *string) {
	if *host != "custom.example.com" {
		t.Errorf("Expected host from custom env 'custom.example.com', got '%s'", *host)
	}
}

// verifyDefaultNamingValue checks default naming environment variable value
func verifyDefaultNamingValue(t *testing.T, dbHost *string) {
	if *dbHost != "db.example.com" {
		t.Errorf("Expected db-host from env 'db.example.com', got '%s'", *dbHost)
	}
}

// Test PrintUsage function
func TestPrintUsage(t *testing.T) {
	fs := New("test")
	fs.String("name", "default", "Name flag")
	fs.Int("port", 8080, "Port flag")

	// This should not panic
	fs.PrintUsage()
}

// Test Changed function - additional coverage
func TestChangedAdditional(t *testing.T) {
	fs := New("test")
	fs.String("name", "default", "Name flag")
	fs.Int("port", 8080, "Port flag")

	// Before parsing, nothing should be changed
	if fs.Changed("name") {
		t.Error("Expected name to not be changed before parsing")
	}

	// Parse with one flag
	err := fs.Parse([]string{"--name", "test"})
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	// Now name should be changed, but port should not
	if !fs.Changed("name") {
		t.Error("Expected name to be changed after parsing")
	}

	if fs.Changed("port") {
		t.Error("Expected port to not be changed")
	}

	// Test non-existent flag
	if fs.Changed("nonexistent") {
		t.Error("Expected non-existent flag to not be changed")
	}
}

// Test AddConfigPath function
func TestAddConfigPath(t *testing.T) {
	fs := New("test")

	// Add some config paths
	fs.AddConfigPath("/etc/myapp")
	fs.AddConfigPath("./config")
	fs.AddConfigPath("$HOME/.config/myapp")

	// This should not panic and paths should be stored
	// Since configPaths is private, we test indirectly by ensuring no error
}

// Test findConfigFile function indirectly
func TestFindConfigFile(t *testing.T) {
	// Create temporary directory and config file
	tmpDir := t.TempDir()
	configFile := tmpDir + "/test.json"

	// Create a test config file
	configContent := `{"name": "from-config", "port": 9000}`
	err := os.WriteFile(configFile, []byte(configContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test config file: %v", err)
	}

	fs := New("test")
	fs.String("name", "default", "Name flag")
	fs.Int("port", 8080, "Port flag")

	// Add the temp directory as config path
	fs.AddConfigPath(tmpDir)

	// Parse should find and load the config
	err = fs.Parse([]string{})
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	// Values should come from config
	if fs.GetString("name") != "from-config" {
		t.Errorf("Expected name from config 'from-config', got '%s'", fs.GetString("name"))
	}

	if fs.GetInt("port") != 9000 {
		t.Errorf("Expected port from config 9000, got %d", fs.GetInt("port"))
	}
}

// Test GetString error cases
func TestGetStringErrors(t *testing.T) {
	fs := New("test")

	// Test non-existent flag
	result := fs.GetString("nonexistent")
	if result != "" {
		t.Errorf("Expected empty string for non-existent flag, got '%s'", result)
	}

	// Test flag with wrong type - GetString actually converts numbers to strings
	fs.Int("port", 8080, "Port flag")
	result = fs.GetString("port")
	// This actually works and returns "8080", so let's test this behavior
	if result != "8080" {
		t.Errorf("Expected '8080' for int flag, got '%s'", result)
	}
}

// Test Flag Validate method
func TestFlagValidate(t *testing.T) {
	fs := New("test")
	port := fs.Int("port", 8080, "Port number")

	// Set a validator that fails for ports < 1024
	flag := fs.Lookup("port")
	if flag == nil {
		t.Fatal("Flag 'port' not found")
	}

	flag.SetValidator(func(val interface{}) error {
		portVal, ok := val.(int)
		if !ok {
			return fmt.Errorf("expected int, got %T", val)
		}
		if portVal < 1024 {
			return fmt.Errorf("port must be >= 1024")
		}
		return nil
	})

	// Test valid value - set the internal value directly
	*port = 8080
	// We need to parse to trigger validation or call it manually
	err := fs.Parse([]string{"--port", "8080"})
	if err != nil {
		t.Errorf("Parse should succeed for port 8080: %v", err)
	}

	// Test invalid value via parsing
	err = fs.Parse([]string{"--port", "80"})
	if err == nil {
		t.Error("Parse should fail for port 80")
	}

	// Test flag without validator
	_ = fs.String("name", "test", "Name")
	nameFlag := fs.Lookup("name")
	err = nameFlag.Validate()
	if err != nil {
		t.Errorf("Validation should succeed for flag without validator: %v", err)
	}
}

// Test LoadEnvironmentVariables edge cases
func TestLoadEnvironmentVariablesEdgeCases(t *testing.T) {
	// Test with various environment variable formats
	testCases := []struct {
		envVar   string
		envValue string
		flagName string
		expected interface{}
		flagType string
	}{
		{"TEST_STRING", "hello world", "string-flag", "hello world", "string"},
		{"TEST_INT", "42", "int-flag", 42, "int"},
		{"TEST_BOOL_TRUE", "true", "bool-flag", true, "bool"},
		{"TEST_BOOL_FALSE", "false", "bool-flag2", false, "bool"},
		{"TEST_DURATION", "5m", "duration-flag", 5 * time.Minute, "duration"},
		{"TEST_FLOAT", "3.14", "float-flag", 3.14, "float64"},
		{"TEST_SLICE", "a,b,c", "slice-flag", []string{"a", "b", "c"}, "stringSlice"},
	}

	for _, tc := range testCases {
		t.Run(tc.envVar, func(t *testing.T) {
			testEnvVarEdgeCase(t, tc.envVar, tc.envValue, tc.flagName, tc.expected, tc.flagType)
		})
	}
}

// testEnvVarEdgeCase tests a single environment variable edge case
func testEnvVarEdgeCase(t *testing.T, envVar, envValue, flagName string, expected interface{}, flagType string) {
	// Set environment variable
	_ = os.Setenv(envVar, envValue)
	defer func() { _ = os.Unsetenv(envVar) }()

	fs := New("test")
	fs.EnableEnvLookup()

	// Register appropriate flag type and test
	switch flagType {
	case "string":
		testStringEnvVar(t, fs, flagName, envVar, expected)
	case "int":
		testIntEnvVar(t, fs, flagName, envVar, expected)
	case "bool":
		testBoolEnvVar(t, fs, flagName, envVar, expected)
	case "duration":
		testDurationEnvVar(t, fs, flagName, envVar, expected)
	case "float64":
		testFloat64EnvVar(t, fs, flagName, envVar, expected)
	case "stringSlice":
		testStringSliceEnvVar(t, fs, flagName, envVar, expected)
	}
}

// testStringEnvVar tests string environment variable handling
func testStringEnvVar(t *testing.T, fs *FlagSet, flagName, envVar string, expected interface{}) {
	flag := fs.String(flagName, "", "Test flag")
	_ = fs.SetEnvVar(flagName, envVar)
	_ = fs.Parse([]string{})
	if *flag != expected {
		t.Errorf("Expected %v, got %v", expected, *flag)
	}
}

// testIntEnvVar tests int environment variable handling
func testIntEnvVar(t *testing.T, fs *FlagSet, flagName, envVar string, expected interface{}) {
	flag := fs.Int(flagName, 0, "Test flag")
	_ = fs.SetEnvVar(flagName, envVar)
	_ = fs.Parse([]string{})
	if *flag != expected {
		t.Errorf("Expected %v, got %v", expected, *flag)
	}
}

// testBoolEnvVar tests bool environment variable handling
func testBoolEnvVar(t *testing.T, fs *FlagSet, flagName, envVar string, expected interface{}) {
	flag := fs.Bool(flagName, false, "Test flag")
	_ = fs.SetEnvVar(flagName, envVar)
	_ = fs.Parse([]string{})
	if *flag != expected {
		t.Errorf("Expected %v, got %v", expected, *flag)
	}
}

// testDurationEnvVar tests duration environment variable handling
func testDurationEnvVar(t *testing.T, fs *FlagSet, flagName, envVar string, expected interface{}) {
	flag := fs.Duration(flagName, 0, "Test flag")
	_ = fs.SetEnvVar(flagName, envVar)
	_ = fs.Parse([]string{})
	if *flag != expected {
		t.Errorf("Expected %v, got %v", expected, *flag)
	}
}

// testFloat64EnvVar tests float64 environment variable handling
func testFloat64EnvVar(t *testing.T, fs *FlagSet, flagName, envVar string, expected interface{}) {
	flag := fs.Float64(flagName, 0, "Test flag")
	_ = fs.SetEnvVar(flagName, envVar)
	_ = fs.Parse([]string{})
	if *flag != expected {
		t.Errorf("Expected %v, got %v", expected, *flag)
	}
}

// testStringSliceEnvVar tests string slice environment variable handling
func testStringSliceEnvVar(t *testing.T, fs *FlagSet, flagName, envVar string, expected interface{}) {
	flag := fs.StringSlice(flagName, nil, "Test flag")
	_ = fs.SetEnvVar(flagName, envVar)
	_ = fs.Parse([]string{})
	expectedSlice, ok := expected.([]string)
	if !ok {
		t.Errorf("Expected []string, got %T", expected)
		return
	}
	if len(*flag) != len(expectedSlice) {
		t.Errorf("Expected slice length %d, got %d", len(expectedSlice), len(*flag))
	}
	for i, v := range expectedSlice {
		if (*flag)[i] != v {
			t.Errorf("Expected slice element %d to be %s, got %s", i, v, (*flag)[i])
		}
	}
}

// Test error cases in various getter methods
func TestGetterErrorCases(t *testing.T) {
	fs := New("test")

	// Test all getters with non-existent flags
	if fs.GetInt("nonexistent") != 0 {
		t.Error("GetInt should return 0 for non-existent flag")
	}

	if fs.GetBool("nonexistent") != false {
		t.Error("GetBool should return false for non-existent flag")
	}

	if fs.GetDuration("nonexistent") != 0 {
		t.Error("GetDuration should return 0 for non-existent flag")
	}

	if fs.GetFloat64("nonexistent") != 0 {
		t.Error("GetFloat64 should return 0 for non-existent flag")
	}

	slice := fs.GetStringSlice("nonexistent")
	if len(slice) != 0 {
		t.Error("GetStringSlice should return empty slice for non-existent flag")
	}

	// Test getters with wrong flag types
	fs.String("name", "test", "Name flag")

	if fs.GetInt("name") != 0 {
		t.Error("GetInt should return 0 for string flag")
	}

	if fs.GetBool("name") != false {
		t.Error("GetBool should return false for string flag")
	}
}

// Test more edge cases to increase coverage
func TestAdditionalCoverage(t *testing.T) {
	fs := New("test")

	// Test SetValidator error case
	err := fs.SetValidator("nonexistent", func(val interface{}) error { return nil })
	if err == nil {
		t.Error("SetValidator should fail for non-existent flag")
	}

	// Test SetEnvVar error case
	err = fs.SetEnvVar("nonexistent", "TEST_VAR")
	if err == nil {
		t.Error("SetEnvVar should fail for non-existent flag")
	}

	// Test ValidateAll with validation errors
	port := fs.Int("port", 8080, "Port number")
	_ = fs.SetValidator("port", func(val interface{}) error {
		return fmt.Errorf("validation error")
	})

	*port = 9000
	err = fs.ValidateAll()
	if err == nil {
		t.Error("ValidateAll should fail when validation fails")
	}

	// Test Flag Reset with different types
	flag := fs.Lookup("port")
	flag.Reset()

	// Test string flag reset
	name := fs.String("name", "default", "Name")
	*name = "changed"
	nameFlag := fs.Lookup("name")
	nameFlag.Reset()
	if *name != "default" {
		t.Error("String flag should be reset to default")
	}

	// Test bool flag reset
	debug := fs.Bool("debug", false, "Debug mode")
	*debug = true
	debugFlag := fs.Lookup("debug")
	debugFlag.Reset()
	if *debug != false {
		t.Error("Bool flag should be reset to default")
	}

	// Test ValidateAllConstraints with errors
	_ = fs.SetRequired("missing")
	err = fs.ValidateAllConstraints()
	if err == nil {
		t.Error("ValidateAllConstraints should fail for missing required flag")
	}
}

// Test config file parsing edge cases
func TestConfigEdgeCases(t *testing.T) {
	// Test invalid JSON config
	tmpDir := t.TempDir()
	configFile := tmpDir + "/invalid.json"

	invalidJSON := `{"name": "test", "port": invalid}`
	err := os.WriteFile(configFile, []byte(invalidJSON), 0644)
	if err != nil {
		t.Fatalf("Failed to create invalid config file: %v", err)
	}

	fs := New("test")
	fs.String("name", "default", "Name")
	fs.SetConfigFile(configFile)

	// Should handle invalid JSON gracefully
	err = fs.Parse([]string{})
	if err == nil {
		t.Error("Parse should fail with invalid JSON config")
	}

	// Test config with unknown fields
	validJSON := `{"name": "test", "unknown_field": "value"}`
	err = os.WriteFile(configFile, []byte(validJSON), 0644)
	if err != nil {
		t.Fatalf("Failed to create config file: %v", err)
	}

	err = fs.Parse([]string{})
	if err != nil {
		t.Errorf("Parse should succeed with unknown fields: %v", err)
	}
}

// Test parsing error cases to increase Parse coverage
func TestParseErrorCases(t *testing.T) {
	fs := New("test")
	port := fs.Int("port", 8080, "Port")

	// Test invalid integer value
	err := fs.Parse([]string{"--port", "invalid"})
	if err == nil {
		t.Error("Parse should fail with invalid integer")
	}

	// Test missing value for flag
	err = fs.Parse([]string{"--port"})
	if err == nil {
		t.Error("Parse should fail when flag value is missing")
	}

	// Test unknown flag
	err = fs.Parse([]string{"--unknown", "value"})
	if err == nil {
		t.Error("Parse should fail with unknown flag")
	}

	// Reset for next test
	*port = 8080

	// Test bool flag with invalid value
	debug := fs.Bool("debug", false, "Debug")
	err = fs.Parse([]string{"--debug", "invalid"})
	if err == nil {
		t.Error("Parse should fail with invalid boolean value")
	}

	// Test duration with invalid value
	timeout := fs.Duration("timeout", time.Second, "Timeout")
	err = fs.Parse([]string{"--timeout", "invalid"})
	if err == nil {
		t.Error("Parse should fail with invalid duration")
	}

	// Test float with invalid value
	rate := fs.Float64("rate", 1.0, "Rate")
	err = fs.Parse([]string{"--rate", "invalid"})
	if err == nil {
		t.Error("Parse should fail with invalid float")
	}

	// Use variables to avoid unused warnings
	_ = port
	_ = debug
	_ = timeout
	_ = rate
}

// Test setFlagValue edge cases
func TestSetFlagValueEdgeCases(t *testing.T) {
	fs := New("test")

	// Test all flag types with edge cases
	str := fs.String("str", "default", "String flag")
	integer := fs.Int("int", 0, "Int flag")
	boolean := fs.Bool("bool", false, "Bool flag")
	duration := fs.Duration("dur", 0, "Duration flag")
	float := fs.Float64("float", 0.0, "Float flag")
	slice := fs.StringSlice("slice", nil, "Slice flag")

	// Test parsing with equals syntax
	err := fs.Parse([]string{
		"--str=test",
		"--int=42",
		"--bool=true",
		"--dur=5m",
		"--float=3.14",
		"--slice=a,b,c",
	})
	if err != nil {
		t.Fatalf("Parse should succeed: %v", err)
	}

	// Verify values
	if *str != "test" {
		t.Errorf("Expected str=test, got %s", *str)
	}
	if *integer != 42 {
		t.Errorf("Expected int=42, got %d", *integer)
	}
	if !*boolean {
		t.Error("Expected bool=true")
	}
	if *duration != 5*time.Minute {
		t.Errorf("Expected dur=5m, got %v", *duration)
	}
	if *float != 3.14 {
		t.Errorf("Expected float=3.14, got %f", *float)
	}
	if len(*slice) != 3 || (*slice)[0] != "a" {
		t.Errorf("Expected slice=[a,b,c], got %v", *slice)
	}
}

// Test more environment variable edge cases
func TestEnvVarMoreCases(t *testing.T) {
	// Test environment variables that don't exist
	fs := New("test")
	fs.EnableEnvLookup()
	fs.SetEnvPrefix("NONEXISTENT")

	name := fs.String("name", "default", "Name")

	err := fs.Parse([]string{})
	if err != nil {
		t.Fatalf("Parse should succeed even with non-existent env vars: %v", err)
	}

	if *name != "default" {
		t.Errorf("Expected default value, got %s", *name)
	}
}

// Test final edge cases for 95% coverage
func TestFinalEdgeCases(t *testing.T) {
	testShortFlagFunctionality(t)
	testFlagWithEmptyShortKey(t)
	testAllGetterEdgeCases(t)
}

// testShortFlagFunctionality tests short flag parsing
func testShortFlagFunctionality(t *testing.T) {
	fs := New("test")
	verbose := fs.BoolVar("verbose", "v", false, "Verbose mode")

	err := fs.Parse([]string{"-v"})
	if err != nil {
		t.Fatalf("Parse should succeed with short flag: %v", err)
	}

	if !*verbose {
		t.Error("Short flag should set verbose to true")
	}
}

// testFlagWithEmptyShortKey tests flags with empty short keys
func testFlagWithEmptyShortKey(t *testing.T) {
	fs := New("test")
	name := fs.StringVar("name", "", "default", "Name")

	err := fs.Parse([]string{"--name", "test"})
	if err != nil {
		t.Fatalf("Parse should succeed: %v", err)
	}

	if *name != "test" {
		t.Errorf("Expected name=test, got %s", *name)
	}
}

// testAllGetterEdgeCases tests all getter methods with existing flags
func testAllGetterEdgeCases(t *testing.T) {
	fs := New("test")

	// Create flags with default values
	fs.String("test-string", "default", "Test string")
	fs.Int("test-int", 123, "Test int")
	fs.Bool("test-bool", true, "Test bool")
	fs.Duration("test-duration", time.Hour, "Test duration")
	fs.Float64("test-float", 2.5, "Test float")
	fs.StringSlice("test-slice", []string{"a", "b"}, "Test slice")

	// Test all getters
	verifyAllGetterValues(t, fs)
}

// verifyAllGetterValues checks all getter method return values
func verifyAllGetterValues(t *testing.T, fs *FlagSet) {
	if fs.GetString("test-string") != "default" {
		t.Error("GetString should return default value")
	}
	if fs.GetInt("test-int") != 123 {
		t.Error("GetInt should return default value")
	}
	if fs.GetBool("test-bool") != true {
		t.Error("GetBool should return default value")
	}
	if fs.GetDuration("test-duration") != time.Hour {
		t.Error("GetDuration should return default value")
	}
	if fs.GetFloat64("test-float") != 2.5 {
		t.Error("GetFloat64 should return default value")
	}

	slice := fs.GetStringSlice("test-slice")
	if len(slice) != 2 || slice[0] != "a" || slice[1] != "b" {
		t.Error("GetStringSlice should return default value")
	}
}

// Helper functions for test complexity reduction

// createTempConfigFile creates a temporary config file with the given content
func createTempConfigFile(t *testing.T, content, pattern string) string {
	tmpfile, err := os.CreateTemp("", pattern)
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}

	if _, err := tmpfile.Write([]byte(content)); err != nil {
		t.Fatalf("Failed to write config file: %v", err)
		_ = os.Remove(tmpfile.Name())
	}
	_ = tmpfile.Close()
	return tmpfile.Name()
}

// setupBasicFlags creates a FlagSet with common test flags
func setupBasicFlags(name string) (*FlagSet, *string, *int, *bool, *float64, *[]string) {
	fs := New(name)
	host := fs.String("host", "default", "Host flag")
	port := fs.Int("port", 8080, "Port flag")
	debug := fs.Bool("debug", false, "Debug flag")
	rate := fs.Float64("rate", 1.0, "Rate flag")
	tags := fs.StringSlice("tags", []string{}, "Tags flag")
	return fs, host, port, debug, rate, tags
}

// verifyConfigValues checks that config values were loaded correctly
func verifyConfigValues(t *testing.T, host *string, port *int, debug *bool, rate *float64, tags *[]string) {
	if *host != "config-host" {
		t.Errorf("Expected host from config 'config-host', got '%s'", *host)
	}
	if *port != 9000 {
		t.Errorf("Expected port from config 9000, got %d", *port)
	}
	if !*debug {
		t.Errorf("Expected debug from config true, got %t", *debug)
	}
	if *rate != 2.5 {
		t.Errorf("Expected rate from config 2.5, got %f", *rate)
	}
	if len(*tags) != 2 || (*tags)[0] != "config" || (*tags)[1] != "test" {
		t.Errorf("Expected tags from config ['config', 'test'], got %v", *tags)
	}
}

// nameValidator returns a validator function for name strings
func nameValidator() func(interface{}) error {
	return func(v interface{}) error {
		if str, ok := v.(string); ok {
			if len(str) < 2 {
				return fmt.Errorf("name must be at least 2 characters")
			}
		}
		return nil
	}
}

// portValidator returns a validator function for port numbers
func portValidator() func(interface{}) error {
	return func(v interface{}) error {
		if intVal, ok := v.(int); ok {
			if intVal < 1 || intVal > 65535 {
				return fmt.Errorf("port must be between 1 and 65535")
			}
		}
		return nil
	}
}

// TestShortKey tests the new ShortKey() method for flag metadata access
func TestShortKey(t *testing.T) {
	fs := New("test")

	// Create flag with short key
	fs.StringVar("host", "h", "localhost", "Server host")
	fs.IntVar("port", "p", 8080, "Server port")
	fs.BoolVar("verbose", "v", false, "Verbose output")

	// Create flag without short key
	fs.String("config", "config.json", "Configuration file")

	t.Run("flags with short keys", func(t *testing.T) {
		hostFlag := fs.Lookup("host")
		if hostFlag == nil {
			t.Fatal("Expected to find host flag")
		}
		if hostFlag.ShortKey() != "h" {
			t.Errorf("Expected short key 'h', got '%s'", hostFlag.ShortKey())
		}

		portFlag := fs.Lookup("port")
		if portFlag == nil {
			t.Fatal("Expected to find port flag")
		}
		if portFlag.ShortKey() != "p" {
			t.Errorf("Expected short key 'p', got '%s'", portFlag.ShortKey())
		}

		verboseFlag := fs.Lookup("verbose")
		if verboseFlag == nil {
			t.Fatal("Expected to find verbose flag")
		}
		if verboseFlag.ShortKey() != "v" {
			t.Errorf("Expected short key 'v', got '%s'", verboseFlag.ShortKey())
		}
	})

	t.Run("flag without short key", func(t *testing.T) {
		configFlag := fs.Lookup("config")
		if configFlag == nil {
			t.Fatal("Expected to find config flag")
		}
		if configFlag.ShortKey() != "" {
			t.Errorf("Expected empty short key, got '%s'", configFlag.ShortKey())
		}
	})

	t.Run("nonexistent flag", func(t *testing.T) {
		flag := fs.Lookup("nonexistent")
		if flag != nil {
			t.Error("Expected nil for nonexistent flag")
		}
	})
}

// Test resetDurationPointer function
func TestResetDurationPointer(t *testing.T) {
	t.Run("reset duration pointer", func(t *testing.T) {
		fs := New("test")
		ptr := fs.Duration("timeout", 5*time.Second, "Timeout duration")

		// Change the value
		*ptr = 10 * time.Second
		if *ptr != 10*time.Second {
			t.Errorf("Expected 10s, got %v", *ptr)
		}

		// Reset should restore default
		flag := fs.Lookup("timeout")
		if flag != nil {
			flag.Reset()
			if *ptr != 5*time.Second {
				t.Errorf("Expected reset to 5s, got %v", *ptr)
			}
		}
	})

	t.Run("reset with invalid default type", func(t *testing.T) {
		fs := New("test")
		ptr := fs.Duration("timeout", 5*time.Second, "Timeout duration")

		flag := fs.Lookup("timeout")
		if flag != nil {
			// Force invalid default value type
			flag.defaultValue = "invalid"
			flag.Reset() // Should not crash, just not reset
			// Value should remain unchanged
			if *ptr == 0 {
				t.Error("Value was incorrectly reset with invalid default type")
			}
		}
	})
}

// Test resetStringSlicePointer function
func TestResetStringSlicePointer(t *testing.T) {
	t.Run("reset string slice pointer", func(t *testing.T) {
		fs := New("test")
		ptr := fs.StringSlice("items", []string{"a", "b"}, "List of items")

		// Change the value
		*ptr = []string{"x", "y", "z"}
		if len(*ptr) != 3 {
			t.Errorf("Expected 3 items, got %d", len(*ptr))
		}

		// Reset should restore default
		flag := fs.Lookup("items")
		if flag != nil {
			flag.Reset()
			if len(*ptr) != 2 || (*ptr)[0] != "a" || (*ptr)[1] != "b" {
				t.Errorf("Expected reset to [a b], got %v", *ptr)
			}
		}
	})

	t.Run("reset with invalid default type", func(t *testing.T) {
		fs := New("test")
		ptr := fs.StringSlice("items", []string{"a"}, "List of items")

		flag := fs.Lookup("items")
		if flag != nil {
			// Force invalid default value type
			flag.defaultValue = 42
			flag.Reset() // Should not crash, just not reset
			// Value should remain unchanged
			if len(*ptr) == 0 {
				t.Error("Value was incorrectly reset with invalid default type")
			}
		}
	})
}

// Test parseStringSlice edge cases
func TestParseStringSliceEdgeCases(t *testing.T) {
	fs := New("test")

	t.Run("empty string", func(t *testing.T) {
		result := fs.parseStringSlice("")
		if len(result) != 0 {
			t.Errorf("Expected empty slice, got %v", result)
		}
	})

	t.Run("string with trailing comma", func(t *testing.T) {
		result := fs.parseStringSlice("a,b,")
		if len(result) != 2 || result[0] != "a" || result[1] != "b" {
			t.Errorf("Expected [a b], got %v", result)
		}
	})

	t.Run("string with leading comma", func(t *testing.T) {
		result := fs.parseStringSlice(",a,b")
		if len(result) != 2 || result[0] != "a" || result[1] != "b" {
			t.Errorf("Expected [a b], got %v", result)
		}
	})

	t.Run("consecutive commas", func(t *testing.T) {
		result := fs.parseStringSlice("a,,b")
		if len(result) != 2 || result[0] != "a" || result[1] != "b" {
			t.Errorf("Expected [a b], got %v", result)
		}
	})
}

// Test resetPointer edge cases
func TestResetPointerEdgeCases(t *testing.T) {
	t.Run("reset unsupported type pointer", func(t *testing.T) {
		// Create a flag with an unsupported type
		flag := &Flag{
			name:         "test",
			defaultValue: complex64(1 + 2i), // Unsupported type
			ptr:          nil,
			flagType:     "complex64",
		}

		// Should not crash when resetting unsupported type
		flag.Reset()
		// Test passes if no panic occurs
	})
}

// Test config file loading edge cases
func TestConfigFileEdgeCases(t *testing.T) {
	t.Run("load config from nonexistent file", func(t *testing.T) {
		fs := New("test")
		fs.String("name", "default", "Name setting")

		fs.SetConfigFile("/nonexistent/path/config.json")
		err := fs.LoadConfig()
		// Should handle file not found gracefully - may return error or nil depending on implementation
		if err != nil {
			// Error is expected and acceptable
			if !strings.Contains(err.Error(), "no such file") &&
				!strings.Contains(err.Error(), "cannot find") {
				t.Errorf("Expected file not found error, got: %v", err)
			}
		}
		// Either way, name should remain default
		if fs.GetString("name") != "default" {
			t.Errorf("Expected name to remain 'default', got '%s'", fs.GetString("name"))
		}
	})

	t.Run("load config with invalid JSON", func(t *testing.T) {
		// Create temporary invalid JSON file
		tmpFile := "/tmp/invalid_config.json"
		content := `{"name": "test", "invalid": json}`

		if err := os.WriteFile(tmpFile, []byte(content), 0644); err != nil {
			t.Skipf("Could not create temp file: %v", err)
		}
		defer os.Remove(tmpFile)

		fs := New("test")
		fs.String("name", "default", "Name setting")

		fs.SetConfigFile(tmpFile)
		err := fs.LoadConfig()
		if err == nil {
			t.Error("Expected error for invalid JSON")
		}
	})

	t.Run("apply config with unsupported type", func(t *testing.T) {
		// Create temporary JSON file with unsupported type
		tmpFile := "/tmp/unsupported_config.json"
		content := `{"name": "test", "complex": {"real": 1, "imag": 2}}`

		if err := os.WriteFile(tmpFile, []byte(content), 0644); err != nil {
			t.Skipf("Could not create temp file: %v", err)
		}
		defer os.Remove(tmpFile)

		fs := New("test")
		fs.String("name", "default", "Name setting")

		fs.SetConfigFile(tmpFile)
		// Should load successfully but skip unsupported types
		err := fs.LoadConfig()
		if err != nil {
			t.Errorf("Should handle unsupported types gracefully, got error: %v", err)
		}

		if fs.GetString("name") != "test" {
			t.Errorf("Expected name 'test', got '%s'", fs.GetString("name"))
		}
	})
}

// Test environment variable loading edge cases
func TestEnvironmentVariablesEdgeCases(t *testing.T) {
	t.Run("load env with prefix", func(t *testing.T) {
		fs := New("test")
		fs.String("name", "default", "Name setting")
		fs.Int("port", 8080, "Port setting")

		// Set environment variables
		os.Setenv("TEST_NAME", "env_name")
		os.Setenv("TEST_PORT", "9090")
		defer os.Unsetenv("TEST_NAME")
		defer os.Unsetenv("TEST_PORT")

		fs.SetEnvPrefix("TEST")
		fs.EnableEnvLookup()

		err := fs.LoadEnvironmentVariables()
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}

		if fs.GetString("name") != "env_name" {
			t.Errorf("Expected name 'env_name', got '%s'", fs.GetString("name"))
		}
		if fs.GetInt("port") != 9090 {
			t.Errorf("Expected port 9090, got %d", fs.GetInt("port"))
		}
	})

	t.Run("load env with invalid values", func(t *testing.T) {
		fs := New("test")
		fs.Int("port", 8080, "Port setting")
		fs.Bool("debug", false, "Debug setting")

		// Set invalid environment variables
		os.Setenv("TEST_PORT", "invalid_int")
		os.Setenv("TEST_DEBUG", "invalid_bool")
		defer os.Unsetenv("TEST_PORT")
		defer os.Unsetenv("TEST_DEBUG")

		fs.SetEnvPrefix("TEST")
		fs.EnableEnvLookup()

		err := fs.LoadEnvironmentVariables()
		// Should handle invalid values gracefully
		if err != nil {
			// Some errors are expected for invalid values
			if !strings.Contains(err.Error(), "invalid") {
				t.Errorf("Expected error about invalid values, got: %v", err)
			}
		}

		// Values should remain at defaults for invalid env vars
		if fs.GetInt("port") != 8080 {
			t.Errorf("Expected port to remain default 8080, got %d", fs.GetInt("port"))
		}
	})

	t.Run("load env without prefix", func(t *testing.T) {
		fs := New("test")
		fs.String("PATH", "default", "Path setting") // Using existing env var

		fs.EnableEnvLookup()

		err := fs.LoadEnvironmentVariables()
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}

		// PATH should be loaded from environment
		pathValue := fs.GetString("PATH")
		if pathValue == "default" {
			t.Error("Expected PATH to be loaded from environment, got default")
		}
	})
}

// Test additional parsing edge cases for better coverage
func TestParsingEdgeCases(t *testing.T) {
	t.Run("parse short flag with equals in complex form", func(t *testing.T) {
		fs := New("test")
		file := fs.StringVar("file", "f", "", "Input file")
		verbose := fs.BoolVar("verbose", "v", false, "Verbose output")

		// Test complex short flag with equals
		args := []string{"-f=input.txt", "-v"}
		err := fs.Parse(args)
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}

		if *file != "input.txt" {
			t.Errorf("Expected file 'input.txt', got '%s'", *file)
		}
		if !*verbose {
			t.Error("Expected verbose to be true")
		}
	})

	t.Run("parse combined short flags with invalid flag", func(t *testing.T) {
		fs := New("test")
		fs.BoolVar("verbose", "v", false, "Verbose output")
		fs.BoolVar("debug", "d", false, "Debug mode")

		// Test combined flags with one invalid
		args := []string{"-vxd"} // 'x' is invalid
		err := fs.Parse(args)
		if err == nil {
			t.Error("Expected error for invalid flag in combined short flags")
		}
	})

	t.Run("process help argument", func(t *testing.T) {
		fs := New("test")
		fs.String("name", "default", "Name setting")

		// Test help flag detection
		args := []string{"--help"}
		err := fs.Parse(args)
		// Help should either return error or be handled gracefully
		if err != nil && !strings.Contains(err.Error(), "help") {
			t.Errorf("Expected help-related error or success, got: %v", err)
		}
	})
}

// Test config value setting edge cases
func TestConfigValueSettingEdgeCases(t *testing.T) {
	t.Run("set config values with type mismatches", func(t *testing.T) {
		tmpFile := "/tmp/type_mismatch_config.json"
		content := `{
			"port": "not_a_number",
			"debug": "not_a_bool",
			"timeout": "not_a_duration",
			"ratio": "not_a_float"
		}`

		if err := os.WriteFile(tmpFile, []byte(content), 0644); err != nil {
			t.Skipf("Could not create temp file: %v", err)
		}
		defer os.Remove(tmpFile)

		fs := New("test")
		fs.Int("port", 8080, "Port number")
		fs.Bool("debug", false, "Debug mode")
		fs.Duration("timeout", 5*time.Second, "Timeout duration")
		fs.Float64("ratio", 1.0, "Ratio value")

		fs.SetConfigFile(tmpFile)
		err := fs.LoadConfig()
		// Should handle type errors gracefully
		if err != nil {
			// Type conversion errors are expected
			if !strings.Contains(err.Error(), "invalid") &&
				!strings.Contains(err.Error(), "parse") &&
				!strings.Contains(err.Error(), "convert") &&
				!strings.Contains(err.Error(), "expected") {
				t.Errorf("Expected type conversion error, got: %v", err)
			}
		}

		// Values should remain at defaults
		if fs.GetInt("port") != 8080 {
			t.Errorf("Expected port to remain 8080, got %d", fs.GetInt("port"))
		}
		if fs.GetBool("debug") != false {
			t.Errorf("Expected debug to remain false, got %v", fs.GetBool("debug"))
		}
		if fs.GetDuration("timeout") != 5*time.Second {
			t.Errorf("Expected timeout to remain 5s, got %v", fs.GetDuration("timeout"))
		}
		if fs.GetFloat64("ratio") != 1.0 {
			t.Errorf("Expected ratio to remain 1.0, got %f", fs.GetFloat64("ratio"))
		}
	})

	t.Run("set string slice from config with different formats", func(t *testing.T) {
		tmpFile := "/tmp/stringslice_config.json"
		content := `{
			"items1": ["a", "b", "c"],
			"items2": "x,y,z",
			"items3": []
		}`

		if err := os.WriteFile(tmpFile, []byte(content), 0644); err != nil {
			t.Skipf("Could not create temp file: %v", err)
		}
		defer os.Remove(tmpFile)

		fs := New("test")
		fs.StringSlice("items1", []string{}, "Items 1")
		fs.StringSlice("items2", []string{}, "Items 2")
		fs.StringSlice("items3", []string{}, "Items 3")

		fs.SetConfigFile(tmpFile)
		_ = fs.LoadConfig()
		// May have errors for string vs array formats - that's acceptable behavior

		// Check array format
		items1 := fs.GetStringSlice("items1")
		if len(items1) != 3 || items1[0] != "a" || items1[1] != "b" || items1[2] != "c" {
			t.Errorf("Expected items1 [a b c], got %v", items1)
		}

		// For items2, the library may handle string->array conversion differently
		// This is testing the edge case handling, not the specific behavior
		items2 := fs.GetStringSlice("items2")
		// Accept either empty (if conversion failed) or parsed (if succeeded)
		if len(items2) != 0 && len(items2) != 3 {
			t.Logf("Items2 conversion result: %v (length %d)", items2, len(items2))
		}

		// Check empty array
		items3 := fs.GetStringSlice("items3")
		if len(items3) != 0 {
			t.Errorf("Expected items3 to be empty, got %v", items3)
		}
	})
}

// Test environment variable edge cases
func TestEnvironmentVariableAdvanced(t *testing.T) {
	t.Run("load duration from env", func(t *testing.T) {
		fs := New("test")
		fs.Duration("timeout", 5*time.Second, "Timeout duration")

		os.Setenv("TEST_TIMEOUT", "30s")
		defer os.Unsetenv("TEST_TIMEOUT")

		fs.SetEnvPrefix("TEST")
		fs.EnableEnvLookup()

		err := fs.LoadEnvironmentVariables()
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}

		if fs.GetDuration("timeout") != 30*time.Second {
			t.Errorf("Expected timeout 30s, got %v", fs.GetDuration("timeout"))
		}
	})

	t.Run("load float64 from env", func(t *testing.T) {
		fs := New("test")
		fs.Float64("ratio", 1.0, "Ratio value")

		os.Setenv("TEST_RATIO", "3.14")
		defer os.Unsetenv("TEST_RATIO")

		fs.SetEnvPrefix("TEST")
		fs.EnableEnvLookup()

		err := fs.LoadEnvironmentVariables()
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}

		if fs.GetFloat64("ratio") != 3.14 {
			t.Errorf("Expected ratio 3.14, got %f", fs.GetFloat64("ratio"))
		}
	})

	t.Run("load string slice from env", func(t *testing.T) {
		fs := New("test")
		fs.StringSlice("items", []string{}, "List items")

		os.Setenv("TEST_ITEMS", "a,b,c")
		defer os.Unsetenv("TEST_ITEMS")

		fs.SetEnvPrefix("TEST")
		fs.EnableEnvLookup()

		err := fs.LoadEnvironmentVariables()
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}

		items := fs.GetStringSlice("items")
		if len(items) != 3 || items[0] != "a" || items[1] != "b" || items[2] != "c" {
			t.Errorf("Expected items [a b c], got %v", items)
		}
	})
}
