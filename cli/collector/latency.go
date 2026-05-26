package collector

import "sort"

const mcpLatencyWindowSize = 100

func appendLatencyWindow(window []int, latencyMS int) []int {
	window = append(window, latencyMS)
	if len(window) > mcpLatencyWindowSize {
		window = window[len(window)-mcpLatencyWindowSize:]
	}
	return window
}

func computeP95(latencies []int) float64 {
	if len(latencies) == 0 {
		return 0
	}
	sorted := make([]int, len(latencies))
	copy(sorted, latencies)
	sort.Ints(sorted)
	idx := int(float64(len(sorted))*0.95 + 0.9999999) - 1
	if idx < 0 {
		idx = 0
	}
	if idx >= len(sorted) {
		idx = len(sorted) - 1
	}
	return float64(sorted[idx])
}
