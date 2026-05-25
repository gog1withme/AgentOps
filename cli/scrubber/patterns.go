package scrubber

import (
	"embed"
	"strings"
)

//go:embed patterns.txt
var defaultPatterns embed.FS

func LoadBuiltinPatterns() ([]string, error) {
	data, err := defaultPatterns.ReadFile("patterns.txt")
	if err != nil {
		return nil, err
	}
	return parsePatternLines(string(data)), nil
}

func parsePatternLines(content string) []string {
	var patterns []string
	for _, line := range strings.Split(content, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		patterns = append(patterns, line)
	}
	return patterns
}

func MergePatterns(builtin, user []string) []string {
	seen := make(map[string]struct{})
	var out []string
	for _, p := range append(builtin, user...) {
		if _, ok := seen[p]; ok {
			continue
		}
		seen[p] = struct{}{}
		out = append(out, p)
	}
	return out
}
