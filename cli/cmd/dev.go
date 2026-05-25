package cmd

import (
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"

	"github.com/gog1withme/AgentOps/cli/collector"
	"github.com/gog1withme/AgentOps/cli/detector"
	"github.com/gog1withme/AgentOps/cli/hooks"
	"github.com/gog1withme/AgentOps/cli/internal/config"
	"github.com/gog1withme/AgentOps/cli/internal/paths"
	"github.com/gog1withme/AgentOps/cli/scrubber"
	"github.com/gog1withme/AgentOps/cli/server"
	"github.com/gog1withme/AgentOps/cli/snapshots"
	"github.com/gog1withme/AgentOps/cli/store"
	"github.com/spf13/cobra"
)

var devApplyEnv bool

var devCmd = &cobra.Command{
	Use:   "dev",
	Short: "Start collector daemon and dashboard",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			return err
		}
		if cfg.SessionID == "" {
			cfg = config.Default()
		}
		workDir, _ := os.Getwd()
		cfg.WorkDir = workDir
		_ = config.Save(cfg)

		st, err := store.Open()
		if err != nil {
			return err
		}
		defer st.Close()

		sc, err := scrubber.New(paths.ScrubPatternsPath())
		if err != nil {
			return err
		}
		c := collector.New(st, sc, cfg)
		defer c.Close()

		snapMgr, err := snapshots.InitSession(cfg.SessionID, workDir, st)
		if err != nil {
			return fmt.Errorf("snapshots: %w", err)
		}
		defer snapMgr.Close()
		c.SetSnapshotCallback(snapMgr.OnFileEdit)

		fw, err := collector.NewFSWatcher(c, workDir)
		if err == nil {
			_ = fw.Start()
			defer fw.Close()
		}

		port := config.Port()
		addr := fmt.Sprintf("127.0.0.1:%d", port)

		if devApplyEnv {
			_ = hooks.ApplyEnv(port)
			fmt.Println("✓ Applied proxy env vars to this process")
		}

		dashboardDir := server.DashboardOutDir()
		srv := server.New(c, st, workDir, snapMgr, dashboardDir)
		go func() {
			if err := srv.Listen(addr); err != nil {
				fmt.Fprintf(os.Stderr, "server error: %v\n", err)
			}
		}()

		if os.Getenv("AGENTOPS_NO_BROWSER") != "1" {
			openBrowser("http://" + addr)
		}

		printDevBanner(cfg, sc.PatternCount(), addr, workDir, srv.MCPServers())

		sig := make(chan os.Signal, 1)
		signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	<-sig
		return nil
	},
}

func init() {
	devCmd.Flags().BoolVar(&devApplyEnv, "apply-env", false, "Set OPENAI_BASE_URL and ANTHROPIC_BASE_URL in this process")
}

func printDevBanner(cfg *config.Config, rules int, addr, workDir string, mcpServers []string) {
	fmt.Println("  agentops dev")
	fmt.Println("  ─────────────────────────────────────────")
	fmt.Printf("  Dashboard  →  http://%s\n", addr)
	fmt.Println("  Collector  →  listening")
	fmt.Printf("  Proxy      →  http://%s/proxy\n", addr)
	fmt.Println()
	fmt.Printf("  Watching:    %s\n", formatTools(cfg.DetectedTools))
	fmt.Printf("  Privacy:     active (%d rules)\n", rules)
	if cfg.Budget.CostLimitUSD > 0 || cfg.Budget.ToolCallLimit > 0 {
		fmt.Printf("  Budget:      $%.2f / %d tools  [action: %s]\n", cfg.Budget.CostLimitUSD, cfg.Budget.ToolCallLimit, cfg.Budget.Action)
	} else {
		fmt.Println("  Budget:      not configured")
	}
	fmt.Printf("  Snapshots:   %s\n", filepath.Join(paths.SnapshotsDir(), cfg.SessionID))
	if len(mcpServers) > 0 {
		fmt.Printf("  MCP proxy:   %d server(s): %s\n", len(mcpServers), strings.Join(mcpServers, ", "))
	} else {
		fmt.Println("  MCP proxy:   no URL-based MCP servers configured")
	}
	fmt.Printf("  Env setup:   run `agentops env` and restart Cursor\n")
	fmt.Println()
	fmt.Println("  Press Ctrl+C to stop.")
}

func formatTools(tools []string) string {
	if len(tools) == 0 {
		return detector.DetectShell() + ", filesystem"
	}
	out := ""
	for i, t := range tools {
		if i > 0 {
			out += ", "
		}
		out += t
	}
	out += ", shell, filesystem"
	return out
}

func openBrowser(url string) {
	var c *exec.Cmd
	switch runtime.GOOS {
	case "windows":
		c = exec.Command("cmd", "/c", "start", url)
	case "darwin":
		c = exec.Command("open", url)
	default:
		c = exec.Command("xdg-open", url)
	}
	_ = c.Start()
}

func healthCheck(addr string) {
	_, _ = http.Get("http://" + addr + "/api/stats")
}
