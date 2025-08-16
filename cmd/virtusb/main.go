package main

import (
	"fmt"
	"os"

	"virtusb/internal/cli"
)

var version = "1.0.0"

func main() {
	app, err := cli.NewCLI()
	if err != nil {
		fmt.Fprintf(os.Stderr, "❌ Initialization error: %v\n", err)
		os.Exit(1)
	}

	if err := app.Run(os.Args[1:]); err != nil {
		fmt.Fprintf(os.Stderr, "❌ Error: %v\n", err)
		os.Exit(1)
	}
}
