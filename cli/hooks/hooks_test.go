package hooks

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestSanitizeName(t *testing.T) {
	if got := SanitizeName("my server"); got != "my_server" {
		t.Fatalf("expected my_server, got %q", got)
	}
}

func TestPatchMCPFilePreservesBackupAndUpstream(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "mcp.json")
	original := `{
  "mcpServers": {
    "My Server": {
      "url": "https://example.com/mcp"
    }
  }
}`
	if err := os.WriteFile(path, []byte(original), 0o644); err != nil {
		t.Fatal(err)
	}

	if err := patchMCPFile(path, 4318); err != nil {
		t.Fatal(err)
	}

	backupPath := path + ".agentops.bak"
	backup, err := os.ReadFile(backupPath)
	if err != nil {
		t.Fatal(err)
	}
	if string(backup) != original {
		t.Fatal("backup should preserve original config")
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var parsed struct {
		MCPServers map[string]map[string]any `json:"mcpServers"`
	}
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatal(err)
	}
	cfg := parsed.MCPServers["My Server"]
	if cfg["_agentops_upstream"] != "https://example.com/mcp" {
		t.Fatalf("expected upstream preserved, got %v", cfg["_agentops_upstream"])
	}
	url, _ := cfg["url"].(string)
	if url != "http://127.0.0.1:4318/mcp/My_Server" {
		t.Fatalf("unexpected proxied url %q", url)
	}

	// Patching again should not overwrite backup.
	if err := patchMCPFile(path, 4318); err != nil {
		t.Fatal(err)
	}
	backup2, err := os.ReadFile(backupPath)
	if err != nil {
		t.Fatal(err)
	}
	if string(backup2) != original {
		t.Fatal("backup should remain original after second patch")
	}
}

func TestApplyEnvWritesEnvFile(t *testing.T) {
	home := t.TempDir()
	t.Setenv("USERPROFILE", home)
	t.Setenv("HOME", home)

	if err := ApplyEnv(4318); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(EnvFilePath())
	if err != nil {
		t.Fatal(err)
	}
	content := string(data)
	if !strings.Contains(content, "OPENAI_BASE_URL") {
		t.Fatalf("expected OPENAI_BASE_URL in env file: %s", content)
	}
	if !strings.Contains(content, "4318") {
		t.Fatalf("expected port in env file: %s", content)
	}
	if !EnvFileConfigured() {
		t.Fatal("EnvFileConfigured should be true after ApplyEnv")
	}
}
