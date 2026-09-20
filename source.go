// source.go: configuration source precedence for flash-flags
//
// Copyright (c) 2025 AGILira - A. Giordano
// Series: an AGILira library
// SPDX-License-Identifier: MPL-2.0

package flashflags

// flagSource identifies which configuration layer supplied a flag's current
// value. The zero value is sourceDefault, so a freshly declared flag correctly
// reports that it still holds its default.
//
// The constants are ordered by ascending priority, which is what makes the
// precedence rules a plain comparison: a layer may write a flag only when its
// own source is greater than or equal to the flag's current source. That single
// invariant is why a config file can no longer suppress an environment
// variable, and why re-running a loader after Parse cannot downgrade a value.
//
//	default < config < env < cli
type flagSource uint8

const (
	// sourceDefault means the flag still holds the default given at declaration.
	sourceDefault flagSource = iota
	// sourceConfig means the value came from a configuration file.
	sourceConfig
	// sourceEnv means the value came from an environment variable.
	sourceEnv
	// sourceCLI means the value came from a command-line argument.
	sourceCLI
)

// String returns the lowercase name of the source, as reported by the public
// Flag.Source and FlagSet.Source accessors.
func (s flagSource) String() string {
	switch s {
	case sourceConfig:
		return "config"
	case sourceEnv:
		return "env"
	case sourceCLI:
		return "cli"
	default:
		return "default"
	}
}

// canSet reports whether a value coming from src may overwrite the flag's
// current value. This is the single definition of the precedence rule: a source
// writes only when it ranks at least as high as the one that currently owns the
// value.
//
// The comparison is >= rather than >, so a source can overwrite itself. That is
// what makes a repeated command-line flag keep its last value, and what lets a
// loader be re-run without its own earlier result blocking it.
func (f *Flag) canSet(src flagSource) bool { return src >= f.source }
