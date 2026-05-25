package main

import (
	"os"

	"github.com/gog1withme/AgentOps/cli/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
