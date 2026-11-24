// help.go: help and usage output
//
// Copyright (c) 2025 AGILira - A. Giordano
// Series: an AGILira library
// SPDX-License-Identifier: MPL-2.0

package flashflags

import (
	"fmt"
	"strings"
)

// PrintUsage prints basic usage information for all flags to stdout.
// This provides a simpler flag listing without the full help formatting, groups, or descriptions.
//
// Output format for each flag:
//
//	--flagname, -s
//	      Description text (type: flagtype)
//
// Example:
//
//	fs := flashflags.New("myapp")
//	fs.StringVar("host", "h", "localhost", "Server host")
//	fs.IntVar("port", "p", 8080, "Server port")
//
//	fs.PrintUsage()
//	// Output:
//	// Usage of myapp:
//	//   --host, -h
//	//         Server host (type: string)
//	//   --port, -p
//	//         Server port (type: int)
//
// Use PrintHelp() for complete help with grouping, defaults, and requirements.
func (fs *FlagSet) PrintUsage() {
	fmt.Printf("Usage of %s:\n", fs.name)
	for _, flag := range fs.order {
		name := flag.name
		fmt.Printf("  --%s", name)
		if flag.shortKey != "" {
			fmt.Printf(", -%s", flag.shortKey)
		}
		fmt.Printf("\n")
		fmt.Printf("        %s (type: %s)\n", flag.usage, flag.kind.name)
	}
}

// SetDescription sets the program description displayed at the top of help output.
// The description should briefly explain what the program does.
//
// Example:
//
//	fs := flashflags.New("webserver")
//	fs.SetDescription("High-performance HTTP server with advanced configuration options")
//
//	// Help output will show:
//	// High-performance HTTP server with advanced configuration options
//	//
//	// Usage: webserver [options]
//	// ...
//
// Call before Parse() to ensure it's included in help text.
func (fs *FlagSet) SetDescription(description string) {
	fs.description = description
}

// SetVersion sets the program version displayed in help output.
// The version should follow semantic versioning (e.g., "v1.2.3").
//
// Example:
//
//	fs := flashflags.New("myapp")
//	fs.SetVersion("v2.1.0")
//
//	// Help output will show:
//	// Usage: myapp [options]
//	//
//	// Version: v2.1.0
//	// ...
//
// Call before Parse() to ensure it's included in help text.
func (fs *FlagSet) SetVersion(version string) {
	fs.version = version
}

// SetGroup sets the group name for a flag to organize help output.
// Flags with the same group will be displayed together under a group heading.
//
// Example:
//
//	fs := flashflags.New("server")
//	host := fs.String("host", "localhost", "Server host")
//	port := fs.Int("port", 8080, "Server port")
//	dbHost := fs.String("db-host", "localhost", "Database host")
//
//	fs.SetGroup("host", "Server Options")
//	fs.SetGroup("port", "Server Options")
//	fs.SetGroup("db-host", "Database Options")
//
//	// Help output will show:
//	// Server Options:
//	//   --host     Server host
//	//   --port     Server port
//	//
//	// Database Options:
//	//   --db-host  Database host
//
// Returns an error if the flag name doesn't exist.
func (fs *FlagSet) SetGroup(name, group string) error {
	flag := fs.Lookup(name)
	if flag == nil {
		return fmt.Errorf("flag not found: %s", name)
	}
	flag.group = group
	return nil
}

// Help generates and returns the complete help text as a string.
// Includes program description, version, usage line, and all flags organized by groups.
//
// The help text format:
//   - Program description (if set with SetDescription)
//   - Usage line: "Usage: {program-name} [options]"
//   - Version info (if set with SetVersion)
//   - Ungrouped flags (if any)
//   - Grouped flags organized by SetGroup() calls
//
// Example:
//
//	fs := flashflags.New("myserver")
//	fs.SetDescription("High-performance web server")
//	fs.SetVersion("v2.1.0")
//
//	port := fs.IntVar("port", "p", 8080, "Server port")
//	host := fs.String("host", "localhost", "Server host")
//
//	fs.SetGroup("port", "Server Options")
//	fs.SetGroup("host", "Server Options")
//
//	helpText := fs.Help()
//	// Contains formatted help with grouped flags, defaults, requirements, etc.
//
// Use PrintHelp() to output directly to stdout.
func (fs *FlagSet) Help() string {
	var help strings.Builder

	// Program name and description
	if fs.description != "" {
		help.WriteString(fs.description)
		help.WriteString("\n\n")
	}

	// Usage line
	help.WriteString("Usage: ")
	help.WriteString(fs.name)
	help.WriteString(" [options]\n\n")

	// Version info
	if fs.version != "" {
		help.WriteString("Version: ")
		help.WriteString(fs.version)
		help.WriteString("\n\n")
	}

	// Group flags by group name
	groups := make(map[string][]*Flag)
	var groupOrder []string
	ungrouped := []*Flag{}

	for _, flag := range fs.order {
		if flag.group != "" {
			if _, seen := groups[flag.group]; !seen {
				groupOrder = append(groupOrder, flag.group)
			}
			groups[flag.group] = append(groups[flag.group], flag)
		} else {
			ungrouped = append(ungrouped, flag)
		}
	}

	// Display ungrouped flags first
	if len(ungrouped) > 0 {
		help.WriteString("Options:\n")
		for _, flag := range ungrouped {
			help.WriteString(fs.formatFlagHelp(flag))
		}
		help.WriteString("\n")
	}

	// Display grouped flags
	for _, groupName := range groupOrder {
		help.WriteString(groupName)
		help.WriteString(":\n")
		for _, flag := range groups[groupName] {
			help.WriteString(fs.formatFlagHelp(flag))
		}
		help.WriteString("\n")
	}

	return help.String()
}

// formatFlagHelp formats a single flag for help output
func (fs *FlagSet) formatFlagHelp(flag *Flag) string {
	var line strings.Builder

	// Build flag name with short key
	fs.buildFlagName(&line, flag)

	// Add type info for non-bool flags
	fs.addTypeInfo(&line, flag)

	// Pad to align descriptions
	fs.padForAlignment(&line)

	// Add description and modifiers
	fs.addDescriptionAndModifiers(&line, flag)

	line.WriteString("\n")
	return line.String()
}

// buildFlagName builds the flag name part of help output
func (fs *FlagSet) buildFlagName(line *strings.Builder, flag *Flag) {
	line.WriteString("  ")
	if flag.shortKey != "" {
		line.WriteString("-")
		line.WriteString(flag.shortKey)
		line.WriteString(", ")
	}
	line.WriteString("--")
	line.WriteString(flag.name)
}

// addTypeInfo adds type information for non-bool flags
func (fs *FlagSet) addTypeInfo(line *strings.Builder, flag *Flag) {
	if flag.kind.name != "bool" {
		line.WriteString(" ")
		line.WriteString(strings.ToUpper(flag.kind.name))
	}
}

// padForAlignment pads the line to align descriptions
func (fs *FlagSet) padForAlignment(line *strings.Builder) {
	for line.Len() < 30 {
		line.WriteString(" ")
	}
}

// addDescriptionAndModifiers adds description, default value, required indicator, and dependencies
func (fs *FlagSet) addDescriptionAndModifiers(line *strings.Builder, flag *Flag) {
	// Add description
	line.WriteString(flag.usage)

	// Add default value
	if flag.defaultValue != nil && flag.kind.name != "bool" {
		line.WriteString(" (default: ")
		line.WriteString(fmt.Sprintf("%v", flag.defaultValue))
		line.WriteString(")")
	}

	// Add required indicator
	if flag.required {
		line.WriteString(" [REQUIRED]")
	}

	// Add dependencies
	fs.addDependencies(line, flag)
}

// addDependencies adds dependency information to help output
func (fs *FlagSet) addDependencies(line *strings.Builder, flag *Flag) {
	if len(flag.dependencies) > 0 {
		line.WriteString(" [depends on: ")
		for i, dep := range flag.dependencies {
			if i > 0 {
				line.WriteString(", ")
			}
			line.WriteString(dep)
		}
		line.WriteString("]")
	}
}

func (fs *FlagSet) PrintHelp() {
	fmt.Print(fs.Help())
}
