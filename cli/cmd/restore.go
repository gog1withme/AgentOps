package cmd

import (
	"fmt"
	"time"

	"github.com/gog1withme/AgentOps/cli/internal/config"
	"github.com/gog1withme/AgentOps/cli/snapshots"
	"github.com/gog1withme/AgentOps/cli/store"
	"github.com/spf13/cobra"
)

var restoreCmd = &cobra.Command{
	Use:   "restore [run_id]",
	Short: "List or restore snapshots",
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

		sessionID := cfg.SessionID
		if len(args) > 0 {
			sessionID = args[0]
		}
		at, _ := cmd.Flags().GetString("at")
		snapshotID, _ := cmd.Flags().GetString("snapshot")
		dryRun, _ := cmd.Flags().GetBool("dry-run")
		workDir := cfg.WorkDir
		if workDir == "" {
			workDir = "."
		}

		if at != "" {
			ts, err := time.Parse(time.RFC3339, at)
			if err != nil {
				return err
			}
			return snapshots.RestoreAt(st, sessionID, ts, workDir, dryRun)
		}
		if snapshotID != "" {
			return snapshots.Restore(st, sessionID, snapshotID, workDir, dryRun)
		}

		snaps, err := st.ListSnapshots(sessionID)
		if err != nil {
			return err
		}
		if len(snaps) == 0 {
			fmt.Println("No snapshots found.")
			return nil
		}
		fmt.Printf("Snapshots for session %s:\n", sessionID)
		for _, s := range snaps {
			fmt.Printf("  %s  %s  trigger=%s  files=%d\n", s.ID, s.Timestamp.Format(time.RFC3339), s.Trigger, s.FileCount)
		}
		return nil
	},
}

func init() {
	restoreCmd.Flags().String("at", "", "Restore to closest snapshot before timestamp (RFC3339)")
	restoreCmd.Flags().String("snapshot", "", "Restore exact snapshot id")
	restoreCmd.Flags().Bool("dry-run", false, "Show what would change")
}
