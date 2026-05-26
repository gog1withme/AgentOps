package cmd

import (
	"fmt"
	"maps"
	"time"

	"github.com/gog1withme/AgentOps/cli/store"
	"github.com/gog1withme/AgentOps/schema"
	"github.com/spf13/cobra"
)

var traceCmd = &cobra.Command{
	Use:   "trace",
	Short: "List and inspect traces",
}

var traceListCmd = &cobra.Command{
	Use:   "list",
	Short: "List sessions",
	RunE:  runTraceList,
}

var replayCmd = &cobra.Command{
	Use:   "replay [run_id]",
	Short: "Replay session timeline",
	Args:  cobra.ExactArgs(1),
	RunE:  runReplay,
}

var diffCmd = &cobra.Command{
	Use:   "diff [run_a] [run_b]",
	Short: "Compare two sessions",
	Args:  cobra.ExactArgs(2),
	RunE:  runDiff,
}

func runTraceList(cmd *cobra.Command, args []string) error {
	st, err := store.Open()
	if err != nil {
		return err
	}
	defer st.Close()
	sessions, err := st.ListSessions()
	if err != nil {
		return err
	}
	for _, s := range sessions {
		fmt.Printf("%s  events=%d  cost=$%.4f  %s → %s\n",
			s.SessionID, s.EventCount, s.TotalCost,
			s.FirstEvent.Format(time.RFC3339), s.LastEvent.Format(time.RFC3339))
	}
	return nil
}

func runReplay(cmd *cobra.Command, args []string) error {
	st, err := store.Open()
	if err != nil {
		return err
	}
	defer st.Close()
	events, err := st.ReplaySession(args[0])
	if err != nil {
		return err
	}
	for _, e := range events {
		fmt.Printf("[%s] %s %s %s\n", e.Timestamp.Format("15:04:05"), e.Type, e.Source, eventSummary(e))
	}
	return nil
}

func runDiff(cmd *cobra.Command, args []string) error {
	st, err := store.Open()
	if err != nil {
		return err
	}
	defer st.Close()
	a, err := st.ReplaySession(args[0])
	if err != nil {
		return err
	}
	b, err := st.ReplaySession(args[1])
	if err != nil {
		return err
	}

	fmt.Printf("Session A (%s): %d events\n", args[0], len(a))
	fmt.Printf("Session B (%s): %d events\n", args[1], len(b))

	countsA := countByType(a)
	countsB := countByType(b)
	fmt.Println("\nEvent counts by type:")
	for _, typ := range []schema.EventType{
		schema.EventLLMCall, schema.EventFileEdit, schema.EventShellCmd, schema.EventMCPCall, schema.EventToolCall,
	} {
		if countsA[typ] == 0 && countsB[typ] == 0 {
			continue
		}
		fmt.Printf("  %-14s  A=%d  B=%d\n", typ, countsA[typ], countsB[typ])
	}

	filesA := uniqueFiles(a)
	filesB := uniqueFiles(b)
	overlap := intersectSets(filesA, filesB)
	fmt.Printf("\nFiles touched: A=%d  B=%d  overlap=%d\n", len(filesA), len(filesB), len(overlap))

	costA, toolsA, llmsA, _ := st.SessionSpend(args[0], time.Time{})
	costB, toolsB, llmsB, _ := st.SessionSpend(args[1], time.Time{})
	fmt.Printf("\nSpend:\n  A: cost=$%.4f tools=%d llm=%d\n", costA, toolsA, llmsA)
	fmt.Printf("  B: cost=$%.4f tools=%d llm=%d\n", costB, toolsB, llmsB)

	modelsA, effA := llmBreakdown(a)
	modelsB, effB := llmBreakdown(b)
	fmt.Println("\nLLM models:")
	allModels := maps.Clone(modelsA)
	for m, c := range modelsB {
		allModels[m] += c
	}
	for model := range allModels {
		fmt.Printf("  %-24s  A=%d  B=%d\n", model, modelsA[model], modelsB[model])
	}
	if effA > 0 || effB > 0 {
		fmt.Printf("\nAvg efficiency:  A=%.0f%%  B=%.0f%%  delta=%+.0f%%\n", effA, effB, effB-effA)
	}
	return nil
}

func countByType(events []schema.Event) map[schema.EventType]int {
	out := make(map[schema.EventType]int)
	for _, e := range events {
		out[e.Type]++
	}
	return out
}

func uniqueFiles(events []schema.Event) map[string]struct{} {
	out := make(map[string]struct{})
	for _, e := range events {
		if e.Type == schema.EventFileEdit && e.FilePath != "" {
			out[e.FilePath] = struct{}{}
		}
	}
	return out
}

func intersectSets(a, b map[string]struct{}) map[string]struct{} {
	out := make(map[string]struct{})
	for k := range a {
		if _, ok := b[k]; ok {
			out[k] = struct{}{}
		}
	}
	return out
}

func llmBreakdown(events []schema.Event) (models map[string]int, avgEff float64) {
	models = make(map[string]int)
	var effSum float64
	var effCount int
	for _, e := range events {
		if e.Type != schema.EventLLMCall {
			continue
		}
		model := e.Model
		if model == "" {
			model = "(unknown)"
		}
		models[model]++
		if e.EfficiencyScore > 0 {
			effSum += e.EfficiencyScore
			effCount++
		}
	}
	if effCount > 0 {
		avgEff = effSum / float64(effCount)
	}
	return models, avgEff
}

func eventSummary(e schema.Event) string {
	switch e.Type {
	case schema.EventLLMCall:
		return fmt.Sprintf("model=%s tokens=%d", e.Model, e.PromptTokens+e.OutputTokens)
	case schema.EventFileEdit:
		return e.FilePath
	case schema.EventShellCmd:
		return e.ShellCommand
	default:
		return string(e.Type)
	}
}

func init() {
	traceCmd.AddCommand(traceListCmd)
}
