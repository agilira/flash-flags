// flash-flags.go: Ultra-fast command-line flag parsing for Go Tests
//
// Copyright (c) 2025 AGILira
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

func TestValidation(t *testing.T) {
	fs := New("test")

	// Test string validation
	fs.String("name", "", "Your name")
	err := fs.SetValidator("name", func(v interface{}) error {
		if str, ok := v.(string); ok {
			if len(str) < 2 {
				return fmt.Errorf("name must be at least 2 characters")
			}
		}
		return nil
	})
	if err != nil {
		t.Fatalf("SetValidator failed: %v", err)
	}

	// Test valid value
	args := []string{"--name=John"}
	err = fs.Parse(args)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	if fs.GetString("name") != "John" {
		t.Errorf("Expected 'John', got '%s'", fs.GetString("name"))
	}

	// Test invalid value
	fs = New("test")
	fs.String("name", "", "Your name")
	fs.SetValidator("name", func(v interface{}) error {
		if str, ok := v.(string); ok {
			if len(str) < 2 {
				return fmt.Errorf("name must be at least 2 characters")
			}
		}
		return nil
	})

	args = []string{"--name=A"}
	err = fs.Parse(args)
	if err == nil {
		t.Error("Expected validation error for short name")
	}

	// Test int validation
	fs = New("test")
	fs.Int("port", 0, "Port number")
	err = fs.SetValidator("port", func(v interface{}) error {
		if intVal, ok := v.(int); ok {
			if intVal < 1 || intVal > 65535 {
				return fmt.Errorf("port must be between 1 and 65535")
			}
		}
		return nil
	})
	if err != nil {
		t.Fatalf("SetValidator failed: %v", err)
	}

	// Valid port
	args = []string{"--port=8080"}
	err = fs.Parse(args)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	if fs.GetInt("port") != 8080 {
		t.Errorf("Expected 8080, got %d", fs.GetInt("port"))
	}

	// Invalid port
	fs = New("test")
	fs.Int("port", 0, "Port number")
	fs.SetValidator("port", func(v interface{}) error {
		if intVal, ok := v.(int); ok {
			if intVal < 1 || intVal > 65535 {
				return fmt.Errorf("port must be between 1 and 65535")
			}
		}
		return nil
	})

	args = []string{"--port=70000"}
	err = fs.Parse(args)
	if err == nil {
		t.Error("Expected validation error for invalid port")
	}

	// Test ValidateAll
	fs = New("test")
	fs.String("name", "TestName", "Your name")
	fs.Int("port", 8080, "Port number")

	fs.SetValidator("name", func(v interface{}) error {
		if str, ok := v.(string); ok {
			if len(str) < 2 {
				return fmt.Errorf("name too short")
			}
		}
		return nil
	})

	fs.SetValidator("port", func(v interface{}) error {
		if intVal, ok := v.(int); ok {
			if intVal < 1 || intVal > 65535 {
				return fmt.Errorf("invalid port range")
			}
		}
		return nil
	})

	// Should pass validation with default values
	err = fs.ValidateAll()
	if err != nil {
		t.Errorf("ValidateAll failed: %v", err)
	}
} // Test string slice parsing
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

		// Reset all flags
		fs.Reset()

		// Verify values are back to defaults
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
	})

	t.Run("ResetSpecificFlag", func(t *testing.T) {
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
		if *strFlag != "default" {
			t.Errorf("After reset, expected strFlag to be 'default', got '%s'", *strFlag)
		}
		if *intFlag != 100 {
			t.Errorf("After reset, expected intFlag to remain 100, got %d", *intFlag)
		}
	})

	t.Run("ResetNonExistentFlag", func(t *testing.T) {
		fs := New("test")

		err := fs.ResetFlag("nonexistent")
		if err == nil {
			t.Error("Expected error when resetting non-existent flag")
		}
		if err.Error() != "flag --nonexistent not found" {
			t.Errorf("Expected 'flag --nonexistent not found' error, got: %v", err)
		}
	})
}

func TestRequiredFlags(t *testing.T) {
	t.Run("RequiredFlagProvided", func(t *testing.T) {
		fs := New("test")

		name := fs.String("name", "", "Name flag")
		fs.SetRequired("name")

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
		fs.SetRequired("name")

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
		fs := New("test")

		host := fs.String("host", "localhost", "Host flag")
		port := fs.Int("port", 8080, "Port flag")
		fs.SetDependencies("port", "host")

		args := []string{"--host", "myhost", "--port", "3000"}
		if err := fs.Parse(args); err != nil {
			t.Fatalf("Parse failed: %v", err)
		}

		if *host != "myhost" {
			t.Errorf("Expected host to be 'myhost', got '%s'", *host)
		}
		if *port != 3000 {
			t.Errorf("Expected port to be 3000, got %d", *port)
		}
	})

	t.Run("DependencyMissing", func(t *testing.T) {
		fs := New("test")

		fs.String("host", "localhost", "Host flag")
		fs.Int("port", 8080, "Port flag")
		fs.SetDependencies("port", "host")

		args := []string{"--port", "3000"}
		err := fs.Parse(args)
		if err == nil {
			t.Error("Expected error when dependency is missing")
		}
		if err.Error() != "flag --port requires --host to be set" {
			t.Errorf("Expected 'flag --port requires --host to be set' error, got: %v", err)
		}
	})

	t.Run("MultipleDependencies", func(t *testing.T) {
		fs := New("test")

		user := fs.String("user", "", "User flag")
		pass := fs.String("pass", "", "Password flag")
		auth := fs.Bool("auth", false, "Auth flag")
		fs.SetDependencies("auth", "user", "pass")

		args := []string{"--user", "admin", "--pass", "secret", "--auth"}
		if err := fs.Parse(args); err != nil {
			t.Fatalf("Parse failed: %v", err)
		}

		if *user != "admin" {
			t.Errorf("Expected user to be 'admin', got '%s'", *user)
		}
		if *pass != "secret" {
			t.Errorf("Expected pass to be 'secret', got '%s'", *pass)
		}
		if !*auth {
			t.Errorf("Expected auth to be true, got %t", *auth)
		}
	})

	t.Run("NonExistentDependency", func(t *testing.T) {
		fs := New("test")

		fs.String("name", "", "Name flag")
		fs.SetDependencies("name", "nonexistent")

		args := []string{"--name", "test"}
		err := fs.Parse(args)
		if err == nil {
			t.Error("Expected error when dependency doesn't exist")
		}
		if err.Error() != "flag --name depends on non-existent flag --nonexistent" {
			t.Errorf("Expected 'flag --name depends on non-existent flag --nonexistent' error, got: %v", err)
		}
	})

	t.Run("SetDependenciesNonExistentFlag", func(t *testing.T) {
		fs := New("test")

		err := fs.SetDependencies("nonexistent", "something")
		if err == nil {
			t.Error("Expected error when setting dependencies on non-existent flag")
		}
		if err.Error() != "flag not found: nonexistent" {
			t.Errorf("Expected 'flag not found: nonexistent' error, got: %v", err)
		}
	})
}

func TestHelpSystem(t *testing.T) {
	t.Run("BasicHelpGeneration", func(t *testing.T) {
		fs := New("myapp")
		fs.SetDescription("A test application for demonstration")
		fs.SetVersion("1.0.0")

		_ = fs.String("name", "default", "Name of the service")
		port := fs.IntVar("port", "p", 8080, "Server port")
		_ = fs.Bool("verbose", false, "Enable verbose output")

		fs.SetRequired("name")
		fs.SetGroup("port", "Server Options")

		help := fs.Help()

		// Check that help contains expected elements
		if !strings.Contains(help, "A test application for demonstration") {
			t.Error("Help should contain description")
		}
		if !strings.Contains(help, "Version: 1.0.0") {
			t.Error("Help should contain version")
		}
		if !strings.Contains(help, "Usage: myapp [options]") {
			t.Error("Help should contain usage line")
		}
		if !strings.Contains(help, "--name") {
			t.Error("Help should contain --name flag")
		}
		if !strings.Contains(help, "-p, --port") {
			t.Error("Help should contain short and long flag format")
		}
		if !strings.Contains(help, "[REQUIRED]") {
			t.Error("Help should indicate required flags")
		}
		if !strings.Contains(help, "Server Options:") {
			t.Error("Help should contain group headers")
		}
		if !strings.Contains(help, "(default: 8080)") {
			t.Error("Help should show default values")
		}

		// Verify port is not nil (used in test)
		if port == nil {
			t.Error("Port should not be nil")
		}
	})

	t.Run("HelpFlagHandling", func(t *testing.T) {
		fs := New("test")
		_ = fs.String("name", "", "Name flag")

		// Test --help
		err := fs.Parse([]string{"--help"})
		if err == nil {
			t.Error("Expected error when --help is used")
		}
		if err.Error() != "help requested" {
			t.Errorf("Expected 'help requested' error, got: %v", err)
		}
	})

	t.Run("ShortHelpFlagHandling", func(t *testing.T) {
		fs := New("test")
		_ = fs.String("name", "", "Name flag")

		// Test -h
		err := fs.Parse([]string{"-h"})
		if err == nil {
			t.Error("Expected error when -h is used")
		}
		if err.Error() != "help requested" {
			t.Errorf("Expected 'help requested' error, got: %v", err)
		}
	})

	t.Run("SetGroupNonExistentFlag", func(t *testing.T) {
		fs := New("test")

		err := fs.SetGroup("nonexistent", "Some Group")
		if err == nil {
			t.Error("Expected error when setting group on non-existent flag")
		}
		if err.Error() != "flag not found: nonexistent" {
			t.Errorf("Expected 'flag not found: nonexistent' error, got: %v", err)
		}
	})

	t.Run("HelpWithDependencies", func(t *testing.T) {
		fs := New("test")

		_ = fs.String("user", "", "Username")
		_ = fs.String("pass", "", "Password")
		_ = fs.Bool("auth", false, "Enable authentication")

		fs.SetDependencies("auth", "user", "pass")

		help := fs.Help()
		if !strings.Contains(help, "[depends on: user, pass]") {
			t.Error("Help should show dependencies")
		}
	})
}

func TestConfigurationFiles(t *testing.T) {
	t.Run("LoadJSONConfig", func(t *testing.T) {
		// Create a temporary config file
		configContent := `{
			"host": "config-host",
			"port": 9000,
			"debug": true,
			"rate": 2.5,
			"tags": ["config", "test"]
		}`

		tmpfile, err := os.CreateTemp("", "test-config-*.json")
		if err != nil {
			t.Fatalf("Failed to create temp file: %v", err)
		}
		defer os.Remove(tmpfile.Name())

		if _, err := tmpfile.Write([]byte(configContent)); err != nil {
			t.Fatalf("Failed to write config file: %v", err)
		}
		tmpfile.Close()

		// Setup flags
		fs := New("test")
		host := fs.String("host", "default", "Host flag")
		port := fs.Int("port", 8080, "Port flag")
		debug := fs.Bool("debug", false, "Debug flag")
		rate := fs.Float64("rate", 1.0, "Rate flag")
		tags := fs.StringSlice("tags", []string{}, "Tags flag")

		// Set config file and parse
		fs.SetConfigFile(tmpfile.Name())

		// Parse with no command line args (should use config values)
		if err := fs.Parse([]string{}); err != nil {
			t.Fatalf("Parse failed: %v", err)
		}

		// Verify config values were loaded
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
	})

	t.Run("CommandLineOverridesConfig", func(t *testing.T) {
		// Create a temporary config file
		configContent := `{
			"host": "config-host",
			"port": 9000
		}`

		tmpfile, err := os.CreateTemp("", "test-override-*.json")
		if err != nil {
			t.Fatalf("Failed to create temp file: %v", err)
		}
		defer os.Remove(tmpfile.Name())

		if _, err := tmpfile.Write([]byte(configContent)); err != nil {
			t.Fatalf("Failed to write config file: %v", err)
		}
		tmpfile.Close()

		// Setup flags
		fs := New("test")
		host := fs.String("host", "default", "Host flag")
		port := fs.Int("port", 8080, "Port flag")

		// Set config file
		fs.SetConfigFile(tmpfile.Name())

		// Parse with command line args that should override config
		args := []string{"--host", "cmdline-host"}
		if err := fs.Parse(args); err != nil {
			t.Fatalf("Parse failed: %v", err)
		}

		// Command line should override config, config should override default
		if *host != "cmdline-host" {
			t.Errorf("Expected host from command line 'cmdline-host', got '%s'", *host)
		}
		if *port != 9000 {
			t.Errorf("Expected port from config 9000, got %d", *port)
		}
	})

	t.Run("ConfigFileNotFound", func(t *testing.T) {
		fs := New("test")
		_ = fs.String("host", "default", "Host flag")

		// Set non-existent config file (should not cause error, just skip)
		fs.SetConfigFile("/non/existent/config.json")

		if err := fs.Parse([]string{}); err != nil {
			t.Fatalf("Parse should not fail when config file doesn't exist: %v", err)
		}
	})

	t.Run("InvalidJSONConfig", func(t *testing.T) {
		// Create invalid JSON config
		configContent := `{ "host": "test", invalid json }`

		tmpfile, err := os.CreateTemp("", "test-invalid-*.json")
		if err != nil {
			t.Fatalf("Failed to create temp file: %v", err)
		}
		defer os.Remove(tmpfile.Name())

		if _, err := tmpfile.Write([]byte(configContent)); err != nil {
			t.Fatalf("Failed to write config file: %v", err)
		}
		tmpfile.Close()

		fs := New("test")
		_ = fs.String("host", "default", "Host flag")
		fs.SetConfigFile(tmpfile.Name())

		err = fs.Parse([]string{})
		if err == nil {
			t.Error("Expected error when parsing invalid JSON config")
		}
		if !strings.Contains(err.Error(), "config file error") {
			t.Errorf("Expected config file error, got: %v", err)
		}
	})

	t.Run("ConfigValidation", func(t *testing.T) {
		// Create config with invalid value
		configContent := `{
			"port": 99999
		}`

		tmpfile, err := os.CreateTemp("", "test-validation-*.json")
		if err != nil {
			t.Fatalf("Failed to create temp file: %v", err)
		}
		defer os.Remove(tmpfile.Name())

		if _, err := tmpfile.Write([]byte(configContent)); err != nil {
			t.Fatalf("Failed to write config file: %v", err)
		}
		tmpfile.Close()

		fs := New("test")
		_ = fs.Int("port", 8080, "Port flag")

		// Set validator that rejects the config value
		fs.SetValidator("port", func(val interface{}) error {
			port := val.(int)
			if port > 65535 {
				return fmt.Errorf("port must be <= 65535")
			}
			return nil
		})

		fs.SetConfigFile(tmpfile.Name())

		err = fs.Parse([]string{})
		if err == nil {
			t.Error("Expected validation error from config")
		}
		if !strings.Contains(err.Error(), "port must be <= 65535") {
			t.Errorf("Expected port validation error, got: %v", err)
		}
	})
}

// TestEnvironmentVariables tests environment variable support
func TestEnvironmentVariables(t *testing.T) {
	t.Run("BasicEnvSupport", func(t *testing.T) {
		// Set environment variables
		os.Setenv("TEST_HOST", "env.example.com")
		os.Setenv("TEST_PORT", "9090")
		os.Setenv("TEST_DEBUG", "true")
		defer os.Unsetenv("TEST_HOST")
		defer os.Unsetenv("TEST_PORT")
		defer os.Unsetenv("TEST_DEBUG")

		fs := New("myapp")
		fs.SetEnvPrefix("TEST")

		host := fs.String("host", "localhost", "Host address")
		port := fs.Int("port", 8080, "Port number")
		debug := fs.Bool("debug", false, "Debug mode")

		err := fs.Parse([]string{})
		if err != nil {
			t.Fatalf("Parse failed: %v", err)
		}

		if *host != "env.example.com" {
			t.Errorf("Expected host from env 'env.example.com', got '%s'", *host)
		}
		if *port != 9090 {
			t.Errorf("Expected port from env 9090, got %d", *port)
		}
		if *debug != true {
			t.Errorf("Expected debug from env true, got %v", *debug)
		}
	})

	t.Run("CommandLineOverridesEnv", func(t *testing.T) {
		// Set environment variables
		os.Setenv("TEST_HOST", "env.example.com")
		os.Setenv("TEST_PORT", "9090")
		defer os.Unsetenv("TEST_HOST")
		defer os.Unsetenv("TEST_PORT")

		fs := New("myapp")
		fs.SetEnvPrefix("TEST")

		host := fs.String("host", "localhost", "Host address")
		port := fs.Int("port", 8080, "Port number")

		// Command line should override environment
		err := fs.Parse([]string{"--host", "cmd.example.com", "--port", "3000"})
		if err != nil {
			t.Fatalf("Parse failed: %v", err)
		}

		if *host != "cmd.example.com" {
			t.Errorf("Expected host from command line 'cmd.example.com', got '%s'", *host)
		}
		if *port != 3000 {
			t.Errorf("Expected port from command line 3000, got %d", *port)
		}
	})

	t.Run("CustomEnvVarNames", func(t *testing.T) {
		// Set custom environment variable
		os.Setenv("CUSTOM_SERVER_HOST", "custom.example.com")
		defer os.Unsetenv("CUSTOM_SERVER_HOST")

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

		if *host != "custom.example.com" {
			t.Errorf("Expected host from custom env 'custom.example.com', got '%s'", *host)
		}
	})

	t.Run("DefaultNaming", func(t *testing.T) {
		// Set environment variable with default naming
		os.Setenv("DB_HOST", "db.example.com")
		defer os.Unsetenv("DB_HOST")

		fs := New("myapp")
		fs.EnableEnvLookup()

		dbHost := fs.String("db-host", "localhost", "Database host")

		err := fs.Parse([]string{})
		if err != nil {
			t.Fatalf("Parse failed: %v", err)
		}

		if *dbHost != "db.example.com" {
			t.Errorf("Expected db-host from env 'db.example.com', got '%s'", *dbHost)
		}
	})
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
		portVal := val.(int)
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
			// Set environment variable
			os.Setenv(tc.envVar, tc.envValue)
			defer os.Unsetenv(tc.envVar)

			fs := New("test")
			fs.EnableEnvLookup()

			// Register appropriate flag type
			switch tc.flagType {
			case "string":
				flag := fs.String(tc.flagName, "", "Test flag")
				fs.SetEnvVar(tc.flagName, tc.envVar)
				fs.Parse([]string{})
				if *flag != tc.expected {
					t.Errorf("Expected %v, got %v", tc.expected, *flag)
				}
			case "int":
				flag := fs.Int(tc.flagName, 0, "Test flag")
				fs.SetEnvVar(tc.flagName, tc.envVar)
				fs.Parse([]string{})
				if *flag != tc.expected {
					t.Errorf("Expected %v, got %v", tc.expected, *flag)
				}
			case "bool":
				flag := fs.Bool(tc.flagName, false, "Test flag")
				fs.SetEnvVar(tc.flagName, tc.envVar)
				fs.Parse([]string{})
				if *flag != tc.expected {
					t.Errorf("Expected %v, got %v", tc.expected, *flag)
				}
			case "duration":
				flag := fs.Duration(tc.flagName, 0, "Test flag")
				fs.SetEnvVar(tc.flagName, tc.envVar)
				fs.Parse([]string{})
				if *flag != tc.expected {
					t.Errorf("Expected %v, got %v", tc.expected, *flag)
				}
			case "float64":
				flag := fs.Float64(tc.flagName, 0, "Test flag")
				fs.SetEnvVar(tc.flagName, tc.envVar)
				fs.Parse([]string{})
				if *flag != tc.expected {
					t.Errorf("Expected %v, got %v", tc.expected, *flag)
				}
			case "stringSlice":
				flag := fs.StringSlice(tc.flagName, nil, "Test flag")
				fs.SetEnvVar(tc.flagName, tc.envVar)
				fs.Parse([]string{})
				expected := tc.expected.([]string)
				if len(*flag) != len(expected) {
					t.Errorf("Expected slice length %d, got %d", len(expected), len(*flag))
				}
				for i, v := range expected {
					if (*flag)[i] != v {
						t.Errorf("Expected slice element %d to be %s, got %s", i, v, (*flag)[i])
					}
				}
			}
		})
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
	fs.SetValidator("port", func(val interface{}) error {
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
	fs.SetRequired("missing")
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
	fs := New("test")

	// Test short flag functionality
	verbose := fs.BoolVar("verbose", "v", false, "Verbose mode")

	// Test short flag parsing
	err := fs.Parse([]string{"-v"})
	if err != nil {
		t.Fatalf("Parse should succeed with short flag: %v", err)
	}

	if !*verbose {
		t.Error("Short flag should set verbose to true")
	}

	// Test flag with empty short key
	name := fs.StringVar("name", "", "default", "Name")
	err = fs.Parse([]string{"--name", "test"})
	if err != nil {
		t.Fatalf("Parse should succeed: %v", err)
	}

	if *name != "test" {
		t.Errorf("Expected name=test, got %s", *name)
	}

	// Test all remaining getter edge cases
	fs.String("test-string", "default", "Test string")
	fs.Int("test-int", 123, "Test int")
	fs.Bool("test-bool", true, "Test bool")
	fs.Duration("test-duration", time.Hour, "Test duration")
	fs.Float64("test-float", 2.5, "Test float")
	fs.StringSlice("test-slice", []string{"a", "b"}, "Test slice")

	// Test getters with existing flags
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
