// configpath_test.go: configuration file path handling.
//
// The config file path is chosen by the application, never by parsed arguments:
// LoadConfig runs before parseArguments, so no --config flag can reach it. The
// parser therefore does not second-guess the path; it checks that the target is
// something it can sensibly read.
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

// loadFrom writes a config at path and reports what Parse made of it.
func loadFrom(t *testing.T, path string) (string, error) {
	t.Helper()
	fs := New("app")
	fs.String("host", "default", "")
	fs.SetConfigFile(path)
	return fs.GetString("host"), fs.Parse([]string{})
}

// TestConfigPath_AcceptsOrdinaryLocations covers the directories a real program
// keeps its configuration in. Before v1.1.9 an absolute path outside a fixed
// five-prefix allowlist was rejected, which included the user's own home -- the
// location AddConfigPath's own documentation uses as its example.
func TestConfigPath_AcceptsOrdinaryLocations(t *testing.T) {
	home, err := os.UserHomeDir()
	if err != nil {
		t.Skipf("no home directory: %v", err)
	}

	dirs := map[string]string{
		"home subdirectory": filepath.Join(home, ".flashflags-test"),
		"dot in dir name":   filepath.Join(t.TempDir(), "v1..2"),
		"nested temp":       filepath.Join(t.TempDir(), "a", "b"),
	}
	for name, dir := range dirs {
		t.Run(name, func(t *testing.T) {
			if err := os.MkdirAll(dir, 0o755); err != nil {
				t.Fatalf("mkdir: %v", err)
			}
			defer os.RemoveAll(dir)
			path := filepath.Join(dir, "app.json")
			if err := os.WriteFile(path, []byte(`{"host":"loaded"}`), 0o600); err != nil {
				t.Fatalf("write: %v", err)
			}

			fs := New("app")
			fs.String("host", "default", "")
			fs.SetConfigFile(path)
			if err := fs.Parse([]string{}); err != nil {
				t.Fatalf("Parse(%s): %v", path, err)
			}
			if got := fs.GetString("host"); got != "loaded" {
				t.Errorf("host = %q, want loaded", got)
			}
		})
	}
}

// TestConfigPath_AcceptsRelativeAndTraversal documents that a path the program
// composed itself is honoured even when it walks upwards. The application chose
// it; rejecting it on the spelling of the string protected nothing, since the
// equivalent absolute path was accepted.
func TestConfigPath_AcceptsRelativeAndTraversal(t *testing.T) {
	base := t.TempDir()
	if err := os.MkdirAll(filepath.Join(base, "sub"), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(base, "app.json"), []byte(`{"host":"loaded"}`), 0o600); err != nil {
		t.Fatalf("write: %v", err)
	}

	path := filepath.Join(base, "sub", "..", "app.json")
	fs := New("app")
	fs.String("host", "default", "")
	fs.SetConfigFile(path)
	if err := fs.Parse([]string{}); err != nil {
		t.Fatalf("Parse(%s): %v", path, err)
	}
	if got := fs.GetString("host"); got != "loaded" {
		t.Errorf("host = %q, want loaded", got)
	}
}

// TestConfigPath_RejectsNonRegularFiles covers what the check is actually for:
// reading a device or a FIFO as configuration hangs or exhausts memory, and a
// directory is never a config file.
func TestConfigPath_RejectsNonRegularFiles(t *testing.T) {
	t.Run("directory", func(t *testing.T) {
		_, err := loadFrom(t, t.TempDir())
		if err == nil {
			t.Fatal("Parse accepted a directory as a config file")
		}
		if !strings.Contains(err.Error(), "not a regular file") {
			t.Errorf("error = %v, want it to mention %q", err, "not a regular file")
		}
	})

	t.Run("fifo", func(t *testing.T) {
		if runtime.GOOS == "windows" {
			t.Skip("no FIFOs on Windows")
		}
		fifo := filepath.Join(t.TempDir(), "app.json")
		if err := mkfifo(fifo); err != nil {
			t.Skipf("mkfifo unavailable: %v", err)
		}
		_, err := loadFrom(t, fifo)
		if err == nil {
			t.Fatal("Parse accepted a FIFO as a config file")
		}
		if !strings.Contains(err.Error(), "not a regular file") {
			t.Errorf("error = %v, want it to mention %q", err, "not a regular file")
		}
	})

	t.Run("device", func(t *testing.T) {
		if runtime.GOOS == "windows" {
			t.Skip("no /dev/zero on Windows")
		}
		if _, err := os.Stat("/dev/zero"); err != nil {
			t.Skip("/dev/zero unavailable")
		}
		fs := New("app")
		fs.String("host", "default", "")
		fs.SetConfigFile("/dev/zero")
		if err := fs.Parse([]string{}); err == nil {
			t.Fatal("Parse accepted /dev/zero as a config file")
		}
	})
}

// TestConfigPath_SymlinkIsFollowed pins current behaviour explicitly rather than
// leaving it implied: a symlinked config file is read, exactly as it was before
// v1.1.9 when the link lived under an allowlisted prefix. An application that
// reads config from a world-writable directory must not rely on this package to
// notice a link planted there.
func TestConfigPath_SymlinkIsFollowed(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("symlink creation needs privileges on Windows")
	}
	target := filepath.Join(t.TempDir(), "real.json")
	if err := os.WriteFile(target, []byte(`{"host":"via-symlink"}`), 0o600); err != nil {
		t.Fatalf("write: %v", err)
	}
	link := filepath.Join(t.TempDir(), "app.json")
	if err := os.Symlink(target, link); err != nil {
		t.Skipf("symlink: %v", err)
	}

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

// TestConfigPath_MissingFile keeps the existing not-found behaviour covered:
// an explicit config file that does not exist is silently skipped by
// findConfigFile, so Parse succeeds with defaults.
func TestConfigPath_MissingFile(t *testing.T) {
	fs := New("app")
	fs.String("host", "default", "")
	fs.SetConfigFile(filepath.Join(t.TempDir(), "absent.json"))
	if err := fs.Parse([]string{}); err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if got := fs.GetString("host"); got != "default" {
		t.Errorf("host = %q, want default", got)
	}
}

// TestConfigPath_MalformedJSON keeps the parse-error path covered.
func TestConfigPath_MalformedJSON(t *testing.T) {
	_, err := loadFrom(t, writeConfig(t, `{"host": `))
	if err == nil {
		t.Fatal("Parse accepted malformed JSON")
	}
}

// --- Auto-discovery ---------------------------------------------------------

// TestLoadConfig_RequiresExplicitConfiguration pins the real contract: nothing
// is discovered unless the application asked for it with SetConfigFile or
// AddConfigPath. LoadConfig returns early otherwise, so a config file sitting
// in the working directory or the home directory is never picked up by
// accident -- which is what makes flag values predictable for a program that
// never opted into configuration files at all.
func TestLoadConfig_RequiresExplicitConfiguration(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)
	if err := os.WriteFile(filepath.Join(home, "app.json"),
		[]byte(`{"host":"from-home"}`), 0o600); err != nil {
		t.Fatalf("write: %v", err)
	}

	cwd := t.TempDir()
	if err := os.WriteFile(filepath.Join(cwd, "app.json"),
		[]byte(`{"host":"from-cwd"}`), 0o600); err != nil {
		t.Fatalf("write: %v", err)
	}
	t.Chdir(cwd)

	fs := New("app")
	fs.String("host", "default", "")
	if err := fs.Parse([]string{}); err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if got := fs.GetString("host"); got != "default" {
		t.Errorf("host = %q, want default: a program that configured no config "+
			"source must not load one", got)
	}
}

// TestAddConfigPath_Home covers the portable way to read config from the user's
// home directory: pass it explicitly. os.UserHomeDir is the right call because
// HOME is normally unset on Windows, where the home lives in USERPROFILE.
func TestAddConfigPath_Home(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)
	if err := os.WriteFile(filepath.Join(home, "app.json"),
		[]byte(`{"host":"from-home"}`), 0o600); err != nil {
		t.Fatalf("write: %v", err)
	}

	resolved, err := os.UserHomeDir()
	if err != nil {
		t.Skipf("no home directory: %v", err)
	}

	fs := New("app")
	fs.String("host", "default", "")
	fs.AddConfigPath(resolved)
	if err := fs.Parse([]string{}); err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if got := fs.GetString("host"); got != "from-home" {
		t.Errorf("host = %q, want from-home", got)
	}
}

// TestAddConfigPath_SearchOrder checks that added paths are searched in the
// order they were added, and that each is probed for every candidate filename.
func TestAddConfigPath_SearchOrder(t *testing.T) {
	first, second := t.TempDir(), t.TempDir()
	if err := os.WriteFile(filepath.Join(second, "app.json"),
		[]byte(`{"host":"second"}`), 0o600); err != nil {
		t.Fatalf("write: %v", err)
	}

	fs := New("app")
	fs.String("host", "default", "")
	fs.AddConfigPath(first) // empty, must fall through
	fs.AddConfigPath(second)
	if err := fs.Parse([]string{}); err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if got := fs.GetString("host"); got != "second" {
		t.Errorf("host = %q, want second", got)
	}

	// An earlier path wins once it holds a match.
	if err := os.WriteFile(filepath.Join(first, "config.json"),
		[]byte(`{"host":"first"}`), 0o600); err != nil {
		t.Fatalf("write: %v", err)
	}
	fs2 := New("app")
	fs2.String("host", "default", "")
	fs2.AddConfigPath(first)
	fs2.AddConfigPath(second)
	if err := fs2.Parse([]string{}); err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if got := fs2.GetString("host"); got != "first" {
		t.Errorf("host = %q, want first", got)
	}
}
