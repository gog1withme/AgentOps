package cmd

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/gog1withme/AgentOps/cli/store"
	"github.com/spf13/cobra"
)

var mcpCmd = &cobra.Command{
	Use:   "mcp",
	Short: "MCP server health",
}

var mcpListCmd = &cobra.Command{
	Use:   "list",
	Short: "List detected MCP servers and health",
	RunE: func(cmd *cobra.Command, args []string) error {
		st, err := store.Open()
		if err != nil {
			return err
		}
		defer st.Close()
		servers, err := st.ListMCPServers()
		if err != nil {
			return err
		}
		if len(servers) == 0 {
			fmt.Println("No MCP servers recorded yet.")
			return nil
		}
		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, "NAME\tCALLS\tERRORS\tAVG_MS\tP95_MS")
		for _, s := range servers {
			fmt.Fprintf(w, "%s\t%d\t%d\t%.0f\t%.0f\n",
				s.Name, s.TotalCalls, s.ErrorCount, s.AvgLatencyMS, s.P95LatencyMS)
		}
		return w.Flush()
	},
}

func init() {
	mcpCmd.AddCommand(mcpListCmd)
}
