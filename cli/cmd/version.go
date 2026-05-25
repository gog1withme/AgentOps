package cmd

import (
	"fmt"

	"github.com/gog1withme/AgentOps/cli/version"
	"github.com/spf13/cobra"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the AgentOps version",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println(version.Version)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
	rootCmd.Version = version.Version
}
