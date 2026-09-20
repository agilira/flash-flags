// shorthandvars_test.go: short-key variants for the flag types that lacked them.
//
// Copyright (c) 2025 AGILira - A. Giordano
// Series: an AGILira library
// SPDX-License-Identifier: MPL-2.0

package flashflags

import (
	"testing"
	"time"
)

func TestFloat64VarShortKey(t *testing.T) {
	fs := New("app")
	ratio := fs.Float64Var("ratio", "r", 1.5, "sampling ratio")

	if err := fs.Parse([]string{"-r", "0.25"}); err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if *ratio != 0.25 {
		t.Fatalf("ratio = %v, want 0.25", *ratio)
	}
	if got := fs.Lookup("ratio").ShortKey(); got != "r" {
		t.Fatalf("ShortKey = %q, want %q", got, "r")
	}
}

func TestStringSliceVarShortKey(t *testing.T) {
	fs := New("app")
	tags := fs.StringSliceVar("tags", "t", []string{"default"}, "service tags")

	if err := fs.Parse([]string{"-t", "web,api"}); err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if len(*tags) != 2 || (*tags)[0] != "web" || (*tags)[1] != "api" {
		t.Fatalf("tags = %q, want [web api]", *tags)
	}
}

func TestDurationVarShortKey(t *testing.T) {
	fs := New("app")
	timeout := fs.DurationVar("timeout", "T", 30*time.Second, "request timeout")

	if err := fs.Parse([]string{"-T", "1m30s"}); err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if *timeout != 90*time.Second {
		t.Fatalf("timeout = %v, want 1m30s", *timeout)
	}
}

// An empty short key registers the long form only, matching StringVar.
func TestVarsWithEmptyShortKeyRegisterLongFormOnly(t *testing.T) {
	fs := New("app")
	ratio := fs.Float64Var("ratio", "", 1.0, "sampling ratio")
	tags := fs.StringSliceVar("tags", "", nil, "service tags")
	timeout := fs.DurationVar("timeout", "", time.Second, "request timeout")

	if err := fs.Parse([]string{"--ratio=0.5", "--tags=a,b", "--timeout=2s"}); err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if *ratio != 0.5 || len(*tags) != 2 || *timeout != 2*time.Second {
		t.Fatalf("got ratio=%v tags=%q timeout=%v", *ratio, *tags, *timeout)
	}
	for _, name := range []string{"ratio", "tags", "timeout"} {
		if got := fs.Lookup(name).ShortKey(); got != "" {
			t.Fatalf("%s: ShortKey = %q, want empty", name, got)
		}
	}
}

// The defaults these return must not alias the caller's slice.
func TestStringSliceVarDefaultIsCopied(t *testing.T) {
	original := []string{"a", "b"}
	fs := New("app")
	tags := fs.StringSliceVar("tags", "t", original, "service tags")

	(*tags)[0] = "mutated"
	if original[0] != "a" {
		t.Fatalf("caller's slice was aliased: original = %q", original)
	}
}
