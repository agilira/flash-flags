// parsegen_test.go: a Parse allocates nothing of its own, and a repeated
// slice flag accumulates within one command line
//
// Copyright (c) 2026 AGILira - A. Giordano
// SPDX-License-Identifier: MPL-2.0

package flashflags

import (
	"reflect"
	"testing"
)

// Parse used to allocate a map on every call to remember the slice flags
// given; a bool value is stored without boxing, so a command line of bool
// flags must now cost nothing.
func TestParseAllocatesNothingForBools(t *testing.T) {
	fs := New("app")
	fs.Bool("verbose", false, "")
	fs.Bool("dry-run", false, "")
	args := []string{"--verbose", "--dry-run"}
	if n := testing.AllocsPerRun(100, func() { _ = fs.Parse(args) }); n != 0 {
		t.Errorf("Parse allocated %v times, want 0", n)
	}
}

// A repetition within one command line appends; the same flag in the next
// Parse starts over.
func TestRepeatedSliceAccumulatesPerCommandLine(t *testing.T) {
	fs := New("app")
	tags := fs.StringSlice("tag", nil, "")
	if err := fs.Parse([]string{"--tag", "a", "--tag", "b"}); err != nil {
		t.Fatal(err)
	}
	if err := fs.Parse([]string{"--tag", "c"}); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(*tags, []string{"c"}) {
		t.Errorf("tag = %v, want [c]", *tags)
	}
}
