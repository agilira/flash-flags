// errorpaths_test.go: every rejection a flag kind or the parser can return
//
// Copyright (c) 2026 AGILira - A. Giordano
// SPDX-License-Identifier: MPL-2.0

package flashflags

import (
	"io"
	"os"
	"strings"
	"testing"
)

// newKinds declares one flag of every kind, with a short key on the ones the
// short-syntax cases use.
func newKinds() *FlagSet {
	fs := New("app")
	fs.String("name", "", "")
	fs.IntVar("port", "p", 0, "")
	fs.BoolVar("debug", "d", false, "")
	fs.Float64("rate", 0, "")
	fs.Duration("timeout", 0, "")
	fs.StringSlice("tags", nil, "")
	return fs
}

// A value of the wrong JSON type is refused with the flag named, for every kind.
func TestConfigRejectsWrongType(t *testing.T) {
	for flag, value := range map[string]interface{}{
		"name":    42.0,
		"port":    "8080",
		"debug":   "yes",
		"rate":    "fast",
		"timeout": true,
		"tags":    "a,b",
	} {
		err := newKinds().setFlagValueFromConfig(flag, value)
		if err == nil || !strings.Contains(err.Error(), flag) {
			t.Errorf("%s = %#v: err = %v, want a type error naming the flag", flag, value, err)
		}
	}
}

// An int reaches a numeric flag as is when the config map is built in code
// rather than decoded from JSON, where every number is a float64.
func TestConfigAcceptsGoInt(t *testing.T) {
	fs := newKinds()
	if err := fs.applyConfig(map[string]interface{}{"port": 8080, "rate": 2}); err != nil {
		t.Fatal(err)
	}
	if fs.GetInt("port") != 8080 || fs.GetFloat64("rate") != 2 {
		t.Errorf("port=%d rate=%v, want 8080 and 2", fs.GetInt("port"), fs.GetFloat64("rate"))
	}
}

// An empty string from the config file passes screening and sets the flag.
func TestConfigEmptyString(t *testing.T) {
	fs := New("app")
	name := fs.String("name", "default", "")
	if err := fs.applyConfig(map[string]interface{}{"name": ""}); err != nil {
		t.Fatal(err)
	}
	if *name != "" || fs.Source("name") != "config" {
		t.Errorf("name=%q source=%s, want empty from config", *name, fs.Source("name"))
	}
}

// Every command-line syntax that carries a value reports a value its kind
// cannot parse.
func TestParseRejectsBadValue(t *testing.T) {
	for _, args := range [][]string{
		{"--debug=maybe"}, // bool parser
		{"-p", "abc"},     // short flag, value in the next argument
		{"-d=maybe"},      // short flag with equals
		{"-dp", "abc"},    // combined, last flag takes the next argument
	} {
		if err := newKinds().Parse(args); err == nil || !strings.Contains(err.Error(), "invalid") {
			t.Errorf("Parse(%q) = %v, want an invalid-value error", args, err)
		}
	}
}

func TestParseRejectsUnknownShortWithEquals(t *testing.T) {
	if err := newKinds().Parse([]string{"-x=1"}); err == nil || !strings.Contains(err.Error(), "unknown flag: -x") {
		t.Errorf("err = %v, want unknown flag -x", err)
	}
}

// The item cap holds across repetitions: each value is under the length
// limit, together they exceed maxSliceItems.
func TestRepeatedSliceCap(t *testing.T) {
	chunk := strings.Repeat("a,", maxSliceItems/2-1) + "a" // maxSliceItems/2 items
	args := []string{"--tags", chunk, "--tags", chunk, "--tags", chunk}
	if err := newKinds().Parse(args); err == nil || !strings.Contains(err.Error(), "too large") {
		t.Errorf("err = %v, want the slice cap", err)
	}
}

// Args before any Parse is an empty slice, not nil.
func TestArgsBeforeParse(t *testing.T) {
	if a := New("app").Args(); a == nil || len(a) != 0 {
		t.Errorf("Args() = %#v, want an empty slice", a)
	}
}

// Auto-discovery with no matching file loads nothing and is not an error.
func TestConfigPathWithoutFile(t *testing.T) {
	fs := New("app")
	name := fs.String("name", "default", "")
	fs.AddConfigPath(t.TempDir())
	if err := fs.LoadConfig(); err != nil {
		t.Fatal(err)
	}
	if *name != "default" {
		t.Errorf("name = %q, want the default", *name)
	}
}

// A flag whose name maps to an empty environment variable name is skipped,
// not looked up as "".
func TestEnvSkipsEmptyVarName(t *testing.T) {
	fs := New("app")
	fs.String("", "default", "")
	fs.EnableEnvLookup()
	if err := fs.LoadEnvironmentVariables(); err != nil {
		t.Fatal(err)
	}
	if fs.Changed("") {
		t.Error("a flag with no environment variable name was set")
	}
}

// PrintUsage lists the short key next to the long name.
func TestPrintUsageShortKey(t *testing.T) {
	fs := New("app")
	fs.IntVar("port", "p", 8080, "Server port")

	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	stdout := os.Stdout
	os.Stdout = w
	fs.PrintUsage()
	os.Stdout = stdout
	_ = w.Close()
	out, _ := io.ReadAll(r)

	if !strings.Contains(string(out), "--port, -p") {
		t.Errorf("PrintUsage output %q lacks \"--port, -p\"", out)
	}
}
