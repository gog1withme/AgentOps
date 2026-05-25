package cmd

import (
	"fmt"
	"time"

	"github.com/gog1withme/AgentOps/cli/internal/config"
	"github.com/gog1withme/AgentOps/cli/store"
	"github.com/gog1withme/AgentOps/schema"
	"github.com/spf13/cobra"
)

var budgetCmd = &cobra.Command{
	Use:   "budget",
	Short: "Manage session budgets",
}

var budgetSetCmd = &cobra.Command{
	Use:   "set",
	Short: "Set budget limits",
	RunE: func(cmd *cobra.Command, args []string) error {
		cost, _ := cmd.Flags().GetFloat64("cost")
		tools, _ := cmd.Flags().GetInt("tools")
		llms, _ := cmd.Flags().GetInt("llms")
		action, _ := cmd.Flags().GetString("action")
		cfg, err := config.Load()
		if err != nil {
			return err
		}
		cfg.Budget.CostLimitUSD = cost
		cfg.Budget.ToolCallLimit = tools
		cfg.Budget.LLMCallLimit = llms
		cfg.Budget.Action = action
		if err := config.Save(cfg); err != nil {
			return err
		}
		st, err := store.Open()
		if err != nil {
			return err
		}
		defer st.Close()
		return st.SaveBudget(&schema.Budget{
			SessionID:     cfg.SessionID,
			CostLimitUSD:  cost,
			ToolCallLimit: tools,
			LLMCallLimit:  llms,
			Action:        action,
			CreatedAt:     time.Now(),
		})
	},
}

var budgetStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show current spend vs limits",
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
		cost, tools, llms, err := st.SessionSpend(cfg.SessionID, time.Now().Add(-24*time.Hour))
		if err != nil {
			return err
		}
		fmt.Printf("Session: %s\n", cfg.SessionID)
		fmt.Printf("Cost:    $%.4f / $%.2f\n", cost, cfg.Budget.CostLimitUSD)
		fmt.Printf("Tools:   %d / %d\n", tools, cfg.Budget.ToolCallLimit)
		fmt.Printf("LLM:     %d / %d\n", llms, cfg.Budget.LLMCallLimit)
		fmt.Printf("Action:  %s\n", cfg.Budget.Action)
		return nil
	},
}

func init() {
	budgetSetCmd.Flags().Float64("cost", 0, "Cost limit in USD")
	budgetSetCmd.Flags().Int("tools", 0, "Tool call limit")
	budgetSetCmd.Flags().Int("llms", 0, "LLM call limit")
	budgetSetCmd.Flags().String("action", "alert", "Action: alert|pause|kill")
	budgetCmd.AddCommand(budgetSetCmd)
	budgetCmd.AddCommand(budgetStatusCmd)
}
