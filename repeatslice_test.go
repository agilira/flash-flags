// repeatslice_test.go: repeated string slice flags accumulate
//
// Copyright (c) 2026 AGILira - A. Giordano
// SPDX-License-Identifier: MPL-2.0

package flashflags

import (
	"reflect"
	"testing"
)

func TestRepeatedStringSliceFlagAccumulates(t *testing.T) {
	fs := New("test")
	fs.StringSliceVar("mount", "m", []string{"default"}, "mounts")
	if err := fs.Parse([]string{"--mount", "a", "-m", "b,c", "--mount=d"}); err != nil {
		t.Fatal(err)
	}
	if got := fs.GetStringSlice("mount"); !reflect.DeepEqual(got, []string{"a", "b", "c", "d"}) {
		t.Fatalf("got %v", got)
	}
}

func TestSingleStringSliceFlagReplacesDefault(t *testing.T) {
	fs := New("test")
	ptr := fs.StringSliceVar("tag", "", []string{"x", "y"}, "tags")
	if err := fs.Parse([]string{"--tag", "z"}); err != nil {
		t.Fatal(err)
	}
	if got := fs.GetStringSlice("tag"); !reflect.DeepEqual(got, []string{"z"}) || !reflect.DeepEqual(*ptr, []string{"z"}) {
		t.Fatalf("got %v ptr %v", got, *ptr)
	}
}

func TestStringSliceAccumulationRestartsAfterReset(t *testing.T) {
	fs := New("test")
	fs.StringSlice("tag", nil, "tags")
	if err := fs.Parse([]string{"--tag", "a"}); err != nil {
		t.Fatal(err)
	}
	fs.Reset()
	if err := fs.Parse([]string{"--tag", "b"}); err != nil {
		t.Fatal(err)
	}
	if got := fs.GetStringSlice("tag"); !reflect.DeepEqual(got, []string{"b"}) {
		t.Fatalf("got %v", got)
	}
}

func TestNewParseStartsOverWithoutReset(t *testing.T) {
	fs := New("test")
	fs.StringSlice("tag", nil, "tags")
	for _, arg := range []string{"a", "b"} {
		if err := fs.Parse([]string{"--tag", arg}); err != nil {
			t.Fatal(err)
		}
	}
	if got := fs.GetStringSlice("tag"); !reflect.DeepEqual(got, []string{"b"}) {
		t.Fatalf("got %v", got)
	}
}
