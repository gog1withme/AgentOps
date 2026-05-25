package cmd

import (
	"fmt"
	"os"

	"github.com/gog1withme/AgentOps/cli/detector"
	"github.com/gog1withme/AgentOps/cli/hooks"
	"github.com/gog1withme/AgentOps/cli/internal/config"
	"github.com/gog1withme/AgentOps/cli/internal/paths"
	"github.com/gog1withme/AgentOps/cli/scrubber"
	"github.com/spf13/cobra"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Detect tools, install hooks, activate privacy filter",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := paths.EnsureDirs(); err != nil {
			return err
		}
		cfg := config.Default()
		if existing, err := config.Load(); err == nil && existing.SessionID != "" {
			cfg = existing
		}
		tools := detector.DetectAll()
		cfg.DetectedTools = nil
		for _, t := range tools {
			if t.Found {
				cfg.DetectedTools = append(cfg.DetectedTools, t.Name)
			}
			mark := "✗"
			if t.Found {
				mark = "✓"
			}
			label := t.Name
			switch t.Name {
			case "claude":
				label = "Detected Claude Code"
			case "cursor":
				label = "Detected Cursor"
			case "copilot":
				label = "Copilot"
			case "aider":
				label = "Aider"
			case "openai_sdk":
				label = "OpenAI SDK"
			}
			detail := ""
			if t.Details != "" {
				detail = " (" + t.Details + ")"
			}
			if t.Name == "copilot" || t.Name == "aider" || t.Name == "openai_sdk" {
				if t.Found {
					fmt.Printf("%s %s%s\n", mark, label, detail)
				} else {
					fmt.Printf("%s %s not found\n", mark, label)
				}
			} else {
				fmt.Printf("%s %s%s\n", mark, label, detail)
			}
		}
		sc, err := scrubber.New(paths.ScrubPatternsPath())
		if err != nil {
			return err
		}
		fmt.Printf("✓ Shell: %s\n", detector.DetectShell())
		fmt.Printf("✓ Privacy filter: active (%d pattern rules)\n", sc.PatternCount())
		if cfg.Budget.CostLimitUSD <= 0 && cfg.Budget.ToolCallLimit <= 0 {
			fmt.Println("✓ Budget: not set (run `agentops budget set` to configure)")
		} else {
			fmt.Printf("✓ Budget: $%.2f / %d tools [action: %s]\n", cfg.Budget.CostLimitUSD, cfg.Budget.ToolCallLimit, cfg.Budget.Action)
		}
		fmt.Println("→ Installing hooks...")
		if err := hooks.Install(cfg.ProxyPort); err != nil {
			fmt.Fprintf(os.Stderr, "warning: hook install: %v\n", err)
		}
		if n, err := hooks.PatchMCPConfigs(cfg.ProxyPort); err == nil && n > 0 {
			fmt.Printf("✓ Patched %d MCP config(s)\n", n)
		}
		cfg.HooksInstalled = true
		if err := config.Save(cfg); err != nil {
			return err
		}
		fmt.Println("✓ Done. Run `agentops dev` to start.")
		fmt.Println()
		fmt.Println("Next: load proxy env vars so Cursor routes LLM calls through AgentOps:")
		fmt.Printf("  %s\n", hooks.EnvSourceCommand())
		fmt.Println("  Then restart Cursor.")
		return nil
	},
}
