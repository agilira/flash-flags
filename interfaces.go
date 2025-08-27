// flash-flags.go: flash-flags interfaces
//
// Copyright (c) 2025 AGILira
// Series: an AGILira library
// SPDX-License-Identifier: MPL-2.0

package flashflags

// ConfigFlag represents a flag interface for configuration management integration.
// This interface allows flash-flags to integrate seamlessly with configuration
// management systems like Argus, Viper, or custom solutions.
type ConfigFlag interface {
	Name() string
	Value() interface{}
	Type() string
	Changed() bool
	Usage() string
}

// ConfigFlagSet represents a collection of flags for configuration integration.
// This interface provides the standard contract expected by configuration managers.
type ConfigFlagSet interface {
	VisitAll(func(ConfigFlag))
	Lookup(name string) ConfigFlag
}

// FlagSetAdapter wraps a FlagSet to implement ConfigFlagSet interface.
// This allows seamless integration with configuration management systems.
type FlagSetAdapter struct {
	*FlagSet
}

// NewAdapter creates a new FlagSetAdapter for configuration integration.
func NewAdapter(fs *FlagSet) *FlagSetAdapter {
	return &FlagSetAdapter{FlagSet: fs}
}

// VisitAll implements ConfigFlagSet interface.
func (fsa *FlagSetAdapter) VisitAll(fn func(ConfigFlag)) {
	fsa.FlagSet.VisitAll(func(flag *Flag) {
		fn(flag)
	})
}

// Lookup implements ConfigFlagSet interface.
func (fsa *FlagSetAdapter) Lookup(name string) ConfigFlag {
	if flag := fsa.FlagSet.Lookup(name); flag != nil {
		return flag
	}
	return nil
}

// Ensure our types implement the interfaces
var (
	_ ConfigFlag    = (*Flag)(nil)
	_ ConfigFlagSet = (*FlagSetAdapter)(nil)
)
