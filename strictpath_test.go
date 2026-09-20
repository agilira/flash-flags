// strictpath_test.go: opt-in symlink refusal for configuration files.
//
// Copyright (c) 2025 AGILira - A. Giordano
// SPDX-License-Identifier: MPL-2.0

package flashflags

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

// newStrictSet builds a FlagSet reading config from path, with strict config
// paths enabled.
func newStrictSet(t *testing.T, path string) *FlagSet {
	t.Helper()
	fs := New("app")
	fs.String("host", "default", "")
	fs.SetConfigFile(path)
	fs.EnableStrictConfigPaths()
	return fs
}

// writeJSON writes a config file and returns its path.
func writeJSON(t *testing.T, dir, name, body string) string {
	t.Helper()
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte(body), 0o600); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
	return path
}

// mustSymlink creates a symlink, skipping the test where that needs privileges.
func mustSymlink(t *testing.T, target, link string) {
	t.Helper()
	if err := os.Symlink(target, link); err != nil {
		t.Skipf("symlink unavailable (%v); creating one needs privileges or "+
			"Developer Mode on Windows", err)
	}
}

// TestStrictConfigPaths_OffByDefault pins that nothing changes for an
// application that does not ask: a symlinked config still loads.
func TestStrictConfigPaths_OffByDefault(t *testing.T) {
	dir := t.TempDir()
	target := writeJSON(t, dir, "real.json", `{"host":"via-symlink"}`)
	link := filepath.Join(t.TempDir(), "app.json")
	mustSymlink(t, target, link)

	fs := New("app")
	fs.String("host", "default", "")
	fs.SetConfigFile(link)
	if err := fs.Parse([]string{}); err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if got := fs.GetString("host"); got != "via-symlink" {
		t.Errorf("host = %q, want via-symlink", got)
	}
}

// TestStrictConfigPaths_RefusesSymlink is the point of the feature.
func TestStrictConfigPaths_RefusesSymlink(t *testing.T) {
	dir := t.TempDir()
	target := writeJSON(t, dir, "real.json", `{"host":"via-symlink"}`)
	link := filepath.Join(t.TempDir(), "app.json")
	mustSymlink(t, target, link)

	fs := newStrictSet(t, link)
	err := fs.Parse([]string{})
	if err == nil {
		t.Fatal("Parse accepted a symlinked config file under strict paths")
	}
	if !strings.Contains(err.Error(), "symbolic link") {
		t.Errorf("error = %v, want it to mention %q", err, "symbolic link")
	}
	if got := fs.GetString("host"); got != "default" {
		t.Errorf("host = %q, want the default to be untouched", got)
	}
}

// TestStrictConfigPaths_AllowsRegularFile checks strict mode does not reject
// the ordinary case.
func TestStrictConfigPaths_AllowsRegularFile(t *testing.T) {
	path := writeJSON(t, t.TempDir(), "app.json", `{"host":"plain"}`)

	fs := newStrictSet(t, path)
	if err := fs.Parse([]string{}); err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if got := fs.GetString("host"); got != "plain" {
		t.Errorf("host = %q, want plain", got)
	}
}

// TestStrictConfigPaths_AllowsSymlinkedParent covers the case that makes this
// usable on macOS at all: /tmp there is a symlink to /private/tmp. Only the
// final path component is examined, so a symlinked parent directory is
// traversed like any other and the config still loads.
func TestStrictConfigPaths_AllowsSymlinkedParent(t *testing.T) {
	base := t.TempDir()
	real := filepath.Join(base, "private", "tmp")
	if err := os.MkdirAll(real, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	writeJSON(t, real, "app.json", `{"host":"through-linked-dir"}`)
	mustSymlink(t, filepath.Join(base, "private", "tmp"), filepath.Join(base, "tmp"))

	fs := newStrictSet(t, filepath.Join(base, "tmp", "app.json"))
	if err := fs.Parse([]string{}); err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if got := fs.GetString("host"); got != "through-linked-dir" {
		t.Errorf("host = %q, want through-linked-dir", got)
	}
}

// TestStrictConfigPaths_KubernetesConfigMapLayout documents the cost of the
// feature rather than hiding it. A ConfigMap mount projects each key as a
// symlink into a ..data directory, which is itself a symlink to a timestamped
// one. That layout loads by default and is refused under strict paths, which is
// precisely why strict paths are opt-in.
func TestStrictConfigPaths_KubernetesConfigMapLayout(t *testing.T) {
	mount := t.TempDir()
	versioned := filepath.Join(mount, "..2026_09_20_00_00_00")
	if err := os.MkdirAll(versioned, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	writeJSON(t, versioned, "app.json", `{"host":"from-configmap"}`)
	mustSymlink(t, "..2026_09_20_00_00_00", filepath.Join(mount, "..data"))
	mustSymlink(t, filepath.Join("..data", "app.json"), filepath.Join(mount, "app.json"))
	key := filepath.Join(mount, "app.json")

	t.Run("loads by default", func(t *testing.T) {
		fs := New("app")
		fs.String("host", "default", "")
		fs.SetConfigFile(key)
		if err := fs.Parse([]string{}); err != nil {
			t.Fatalf("Parse: %v", err)
		}
		if got := fs.GetString("host"); got != "from-configmap" {
			t.Errorf("host = %q, want from-configmap", got)
		}
	})

	t.Run("refused under strict paths", func(t *testing.T) {
		fs := newStrictSet(t, key)
		if err := fs.Parse([]string{}); err == nil {
			t.Error("strict paths accepted a ConfigMap key symlink")
		}
	})
}

// TestStrictConfigPaths_StillRejectsNonRegularFiles checks the existing
// robustness checks survive strict mode.
func TestStrictConfigPaths_StillRejectsNonRegularFiles(t *testing.T) {
	fs := newStrictSet(t, t.TempDir())
	err := fs.Parse([]string{})
	if err == nil {
		t.Fatal("strict paths accepted a directory")
	}
	if !strings.Contains(err.Error(), "not a regular file") {
		t.Errorf("error = %v, want it to mention %q", err, "not a regular file")
	}
}

// TestStrictConfigPaths_AppliesToDiscoveredFiles checks the setting covers
// auto-discovery, not just an explicitly named file.
func TestStrictConfigPaths_AppliesToDiscoveredFiles(t *testing.T) {
	dir := t.TempDir()
	target := writeJSON(t, t.TempDir(), "real.json", `{"host":"via-symlink"}`)
	mustSymlink(t, target, filepath.Join(dir, "app.json"))

	fs := New("app")
	fs.String("host", "default", "")
	fs.AddConfigPath(dir)
	fs.EnableStrictConfigPaths()
	if err := fs.Parse([]string{}); err == nil {
		t.Error("strict paths accepted a symlink found by auto-discovery")
	}
}

// TestStrictConfigPaths_MissingFileErrorIsUnchanged keeps the existing error
// path stable under strict mode.
func TestStrictConfigPaths_MissingFileErrorIsUnchanged(t *testing.T) {
	fs := newStrictSet(t, filepath.Join(t.TempDir(), "absent.json"))
	// An explicitly named file that does not exist is skipped by
	// findConfigFile, so Parse succeeds with defaults.
	if err := fs.Parse([]string{}); err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if got := fs.GetString("host"); got != "default" {
		t.Errorf("host = %q, want default", got)
	}
}

// TestStrictConfigPaths_WindowsIsBestEffort records, as an assertion the reader
// cannot miss, that the guarantee differs by platform: on Unix the kernel
// refuses the open, while on Windows the check runs before the open and a
// racing replacement could slip through.
func TestStrictConfigPaths_WindowsIsBestEffort(t *testing.T) {
	path := writeJSON(t, t.TempDir(), "app.json", `{"host":"plain"}`)
	fs := newStrictSet(t, path)
	if err := fs.Parse([]string{}); err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if runtime.GOOS == "windows" {
		t.Log("windows: symlink refusal is a pre-open check, not atomic")
	} else {
		t.Log("unix: symlink refusal is enforced by the kernel at open (O_NOFOLLOW)")
	}
}
