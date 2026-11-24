// helporder_test.go: help lists flags and groups in declaration order
//
// Copyright (c) 2026 AGILira - A. Giordano
// SPDX-License-Identifier: MPL-2.0

package flashflags

import (
	"strings"
	"testing"
)

// Help walked the flag map, so flags and groups came out in a different
// order on every run. They now follow the declarations: ungrouped flags
// first, then each group where its first flag was declared.
func TestHelpFollowsDeclarationOrder(t *testing.T) {
	fs := New("app")
	fs.String("zeta", "", "")
	fs.String("alpha", "", "")
	fs.String("port", "", "")
	fs.String("host", "", "")
	fs.String("verbose", "", "")
	fs.String("timeout", "", "")
	_ = fs.SetGroup("port", "Network")
	_ = fs.SetGroup("verbose", "Runtime")
	_ = fs.SetGroup("host", "Network")
	_ = fs.SetGroup("timeout", "Runtime")

	want := []string{"Options:", "--zeta", "--alpha", "Network:", "--port", "--host", "Runtime:", "--verbose", "--timeout"}
	for i := 0; i < 20; i++ {
		help := fs.Help()
		pos := -1
		for _, w := range want {
			next := strings.Index(help[pos+1:], w)
			if next < 0 {
				t.Fatalf("help lacks %q after position %d:\n%s", w, pos, help)
			}
			pos += 1 + next
		}
	}
}

// A flag declared twice keeps its first position and the latest definition.
func TestRedeclaredFlagKeepsPosition(t *testing.T) {
	fs := New("app")
	fs.String("a", "", "")
	fs.String("b", "", "first")
	fs.String("c", "", "")
	fs.String("b", "", "second")
	var names []string
	fs.VisitAll(func(f *Flag) { names = append(names, f.Name()+":"+f.Usage()) })
	if got := strings.Join(names, " "); got != "a: b:second c:" {
		t.Errorf("VisitAll order = %q, want %q", got, "a: b:second c:")
	}
}
