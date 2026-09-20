// configopen_windows.go: symlink-refusing config file open for Windows.
//
// Copyright (c) 2025 AGILira - A. Giordano
// Series: an AGILira library
// SPDX-License-Identifier: MPL-2.0

//go:build windows

package flashflags

import "os"

// openConfigFile opens path for reading. When strict is set, the final path
// component is checked for a reparse point before the open.
//
// WHY this is weaker than the Unix path: Go's os package exposes no portable
// O_NOFOLLOW equivalent on Windows, so the check cannot be folded into the open
// and a racing replacement between the two could still be followed. The
// exposure is narrower than it looks -- creating a symbolic link on Windows
// requires SeCreateSymbolicLinkPrivilege or Developer Mode -- but it is a
// best-effort check rather than the guarantee the Unix build gives, and
// EnableStrictConfigPaths says so.
//
// Only the final component is examined, matching the Unix behaviour: Lstat does
// not report on a symlinked parent directory.
func openConfigFile(path string, strict bool) (*os.File, error) {
	if strict {
		info, err := os.Lstat(path)
		if err != nil {
			return nil, err
		}
		if info.Mode()&os.ModeSymlink != 0 {
			return nil, errSymlinkRefused
		}
	}
	return os.Open(path) // #nosec G304 -- application-supplied path, never parsed input
}
