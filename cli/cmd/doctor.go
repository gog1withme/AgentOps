package cmd

import (
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"time"

	"github.com/gog1withme/AgentOps/cli/detector"
	"github.com/gog1withme/AgentOps/cli/hooks"
	"github.com/gog1withme/AgentOps/cli/internal/config"
	"github.com/gog1withme/AgentOps/cli/internal/paths"
	"github.com/gog1withme/AgentOps/cli/scrubber"
	"github.com/gog1withme/AgentOps/cli/server"
	"github.com/gog1withme/AgentOps/cli/store"
	"github.com/spf13/cobra"
)

var doctorVerbose bool

var doctorCmd = &cobra.Command{
	Use:   "doctor",
	Short: "Diagnose integration health",
	RunE: func(cmd *cobra.Command, args []string) error {
		ok := true
		check := func(name string, pass bool, detail string) {
			mark := "✓"
			if !pass {
				mark = "✗"
				ok = false
			}
			if detail != "" {
				fmt.Printf("%s %s: %s\n", mark, name, detail)
			} else {
				fmt.Printf("%s %s\n", mark, name)
			}
		}
		hint := func(msg string) {
			if doctorVerbose {
				fmt.Printf("    → %s\n", msg)
			}
		}

		check("Platform", true, runtime.GOOS+"/"+runtime.GOARCH)
		_, err := scrubber.New(paths.ScrubPatternsPath())
		check("Privacy scrubber", err == nil, "")
		if err != nil && doctorVerbose {
			hint("Run `agentops init` to create default scrub patterns.")
		}
		check("Data directory", paths.EnsureDirs() == nil, paths.DataDir())

		if _, err := exec.LookPath("git"); err != nil {
			check("Git", false, "git not found in PATH")
		} else {
			check("Git", true, "")
		}

		tools := detector.DetectAll()
		found := 0
		for _, t := range tools {
			if t.Found {
				found++
			}
		}
		check("Agent tools", found > 0, fmt.Sprintf("%d detected", found))

		check("Shell", true, detector.DetectShell())
		check("Shell hook", hooks.HookInstalled(), "")
		if !hooks.HookInstalled() && doctorVerbose {
			hint("Run `agentops init` to install the PowerShell/bash hook.")
		}

		check("Env file", hooks.EnvFileConfigured(), hooks.EnvFilePath())
		if !hooks.EnvFileConfigured() && doctorVerbose {
			hint("Run `agentops init` then `" + hooks.EnvSourceCommand() + "` and restart Cursor.")
		}

		urlBased := hooks.URLBasedMCPServerCount()
		mcpCount := hooks.PatchedMCPServerCount()
		upstreams := hooks.LoadMCPServerUpstreams()
		mcpOK := urlBased == 0 || mcpCount > 0 || len(upstreams) > 0
		check("MCP configs", mcpOK, fmt.Sprintf("%d patched, %d upstream(s), %d URL-based", mcpCount, len(upstreams), urlBased))
		if !mcpOK && doctorVerbose {
			hint("Run `agentops init` to patch URL-based MCP servers in Cursor mcp.json.")
		} else if urlBased == 0 && doctorVerbose {
			hint("Optional: add URL-based MCP servers to Cursor mcp.json, then run `agentops init`.")
		}

		dashboardDir := server.DashboardOutDir()
		dashboardBuilt := dashboardDir != "" && fileExists(filepath.Join(dashboardDir, "index.html"))
		check("Dashboard build", dashboardBuilt, dashboardDir)
		if !dashboardBuilt && doctorVerbose {
			hint("Run `build.ps1 dashboard-build` (Windows) or `cd dashboard && npm run build`.")
		}

		st, err := store.Open()
		if err != nil {
			check("Store", false, err.Error())
		} else {
			defer st.Close()
			check("Store backend", true, store.BackendName())
		}

		port := config.Port()
		proxyOK := doctorProbe(fmt.Sprintf("http://127.0.0.1:%d/api/stats", port))
		check("Daemon reachable", proxyOK, fmt.Sprintf("http://127.0.0.1:%d", port))
		if !proxyOK && doctorVerbose {
			hint("Start the collector with `agentops dev` before dogfooding.")
		}

		if !ok {
			if doctorVerbose {
				fmt.Println("\nSome checks failed. Fix the hints above, then re-run `agentops doctor --verbose`.")
			} else {
				fmt.Println("\nSome checks failed. Re-run with `agentops doctor --verbose` for fix hints.")
			}
			return fmt.Errorf("some checks failed")
		}
		fmt.Println("\nAll checks passed.")
		return nil
	},
}

func init() {
	doctorCmd.Flags().BoolVar(&doctorVerbose, "verbose", false, "Show actionable fix hints")
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func doctorProbe(url string) bool {
	client := &http.Client{Timeout: 500 * time.Millisecond}
	resp, err := client.Get(url)
	if err != nil {
		return false
	}
	defer resp.Body.Close()
	return resp.StatusCode == http.StatusOK
}
