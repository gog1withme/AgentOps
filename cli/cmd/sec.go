package cmd

import (
	"fmt"

	"github.com/gog1withme/AgentOps/cli/internal/config"
	"github.com/gog1withme/AgentOps/cli/scrubber"
	"github.com/gog1withme/AgentOps/cli/store"
	"github.com/spf13/cobra"
)

var secCmd = &cobra.Command{
	Use:   "sec",
	Short: "Security scanning",
}

var secScanCmd = &cobra.Command{
	Use:   "scan",
	Short: "Scan recent events for security issues",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			return err
		}
		st, err := store.Open()
		if err != nil {
			return err
		}
		defer st.Close()
		sc, err := scrubber.New(cfg.ScrubPatternsFile)
		if err != nil {
			return err
		}
		events, err := st.ListEvents(500, config.SessionStart(), "", "")
		if err != nil {
			return err
		}
		issues := 0
		for _, e := range events {
			if e.Type == "shell_command" {
				for _, d := range []string{"rm -rf", "mkfs", "format c:"} {
					if contains(e.ShellCommand, d) {
						fmt.Printf("⚠ Dangerous command: %s\n", truncate(e.ShellCommand, 80))
						issues++
					}
				}
			}
			scrubbed := sc.ScrubString(e.ToolInput + e.ToolOutput)
			if scrubbed != e.ToolInput+e.ToolOutput && e.ToolInput+e.ToolOutput != "" {
				// content had secrets before scrub - flag if stored unscrubbed (shouldn't happen)
			}
		}
		alerts, _ := st.ListAlerts(true)
		fmt.Printf("\nScan complete: %d shell warnings, %d active alerts\n", issues, len(alerts))
		return nil
	},
}

func contains(s, sub string) bool {
	return len(s) >= len(sub) && (s == sub || len(sub) == 0 ||
		(len(s) > 0 && (stringIndex(s, sub) >= 0)))
}

func stringIndex(s, sub string) int {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return i
		}
	}
	return -1
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "..."
}

func init() {
	secCmd.AddCommand(secScanCmd)
}
