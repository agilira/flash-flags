// confine.go: confining a path-valued flag to a directory.
//
// Copyright (c) 2025 AGILira - A. Giordano
// Series: an AGILira library
// SPDX-License-Identifier: MPL-2.0

package flashflags

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// ConfinePath restricts a string or string-slice flag to paths inside dir.
// Parse then rejects a value that resolves outside it, whichever source
// supplied the value.
//
// WHY this is not the denylist this package removed in v1.1.9: that one tried
// to infer from a value's contents whether it was dangerous, which a parser
// cannot know. Here the application states the rule -- this flag names a path,
// and it belongs under this directory -- and the library enforces the rule it
// was given. Containment is also a well-defined operation that is easy to get
// subtly wrong by hand, which is what makes it worth providing.
//
// What the check does:
//
//   - Resolves both the base and the value, following symlinks, so a link
//     inside the base pointing outside it is caught. Both are resolved, not
//     just the value: on macOS /tmp is a symlink to /private/tmp, and resolving
//     only one side would reject every path under such a base.
//   - Handles a value that does not exist yet, by resolving its deepest
//     existing ancestor and re-appending the rest. A flag naming a file to be
//     created is ordinary, and filepath.EvalSymlinks alone fails on it.
//   - Compares whole path elements, so a sibling directory sharing the base's
//     name as a prefix does not pass. That is the mistake a strings.HasPrefix
//     comparison makes.
//
// A relative value is resolved against the working directory, the way the
// operating system will resolve it when the file is opened -- not against dir.
// Resolving it against dir would make the check disagree with what actually
// gets opened, which is worse than not checking. An empty value is skipped: it
// names nothing, and resolving it would quietly mean the working directory.
//
// # The gap this cannot close
//
// The check answers for the moment it runs. Between Parse and the moment the
// application opens the path, the filesystem can change, and a symlink planted
// in that window is followed. This stops mistakes, typos and plain traversal;
// it does not stop a local attacker racing the process. Closing that window
// requires performing the open under the same constraint, which is what
// OpenConfined does.
//
// Example:
//
//	fs := flashflags.New("myapp")
//	fs.String("config", "", "Config file")
//	if err := fs.ConfinePath("config", "/etc/myapp"); err != nil {
//		log.Fatal(err)
//	}
//
//	// Parse now rejects --config /etc/shadow, --config ../../etc/shadow,
//	// and a symlink under /etc/myapp pointing anywhere else.
//
// Returns an error if the flag does not exist, is not a string or string
// slice, or dir is empty.
func (fs *FlagSet) ConfinePath(name, dir string) error {
	flag := fs.Lookup(name)
	if flag == nil {
		return fmt.Errorf("flag not found: %s", name)
	}
	if flag.flagType != "string" && flag.flagType != "stringSlice" {
		return fmt.Errorf("flag --%s is a %s flag; only string and stringSlice flags can be confined to a path",
			name, flag.flagType)
	}
	if dir == "" {
		return fmt.Errorf("flag --%s: confinement directory must not be empty", name)
	}
	flag.confineTo = dir
	return nil
}

// checkConfinement verifies a flag's current value against its confinement
// base, if one was set. It is called for every source, so the guarantee does
// not depend on where a value came from.
func (fs *FlagSet) checkConfinement(flag *Flag, name string) error {
	if flag.confineTo == "" {
		return nil
	}
	switch v := flag.value.(type) {
	case string:
		return confinePathValue(name, flag.confineTo, v)
	case []string:
		for _, item := range v {
			if err := confinePathValue(name, flag.confineTo, item); err != nil {
				return err
			}
		}
	}
	return nil
}

// confinePathValue reports whether value resolves inside base.
func confinePathValue(name, base, value string) error {
	if value == "" {
		return nil
	}
	resolvedBase, err := resolvePath(base)
	if err != nil {
		return fmt.Errorf("flag --%s: cannot resolve confinement directory %s: %v", name, base, err)
	}
	resolvedValue, err := resolvePath(value)
	if err != nil {
		return fmt.Errorf("flag --%s: cannot resolve %s: %v", name, value, err)
	}

	// filepath.Rel already compares whole elements, so a sibling such as
	// "<base>-evil" comes back as "../<base>-evil" rather than as a child. A
	// Rel that fails outright -- different Windows volumes, for instance --
	// means the value cannot be under the base either, so both answers are the
	// same refusal.
	rel, err := filepath.Rel(resolvedBase, resolvedValue)
	if err != nil || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return fmt.Errorf("flag --%s: %s is outside %s", name, value, base)
	}
	return nil
}

// resolvePath makes p absolute and follows every symlink it can, including for
// a path that does not exist yet: it resolves the deepest existing ancestor
// and re-appends the components below it.
//
// WHY not filepath.EvalSymlinks alone: it returns an error when any component
// is missing, which would reject every flag naming a file the program is about
// to create.
func resolvePath(p string) (string, error) {
	abs, err := filepath.Abs(p)
	if err != nil {
		return "", err
	}
	current := filepath.Clean(abs)
	var tail string
	for {
		resolved, err := filepath.EvalSymlinks(current)
		if err == nil {
			if tail == "" {
				return resolved, nil
			}
			return filepath.Join(resolved, tail), nil
		}
		parent := filepath.Dir(current)
		if parent == current {
			// Reached the root without finding anything that exists; the
			// cleaned absolute path is the best answer available.
			return filepath.Clean(abs), nil
		}
		tail = filepath.Join(filepath.Base(current), tail)
		current = parent
	}
}

// OpenConfined opens the file named by a confined string flag, for reading,
// without the window ConfinePath leaves open.
//
// The open is performed through os.Root, so containment is enforced while the
// path is being resolved rather than checked beforehand. A symlink planted
// after Parse, or swapped in mid-resolution, cannot redirect the open outside
// the confinement directory: on Linux the kernel enforces it, and elsewhere Go
// resolves the path one element at a time under the same constraint.
//
// This is the only way to close that window. ConfinePath checks a string and
// the application opens it later; whatever happens in between is invisible to
// a check that has already run. Use ConfinePath to fail early with a clear
// message, and OpenConfined when the guarantee has to hold.
//
// Example:
//
//	fs.ConfinePath("config", "/etc/myapp")
//	// ...
//	f, err := fs.OpenConfined("config")
//	if err != nil {
//		log.Fatal(err)
//	}
//	defer f.Close()
//
// Returns an error if the flag does not exist, is not a string flag, has no
// confinement directory, holds no value, or cannot be opened.
func (fs *FlagSet) OpenConfined(name string) (*os.File, error) {
	flag := fs.Lookup(name)
	if flag == nil {
		return nil, fmt.Errorf("flag not found: %s", name)
	}
	if flag.confineTo == "" {
		return nil, fmt.Errorf("flag --%s is not confined; call ConfinePath before OpenConfined", name)
	}
	value, ok := flag.value.(string)
	if !ok {
		return nil, fmt.Errorf("flag --%s is a %s flag; OpenConfined reads a single path", name, flag.flagType)
	}
	if value == "" {
		return nil, fmt.Errorf("flag --%s has no value to open", name)
	}

	root, err := os.OpenRoot(flag.confineTo)
	if err != nil {
		return nil, fmt.Errorf("flag --%s: cannot open confinement directory %s: %v", name, flag.confineTo, err)
	}
	defer func() { _ = root.Close() }()

	// os.Root takes a name relative to the root, so express the value that way
	// when it can be done. This is a convenience, not the guarantee: os.Root
	// refuses an escape itself during resolution. If the value cannot be made
	// relative -- a broken working directory, different Windows volumes -- hand
	// it over unchanged and let os.Root decide, which fails closed.
	//
	// WHY both sides go through resolvePath: resolving only the base makes the
	// two disagree about what the flag names whenever the path is reached
	// through a link. On macOS /var/folders resolves via /private, and on
	// Windows EvalSymlinks expands an 8.3 short name such as RUNNER~1, so the
	// same file spelled two ways came out as an escape. Linux under /tmp hid it,
	// because nothing there needs resolving.
	target := value
	if resolvedBase, err := resolvePath(flag.confineTo); err == nil {
		if resolvedValue, err := resolvePath(value); err == nil {
			if rel, err := filepath.Rel(resolvedBase, resolvedValue); err == nil {
				target = rel
			}
		}
	}

	f, err := root.Open(target)
	if err != nil {
		return nil, fmt.Errorf("flag --%s: %v", name, err)
	}
	return f, nil
}
