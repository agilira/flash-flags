// lonedash_test.go: a lone "-" is an operand
//
// Copyright (c) 2026 AGILira - A. Giordano
// SPDX-License-Identifier: MPL-2.0

package flashflags

import (
	"reflect"
	"testing"
)

// A lone "-" (standard input, by convention) was dropped: it starts with
// '-' but is neither a short nor a long flag.
func TestLoneDashIsAnOperand(t *testing.T) {
	for name, c := range map[string]struct {
		args []string
		want []string
	}{
		"alone":            {[]string{"-"}, []string{"-"}},
		"after a flag":     {[]string{"--when", "daily", "-"}, []string{"-"}},
		"among operands":   {[]string{"job", "-", "more"}, []string{"job", "-", "more"}},
		"after --":         {[]string{"--", "-"}, []string{"-"}},
		"before a flag":    {[]string{"-", "--when", "daily"}, []string{"-"}},
		"short flag value": {[]string{"-w", "daily", "-"}, []string{"-"}},
	} {
		fs := New("t")
		when := fs.StringVar("when", "w", "", "when")
		if err := fs.Parse(c.args); err != nil {
			t.Fatalf("%s: %v", name, err)
		}
		if !reflect.DeepEqual(fs.Args(), c.want) {
			t.Errorf("%s: args %q, want %q", name, fs.Args(), c.want)
		}
		if len(c.args) > 1 && c.args[0] != "--" && c.args[0] != "job" && *when != "daily" {
			t.Errorf("%s: --when %q", name, *when)
		}
	}
}

// A value flag still takes "-" as its value.
func TestDashAsAFlagValue(t *testing.T) {
	fs := New("t")
	file := fs.String("file", "", "file")
	if err := fs.Parse([]string{"--file", "-", "x"}); err != nil {
		t.Fatal(err)
	}
	if *file != "-" || !reflect.DeepEqual(fs.Args(), []string{"x"}) {
		t.Fatalf("file %q args %q", *file, fs.Args())
	}
}
