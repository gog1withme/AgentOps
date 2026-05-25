package cmd

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/gog1withme/AgentOps/cli/store"
	"github.com/sergi/go-diff/diffmatchpatch"
	"github.com/spf13/cobra"
)

var promptCmd = &cobra.Command{
	Use:   "prompt",
	Short: "Manage prompt versions",
}

var promptListCmd = &cobra.Command{
	Use:   "list",
	Short: "List captured prompt versions",
	RunE: func(cmd *cobra.Command, args []string) error {
		st, err := store.Open()
		if err != nil {
			return err
		}
		defer st.Close()
		prompts, err := st.ListPrompts()
		if err != nil {
			return err
		}
		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, "HASH\tSESSIONS\tAVG_COST\tAVG_TOKENS\tEFFICIENCY")
		for _, p := range prompts {
			hash := p.Hash
			if len(hash) > 8 {
				hash = hash[:8]
			}
			fmt.Fprintf(w, "%s\t%d\t$%.4f\t%d\t%.1f%%\n",
				hash, p.SessionCount, p.AvgCostUSD, p.AvgPromptTokens, p.AvgEfficiencyScore)
		}
		return w.Flush()
	},
}

var promptDiffCmd = &cobra.Command{
	Use:   "diff [hash_a] [hash_b]",
	Short: "Diff two prompt versions",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		st, err := store.Open()
		if err != nil {
			return err
		}
		defer st.Close()
		a, err := st.GetPrompt(resolveHash(st, args[0]))
		if err != nil {
			return err
		}
		b, err := st.GetPrompt(resolveHash(st, args[1]))
		if err != nil {
			return err
		}
		dmp := diffmatchpatch.New()
		diffs := dmp.DiffMain(a.Content, b.Content, false)
		fmt.Println(dmp.DiffPrettyText(diffs))
		return nil
	},
}

func resolveHash(st *store.Store, prefix string) string {
	prompts, err := st.ListPrompts()
	if err != nil {
		return prefix
	}
	for _, p := range prompts {
		if p.Hash == prefix || (len(p.Hash) >= len(prefix) && p.Hash[:len(prefix)] == prefix) {
			return p.Hash
		}
	}
	return prefix
}

func init() {
	promptCmd.AddCommand(promptListCmd)
	promptCmd.AddCommand(promptDiffCmd)
}
