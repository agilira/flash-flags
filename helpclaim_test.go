// helpclaim_test.go: --help/-h are conveniences, not reserved names.
//
// Copyright (c) 2025 AGILira - A. Giordano
// Series: an AGILira library
// SPDX-License-Identifier: MPL-2.0

package flashflags

import "testing"

// A flag registered under the short key "h" must be reachable with -h.
func TestShortHelpKeyYieldsToRegisteredFlag(t *testing.T) {
	fs := New("app")
	host := fs.StringVar("host", "h", "localhost", "server host")

	if err := fs.Parse([]string{"-h", "example.com"}); err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if *host != "example.com" {
		t.Fatalf("host = %q, want %q", *host, "example.com")
	}
	if got := fs.Source("host"); got != "cli" {
		t.Fatalf("source = %q, want cli", got)
	}
}

// Same for a flag named "help".
func TestLongHelpNameYieldsToRegisteredFlag(t *testing.T) {
	fs := New("app")
	topic := fs.String("help", "", "help topic to print")

	if err := fs.Parse([]string{"--help", "networking"}); err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if *topic != "networking" {
		t.Fatalf("help = %q, want %q", *topic, "networking")
	}
}

// The two spellings are judged independently.
func TestUnclaimedHelpSpellingStillPrintsHelp(t *testing.T) {
	fs := New("app")
	fs.StringVar("host", "h", "localhost", "server host") // claims -h, not --help

	if err := fs.Parse([]string{"--help"}); err == nil || err.Error() != "help requested" {
		t.Fatalf("--help: err = %v, want \"help requested\"", err)
	}

	fs2 := New("app")
	fs2.String("help", "", "help topic") // claims --help, not -h

	if err := fs2.Parse([]string{"-h"}); err == nil || err.Error() != "help requested" {
		t.Fatalf("-h: err = %v, want \"help requested\"", err)
	}
}

// With nothing registered, both spellings keep working.
func TestHelpFlagStillWorksWhenUnclaimed(t *testing.T) {
	for _, arg := range []string{"--help", "-h"} {
		fs := New("app")
		fs.String("host", "localhost", "server host")
		if err := fs.Parse([]string{arg}); err == nil || err.Error() != "help requested" {
			t.Fatalf("%s: err = %v, want \"help requested\"", arg, err)
		}
	}
}
