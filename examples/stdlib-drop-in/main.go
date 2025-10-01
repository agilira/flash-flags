// Example demonstrates drop-in replacement for standard library flag package.
// This example shows how existing code can migrate to flash-flags with ZERO changes
// while gaining advanced features like config file support and environment variables.
package main

import (
	"fmt"

	// Drop-in replacement: just change the import!
	// OLD: import "flag"
	// NEW: import flag "github.com/agilira/flash-flags/stdlib"
	flag "github.com/agilira/flash-flags/stdlib"
)

func main() {
	// This is EXACTLY the same code you would write with stdlib flag!
	var (
		name    = flag.String("name", "world", "Name to greet")
		port    = flag.Int("port", 8080, "Server port")
		debug   = flag.Bool("debug", false, "Enable debug mode")
		timeout = flag.Duration("timeout", 0, "Request timeout")
		ratio   = flag.Float64("ratio", 1.0, "Processing ratio")
	)

	// Variables with *Var functions (also exactly the same as stdlib)
	var (
		host     string
		workers  int
		verbose  bool
		maxConns int64
		threads  uint
		buffer   uint64
	)

	flag.StringVar(&host, "host", "localhost", "Server host")
	flag.IntVar(&workers, "workers", 4, "Number of workers")
	flag.BoolVar(&verbose, "verbose", false, "Verbose logging")
	flag.Int64Var(&maxConns, "max-conns", 1000, "Maximum connections")
	flag.UintVar(&threads, "threads", 2, "Number of threads")
	flag.Uint64Var(&buffer, "buffer", 4096, "Buffer size")

	// Parse flags - exactly the same!
	flag.Parse()

	// Use the flags - exactly the same!
	fmt.Printf("  Server Configuration (powered by flash-flags):\n")
	fmt.Printf("  Address: %s:%d\n", host, *port)
	fmt.Printf("  Greeting: Hello, %s!\n", *name)
	fmt.Printf("  Workers: %d\n", workers)
	fmt.Printf("  Max Connections: %d\n", maxConns)
	fmt.Printf("  Threads: %d\n", threads)
	fmt.Printf("  Buffer Size: %d bytes\n", buffer)
	fmt.Printf("  Ratio: %.2f\n", *ratio)
	fmt.Printf("  Debug: %v\n", *debug)
	fmt.Printf("  Verbose: %v\n", verbose)
	if *timeout > 0 {
		fmt.Printf("  Timeout: %v\n", *timeout)
	}

	// Demonstrate stdlib compatibility functions
	fmt.Printf("\n Flag Statistics:\n")
	fmt.Printf("  Flags set: %d\n", flag.NFlag())
	fmt.Printf("  Remaining args: %d\n", flag.NArg())

	if flag.NArg() > 0 {
		fmt.Printf("  Args: %v\n", flag.Args())
	}

	// Show that we can lookup flags (stdlib API)
	if hostFlag := flag.Lookup("host"); hostFlag != nil {
		fmt.Printf("  Host flag usage: %s\n", hostFlag.Usage)
	}

	fmt.Printf("\n This example works with ZERO changes from stdlib flag!\n")
	fmt.Printf("   But you get all flash-flags benefits:\n")
	fmt.Printf("   • 1.5x faster parsing\n")
	fmt.Printf("   • Short flags support (-h, -p, etc.)\n")
	fmt.Printf("   • Combined flags (-vd for -v -d)\n")
	fmt.Printf("   • Configuration file support\n")
	fmt.Printf("   • Environment variable integration\n")
	fmt.Printf("   • Advanced validation\n")
	fmt.Printf("   • Professional help output\n")

	// Demo advanced usage (these are flash-flags extensions)
	fmt.Printf("\n Try these advanced features:\n")
	fmt.Printf("   ./stdlib-drop-in --help                    # Professional help\n")
	fmt.Printf("   ./stdlib-drop-in -h localhost -p 9000      # Short flags\n")
	fmt.Printf("   ./stdlib-drop-in -vd                       # Combined flags\n")
	fmt.Printf("   MYAPP_HOST=remote ./stdlib-drop-in         # Environment variables\n")

	fmt.Printf("\n Migration path:\n")
	fmt.Printf("   1. Change import: flag → github.com/agilira/flash-flags/stdlib\n")
	fmt.Printf("   2. That's it! Zero code changes needed.\n")
	fmt.Printf("   3. Gradually add flash-flags features as needed.\n")
}
