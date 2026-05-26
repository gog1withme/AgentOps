package cmd

import (
	"os"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "agentops",
	Short: "Local observability for AI coding agents",
	Long:  "AgentOps — passive telemetry and observability for Claude Code, Cursor, Copilot, and more.",
}

func Execute() error {
	setupLogging()
	return rootCmd.Execute()
}

func setupLogging() {
	level := zerolog.InfoLevel
	switch os.Getenv("AGENTOPS_LOG_LEVEL") {
	case "debug":
		level = zerolog.DebugLevel
	case "warn":
		level = zerolog.WarnLevel
	case "error":
		level = zerolog.ErrorLevel
	}
	log.Logger = zerolog.New(os.Stderr).With().Timestamp().Logger().Level(level)
}

func init() {
	rootCmd.AddCommand(initCmd)
	rootCmd.AddCommand(devCmd)
	rootCmd.AddCommand(budgetCmd)
	rootCmd.AddCommand(restoreCmd)
	rootCmd.AddCommand(traceCmd)
	rootCmd.AddCommand(replayCmd)
	rootCmd.AddCommand(diffCmd)
	rootCmd.AddCommand(blameCmd)
	rootCmd.AddCommand(promptCmd)
	rootCmd.AddCommand(mcpCmd)
	rootCmd.AddCommand(alertsCmd)
	rootCmd.AddCommand(doctorCmd)
	rootCmd.AddCommand(envCmd)
	rootCmd.AddCommand(secCmd)
	rootCmd.AddCommand(contextCmd)
	rootCmd.AddCommand(upgradeCmd)
}
