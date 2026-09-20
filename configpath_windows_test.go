//go:build windows

package flashflags

import "errors"

func mkfifo(string) error { return errors.New("not supported on windows") }
