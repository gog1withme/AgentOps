package collector

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gog1withme/AgentOps/cli/internal/config"
	"github.com/gog1withme/AgentOps/cli/scrubber"
	"github.com/gog1withme/AgentOps/cli/store"
	"github.com/gog1withme/AgentOps/schema"
)

func TestProxyOpenAIEmitsLLMEvent(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("AGENTOPS_DATA_DIR", dir)

	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/chat/completions" {
			t.Fatalf("unexpected path %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"choices":[{"message":{"content":"hi"}}],"usage":{"prompt_tokens":10,"completion_tokens":5}}`)
	}))
	defer upstream.Close()

	st, err := store.Open()
	if err != nil {
		t.Fatal(err)
	}
	defer st.Close()

	cfg := config.Default()
	cfg.SessionID = "proxy-test"
	sc, err := scrubber.New("")
	if err != nil {
		t.Fatal(err)
	}
	col := New(st, sc, cfg)
	defer col.Close()

	proxy := NewProxy(col, upstream.URL)
	body := `{"model":"gpt-4o-mini","messages":[{"role":"system","content":"You are helpful"},{"role":"user","content":"hello"}]}`
	req := httptest.NewRequest(http.MethodPost, "/v1/chat/completions", strings.NewReader(body))
	rec := httptest.NewRecorder()
	proxy.Handler(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", rec.Code, rec.Body.String())
	}

	deadline := time.Now().Add(2 * time.Second)
	for {
		events, err := st.ListEvents(10, time.Now().Add(-time.Minute), "", string(schema.EventLLMCall))
		if err != nil {
			t.Fatal(err)
		}
		if len(events) > 0 {
			ev := events[0]
			if ev.PromptHash == "" {
				t.Fatal("expected prompt hash on llm_call event")
			}
			if ev.PromptTokens != 10 || ev.OutputTokens != 5 {
				t.Fatalf("unexpected token counts: pt=%d ot=%d", ev.PromptTokens, ev.OutputTokens)
			}
			return
		}
		if time.Now().After(deadline) {
			t.Fatal("timed out waiting for llm_call event")
		}
		time.Sleep(50 * time.Millisecond)
	}
}

func TestProxyAnthropicMessages(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("AGENTOPS_DATA_DIR", dir)

	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/messages" {
			t.Fatalf("unexpected path %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"content":[{"type":"text","text":"hello"}],"usage":{"input_tokens":7,"output_tokens":3}}`)
	}))
	defer upstream.Close()

	st, err := store.Open()
	if err != nil {
		t.Fatal(err)
	}
	defer st.Close()

	cfg := config.Default()
	cfg.SessionID = "anthropic-test"
	sc, _ := scrubber.New("")
	col := New(st, sc, cfg)
	defer col.Close()

	t.Setenv("ANTHROPIC_UPSTREAM", upstream.URL)
	proxy := NewProxy(col, "https://api.openai.com")
	body := `{"model":"claude-3-5-sonnet-20241022","system":"Be concise","messages":[{"role":"user","content":"hi"}]}`
	req := httptest.NewRequest(http.MethodPost, "/v1/messages", strings.NewReader(body))
	req.Header.Set("anthropic-version", "2023-06-01")
	rec := httptest.NewRecorder()
	proxy.Handler(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	var resp map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatal(err)
	}
	if _, ok := resp["content"]; !ok {
		t.Fatal("expected anthropic content in proxied response")
	}
}

func TestProxyEmitsEventOnUpstreamError(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("AGENTOPS_DATA_DIR", dir)

	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "bad request", http.StatusBadRequest)
	}))
	defer upstream.Close()

	st, err := store.Open()
	if err != nil {
		t.Fatal(err)
	}
	defer st.Close()

	cfg := config.Default()
	cfg.SessionID = "err-test"
	sc, _ := scrubber.New("")
	col := New(st, sc, cfg)
	defer col.Close()

	proxy := NewProxy(col, upstream.URL)
	body := `{"model":"gpt-4o-mini","messages":[{"role":"user","content":"x"}]}`
	req := httptest.NewRequest(http.MethodPost, "/v1/chat/completions", strings.NewReader(body))
	rec := httptest.NewRecorder()
	proxy.Handler(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 passthrough, got %d", rec.Code)
	}

	deadline := time.Now().Add(2 * time.Second)
	for {
		events, err := st.ListEvents(5, time.Now().Add(-time.Minute), "", string(schema.EventLLMCall))
		if err != nil {
			t.Fatal(err)
		}
		if len(events) > 0 && events[0].Error != "" {
			return
		}
		if time.Now().After(deadline) {
			t.Fatal("timed out waiting for llm_call error event")
		}
		time.Sleep(50 * time.Millisecond)
	}
}
