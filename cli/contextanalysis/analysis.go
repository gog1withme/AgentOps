package contextanalysis

import (
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/gog1withme/AgentOps/cli/store"
	"github.com/gog1withme/AgentOps/schema"
)

var filePathPattern = regexp.MustCompile(`(?:^|[\s"'` + "`" + `])((?:[\w.-]+/)+[\w.-]+\.(?:go|ts|tsx|js|jsx|py|rs|md|json|yaml|yml|sql|toml|ps1|sh|css|html|vue|rb|java|kt|cs|cpp|h|proto|env|txt))`)

type FileStat struct {
	Path            string  `json:"path"`
	Occurrences     int     `json:"occurrences"`
	TotalTokens     int     `json:"total_tokens"`
	AvgEfficiency   float64 `json:"avg_efficiency"`
	ConsecutiveHits int     `json:"consecutive_hits,omitempty"`
}

type Analysis struct {
	SessionID         string     `json:"session_id"`
	TotalPromptTokens int        `json:"total_prompt_tokens"`
	LLMCallCount      int        `json:"llm_call_count"`
	AvgEfficiency     float64    `json:"avg_efficiency"`
	DuplicateFiles    []FileStat `json:"duplicate_files"`
	NoisyFiles        []FileStat `json:"noisy_files"`
	Callouts          []string   `json:"callouts"`
}

func AnalyzeSession(st *store.Store, sessionID string, limit int) (*Analysis, error) {
	if limit <= 0 {
		limit = 200
	}
	events, err := st.ReplaySession(sessionID)
	if err != nil {
		return nil, err
	}
	return AnalyzeEvents(sessionID, filterLLMEvents(events, limit)), nil
}

func AnalyzeRecent(st *store.Store, sessionID string, since time.Time, limit int) (*Analysis, error) {
	if limit <= 0 {
		limit = 200
	}
	events, err := st.ListEvents(limit, since, "", string(schema.EventLLMCall))
	if err != nil {
		return nil, err
	}
	if sessionID != "" {
		filtered := events[:0]
		for _, e := range events {
			if e.SessionID == sessionID {
				filtered = append(filtered, e)
			}
		}
		events = filtered
	}
	return AnalyzeEvents(sessionID, events), nil
}

func AnalyzeEvents(sessionID string, llm []schema.Event) *Analysis {
	out := &Analysis{
		SessionID:      sessionID,
		DuplicateFiles: []FileStat{},
		NoisyFiles:     []FileStat{},
		Callouts:       []string{},
	}
	if len(llm) == 0 {
		return out
	}

	fileStats := make(map[string]*fileAccumulator)
	var prevPaths []string
	var totalEff float64
	var effCount int

	for _, e := range llm {
		out.TotalPromptTokens += e.PromptTokens
		if e.EfficiencyScore > 0 {
			totalEff += e.EfficiencyScore
			effCount++
		}
		paths := extractReferencedPaths(e)
		pathSet := make(map[string]struct{})
		for _, p := range paths {
			pathSet[p] = struct{}{}
			acc := fileStats[p]
			if acc == nil {
				acc = &fileAccumulator{}
				fileStats[p] = acc
			}
			acc.occurrences++
			acc.totalTokens += e.PromptTokens
			if e.EfficiencyScore > 0 {
				acc.effSum += e.EfficiencyScore
				acc.effCount++
			}
		}
		if overlapCount(prevPaths, paths) > 0 {
			for p := range pathSet {
				fileStats[p].consecutiveHits++
			}
		}
		prevPaths = paths
	}

	out.LLMCallCount = len(llm)
	if effCount > 0 {
		out.AvgEfficiency = totalEff / float64(effCount)
	}

	var stats []FileStat
	for path, acc := range fileStats {
		if acc.occurrences < 2 {
			continue
		}
		fs := FileStat{
			Path:            path,
			Occurrences:     acc.occurrences,
			TotalTokens:     acc.totalTokens,
			ConsecutiveHits: acc.consecutiveHits,
		}
		if acc.effCount > 0 {
			fs.AvgEfficiency = acc.effSum / float64(acc.effCount)
		}
		stats = append(stats, fs)
	}

	sort.Slice(stats, func(i, j int) bool {
		return stats[i].TotalTokens > stats[j].TotalTokens
	})

	for _, fs := range stats {
		if fs.ConsecutiveHits >= 1 || fs.Occurrences >= 2 {
			out.DuplicateFiles = append(out.DuplicateFiles, fs)
		}
		if fs.AvgEfficiency > 0 && fs.AvgEfficiency < 40 && fs.TotalTokens >= 1000 {
			out.NoisyFiles = append(out.NoisyFiles, fs)
		}
	}

	if len(out.DuplicateFiles) > 5 {
		out.DuplicateFiles = out.DuplicateFiles[:5]
	}
	if len(out.NoisyFiles) > 5 {
		out.NoisyFiles = out.NoisyFiles[:5]
	}

	out.Callouts = buildCallouts(out)
	return out
}

type fileAccumulator struct {
	occurrences     int
	totalTokens     int
	effSum          float64
	effCount        int
	consecutiveHits int
}

func filterLLMEvents(events []schema.Event, limit int) []schema.Event {
	var llm []schema.Event
	for _, e := range events {
		if e.Type == schema.EventLLMCall {
			llm = append(llm, e)
		}
	}
	if len(llm) > limit {
		llm = llm[len(llm)-limit:]
	}
	return llm
}

func extractReferencedPaths(e schema.Event) []string {
	seen := make(map[string]struct{})
	var paths []string
	add := func(p string) {
		p = strings.TrimSpace(p)
		if p == "" {
			return
		}
		if _, ok := seen[p]; ok {
			return
		}
		seen[p] = struct{}{}
		paths = append(paths, p)
	}
	if e.FilePath != "" {
		add(e.FilePath)
	}
	if e.Metadata != nil {
		if v := e.Metadata["referenced_files"]; v != "" {
			for _, p := range strings.Split(v, ",") {
				add(strings.TrimSpace(p))
			}
		}
	}
	text := e.ToolInput
	if text == "" {
		return paths
	}
	for _, m := range filePathPattern.FindAllStringSubmatch(text, -1) {
		if len(m) > 1 {
			add(m[1])
		}
	}
	return paths
}

func overlapCount(a, b []string) int {
	set := make(map[string]struct{}, len(a))
	for _, p := range a {
		set[p] = struct{}{}
	}
	n := 0
	for _, p := range b {
		if _, ok := set[p]; ok {
			n++
		}
	}
	return n
}

func buildCallouts(a *Analysis) []string {
	callouts := []string{}
	if a.LLMCallCount == 0 {
		return callouts
	}
	if a.AvgEfficiency > 0 && a.AvgEfficiency < 30 {
		callouts = append(callouts, "Session average token efficiency is below 30% — consider trimming context.")
	}
	for _, f := range a.DuplicateFiles {
		if f.ConsecutiveHits >= 1 {
			callouts = append(callouts, f.Path+" loaded repeatedly across consecutive calls — "+formatTokens(f.TotalTokens)+" prompt tokens total.")
			break
		}
	}
	for _, f := range a.NoisyFiles {
		callouts = append(callouts, f.Path+" averages "+formatPercent(f.AvgEfficiency)+" efficiency across "+itoa(f.Occurrences)+" calls — likely noisy context.")
		break
	}
	dupTokens := 0
	for _, f := range a.DuplicateFiles {
		dupTokens += f.TotalTokens
	}
	if dupTokens > 0 && a.TotalPromptTokens > 0 {
		pct := float64(dupTokens) / float64(a.TotalPromptTokens) * 100
		if pct >= 20 {
			callouts = append(callouts, formatPercent(pct)+" of prompt tokens came from repeatedly referenced files.")
		}
	}
	return callouts
}

func formatTokens(n int) string {
	if n >= 1000 {
		return itoa(n/1000) + "k"
	}
	return itoa(n)
}

func formatPercent(v float64) string {
	return itoa(int(v+0.5)) + "%"
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	neg := n < 0
	if neg {
		n = -n
	}
	var b [20]byte
	i := len(b)
	for n > 0 {
		i--
		b[i] = byte('0' + n%10)
		n /= 10
	}
	if neg {
		i--
		b[i] = '-'
	}
	return string(b[i:])
}
