package main

import (
	"fmt"
	"os"

	"virtusb/internal/cli"
)

func main() {
	// Create CLI interface
	app, err := cli.NewCLI()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to initialize CLI: %v\n", err)
		os.Exit(1)
	}

	// Execute command
	if err := app.Run(os.Args[1:]); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
