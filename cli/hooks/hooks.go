package hooks

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/gog1withme/AgentOps/cli/internal/paths"
	"github.com/gog1withme/AgentOps/cli/internal/platform"
)
const marker = "# >>> agentops >>>"
const markerEnd = "# <<< agentops <<<"

func Install(port int) error {
	if err := installShellHook(port); err != nil {
		return err
	}
	return setProxyEnv(port)
}

func installShellHook(port int) error {
	if runtime.GOOS == "windows" {
		return installPowerShellHook(port)
	}
	return installUnixShellHook(port)
}

func installUnixShellHook(port int) error {
	home, _ := os.UserHomeDir()
	var rcPath string
	switch platform.DetectShellType() {
	case "zsh":
		rcPath = filepath.Join(home, ".zshrc")
	case "bash":
		rcPath = filepath.Join(home, ".bashrc")
	default:
		rcPath = filepath.Join(home, ".profile")
	}
	snippet := fmt.Sprintf(`
%s
export AGENTOPS_PORT=%d
export AGENTOPS_ENABLED=1
agentops_hook() {
  local exit_code=$?
  if [ -n "$AGENTOPS_ENABLED" ]; then
    curl -s -X POST "http://127.0.0.1:${AGENTOPS_PORT}/api/ingest/shell" \
      -H "Content-Type: application/json" \
      -d "{\"command\":$(printf '%%s' "$1" | python3 -c 'import json,sys; print(json.dumps(sys.stdin.read()))'),\"exit_code\":$exit_code}" >/dev/null 2>&1 || true
  fi
}
if [ -n "$PROMPT_COMMAND" ]; then
  PROMPT_COMMAND="agentops_hook \"\$BASH_COMMAND\"; $PROMPT_COMMAND"
else
  PROMPT_COMMAND='agentops_hook "$BASH_COMMAND"'
fi
%s
`, marker, port, markerEnd)
	return appendSnippet(rcPath, snippet)
}

func installPowerShellHook(port int) error {
	home, _ := os.UserHomeDir()
	profileDir := filepath.Join(home, "Documents", "PowerShell")
	if err := os.MkdirAll(profileDir, 0o755); err != nil {
		return err
	}
	profile := filepath.Join(profileDir, "Microsoft.PowerShell_profile.ps1")
	snippet := fmt.Sprintf(`
# %s
$env:AGENTOPS_PORT = "%d"
$env:AGENTOPS_ENABLED = "1"
function Global:AgentOps-Record($cmd, $code) {
  if ($env:AGENTOPS_ENABLED -ne "1" -or [string]::IsNullOrWhiteSpace($cmd)) { return }
  $port = if ($env:AGENTOPS_PORT) { $env:AGENTOPS_PORT } else { "%d" }
  $body = @{ command = $cmd; exit_code = [int]$code } | ConvertTo-Json -Compress
  try {
    Invoke-RestMethod -Uri "http://127.0.0.1:$port/api/ingest/shell" -Method Post -Body $body -ContentType "application/json" | Out-Null
  } catch {}
}
if (Get-Module -ListAvailable -Name PSReadLine) {
  Import-Module PSReadLine -ErrorAction SilentlyContinue
  Set-PSReadLineKeyHandler -Key Enter -BriefDescription 'AgentOps capture' -ScriptBlock {
    param($key, $arg)
    $line = $null
    [Microsoft.PowerShell.PSConsoleReadLine]::GetBufferState([ref]$line, [ref]$null)
    if ($line) { $Global:AgentOpsPendingLine = $line }
    [Microsoft.PowerShell.PSConsoleReadLine]::AcceptLine()
  }
}
function Global:prompt {
  if ($Global:AgentOpsPendingLine) {
    AgentOps-Record $Global:AgentOpsPendingLine $LASTEXITCODE
    $Global:AgentOpsPendingLine = $null
  }
  "PS> "
}
# %s
`, marker, port, port, markerEnd)
	return appendSnippet(profile, snippet)
}

func appendSnippet(path, snippet string) error {
	data, _ := os.ReadFile(path)
	content := string(data)
	if strings.Contains(content, marker) {
		return nil
	}
	f, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = f.WriteString("\n" + snippet + "\n")
	return err
}

func setProxyEnv(port int) error {
	home, _ := os.UserHomeDir()
	envFile := filepath.Join(home, ".agentops", "env.sh")
	base := fmt.Sprintf("http://127.0.0.1:%d/proxy", port)
	content := fmt.Sprintf(`export OPENAI_BASE_URL=%q
export ANTHROPIC_BASE_URL=%q
export AGENTOPS_PORT=%d
export AGENTOPS_ENABLED=1
`, base, base, port)
	if runtime.GOOS == "windows" {
		envFile = filepath.Join(home, ".agentops", "env.ps1")
		content = fmt.Sprintf(`$env:OPENAI_BASE_URL=%q
$env:ANTHROPIC_BASE_URL=%q
$env:AGENTOPS_PORT=%d
$env:AGENTOPS_ENABLED="1"
`, base, base, port)
	}
	if err := paths.EnsureDirs(); err != nil {
		return err
	}
	return os.WriteFile(envFile, []byte(content), 0o644)
}

func PatchMCPConfigs(port int) (int, error) {
	pathsList := MCPConfigPaths()
	patched := 0
	for _, p := range pathsList {
		if err := patchMCPFile(p, port); err == nil {
			patched++
		}
	}
	return patched, nil
}

func MCPConfigPaths() []string {
	home, _ := os.UserHomeDir()
	list := []string{
		filepath.Join(home, ".cursor", "mcp.json"),
		filepath.Join(home, ".config", "claude", "mcp.json"),
	}
	if runtime.GOOS == "windows" {
		list = append(list, filepath.Join(home, "AppData", "Roaming", "Cursor", "mcp.json"))
	}
	return list
}

func patchMCPFile(path string, port int) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	backup := path + ".agentops.bak"
	if _, err := os.Stat(backup); os.IsNotExist(err) {
		_ = os.WriteFile(backup, data, 0o644)
	}
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	serversRaw, ok := raw["mcpServers"]
	if !ok {
		return fmt.Errorf("no mcpServers")
	}
	var servers map[string]map[string]any
	if err := json.Unmarshal(serversRaw, &servers); err != nil {
		return err
	}
	for name, cfg := range servers {
		if url, ok := cfg["url"].(string); ok && url != "" {
			cfg["_agentops_upstream"] = url
			cfg["url"] = fmt.Sprintf("http://127.0.0.1:%d/mcp/%s", port, SanitizeName(name))
		}
		servers[name] = cfg
	}
	updated, err := json.MarshalIndent(map[string]any{"mcpServers": servers}, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, updated, 0o644)
}

func SanitizeName(name string) string {
	return strings.ReplaceAll(name, " ", "_")
}
