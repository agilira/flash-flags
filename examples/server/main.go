// main.go: a small server configured from flags, environment and a config file
//
// Copyright (c) 2026 AGILira - A. Giordano
// Series: an AGILira library
// SPDX-License-Identifier: MPL-2.0

// This is the program recorded in assets/demo.gif (see assets/demo.sh).
package main

import (
	"fmt"
	"os"
	"strings"
	"time"

	flashflags "github.com/agilira/flash-flags"
)

func main() {
	fs := flashflags.New("server")
	fs.SetDescription("A small server configured with flash-flags")
	fs.SetVersion("v1.3.4")

	host := fs.StringVar("host", "H", "localhost", "Address to listen on")
	port := fs.IntVar("port", "p", 8080, "Port to listen on")
	timeout := fs.DurationVar("timeout", "t", 30*time.Second, "Request timeout")
	verbose := fs.BoolVar("verbose", "v", false, "Verbose logging")
	tags := fs.StringSliceVar("tag", "", nil, "Tag to attach (repeatable)")

	_ = fs.SetGroup("host", "Network")
	_ = fs.SetGroup("port", "Network")
	_ = fs.SetGroup("timeout", "Runtime")
	_ = fs.SetGroup("verbose", "Runtime")
	_ = fs.SetGroup("tag", "Runtime")

	_ = fs.SetValidator("port", func(v interface{}) error {
		if p, ok := v.(int); !ok || p < 1 || p > 65535 {
			return fmt.Errorf("must be between 1 and 65535")
		}
		return nil
	})

	// Values come from server.json in the working directory, then from
	// SERVER_* environment variables, then from the command line.
	fs.AddConfigPath(".")
	fs.SetEnvPrefix("SERVER")

	if err := fs.Parse(os.Args[1:]); err != nil {
		if err.Error() == "help requested" {
			return
		}
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(2)
	}

	show := func(name string, value interface{}) {
		fmt.Printf("  %-8s %-14v %s\n", name, value, fs.Source(name))
	}
	show("host", *host)
	show("port", *port)
	show("timeout", *timeout)
	show("verbose", *verbose)
	show("tag", "["+strings.Join(*tags, " ")+"]")
}
