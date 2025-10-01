package main

import (
	"fmt"
	"log"
	"os"

	flashflags "github.com/agilira/flash-flags"
)

func main() {
	// Create a new FlagSet
	flags := flashflags.New("advanced-syntax")

	// Define flags with short keys for combined flag usage
	var (
		verbose    = flags.BoolVar("verbose", "v", false, "Enable verbose output")
		debug      = flags.BoolVar("debug", "d", false, "Enable debug mode")
		quiet      = flags.BoolVar("quiet", "q", false, "Suppress output")
		name       = flags.StringVar("name", "n", "default", "Specify a name")
		port       = flags.IntVar("port", "p", 8080, "Server port number")
		timeout    = flags.Duration("timeout", 0, "Request timeout (no short key)")
		configFile = flags.StringVar("config", "c", "", "Configuration file path")
		force      = flags.BoolVar("force", "f", false, "Force operation")
	)

	// Parse command line arguments
	if err := flags.Parse(os.Args[1:]); err != nil {
		log.Fatal(err)
	}

	// Display parsed values
	fmt.Println("=== Flash-flags Advanced Syntax Demo ===")
	fmt.Printf("Verbose: %v\n", *verbose)
	fmt.Printf("Debug: %v\n", *debug)
	fmt.Printf("Quiet: %v\n", *quiet)
	fmt.Printf("Name: %s\n", *name)
	fmt.Printf("Port: %d\n", *port)
	fmt.Printf("Timeout: %v\n", *timeout)
	fmt.Printf("Config: %s\n", *configFile)
	fmt.Printf("Force: %v\n", *force)

	// Show usage examples if no arguments provided
	if len(os.Args) == 1 {
		fmt.Println("\n=== Usage Examples ===")
		fmt.Println("Standard syntax:")
		fmt.Println("  ./advanced-syntax --verbose --name myapp --port 9000")
		fmt.Println("  ./advanced-syntax -v -n myapp -p 9000")
		fmt.Println()
		fmt.Println("Advanced syntax (equals assignment):")
		fmt.Println("  ./advanced-syntax -n=myapp -p=9000 --timeout=30s")
		fmt.Println("  ./advanced-syntax --name=myapp --port=9000 --timeout=30s")
		fmt.Println()
		fmt.Println("Combined short flags:")
		fmt.Println("  ./advanced-syntax -vdf")
		fmt.Println("  ./advanced-syntax -vq -n=testapp")
		fmt.Println("  ./advanced-syntax -vdn myapp")
		fmt.Println("  ./advanced-syntax -vdp 8090")
		fmt.Println()
		fmt.Println("Mixed usage:")
		fmt.Println("  ./advanced-syntax -vd --name=server -p=3000 -c=/etc/config.json")
		fmt.Println("  ./advanced-syntax -vqf --timeout=1m --config=/path/to/config")
		fmt.Println()
		fmt.Println("Try running with different flag combinations to see the parser in action!")
	}
}
