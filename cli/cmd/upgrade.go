package cmd

import (
	"fmt"

	"github.com/gog1withme/AgentOps/cli/upgrade"
	"github.com/spf13/cobra"
)

var upgradeCheck bool
var upgradeForce bool

var upgradeCmd = &cobra.Command{
	Use:   "upgrade",
	Short: "Update AgentOps from GitHub Releases",
	RunE: func(cmd *cobra.Command, args []string) error {
		info, err := upgrade.Check()
		if err != nil {
			return err
		}
		if upgradeCheck {
			if info.UpdateAvailable {
				fmt.Printf("Update available: %s → %s\nRun `agentops upgrade` to install.\n", info.CurrentVersion, info.LatestVersion)
			} else {
				fmt.Printf("AgentOps %s is up to date.\n", info.CurrentVersion)
			}
			return nil
		}
		if !info.UpdateAvailable && !upgradeForce {
			fmt.Printf("AgentOps %s is already up to date.\n", info.CurrentVersion)
			return nil
		}
		fmt.Printf("Upgrading %s → %s...\n", info.CurrentVersion, info.LatestVersion)
		if err := upgrade.Upgrade(upgradeForce); err != nil {
			return err
		}
		fmt.Printf("Upgraded to %s successfully.\n", info.LatestVersion)
		return nil
	},
}

func init() {
	upgradeCmd.Flags().BoolVar(&upgradeCheck, "check", false, "Check for updates without installing")
	upgradeCmd.Flags().BoolVar(&upgradeForce, "force", false, "Reinstall the latest release even if already current")
}
