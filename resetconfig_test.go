// resetconfig_test.go: Reset lets the next Parse load the config file again
//
// Copyright (c) 2026 AGILira - A. Giordano
// SPDX-License-Identifier: MPL-2.0

package flashflags

import (
	"os"
	"path/filepath"
	"testing"
)

// Reset promises the state of a FlagSet never parsed, but it left the config
// file marked as loaded: the next Parse skipped it and every flag the command
// line did not set fell back to its default instead of the config value.
func TestResetReloadsConfig(t *testing.T) {
	path := filepath.Join(t.TempDir(), "app.json")
	if err := os.WriteFile(path, []byte(`{"host": "config-host", "port": 9090}`), 0o600); err != nil {
		t.Fatal(err)
	}
	fs := New("app")
	host := fs.String("host", "default", "")
	port := fs.Int("port", 8080, "")
	fs.SetConfigFile(path)
	if err := fs.Parse(nil); err != nil {
		t.Fatal(err)
	}

	fs.Reset()
	if err := fs.Parse([]string{"--port", "3000"}); err != nil {
		t.Fatal(err)
	}
	if *host != "config-host" || fs.Source("host") != "config" {
		t.Errorf("host = %q from %s, want config-host from config", *host, fs.Source("host"))
	}
	if *port != 3000 || fs.Source("port") != "cli" {
		t.Errorf("port = %d from %s, want 3000 from cli", *port, fs.Source("port"))
	}
}

// Reset also clears the operands of the previous Parse.
func TestResetClearsArgs(t *testing.T) {
	fs := New("app")
	if err := fs.Parse([]string{"file1"}); err != nil {
		t.Fatal(err)
	}
	fs.Reset()
	if fs.NArg() != 0 {
		t.Errorf("Args() = %v after Reset, want none", fs.Args())
	}
}
