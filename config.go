// config.go: configuration file discovery and loading
//
// Copyright (c) 2025 AGILira - A. Giordano
// Series: an AGILira library
// SPDX-License-Identifier: MPL-2.0

package flashflags

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"time"
)

// SetConfigFile sets an explicit configuration file path.
// The config file is loaded automatically during Parse() with lower priority than CLI arguments.
//
// Supported format: JSON with flag names as keys and values matching flag types.
//
// Example config file (myapp.json):
//
//	{
//		"host": "0.0.0.0",
//		"port": 3000,
//		"debug": true,
//		"timeout": "45s",
//		"tags": ["web", "api", "production"]
//	}
//
// Usage:
//
//	fs := flashflags.New("myapp")
//	fs.SetConfigFile("./config/myapp.json")
//	// File will be loaded automatically during Parse()
//
// Security: Path validation prevents directory traversal attacks.
func (fs *FlagSet) SetConfigFile(path string) {
	fs.configFile = path
}

// errSymlinkRefused is returned by openConfigFile when strict config paths are
// enabled and the final path component is a symbolic link. It is a sentinel so
// that both platform implementations report the same thing; loadConfigFromFile
// turns it into an error naming the path.
var errSymlinkRefused = errors.New("configuration file is a symbolic link")

// EnableStrictConfigPaths makes Parse refuse a configuration file whose final
// path component is a symbolic link, for both an explicitly named file and one
// found by auto-discovery. It is off by default.
//
// WHY it is opt-in: following a symlink is what a configuration file is usually
// expected to do. A Kubernetes ConfigMap mount projects every key as a symlink
// into a ..data directory, and dotfile managers such as GNU Stow and chezmoi
// link a config into place from a repository. Refusing symlinks by default
// would break both, on the platform where they are most common.
//
// Turn it on when the program reads configuration from a directory other local
// users can write to, where an attacker can plant a link before startup and
// have a privileged process read a file it chose. That is the classic /tmp
// symlink attack, and it is the only case this setting is for.
//
// Only the final component is examined, so a symlinked parent directory is
// traversed normally. That is deliberate: on macOS /tmp is itself a symlink to
// /private/tmp, and rejecting paths through it would make the setting unusable
// there.
//
// The strength of the guarantee differs by platform, and the difference matters
// if you are relying on it:
//
//   - On Unix the file is opened with O_NOFOLLOW, so the kernel refuses the
//     call. The check and the open are one operation and cannot be raced.
//   - On Windows the final component is checked with Lstat before the open,
//     because Go exposes no portable O_NOFOLLOW there. A replacement racing
//     between the two could still be followed. Creating a symbolic link on
//     Windows needs SeCreateSymbolicLinkPrivilege or Developer Mode, so the
//     exposure is narrower, but this is best-effort rather than a guarantee.
//
// Example:
//
//	fs := flashflags.New("myapp")
//	fs.SetConfigFile("/tmp/myapp.json")
//	fs.EnableStrictConfigPaths()
//
//	if err := fs.Parse(os.Args[1:]); err != nil {
//		// "config file /tmp/myapp.json is a symbolic link (strict config
//		//  paths enabled)"
//		log.Fatal(err)
//	}
func (fs *FlagSet) EnableStrictConfigPaths() {
	fs.strictPaths = true
}

// AddConfigPath adds a directory to search for configuration files during auto-discovery.
// Multiple paths can be added and will be searched in order during Parse().
//
// Auto-discovery searches for these filenames in each path:
//   - {program-name}.json (e.g., "myapp.json")
//   - {program-name}.config.json (e.g., "myapp.config.json")
//   - config.json
//
// Example:
//
//	fs := flashflags.New("myapp")
//	fs.AddConfigPath("./config")        // ./config/myapp.json
//	fs.AddConfigPath("/etc/myapp")      // /etc/myapp/myapp.json
//	home, _ := os.UserHomeDir()
//	fs.AddConfigPath(home)              // $HOME/myapp.json
//
//	// First found config file will be loaded during Parse()
//
// Auto-discovery only runs for paths added here, or for the file named by
// SetConfigFile. A FlagSet with neither loads no configuration at all, so a
// stray config.json in the working directory cannot alter a program that never
// opted into configuration files.
func (fs *FlagSet) AddConfigPath(path string) {
	fs.configPaths = append(fs.configPaths, path)
}

// LoadConfig loads configuration from file and applies it.
// This is called automatically during Parse, but can be called manually if needed.
//
// Possible errors:
//   - File reading errors: permission denied, file not found (for explicit config files)
//   - JSON parsing errors: invalid JSON syntax in configuration file
//   - Path validation errors: unsafe file paths (directory traversal attempts)
//   - Flag validation errors: config values that fail custom validators
//
// Note: Missing auto-discovery config files are not considered errors.
//
// Example error handling:
//
//	if err := fs.LoadConfig(); err != nil {
//		log.Printf("Config error: %v", err)
//		// Continue without config or exit based on your needs
//	}
func (fs *FlagSet) LoadConfig() error {
	// Skip if already loaded
	if fs.configLoaded {
		return nil
	}
	fs.configLoaded = true

	// Skip entirely if no config file specified and no config paths added
	if fs.configFile == "" && len(fs.configPaths) == 0 {
		return nil
	}

	configPath := fs.findConfigFile()
	if configPath == "" {
		return nil // No config file found, not an error
	}

	return fs.loadConfigFromFile(configPath)
}

// findConfigFile finds the configuration file using the specified path or auto-discovery
func (fs *FlagSet) findConfigFile() string {
	// If explicit config file is set, use it
	if fs.configFile != "" {
		if _, err := os.Stat(fs.configFile); err == nil {
			return fs.configFile
		}
		return "" // Explicit file not found
	}

	// Auto-discovery in specified paths
	configNames := []string{
		fs.name + ".json",
		fs.name + ".config.json",
		"config.json",
	}

	// WHY no default search paths: LoadConfig returns early unless the
	// application set a config file or added at least one path, so a default
	// list here was unreachable -- the coverage profile showed the branch at
	// zero hits. Implementing it instead would mean a program that never opted
	// into configuration files silently loading ./<name>.json if one happened
	// to sit in the working directory, which is not a default worth having.
	for _, dir := range fs.configPaths {
		for _, name := range configNames {
			path := filepath.Join(dir, name)
			if _, err := os.Stat(path); err == nil {
				return path
			}
		}
	}

	return ""
}

// loadConfigFromFile loads and applies configuration from a JSON file.
//
// WHY there is no path allowlist: the path comes from the application itself.
// SetConfigFile and AddConfigPath are called by the program, and LoadConfig runs
// before parseArguments, so no command-line value can ever reach here. Until
// v1.1.9 this function rejected any path containing ".." and any absolute path
// outside a five-prefix allowlist. That blocked the user's own home directory --
// the location AddConfigPath's documentation uses as its example -- and a
// directory merely named "v1..2", while still allowing /tmp, the one
// world-writable prefix on the list and therefore the only one where planting a
// symlink is worth anything.
//
// What is checked instead is that the target is a regular file. Reading a
// character device such as /dev/zero as configuration exhausts memory, and
// opening a FIFO blocks Parse until a writer appears -- an unkillable hang that
// the old allowlist permitted, since a FIFO under /tmp passed it. os.Stat
// answers both questions without opening anything.
//
// This is a robustness check, not a security boundary. An application that
// derives the config path from untrusted input must validate and confine it
// before calling SetConfigFile.
//
// Symlinks are followed, because that is what a config file is usually expected
// to do -- a Kubernetes ConfigMap key and a dotfile manager's link both rely on
// it. A program reading configuration from a directory other local users can
// write to should call EnableStrictConfigPaths, which refuses a symlinked final
// component at the open itself.
func (fs *FlagSet) loadConfigFromFile(path string) error {
	info, err := os.Stat(path)
	if err != nil {
		return fmt.Errorf("failed to read config file %s: %v", path, err)
	}
	if !info.Mode().IsRegular() {
		return fmt.Errorf("config file %s is not a regular file", path)
	}

	f, err := openConfigFile(path, fs.strictPaths)
	if err != nil {
		if errors.Is(err, errSymlinkRefused) {
			return fmt.Errorf("config file %s is a symbolic link (strict config paths enabled)", path)
		}
		return fmt.Errorf("failed to read config file %s: %v", path, err)
	}
	defer func() { _ = f.Close() }()

	// Re-check through the open handle: os.Stat above answered before the file
	// was opened, and this one cannot be raced against a replacement.
	if info, err := f.Stat(); err != nil {
		return fmt.Errorf("failed to read config file %s: %v", path, err)
	} else if !info.Mode().IsRegular() {
		return fmt.Errorf("config file %s is not a regular file", path)
	}

	data, err := io.ReadAll(f)
	if err != nil {
		return fmt.Errorf("failed to read config file %s: %v", path, err)
	}

	var config map[string]interface{}
	if err := json.Unmarshal(data, &config); err != nil {
		return fmt.Errorf("failed to parse config file %s: %v", path, err)
	}

	return fs.applyConfig(config)
}

// applyConfig applies configuration values to flags (only if not already set by command line)
func (fs *FlagSet) applyConfig(config map[string]interface{}) error {
	for flagName, value := range config {
		flag := fs.Lookup(flagName)
		if flag == nil {
			continue // Skip unknown flags
		}

		// setFlagValueFromConfig enforces this too; skipping early avoids the
		// conversion work for a value that would be discarded anyway.
		if !flag.canSet(sourceConfig) {
			continue
		}

		// Convert and set the value
		if err := fs.setFlagValueFromConfig(flagName, value); err != nil {
			return fmt.Errorf("failed to set flag %s from config: %v", flagName, err)
		}
	}

	return nil
}

// Config-specific value setters to reduce complexity

func (fs *FlagSet) setStringValueFromConfig(flag *Flag, value interface{}, name string) error {
	if str, ok := value.(string); ok {
		flag.value = str
		if flag.ptr != nil {
			if ptr, ok := flag.ptr.(*string); ok {
				*ptr = str
			}
		}
		return nil
	}
	return fmt.Errorf("expected string for flag %s, got %T", name, value)
}

func (fs *FlagSet) setIntValueFromConfig(flag *Flag, value interface{}, name string) error {
	var intVal int
	switch v := value.(type) {
	case float64: // JSON numbers are float64
		intVal = int(v)
	case int:
		intVal = v
	default:
		return fmt.Errorf("expected number for flag %s, got %T", name, value)
	}
	flag.value = intVal
	if flag.ptr != nil {
		if ptr, ok := flag.ptr.(*int); ok {
			*ptr = intVal
		}
	}
	return nil
}

func (fs *FlagSet) setBoolValueFromConfig(flag *Flag, value interface{}, name string) error {
	if boolVal, ok := value.(bool); ok {
		flag.value = boolVal
		if flag.ptr != nil {
			if ptr, ok := flag.ptr.(*bool); ok {
				*ptr = boolVal
			}
		}
		return nil
	}
	return fmt.Errorf("expected boolean for flag %s, got %T", name, value)
}

func (fs *FlagSet) setFloat64ValueFromConfig(flag *Flag, value interface{}, name string) error {
	var floatVal float64
	switch v := value.(type) {
	case float64:
		floatVal = v
	case int:
		floatVal = float64(v)
	default:
		return fmt.Errorf("expected number for flag %s, got %T", name, value)
	}
	flag.value = floatVal
	if flag.ptr != nil {
		if ptr, ok := flag.ptr.(*float64); ok {
			*ptr = floatVal
		}
	}
	return nil
}

// setDurationValueFromConfig sets a duration flag from a config file value.
// JSON has no duration type, so two encodings are accepted: the human-readable
// string understood by time.ParseDuration ("30s", "1m30s"), and a plain number
// of nanoseconds, which is what encoding/json emits for a time.Duration.
func (fs *FlagSet) setDurationValueFromConfig(flag *Flag, value interface{}, name string) error {
	var duration time.Duration
	switch v := value.(type) {
	case string:
		parsed, err := time.ParseDuration(v)
		if err != nil {
			return fmt.Errorf("invalid duration for flag %s: %v", name, err)
		}
		duration = parsed
	case float64: // JSON numbers decode to float64; treated as nanoseconds
		duration = time.Duration(v)
	default:
		return fmt.Errorf("expected duration string or number for flag %s, got %T", name, value)
	}
	flag.value = duration
	if flag.ptr != nil {
		if ptr, ok := flag.ptr.(*time.Duration); ok {
			*ptr = duration
		}
	}
	return nil
}

func (fs *FlagSet) setStringSliceValueFromConfig(flag *Flag, value interface{}, name string) error {
	if slice, ok := value.([]interface{}); ok {
		strSlice := make([]string, len(slice))
		for i, item := range slice {
			if str, ok := item.(string); ok {
				strSlice[i] = str
			} else {
				return fmt.Errorf("expected string array for flag %s, got %T in array", name, item)
			}
		}
		flag.value = strSlice
		if flag.ptr != nil {
			if ptr, ok := flag.ptr.(*[]string); ok {
				*ptr = strSlice
			}
		}
		return nil
	}
	return fmt.Errorf("expected array for flag %s, got %T", name, value)
}

func (fs *FlagSet) setFlagValueFromConfig(name string, value interface{}) error {
	flag, exists := fs.flags[name]
	if !exists {
		return fmt.Errorf("unknown flag: %s", name)
	}

	if !flag.canSet(sourceConfig) {
		return nil
	}

	if err := fs.screenConfigValue(name, value); err != nil {
		return err
	}

	// Set value based on type using dedicated functions
	if err := fs.setConfigValueByType(flag, value, name); err != nil {
		return err
	}

	// Record the config file as this value's origin.
	flag.source = sourceConfig

	// Validate the value if validator is set
	return fs.validateFlagValue(flag)
}

// screenConfigValue applies input hygiene to a value decoded from a config
// file. JSON carries strings in two shapes, and both reach a flag: a bare
// string, and the elements of an array bound to a string slice.
//
// WHY both: until v1.1.9 only the bare string was screened, so
// {"tags": ["bad\u0000item"]} reached the flag untouched while
// --tags "bad\x00item" was rejected. Whether a value is accepted must not
// depend on which layer supplied it.
//
// Numbers and booleans need no screening: they cannot carry a byte sequence.
func (fs *FlagSet) screenConfigValue(name string, value interface{}) error {
	switch v := value.(type) {
	case string:
		if err := fs.validateInputHygiene(name, v); err != nil {
			return fmt.Errorf("config security validation failed: %v", err)
		}
	case []interface{}:
		for i, item := range v {
			str, ok := item.(string)
			if !ok {
				continue // type mismatch is reported by the element setter
			}
			if err := fs.validateInputHygiene(name+"["+strconv.Itoa(i)+"]", str); err != nil {
				return fmt.Errorf("config security validation failed: %v", err)
			}
		}
	}
	return nil
}

// setConfigValueByType sets the flag value from config based on its type
func (fs *FlagSet) setConfigValueByType(flag *Flag, value interface{}, name string) error {
	switch flag.flagType {
	case "string":
		return fs.setStringValueFromConfig(flag, value, name)
	case "int":
		return fs.setIntValueFromConfig(flag, value, name)
	case "bool":
		return fs.setBoolValueFromConfig(flag, value, name)
	case "float64":
		return fs.setFloat64ValueFromConfig(flag, value, name)
	case "duration":
		return fs.setDurationValueFromConfig(flag, value, name)
	case "stringSlice":
		return fs.setStringSliceValueFromConfig(flag, value, name)
	default:
		return fmt.Errorf("unsupported flag type: %s", flag.flagType)
	}
}

// validateFlagValue validates a flag value using its validator function
func (fs *FlagSet) validateFlagValue(flag *Flag) error {
	if err := fs.checkConfinement(flag, flag.name); err != nil {
		return err
	}
	if flag.validator != nil {
		return flag.validator(flag.value)
	}
	return nil
}
