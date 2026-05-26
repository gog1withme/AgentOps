package contextanalysis

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/gog1withme/AgentOps/schema"
)

func TestAnalyzeEventsEmptySlicesJSON(t *testing.T) {
	a := AnalyzeEvents("sess", []schema.Event{
		{Type: schema.EventLLMCall, PromptTokens: 100, ToolInput: "hello"},
	})
	data, err := json.Marshal(a)
	if err != nil {
		t.Fatal(err)
	}
	body := string(data)
	for _, field := range []string{"duplicate_files", "noisy_files", "callouts"} {
		if strings.Contains(body, `"`+field+`":null`) {
			t.Fatalf("expected %s to be [], got null in %s", field, body)
		}
	}
}
