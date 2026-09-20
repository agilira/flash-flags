// configopen_unix.go: symlink-refusing config file open for Unix systems.
//
// Copyright (c) 2025 AGILira - A. Giordano
// Series: an AGILira library
// SPDX-License-Identifier: MPL-2.0

//go:build !windows

package flashflags

import (
	"errors"
	"os"
	"syscall"
)

// openConfigFile opens path for reading. When strict is set, the open is made
// with O_NOFOLLOW, so the kernel itself refuses the call if the final path
// component is a symbolic link.
//
// WHY the kernel and not a prior Lstat: the check and the open are the same
// operation, so there is no window in which the path can be swapped between
// deciding it is safe and reading it. A separate Lstat would leave exactly that
// race open, which is the race the caller enabled strict paths to close.
//
// Only the final component is examined. A symlinked parent directory is
// traversed like any other, which is what keeps this usable on macOS, where
// /tmp is a symlink to /private/tmp.
func openConfigFile(path string, strict bool) (*os.File, error) {
	flags := os.O_RDONLY
	if strict {
		flags |= syscall.O_NOFOLLOW
	}
	f, err := os.OpenFile(path, flags, 0) // #nosec G304 -- application-supplied path, never parsed input
	if err != nil && strict && errors.Is(err, syscall.ELOOP) {
		return nil, errSymlinkRefused
	}
	return f, err
}
