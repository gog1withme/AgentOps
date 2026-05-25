package detector

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

type ToolResult struct {
	Name    string
	Found   bool
	Details string
}

type MCPConfig struct {
	Servers map[string]json.RawMessage `json:"mcpServers"`
}

func DetectAll() []ToolResult {
	return []ToolResult{
		DetectClaude(),
		DetectCursor(),
		DetectCopilot(),
		DetectAider(),
		DetectOpenAI(),
	}
}

func DetectClaude() ToolResult {
	r := ToolResult{Name: "claude", Found: false}
	home, _ := os.UserHomeDir()
	if _, err := os.Stat(filepath.Join(home, ".claude")); err == nil {
		r.Found = true
		r.Details = "~/.claude/"
	}
	if inPath("claude") {
		r.Found = true
		if r.Details == "" {
			r.Details = "claude in PATH"
		}
	}
	return r
}

func DetectCursor() ToolResult {
	r := ToolResult{Name: "cursor", Found: false}
	home, _ := os.UserHomeDir()
	cursorDir := filepath.Join(home, ".cursor")
	if _, err := os.Stat(cursorDir); err == nil {
		r.Found = true
		r.Details = "~/.cursor/"
	}
	if count, ok := countMCPServers(filepath.Join(cursorDir, "mcp.json")); ok {
		r.Details = strings.TrimPrefix(r.Details+"; MCP config found: "+itoa(count)+" servers", "; ")
	}
	return r
}

func DetectCopilot() ToolResult {
	r := ToolResult{Name: "copilot", Found: false}
	if inPath("gh") {
		cmd := exec.Command("gh", "copilot", "--help")
		if err := cmd.Run(); err == nil {
			r.Found = true
			r.Details = "gh copilot"
		}
	}
	home, _ := os.UserHomeDir()
	extDirs := []string{
		filepath.Join(home, ".vscode", "extensions"),
		filepath.Join(home, "AppData", "Roaming", "Code", "extensions"),
	}
	for _, d := range extDirs {
		if matches, _ := filepath.Glob(filepath.Join(d, "*copilot*")); len(matches) > 0 {
			r.Found = true
			r.Details = "VSCode extension"
			break
		}
	}
	return r
}

func DetectAider() ToolResult {
	r := ToolResult{Name: "aider", Found: false}
	if inPath("aider") {
		r.Found = true
		r.Details = "aider in PATH"
	}
	return r
}

func DetectOpenAI() ToolResult {
	r := ToolResult{Name: "openai_sdk", Found: false}
	if inPath("pip") || inPath("pip3") {
		cmd := exec.Command("pip", "show", "openai")
		if err := cmd.Run(); err == nil {
			r.Found = true
			r.Details = "pip openai"
		}
	}
	if inPath("node") {
		cmd := exec.Command("node", "-e", "require('openai')")
		if err := cmd.Run(); err == nil {
			r.Found = true
			if r.Details == "" {
				r.Details = "node openai"
			}
		}
	}
	return r
}

func DetectShell() string {
	if runtime.GOOS == "windows" {
		if os.Getenv("PSModulePath") != "" {
			return "powershell"
		}
		return "cmd"
	}
	shell := os.Getenv("SHELL")
	if strings.Contains(shell, "zsh") {
		return "zsh"
	}
	if strings.Contains(shell, "bash") {
		return "bash"
	}
	return "sh"
}

func ListMCPServerPaths() []string {
	home, _ := os.UserHomeDir()
	paths := []string{
		filepath.Join(home, ".cursor", "mcp.json"),
		filepath.Join(home, ".config", "claude", "mcp.json"),
	}
	if runtime.GOOS == "windows" {
		paths = append(paths, filepath.Join(home, "AppData", "Roaming", "Cursor", "mcp.json"))
	}
	return paths
}

func countMCPServers(path string) (int, bool) {
	data, err := os.ReadFile(path)
	if err != nil {
		return 0, false
	}
	var cfg MCPConfig
	if err := json.Unmarshal(data, &cfg); err != nil {
		return 0, false
	}
	return len(cfg.Servers), true
}

func inPath(name string) bool {
	_, err := exec.LookPath(name)
	return err == nil
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	var b [20]byte
	i := len(b)
	for n > 0 {
		i--
		b[i] = byte('0' + n%10)
		n /= 10
	}
	return string(b[i:])
}
