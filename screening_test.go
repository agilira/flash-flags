// screening_test.go: input hygiene tests.
//
// The parser screens flag values for input that is *malformed* -- null bytes,
// control characters, absurd length. It deliberately does not try to guess what
// a value will mean to the application that receives it.
//
// Copyright (c) 2025 AGILira - A. Giordano
// SPDX-License-Identifier: MPL-2.0

package flashflags

import (
	"strings"
	"testing"
)

// parseValue feeds one value through the command-line path and returns what the
// flag ended up holding.
func parseValue(t *testing.T, value string) (string, error) {
	t.Helper()
	fs := New("app")
	fs.String("v", "", "")
	err := fs.Parse([]string{"--v", value})
	return fs.GetString("v"), err
}

// TestScreening_AcceptsLegitimateValues covers values a general-purpose CLI must
// be able to receive. Every one of these was rejected before v1.1.9 by a
// denylist that tried to infer intent from a substring.
func TestScreening_AcceptsLegitimateValues(t *testing.T) {
	cases := []struct{ name, value string }{
		{"absolute config path", "/etc/myapp/config.yaml"},
		{"proc path", "/proc/self/status"},
		{"sys path", "/sys/class/net/eth0/address"},
		{"command argument", "rm -rf /tmp/build"},
		{"relative path", "../sibling/data.json"},
		{"windows relative path", `..\sibling\data.json`},
		{"printf template", "%s-%d.log"},
		{"go format verb", "%v"},
		{"percent in prose", "Progress: 50% done"},
		{"company name", "Comp Ltd"},
		{"serial port", "COM3"},
		{"sql statement", "DROP TABLE staging_imports"},
		{"shell substitution in a message", "cost is $(price) euro"},
		{"backtick in markdown", "use `--help` for usage"},
		{"url with query", "https://example.com/p?a=1&b=2"},
		{"connection string", "postgres://user:pw@host:5432/db?sslmode=require"},
		{"jq expression", ".items[] | select(.id)"},
		{"tab inside value", "col1\tcol2"},
		{"newline inside value", "line1\nline2"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := parseValue(t, tc.value)
			if err != nil {
				t.Fatalf("Parse(%q) = %v, want nil", tc.value, err)
			}
			if got != tc.value {
				t.Errorf("value = %q, want %q (values must round-trip unchanged)", got, tc.value)
			}
		})
	}
}

// TestScreening_RejectsMalformedInput covers what the screening still does: it
// rejects input that is broken as a string, whatever it is later used for.
func TestScreening_RejectsMalformedInput(t *testing.T) {
	cases := []struct{ name, value, wantErr string }{
		{"null byte", "abc\x00def", "null byte"},
		{"leading null byte", "\x00abc", "null byte"},
		{"bell control char", "abc\x07def", "control character"},
		{"escape control char", "\x1b[31mred", "control character"},
		{"over length", strings.Repeat("a", 10001), "too long"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := parseValue(t, tc.value)
			if err == nil {
				t.Fatalf("Parse succeeded, want error mentioning %q", tc.wantErr)
			}
			if !strings.Contains(err.Error(), tc.wantErr) {
				t.Errorf("error = %v, want it to mention %q", err, tc.wantErr)
			}
		})
	}
}

// TestScreening_AtLengthBoundary pins the exact cutoff.
func TestScreening_AtLengthBoundary(t *testing.T) {
	if _, err := parseValue(t, strings.Repeat("x", 10000)); err != nil {
		t.Errorf("10000 chars rejected: %v", err)
	}
	if _, err := parseValue(t, strings.Repeat("x", 10001)); err == nil {
		t.Error("10001 chars accepted, want rejection")
	}
}

// TestScreening_DoesNotGuessSemantics is deliberately written as an assertion,
// not a comment. A flag parser cannot know whether a value will reach a shell,
// a SQL driver or a file open, so it does not pretend to. These values pass
// through untouched, and escaping them correctly is the caller's job.
//
// If someone later adds ';' or '|' to a denylist, this test fails and says why.
func TestScreening_DoesNotGuessSemantics(t *testing.T) {
	cases := []string{
		"a; rm -rf ~",
		"deploy && restart",
		"cat file | grep x",
		"value\nsecond-line",
		"..%2f..%2fetc%2fpasswd",
		"%2e%2e%2f",
		"' OR 1=1 --",
		"$IFS$()cat",
		"<script>alert(1)</script>",
	}
	for _, c := range cases {
		got, err := parseValue(t, c)
		if err != nil {
			t.Errorf("Parse(%q) = %v; the parser must not reject on guessed intent", c, err)
			continue
		}
		if got != c {
			t.Errorf("value = %q, want %q unchanged", got, c)
		}
	}
}

// TestScreening_AppliesToEverySource checks the screening covers the config file
// and environment paths, not just the command line.
func TestScreening_AppliesToEverySource(t *testing.T) {
	t.Run("environment", func(t *testing.T) {
		// A null byte cannot be tested here: the OS forbids one inside an
		// environment value, and os.Setenv rejects it outright. A C0 control
		// character carries fine, so that is what the environment path is
		// screened for.
		fs := New("app")
		fs.String("host", "d", "")
		fs.SetEnvPrefix("APP")
		fs.EnableEnvLookup()
		t.Setenv("APP_HOST", "bad\x07value")
		if err := fs.Parse([]string{}); err == nil {
			t.Error("Parse accepted a control character from the environment")
		}
	})

	t.Run("config file", func(t *testing.T) {
		fs := New("app")
		fs.String("host", "d", "")
		fs.SetConfigFile(writeConfig(t, "{\"host\":\"bad\\u0000value\"}"))
		if err := fs.Parse([]string{}); err == nil {
			t.Error("Parse accepted a null byte from the config file")
		}
	})

	t.Run("string slice element", func(t *testing.T) {
		fs := New("app")
		fs.StringSlice("tags", nil, "")
		if err := fs.Parse([]string{"--tags", "ok,bad\x00item"}); err == nil {
			t.Error("Parse accepted a null byte inside a slice element")
		}
	})

	t.Run("string slice legitimate elements", func(t *testing.T) {
		fs := New("app")
		fs.StringSlice("paths", nil, "")
		if err := fs.Parse([]string{"--paths", "/etc/a.conf,/proc/b,../c"}); err != nil {
			t.Fatalf("Parse: %v", err)
		}
		got := fs.GetStringSlice("paths")
		want := []string{"/etc/a.conf", "/proc/b", "../c"}
		if len(got) != len(want) {
			t.Fatalf("paths = %v, want %v", got, want)
		}
		for i := range want {
			if got[i] != want[i] {
				t.Errorf("paths[%d] = %q, want %q", i, got[i], want[i])
			}
		}
	})
}

// TestScreening_FastPath checks that the fast path and the full scan agree: a
// value the fast path waves through must also be one the full scan accepts.
func TestScreening_FastPath(t *testing.T) {
	fs := New("app")
	// Simple alphanumeric, under 100 chars: takes the fast path.
	if !isSimpleAlphanumeric("host-name_1.2:8080") {
		t.Fatal("expected a simple alphanumeric value")
	}
	if err := fs.validateInputHygieneSlow("v", "host-name_1.2:8080"); err != nil {
		t.Errorf("full scan rejected a fast-path value: %v", err)
	}
	if isSimpleAlphanumeric("has space") {
		t.Error("isSimpleAlphanumeric accepted a space")
	}
	if isSimpleAlphanumeric("nul\x00") {
		t.Error("isSimpleAlphanumeric accepted a null byte")
	}
}

// TestScreening_ConfigArrayElements covers a gap that predates v1.1.9: a config
// file value is screened only when it decodes to a string, so the elements of a
// JSON array reached the flag unscreened. The same value supplied on the command
// line was rejected, which made the guarantee depend on which source a value
// arrived from.
func TestScreening_ConfigArrayElements(t *testing.T) {
	tests := []struct {
		name string
		body string
	}{
		{"null byte in element", "{\"tags\":[\"ok\",\"bad\\u0000item\"]}"},
		{"control char in element", "{\"tags\":[\"ok\",\"bad\\u0007item\"]}"},
		{"null byte in first element", "{\"tags\":[\"bad\\u0000item\"]}"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fs := New("app")
			fs.StringSlice("tags", []string{"default"}, "")
			fs.SetConfigFile(writeConfig(t, tt.body))
			if err := fs.Parse([]string{}); err == nil {
				t.Fatalf("Parse accepted %s; got tags=%q", tt.name, fs.GetStringSlice("tags"))
			}
		})
	}
}

// TestScreening_ConfigArrayLegitimateElements checks the fix does not reject
// ordinary array values, including ones the removed denylist used to block.
func TestScreening_ConfigArrayLegitimateElements(t *testing.T) {
	fs := New("app")
	fs.StringSlice("paths", nil, "")
	fs.SetConfigFile(writeConfig(t, `{"paths":["/etc/a.conf","../b","c; d","%s"]}`))
	if err := fs.Parse([]string{}); err != nil {
		t.Fatalf("Parse: %v", err)
	}
	want := []string{"/etc/a.conf", "../b", "c; d", "%s"}
	got := fs.GetStringSlice("paths")
	if len(got) != len(want) {
		t.Fatalf("paths = %q, want %q", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("paths[%d] = %q, want %q", i, got[i], want[i])
		}
	}
}

// TestScreening_SameGuaranteeFromEverySource states the property the gap broke:
// whether a value is accepted must not depend on which layer supplied it.
func TestScreening_SameGuaranteeFromEverySource(t *testing.T) {
	const bad = "bad\x07item"

	fromCLI := func() error {
		fs := New("app")
		fs.StringSlice("tags", nil, "")
		return fs.Parse([]string{"--tags", "ok," + bad})
	}
	fromConfig := func() error {
		fs := New("app")
		fs.StringSlice("tags", nil, "")
		fs.SetConfigFile(writeConfig(t, "{\"tags\":[\"ok\",\"bad\\u0007item\"]}"))
		return fs.Parse([]string{})
	}
	fromEnv := func() error {
		fs := New("app")
		fs.StringSlice("tags", nil, "")
		fs.SetEnvPrefix("APP")
		fs.EnableEnvLookup()
		t.Setenv("APP_TAGS", "ok,"+bad)
		return fs.Parse([]string{})
	}

	for name, run := range map[string]func() error{
		"cli": fromCLI, "config": fromConfig, "env": fromEnv,
	} {
		if err := run(); err == nil {
			t.Errorf("%s: accepted a control character; every source must screen alike", name)
		}
	}
}

// TestScreening_ConfigArrayNonStringElements checks that a non-string element
// falls through screening to the element setter, which is what reports the type
// mismatch. Screening must not shadow that error with one of its own.
func TestScreening_ConfigArrayNonStringElements(t *testing.T) {
	fs := New("app")
	fs.StringSlice("tags", []string{"default"}, "")
	fs.SetConfigFile(writeConfig(t, `{"tags":["ok",42]}`))
	err := fs.Parse([]string{})
	if err == nil {
		t.Fatal("Parse accepted a number inside a string slice")
	}
	if !strings.Contains(err.Error(), "expected string array") {
		t.Errorf("error = %v, want the element setter's type error", err)
	}
}

// TestScreening_LengthCapCountsBytes pins both halves of the length limit: it
// is measured in bytes, and the error says so. Go's len counts bytes, so a
// value of multi-byte runes is rejected well before it reaches 10000
// characters, and reporting that count as "chars" misled the reader into
// thinking the limit was on characters.
func TestScreening_LengthCapCountsBytes(t *testing.T) {
	// "à" is two bytes in UTF-8.
	const rep = 6000
	value := strings.Repeat("à", rep)
	if len(value) != rep*2 {
		t.Fatalf("expected %d bytes, got %d", rep*2, len(value))
	}

	_, err := parseValue(t, value)
	if err == nil {
		t.Fatalf("Parse accepted %d bytes, want rejection above %d", len(value), maxValueLength)
	}
	if !strings.Contains(err.Error(), "bytes") {
		t.Errorf("error = %v, want it to say bytes", err)
	}
	if strings.Contains(err.Error(), "chars") {
		t.Errorf("error = %v, still reports a byte count as characters", err)
	}
	if !strings.Contains(err.Error(), "12000") {
		t.Errorf("error = %v, want it to report the byte count 12000", err)
	}

	// Just under the cap in bytes, but far over it in characters had the limit
	// been on characters: this must be accepted.
	ok := strings.Repeat("à", maxValueLength/2)
	if _, err := parseValue(t, ok); err != nil {
		t.Errorf("Parse rejected %d bytes (%d characters): %v",
			len(ok), len([]rune(ok)), err)
	}
}
