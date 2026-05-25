package platform

import (
	"strings"

	"github.com/shirou/gopsutil/v4/process"
)

func AttributeFileAgent(filePath string) string {
	procs, err := process.Processes()
	if err != nil {
		return "unknown"
	}
	lower := strings.ToLower(filePath)
	for _, p := range procs {
		name, err := p.Name()
		if err != nil {
			continue
		}
		ln := strings.ToLower(name)
		if strings.Contains(ln, "claude") {
			return "claude"
		}
		if strings.Contains(ln, "cursor") {
			return "cursor"
		}
		if strings.Contains(ln, "copilot") {
			return "copilot"
		}
		if strings.Contains(ln, "aider") {
			return "aider"
		}
		_ = lower
	}
	return "unknown"
}

func FindAgentPID(source string) int {
	procs, err := process.Processes()
	if err != nil {
		return 0
	}
	needle := strings.ToLower(source)
	for _, p := range procs {
		name, err := p.Name()
		if err != nil {
			continue
		}
		if strings.Contains(strings.ToLower(name), needle) {
			return int(p.Pid)
		}
	}
	return 0
}
