package contextanalysis

import (
	"testing"

	"github.com/gog1withme/AgentOps/schema"
)

func TestAnalyzeEventsDuplicateContext(t *testing.T) {
	events := []schema.Event{
		{
			Type:         schema.EventLLMCall,
			PromptTokens: 5000,
			ToolInput:    "Please review src/auth.ts and src/db.go",
			EfficiencyScore: 50,
		},
		{
			Type:         schema.EventLLMCall,
			PromptTokens: 4800,
			ToolInput:    "Again look at src/auth.ts for the bug",
			EfficiencyScore: 45,
		},
	}
	a := AnalyzeEvents("sess-1", events)
	if a.LLMCallCount != 2 {
		t.Fatalf("expected 2 llm calls, got %d", a.LLMCallCount)
	}
	if len(a.DuplicateFiles) == 0 {
		t.Fatal("expected duplicate file detection")
	}
	if len(a.Callouts) == 0 {
		t.Fatal("expected callouts")
	}
}

func TestAnalyzeEventsNoisyFile(t *testing.T) {
	events := []schema.Event{
		{
			Type:            schema.EventLLMCall,
			PromptTokens:    6000,
			ToolInput:         "context from src/legacy/old.go",
			EfficiencyScore: 15,
		},
		{
			Type:            schema.EventLLMCall,
			PromptTokens:    5500,
			ToolInput:         "more src/legacy/old.go",
			EfficiencyScore: 12,
		},
	}
	a := AnalyzeEvents("sess-2", events)
	if len(a.NoisyFiles) == 0 {
		t.Fatal("expected noisy file detection")
	}
}
