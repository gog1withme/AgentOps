package hooks

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/gog1withme/AgentOps/cli/internal/paths"
)

// ApplyEnv loads proxy environment variables into the current process.
func ApplyEnv(port int) error {
	base := "http://127.0.0.1:" + itoa(port) + "/proxy"
	_ = os.Setenv("OPENAI_BASE_URL", base)
	_ = os.Setenv("ANTHROPIC_BASE_URL", base)
	_ = os.Setenv("AGENTOPS_PORT", itoa(port))
	_ = os.Setenv("AGENTOPS_ENABLED", "1")
	return setProxyEnv(port)
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

// EnvFilePath returns the platform-specific env snippet path.
func EnvFilePath() string {
	if runtime.GOOS == "windows" {
		return filepath.Join(paths.AgentOpsDir(), "env.ps1")
	}
	return filepath.Join(paths.AgentOpsDir(), "env.sh")
}

// EnvSourceCommand returns the shell command to load proxy env vars.
func EnvSourceCommand() string {
	path := EnvFilePath()
	if runtime.GOOS == "windows" {
		return ". " + path
	}
	return "source " + path
}

// HookInstalled checks whether the agentops marker exists in shell profile.
func HookInstalled() bool {
	home, _ := os.UserHomeDir()
	var paths []string
	if runtime.GOOS == "windows" {
		paths = append(paths, filepath.Join(home, "Documents", "PowerShell", "Microsoft.PowerShell_profile.ps1"))
	} else {
		paths = append(paths,
			filepath.Join(home, ".bashrc"),
			filepath.Join(home, ".zshrc"),
			filepath.Join(home, ".profile"),
		)
	}
	for _, p := range paths {
		data, err := os.ReadFile(p)
		if err == nil && strings.Contains(string(data), marker) {
			return true
		}
	}
	return false
}

// EnvFileConfigured checks env file contains OPENAI_BASE_URL.
func EnvFileConfigured() bool {
	data, err := os.ReadFile(EnvFilePath())
	if err != nil {
		return false
	}
	return strings.Contains(string(data), "OPENAI_BASE_URL")
}
