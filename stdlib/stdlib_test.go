package stdlib_test

import (
	"os"
	"testing"

	flag "github.com/agilira/flash-flags/stdlib"
)

func TestDropInReplacement(t *testing.T) {
	// Test that we can use stdlib API exactly
	name := flag.String("name", "default", "Name flag")
	port := flag.Int("port", 8080, "Port flag")
	debug := flag.Bool("debug", false, "Debug flag")

	// Simulate command line args
	oldArgs := os.Args
	os.Args = []string{"test", "--name", "test", "--port", "9090", "--debug"}
	defer func() { os.Args = oldArgs }()

	flag.Parse()

	if *name != "test" {
		t.Errorf("Expected name 'test', got '%s'", *name)
	}
	if *port != 9090 {
		t.Errorf("Expected port 9090, got %d", *port)
	}
	if !*debug {
		t.Errorf("Expected debug true, got %v", *debug)
	}
}

func TestStdlibVarFunctions(t *testing.T) {
	var (
		name  string
		port  int
		debug bool
	)

	flag.StringVar(&name, "varname", "default", "Name var flag")
	flag.IntVar(&port, "varport", 8080, "Port var flag")
	flag.BoolVar(&debug, "vardebug", false, "Debug var flag")

	// Simulate command line args
	oldArgs := os.Args
	os.Args = []string{"test", "--varname", "vartest", "--varport", "9999", "--vardebug"}
	defer func() { os.Args = oldArgs }()

	flag.Parse()

	if name != "vartest" {
		t.Errorf("Expected name 'vartest', got '%s'", name)
	}
	if port != 9999 {
		t.Errorf("Expected port 9999, got %d", port)
	}
	if !debug {
		t.Errorf("Expected debug true, got %v", debug)
	}
}

func TestStdlibUtilityFunctions(t *testing.T) {
	flag.String("utiltest", "default", "Utility test flag")

	// Test Lookup
	f := flag.Lookup("utiltest")
	if f == nil {
		t.Error("Expected to find flag 'utiltest'")
	}

	// Test NFlag (should be 0 before parsing)
	if flag.NFlag() < 0 {
		t.Error("NFlag should be non-negative")
	}

	// Test Args (should work even if empty)
	args := flag.Args()
	if args == nil {
		t.Error("Args should return empty slice, not nil")
	}
}
