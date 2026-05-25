package cmd

import (
	"fmt"
	"time"

	"github.com/gog1withme/AgentOps/cli/store"
	"github.com/spf13/cobra"
)

var alertsCmd = &cobra.Command{
	Use:   "alerts",
	Short: "View active alerts",
	RunE: func(cmd *cobra.Command, args []string) error {
		st, err := store.Open()
		if err != nil {
			return err
		}
		defer st.Close()
		alerts, err := st.ListAlerts(true)
		if err != nil {
			return err
		}
		if len(alerts) == 0 {
			fmt.Println("No active alerts.")
			return nil
		}
		for _, a := range alerts {
			fmt.Printf("[%s] %s (%s): %s\n",
				a.Timestamp.Format(time.RFC3339), a.Severity, a.Rule, a.Message)
		}
		return nil
	},
}
