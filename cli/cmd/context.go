package cmd

import (
	"fmt"

	"github.com/gog1withme/AgentOps/cli/contextanalysis"
	"github.com/gog1withme/AgentOps/cli/internal/config"
	"github.com/gog1withme/AgentOps/cli/store"
	"github.com/spf13/cobra"
)

var contextCmd = &cobra.Command{
	Use:   "context",
	Short: "Inspect LLM context usage",
}

var contextSummaryCmd = &cobra.Command{
	Use:   "summary",
	Short: "Summarize duplicate and noisy context for the current session",
	RunE:  runContextSummary,
}

func runContextSummary(cmd *cobra.Command, args []string) error {
	st, err := store.Open()
	if err != nil {
		return err
	}
	defer st.Close()

	cfg, _ := config.Load()
	sessionID := cfg.SessionID
	if sessionID == "" {
		sessionID = config.Default().SessionID
	}

	analysis, err := contextanalysis.AnalyzeSession(st, sessionID, 200)
	if err != nil {
		return err
	}

	fmt.Printf("Session: %s\n", analysis.SessionID)
	fmt.Printf("LLM calls: %d\n", analysis.LLMCallCount)
	fmt.Printf("Prompt tokens: %d\n", analysis.TotalPromptTokens)
	if analysis.AvgEfficiency > 0 {
		fmt.Printf("Avg efficiency: %.0f%%\n", analysis.AvgEfficiency)
	}
	fmt.Println()

	for _, c := range analysis.Callouts {
		fmt.Printf("• %s\n", c)
	}
	if len(analysis.Callouts) == 0 {
		fmt.Println("No context issues detected.")
	}

	if len(analysis.DuplicateFiles) > 0 {
		fmt.Println("\nDuplicate context:")
		for _, f := range analysis.DuplicateFiles {
			fmt.Printf("  %s  calls=%d  tokens=%d\n", f.Path, f.Occurrences, f.TotalTokens)
		}
	}
	if len(analysis.NoisyFiles) > 0 {
		fmt.Println("\nNoisy context:")
		for _, f := range analysis.NoisyFiles {
			fmt.Printf("  %s  avg_eff=%.0f%%  tokens=%d\n", f.Path, f.AvgEfficiency, f.TotalTokens)
		}
	}
	return nil
}

func init() {
	contextCmd.AddCommand(contextSummaryCmd)
}
