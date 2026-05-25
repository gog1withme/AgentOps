package hooks

import (
	"encoding/json"
	"os"
	"strings"
)

// LoadMCPServerUpstreams returns sanitized server name -> upstream URL from patched MCP configs.
func LoadMCPServerUpstreams() map[string]string {
	out := make(map[string]string)
	for _, path := range MCPConfigPaths() {
		data, err := os.ReadFile(path)
		if err != nil {
			continue
		}
		var raw struct {
			MCPServers map[string]map[string]any `json:"mcpServers"`
		}
		if err := json.Unmarshal(data, &raw); err != nil {
			continue
		}
		for name, cfg := range raw.MCPServers {
			upstream := mcpUpstreamURL(cfg)
			if upstream == "" {
				continue
			}
			out[SanitizeName(name)] = upstream
		}
	}
	return out
}

func mcpUpstreamURL(cfg map[string]any) string {
	if u, ok := cfg["_agentops_upstream"].(string); ok && u != "" {
		return u
	}
	if u, ok := cfg["url"].(string); ok && u != "" && !strings.Contains(u, "127.0.0.1") && !strings.Contains(u, "localhost") {
		return u
	}
	return ""
}

// URLBasedMCPServerCount counts MCP servers with external URLs in config files.
func URLBasedMCPServerCount() int {
	n := 0
	for _, path := range MCPConfigPaths() {
		data, err := os.ReadFile(path)
		if err != nil {
			continue
		}
		var raw struct {
			MCPServers map[string]map[string]any `json:"mcpServers"`
		}
		if err := json.Unmarshal(data, &raw); err != nil {
			continue
		}
		for _, cfg := range raw.MCPServers {
			if u, ok := cfg["url"].(string); ok && u != "" {
				n++
			}
		}
	}
	return n
}

func PatchedMCPServerCount() int {
	n := 0
	for _, path := range MCPConfigPaths() {
		data, err := os.ReadFile(path)
		if err != nil {
			continue
		}
		var raw struct {
			MCPServers map[string]map[string]any `json:"mcpServers"`
		}
		if err := json.Unmarshal(data, &raw); err != nil {
			continue
		}
		for _, cfg := range raw.MCPServers {
			if _, ok := cfg["_agentops_upstream"]; ok {
				n++
			}
		}
	}
	return n
}
