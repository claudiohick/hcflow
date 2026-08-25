// Command hcflow is the main entrypoint for the hcflow CLI.
package main

import (
	"os"

	"github.com/hickstein/hcflow/cmd/hcflow/commands"
)

func main() {
	if err := commands.Execute(); err != nil {
		os.Exit(1)
	}
}
