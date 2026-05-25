package cmd

import (
	"fmt"
	"runtime"

	"github.com/gog1withme/AgentOps/cli/hooks"
	"github.com/gog1withme/AgentOps/cli/internal/config"
	"github.com/spf13/cobra"
)

var envCmd = &cobra.Command{
	Use:   "env",
	Short: "Print command to load proxy environment variables",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			return err
		}
		port := cfg.ProxyPort
		if port == 0 {
			port = config.Port()
		}
		_ = hooks.ApplyEnv(port)

		fmt.Println("Run this command in your shell, then restart Cursor/Claude:")
		fmt.Println()
		fmt.Printf("  %s\n", hooks.EnvSourceCommand())
		fmt.Println()
		if runtime.GOOS == "windows" {
			fmt.Println("Or set permanently for this session before launching Cursor:")
			fmt.Printf("  $env:OPENAI_BASE_URL=\"http://127.0.0.1:%d/proxy\"\n", port)
			fmt.Printf("  $env:ANTHROPIC_BASE_URL=\"http://127.0.0.1:%d/proxy\"\n", port)
		} else {
			fmt.Printf("  export OPENAI_BASE_URL=http://127.0.0.1:%d/proxy\n", port)
			fmt.Printf("  export ANTHROPIC_BASE_URL=http://127.0.0.1:%d/proxy\n", port)
		}
		fmt.Println()
		fmt.Println("Restart Cursor after sourcing so it picks up the proxy URLs.")
		return nil
	},
}
