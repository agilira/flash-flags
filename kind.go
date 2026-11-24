// kind.go: the per-type behaviour of a flag, kept in one place per type
//
// Copyright (c) 2025 AGILira - A. Giordano
// Series: an AGILira library
// SPDX-License-Identifier: MPL-2.0

package flashflags

import (
	"fmt"
	"strconv"
	"time"
)

// flagKind is everything that differs between flag types: the name Type
// reports, how a command-line or environment value is parsed, how a value
// decoded from the JSON config file is converted, and how a value is written
// into the pointer handed out at declaration.
//
// WHY one value per type: the type used to be a string switched on in four
// places (CLI setter, config setter, pointer reset, pointer update). A fix
// applied to one switch and not the others is a bug waiting to happen, and
// adding a type meant touching all four. Now a type is one declaration here.
type flagKind struct {
	name string
	// parse converts a value from the command line or the environment.
	// again is true when the flag was already given on this command line;
	// only an accumulating kind looks at it.
	//
	// WHY the FlagSet is not passed: through an indirect call the compiler
	// cannot prove it does not escape, and every Parse then pays a heap
	// allocation for the FlagSet itself.
	parse func(f *Flag, raw string, again bool) (interface{}, error)
	// fromConfig converts a value decoded from the JSON config file.
	fromConfig func(f *Flag, v interface{}) (interface{}, error)
	// accumulates marks a kind whose repeated command-line values append
	// rather than replace.
	accumulates bool
	// store writes v into ptr. It does nothing when either has the wrong
	// type, so a flag whose default was replaced by a foreign value is left
	// alone rather than panicking.
	store func(ptr, v interface{})
}

// storeAs writes v into ptr when ptr is a *T and v is a T.
func storeAs[T any](ptr, v interface{}) {
	p, okPtr := ptr.(*T)
	x, okVal := v.(T)
	if okPtr && okVal {
		*p = x
	}
}

var stringKind = &flagKind{
	name: "string",
	parse: func(_ *Flag, raw string, _ bool) (interface{}, error) {
		return raw, nil
	},
	fromConfig: func(f *Flag, v interface{}) (interface{}, error) {
		if str, ok := v.(string); ok {
			return str, nil
		}
		return nil, fmt.Errorf("expected string for flag %s, got %T", f.name, v)
	},
	store: storeAs[string],
}

var intKind = &flagKind{
	name: "int",
	parse: func(f *Flag, raw string, _ bool) (interface{}, error) {
		n, err := strconv.Atoi(raw)
		if err != nil {
			return nil, fmt.Errorf("invalid int value for flag --%s: %s", f.name, raw)
		}
		return n, nil
	},
	fromConfig: func(f *Flag, v interface{}) (interface{}, error) {
		switch n := v.(type) {
		case float64: // JSON numbers are float64
			return int(n), nil
		case int:
			return n, nil
		}
		return nil, fmt.Errorf("expected number for flag %s, got %T", f.name, v)
	},
	store: storeAs[int],
}

var boolKind = &flagKind{
	name: "bool",
	parse: func(f *Flag, raw string, _ bool) (interface{}, error) {
		b, err := strconv.ParseBool(raw)
		if err != nil {
			return nil, fmt.Errorf("invalid bool value for flag --%s: %s", f.name, raw)
		}
		return b, nil
	},
	fromConfig: func(f *Flag, v interface{}) (interface{}, error) {
		if b, ok := v.(bool); ok {
			return b, nil
		}
		return nil, fmt.Errorf("expected boolean for flag %s, got %T", f.name, v)
	},
	store: storeAs[bool],
}

var float64Kind = &flagKind{
	name: "float64",
	parse: func(f *Flag, raw string, _ bool) (interface{}, error) {
		x, err := strconv.ParseFloat(raw, 64)
		if err != nil {
			return nil, fmt.Errorf("invalid float64 value for flag --%s: %s", f.name, raw)
		}
		return x, nil
	},
	fromConfig: func(f *Flag, v interface{}) (interface{}, error) {
		switch x := v.(type) {
		case float64:
			return x, nil
		case int:
			return float64(x), nil
		}
		return nil, fmt.Errorf("expected number for flag %s, got %T", f.name, v)
	},
	store: storeAs[float64],
}

// durationKind accepts two encodings in a config file, because JSON has no
// duration type: the string understood by time.ParseDuration ("30s",
// "1m30s"), and a plain number of nanoseconds, which is what encoding/json
// emits for a time.Duration.
var durationKind = &flagKind{
	name: "duration",
	parse: func(f *Flag, raw string, _ bool) (interface{}, error) {
		d, err := time.ParseDuration(raw)
		if err != nil {
			return nil, fmt.Errorf("invalid duration value for flag --%s: %s", f.name, raw)
		}
		return d, nil
	},
	fromConfig: func(f *Flag, v interface{}) (interface{}, error) {
		switch d := v.(type) {
		case string:
			parsed, err := time.ParseDuration(d)
			if err != nil {
				return nil, fmt.Errorf("invalid duration for flag %s: %v", f.name, err)
			}
			return parsed, nil
		case float64: // JSON numbers decode to float64; treated as nanoseconds
			return time.Duration(d), nil
		}
		return nil, fmt.Errorf("expected duration string or number for flag %s, got %T", f.name, v)
	},
	store: storeAs[time.Duration],
}

// maxSliceItems caps a string slice flag (DoS protection).
const maxSliceItems = 10000

var stringSliceKind = &flagKind{
	name:        "stringSlice",
	parse:       parseStringSliceValue,
	accumulates: true,
	fromConfig: func(f *Flag, v interface{}) (interface{}, error) {
		items, ok := v.([]interface{})
		if !ok {
			return nil, fmt.Errorf("expected array for flag %s, got %T", f.name, v)
		}
		out := make([]string, len(items))
		for i, item := range items {
			str, ok := item.(string)
			if !ok {
				return nil, fmt.Errorf("expected string array for flag %s, got %T in array", f.name, item)
			}
			out[i] = str
		}
		return out, nil
	},
	store: storeAs[[]string],
}

// parseStringSliceValue splits a comma-separated value.
//
// WHY again appends: a slice flag repeated on the command line accumulates
// (--tag a --tag b means [a b]), as the NOTE on splitByComma promises. Only a
// repetition within the command line appends: a value from the command line
// still replaces one from the default, env or config.
//
// WHY no per-element screening here: the caller screened the whole value
// before it was split, and input hygiene is a per-character property, so no
// element can fail a check the joined string already passed. The config file
// path does need it, because a JSON array never passes through this function
// -- see screenConfigValue.
func parseStringSliceValue(f *Flag, raw string, again bool) (interface{}, error) {
	slice := splitStringSlice(raw)
	if again {
		if current, ok := f.value.([]string); ok {
			slice = append(append([]string(nil), current...), slice...)
		}
	}
	if len(slice) > maxSliceItems {
		return nil, fmt.Errorf("string slice too large: %d items (max: %d)", len(slice), maxSliceItems)
	}
	return slice, nil
}
