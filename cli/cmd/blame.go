package cmd

import (
	"fmt"
	"os"
	"text/tabwriter"
	"time"

	"github.com/gog1withme/AgentOps/cli/store"
	"github.com/spf13/cobra"
)

var blameCmd = &cobra.Command{
	Use:   "blame [file]",
	Short: "Show AI attribution log for a file",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		st, err := store.Open()
		if err != nil {
			return err
		}
		defer st.Close()
		rows, err := st.ListAttribution(args[0])
		if err != nil {
			return err
		}
		if len(rows) == 0 {
			fmt.Println("No AI edits recorded for this file.")
			return nil
		}
		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, "TIMESTAMP\tAGENT\t+LINES\t-LINES")
		for _, r := range rows {
			fmt.Fprintf(w, "%s\t%s\t%d\t%d\n",
				r.Timestamp.Format(time.RFC3339), r.Agent, r.LinesAdded, r.LinesRemoved)
		}
		return w.Flush()
	},
}
