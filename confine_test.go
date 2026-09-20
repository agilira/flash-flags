// confine_test.go: path confinement for flag values.
//
// Two layers with different guarantees, and the tests say which is which:
// ConfinePath validates at Parse and is exposed to the gap between check and
// open; OpenConfined closes that gap by performing the open itself.
//
// Copyright (c) 2025 AGILira - A. Giordano
// SPDX-License-Identifier: MPL-2.0

package flashflags

import (
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// confineTree builds a base directory with a file inside it, a secret outside
// it, a symlink from inside to that secret, and a sibling sharing the base's
// name as a prefix.
func confineTree(t *testing.T) (base, secret string) {
	t.Helper()
	root := t.TempDir()
	base = filepath.Join(root, "app")
	if err := os.MkdirAll(filepath.Join(base, "sub"), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(base, "sub", "ok.json"),
		[]byte(`{"host":"inside"}`), 0o600); err != nil {
		t.Fatalf("write: %v", err)
	}
	secretDir := filepath.Join(root, "secret")
	if err := os.MkdirAll(secretDir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	secret = filepath.Join(secretDir, "key")
	if err := os.WriteFile(secret, []byte("SECRET"), 0o600); err != nil {
		t.Fatalf("write: %v", err)
	}
	if err := os.MkdirAll(base+"-evil", 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	return base, secret
}

func confinedSet(t *testing.T, base string) *FlagSet {
	t.Helper()
	fs := New("app")
	fs.String("path", "", "")
	if err := fs.ConfinePath("path", base); err != nil {
		t.Fatalf("ConfinePath: %v", err)
	}
	return fs
}

// TestConfinePath_Accepts covers what must keep working. The non-existent
// cases matter most: a flag naming a file to be created is ordinary, and an
// implementation built on filepath.EvalSymlinks alone rejects all of them,
// because EvalSymlinks fails on a path that is not there yet.
func TestConfinePath_Accepts(t *testing.T) {
	base, _ := confineTree(t)
	cases := map[string]string{
		"existing file":          filepath.Join(base, "sub", "ok.json"),
		"the base itself":        base,
		"non-existent file":      filepath.Join(base, "sub", "new.log"),
		"non-existent deep path": filepath.Join(base, "a", "b", "c.txt"),
		"interior dot-dot":       filepath.Join(base, "sub", "..", "sub", "ok.json"),
		"trailing separator":     base + string(filepath.Separator),
		"redundant separators":   base + "//sub//ok.json",
	}
	for name, path := range cases {
		t.Run(name, func(t *testing.T) {
			fs := confinedSet(t, base)
			if err := fs.Parse([]string{"--path", path}); err != nil {
				t.Errorf("Parse(%q) = %v, want nil", path, err)
			}
		})
	}
}

// TestConfinePath_Rejects covers the escapes. The sibling case is the one a
// naive strings.HasPrefix implementation gets wrong, and the symlink case is
// the one a pure string comparison cannot see at all.
func TestConfinePath_Rejects(t *testing.T) {
	base, secret := confineTree(t)
	link := filepath.Join(base, "escape")
	if err := os.Symlink(secret, link); err != nil {
		t.Skipf("symlink: %v", err)
	}

	cases := map[string]string{
		"explicit traversal":       filepath.Join(base, "..", "secret", "key"),
		"absolute path outside":    secret,
		"symlink pointing outside": link,
		"sibling sharing a prefix": base + "-evil" + string(filepath.Separator) + "x",
		"the parent directory":     filepath.Dir(base),
	}
	for name, path := range cases {
		t.Run(name, func(t *testing.T) {
			fs := confinedSet(t, base)
			err := fs.Parse([]string{"--path", path})
			if err == nil {
				t.Fatalf("Parse(%q) succeeded, want rejection", path)
			}
			if !strings.Contains(err.Error(), "outside") {
				t.Errorf("error = %v, want it to mention %q", err, "outside")
			}
		})
	}
}

// TestConfinePath_BaseReachedThroughSymlink pins the case that makes this
// usable on macOS, where /tmp is a symlink to /private/tmp. Resolving only the
// value and not the base would reject every path under such a base.
func TestConfinePath_BaseReachedThroughSymlink(t *testing.T) {
	base, _ := confineTree(t)
	linkedBase := filepath.Join(filepath.Dir(base), "link-to-app")
	if err := os.Symlink(base, linkedBase); err != nil {
		t.Skipf("symlink: %v", err)
	}

	// Base given through the link, value given directly.
	fs := New("app")
	fs.String("path", "", "")
	if err := fs.ConfinePath("path", linkedBase); err != nil {
		t.Fatalf("ConfinePath: %v", err)
	}
	if err := fs.Parse([]string{"--path", filepath.Join(base, "sub", "ok.json")}); err != nil {
		t.Errorf("Parse: %v", err)
	}

	// Base given directly, value given through the link.
	fs2 := confinedSet(t, base)
	if err := fs2.Parse([]string{"--path", filepath.Join(linkedBase, "sub", "ok.json")}); err != nil {
		t.Errorf("Parse: %v", err)
	}
}

// TestConfinePath_RelativeResolvesAgainstWorkingDirectory states the choice
// explicitly: a relative value is resolved the way the operating system will
// resolve it when the file is opened, not against the confinement base. Any
// other reading would make the check disagree with what actually gets opened.
func TestConfinePath_RelativeResolvesAgainstWorkingDirectory(t *testing.T) {
	base, _ := confineTree(t)

	// From an unrelated directory, a bare name is outside the base.
	t.Chdir(t.TempDir())
	fs := confinedSet(t, base)
	if err := fs.Parse([]string{"--path", "ok.json"}); err == nil {
		t.Error("a relative value resolved against the base, not the working directory")
	}

	// From inside the base, the same bare name is within it.
	t.Chdir(filepath.Join(base, "sub"))
	fs2 := confinedSet(t, base)
	if err := fs2.Parse([]string{"--path", "ok.json"}); err != nil {
		t.Errorf("Parse: %v", err)
	}
}

// TestConfinePath_EmptyValueIsSkipped keeps an unset optional path flag from
// failing: an empty value names nothing, and resolving it would silently mean
// the working directory.
func TestConfinePath_EmptyValueIsSkipped(t *testing.T) {
	base, _ := confineTree(t)
	fs := confinedSet(t, base)
	if err := fs.Parse([]string{}); err != nil {
		t.Errorf("Parse with an unset path flag: %v", err)
	}
}

// TestConfinePath_AppliesToEverySource checks confinement is a property of the
// flag, not of the command line.
func TestConfinePath_AppliesToEverySource(t *testing.T) {
	base, secret := confineTree(t)

	t.Run("environment", func(t *testing.T) {
		fs := New("app")
		fs.String("path", "", "")
		fs.SetEnvPrefix("APP")
		fs.EnableEnvLookup()
		if err := fs.ConfinePath("path", base); err != nil {
			t.Fatalf("ConfinePath: %v", err)
		}
		t.Setenv("APP_PATH", secret)
		if err := fs.Parse([]string{}); err == nil {
			t.Error("an escaping path from the environment was accepted")
		}
	})

	t.Run("config file", func(t *testing.T) {
		fs := New("app")
		fs.String("path", "", "")
		fs.SetConfigFile(writeConfig(t, `{"path":`+quoteJSON(secret)+`}`))
		if err := fs.ConfinePath("path", base); err != nil {
			t.Fatalf("ConfinePath: %v", err)
		}
		if err := fs.Parse([]string{}); err == nil {
			t.Error("an escaping path from the config file was accepted")
		}
	})
}

// TestConfinePath_StringSlice covers a flag carrying several paths: every
// element is confined, and one escape rejects the whole value.
func TestConfinePath_StringSlice(t *testing.T) {
	base, secret := confineTree(t)
	inside := filepath.Join(base, "sub", "ok.json")

	t.Run("all inside", func(t *testing.T) {
		fs := New("app")
		fs.StringSlice("paths", nil, "")
		if err := fs.ConfinePath("paths", base); err != nil {
			t.Fatalf("ConfinePath: %v", err)
		}
		if err := fs.Parse([]string{"--paths", inside + "," + base}); err != nil {
			t.Errorf("Parse: %v", err)
		}
	})

	t.Run("one escapes", func(t *testing.T) {
		fs := New("app")
		fs.StringSlice("paths", nil, "")
		if err := fs.ConfinePath("paths", base); err != nil {
			t.Fatalf("ConfinePath: %v", err)
		}
		err := fs.Parse([]string{"--paths", inside + "," + secret})
		if err == nil {
			t.Fatal("a slice with one escaping element was accepted")
		}
		if !strings.Contains(err.Error(), "outside") {
			t.Errorf("error = %v, want it to mention %q", err, "outside")
		}
	})
}

// TestConfinePath_Errors covers the setup-time failures.
func TestConfinePath_Errors(t *testing.T) {
	base, _ := confineTree(t)

	t.Run("unknown flag", func(t *testing.T) {
		fs := New("app")
		if err := fs.ConfinePath("nope", base); err == nil {
			t.Error("ConfinePath on an unknown flag returned nil")
		}
	})

	t.Run("wrong flag type", func(t *testing.T) {
		fs := New("app")
		fs.Int("port", 0, "")
		if err := fs.ConfinePath("port", base); err == nil {
			t.Error("ConfinePath on an int flag returned nil")
		}
	})

	t.Run("empty base", func(t *testing.T) {
		fs := New("app")
		fs.String("path", "", "")
		if err := fs.ConfinePath("path", ""); err == nil {
			t.Error("ConfinePath with an empty base returned nil")
		}
	})
}

// --- OpenConfined ------------------------------------------------------------

// TestOpenConfined_Reads is the layer that closes the gap: the open itself is
// confined, so nothing can redirect it between the check and the read.
func TestOpenConfined_Reads(t *testing.T) {
	base, _ := confineTree(t)
	fs := confinedSet(t, base)
	if err := fs.Parse([]string{"--path", filepath.Join(base, "sub", "ok.json")}); err != nil {
		t.Fatalf("Parse: %v", err)
	}

	f, err := fs.OpenConfined("path")
	if err != nil {
		t.Fatalf("OpenConfined: %v", err)
	}
	defer func() { _ = f.Close() }()

	data, err := io.ReadAll(f)
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	if string(data) != `{"host":"inside"}` {
		t.Errorf("content = %q", data)
	}
}

// TestOpenConfined_RefusesSymlinkPlantedAfterParse is the reason this method
// exists. The value passes confinement while it names a regular file, and the
// symlink appears afterwards -- exactly the window a validating check cannot
// close. The open must still refuse.
func TestOpenConfined_RefusesSymlinkPlantedAfterParse(t *testing.T) {
	base, secret := confineTree(t)
	target := filepath.Join(base, "swap.json")
	if err := os.WriteFile(target, []byte(`{"host":"inside"}`), 0o600); err != nil {
		t.Fatalf("write: %v", err)
	}

	fs := confinedSet(t, base)
	if err := fs.Parse([]string{"--path", target}); err != nil {
		t.Fatalf("Parse: %v", err)
	}

	// The attacker wins the race: the path is replaced by a link outside.
	if err := os.Remove(target); err != nil {
		t.Fatalf("remove: %v", err)
	}
	if err := os.Symlink(secret, target); err != nil {
		t.Skipf("symlink: %v", err)
	}

	f, err := fs.OpenConfined("path")
	if err == nil {
		data, _ := io.ReadAll(f)
		_ = f.Close()
		t.Fatalf("OpenConfined followed a symlink planted after Parse and read %q", data)
	}

	// And the contrast, which is the whole point of documenting two layers:
	// opening the value directly does follow it.
	direct, err := os.ReadFile(fs.GetString("path"))
	if err != nil {
		t.Fatalf("plain read: %v", err)
	}
	if string(direct) != "SECRET" {
		t.Fatalf("expected the plain read to be redirected, got %q", direct)
	}
}

// TestOpenConfined_RequiresConfinement refuses to guess: a flag with no base
// has nothing to be confined to.
func TestOpenConfined_RequiresConfinement(t *testing.T) {
	fs := New("app")
	fs.String("path", "", "")
	if err := fs.Parse([]string{"--path", "x"}); err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if _, err := fs.OpenConfined("path"); err == nil {
		t.Error("OpenConfined on an unconfined flag returned nil")
	}
	if _, err := fs.OpenConfined("missing"); err == nil {
		t.Error("OpenConfined on an unknown flag returned nil")
	}
}

// TestOpenConfined_MissingFile reports the open error rather than masking it.
func TestOpenConfined_MissingFile(t *testing.T) {
	base, _ := confineTree(t)
	fs := confinedSet(t, base)
	if err := fs.Parse([]string{"--path", filepath.Join(base, "absent.json")}); err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if _, err := fs.OpenConfined("path"); err == nil {
		t.Error("OpenConfined on a missing file returned nil")
	}
}

// quoteJSON renders s as a JSON string literal for the config fixtures.
func quoteJSON(s string) string {
	return `"` + strings.ReplaceAll(s, `\`, `\\`) + `"`
}

// TestConfinePath_EmptyValues covers the two shapes an empty path can take.
//
// A string flag set explicitly to "" reaches the check and is skipped: it names
// nothing, and resolving it would quietly mean the working directory, which is
// almost never inside the base and would turn an unset flag into a failure.
//
// A slice never carries one at all -- the parser drops empty elements, so
// "a,,b" arrives as two items. The test states that rather than leaving a
// reader to assume confinement handles it.
func TestConfinePath_EmptyValues(t *testing.T) {
	base, _ := confineTree(t)
	inside := filepath.Join(base, "sub", "ok.json")

	t.Run("explicit empty string", func(t *testing.T) {
		fs := confinedSet(t, base)
		if err := fs.Parse([]string{"--path", ""}); err != nil {
			t.Errorf("Parse with an explicitly empty path: %v", err)
		}
	})

	t.Run("empty slice element is dropped before the check", func(t *testing.T) {
		fs := New("app")
		fs.StringSlice("paths", nil, "")
		if err := fs.ConfinePath("paths", base); err != nil {
			t.Fatalf("ConfinePath: %v", err)
		}
		if err := fs.Parse([]string{"--paths", inside + ",," + base}); err != nil {
			t.Fatalf("Parse: %v", err)
		}
		if got := fs.GetStringSlice("paths"); len(got) != 2 {
			t.Errorf("paths = %q, want the empty element dropped", got)
		}
	})
}

// TestOpenConfined_Rejections covers the cases where OpenConfined declines
// before ever touching the filesystem, plus the one where the confinement
// directory itself cannot be opened.
func TestOpenConfined_Rejections(t *testing.T) {
	base, _ := confineTree(t)

	t.Run("slice flag", func(t *testing.T) {
		fs := New("app")
		fs.StringSlice("paths", nil, "")
		if err := fs.ConfinePath("paths", base); err != nil {
			t.Fatalf("ConfinePath: %v", err)
		}
		if err := fs.Parse([]string{"--paths", filepath.Join(base, "sub", "ok.json")}); err != nil {
			t.Fatalf("Parse: %v", err)
		}
		_, err := fs.OpenConfined("paths")
		if err == nil {
			t.Fatal("OpenConfined accepted a slice flag")
		}
		if !strings.Contains(err.Error(), "single path") {
			t.Errorf("error = %v, want it to explain it reads a single path", err)
		}
	})

	t.Run("unset value", func(t *testing.T) {
		fs := confinedSet(t, base)
		if err := fs.Parse([]string{}); err != nil {
			t.Fatalf("Parse: %v", err)
		}
		_, err := fs.OpenConfined("path")
		if err == nil {
			t.Fatal("OpenConfined accepted an unset flag")
		}
		if !strings.Contains(err.Error(), "no value") {
			t.Errorf("error = %v, want it to mention the missing value", err)
		}
	})

	t.Run("confinement directory missing", func(t *testing.T) {
		absent := filepath.Join(t.TempDir(), "not-there")
		fs := New("app")
		fs.String("path", "", "")
		if err := fs.ConfinePath("path", absent); err != nil {
			t.Fatalf("ConfinePath: %v", err)
		}
		// The value resolves under the (absent) base, so the check passes and
		// the failure surfaces when the root is opened.
		if err := fs.Parse([]string{"--path", filepath.Join(absent, "x.json")}); err != nil {
			t.Fatalf("Parse: %v", err)
		}
		_, err := fs.OpenConfined("path")
		if err == nil {
			t.Fatal("OpenConfined accepted a missing confinement directory")
		}
		if !strings.Contains(err.Error(), "confinement directory") {
			t.Errorf("error = %v, want it to name the confinement directory", err)
		}
	})
}

// TestConfinePath_UnresolvableWorkingDirectory covers what happens when the
// process cannot resolve a relative path at all, because its working directory
// has been removed underneath it -- a daemon whose cwd is deleted, most
// plausibly.
//
// The point is not the error text but the direction: confinement refuses. The
// tempting alternative, falling back to the path as written, would accept a
// value the library cannot verify, which is worse than rejecting it.
func TestConfinePath_UnresolvableWorkingDirectory(t *testing.T) {
	brokenCwd := func(t *testing.T) {
		t.Helper()
		dead := filepath.Join(t.TempDir(), "dead")
		if err := os.MkdirAll(dead, 0o755); err != nil {
			t.Fatalf("mkdir: %v", err)
		}
		t.Chdir(dead)
		if err := os.RemoveAll(dead); err != nil {
			t.Skipf("cannot remove the working directory: %v", err)
		}
		if _, err := os.Getwd(); err == nil {
			t.Skip("this platform still resolves a deleted working directory")
		}
	}

	t.Run("relative value", func(t *testing.T) {
		base, _ := confineTree(t)
		brokenCwd(t)
		if err := confinePathValue("path", base, "relative.json"); err == nil {
			t.Error("a value that cannot be resolved was accepted")
		}
	})

	t.Run("relative base", func(t *testing.T) {
		brokenCwd(t)
		if err := confinePathValue("path", "relative-base", "/etc/passwd"); err == nil {
			t.Error("a base that cannot be resolved was accepted")
		}
	})

	t.Run("OpenConfined", func(t *testing.T) {
		fs := New("app")
		fs.String("path", "", "")
		if err := fs.ConfinePath("path", "relative-base"); err != nil {
			t.Fatalf("ConfinePath: %v", err)
		}
		fs.Lookup("path").value = "relative.json"
		brokenCwd(t)
		if _, err := fs.OpenConfined("path"); err == nil {
			t.Error("OpenConfined proceeded with an unresolvable path")
		}
	})
}
