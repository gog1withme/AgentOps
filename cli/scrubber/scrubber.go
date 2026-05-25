package scrubber

import (
	"os"
	"regexp"

	"github.com/gog1withme/AgentOps/schema"
)

const defaultReplace = "[REDACTED]"

type Scrubber struct {
	patterns []*regexp.Regexp
	replace  string
	redactAll bool
}

func New(patternFile string) (*Scrubber, error) {
	builtin, err := LoadBuiltinPatterns()
	if err != nil {
		return nil, err
	}
	var user []string
	if patternFile != "" {
		if data, err := os.ReadFile(patternFile); err == nil {
			user = parsePatternLines(string(data))
		}
	}
	merged := MergePatterns(builtin, user)
	var compiled []*regexp.Regexp
	for _, p := range merged {
		re, err := regexp.Compile(p)
		if err != nil {
			continue
		}
		compiled = append(compiled, re)
	}
	redactAll := os.Getenv("AGENTOPS_REDACT_CONTENT") == "1"
	return &Scrubber{patterns: compiled, replace: defaultReplace, redactAll: redactAll}, nil
}

func (s *Scrubber) PatternCount() int {
	return len(s.patterns)
}

func (s *Scrubber) ScrubString(input string) string {
	if input == "" {
		return input
	}
	if s.redactAll {
		return s.replace
	}
	out := input
	for _, re := range s.patterns {
		out = re.ReplaceAllString(out, s.replace)
	}
	return out
}

func (s *Scrubber) ScrubEvent(e *schema.Event) *schema.Event {
	if e == nil {
		return nil
	}
	copy := *e
	if copy.Metadata != nil {
		copy.Metadata = make(map[string]string, len(e.Metadata))
		for k, v := range e.Metadata {
			copy.Metadata[k] = s.ScrubString(v)
		}
	}
	copy.ToolInput = s.scrubField(e.ToolInput)
	copy.ToolOutput = s.scrubField(e.ToolOutput)
	copy.ShellCommand = s.scrubField(e.ShellCommand)
	copy.FileDiff = s.scrubField(e.FileDiff)
	copy.Error = s.scrubField(e.Error)
	return &copy
}

func (s *Scrubber) scrubField(v string) string {
	if s.redactAll && v != "" {
		return s.replace
	}
	return s.ScrubString(v)
}
