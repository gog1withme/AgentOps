package collector

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gog1withme/AgentOps/cli/store"
	"github.com/gog1withme/AgentOps/schema"
	"github.com/rs/zerolog/log"
)

type Proxy struct {
	collector         *Collector
	openAIUpstream    string
	anthropicUpstream string
	client            *http.Client
}

func NewProxy(c *Collector, openAIUpstream string) *Proxy {
	if openAIUpstream == "" {
		openAIUpstream = os.Getenv("OPENAI_UPSTREAM")
	}
	if openAIUpstream == "" {
		openAIUpstream = "https://api.openai.com"
	}
	anthropicUpstream := os.Getenv("ANTHROPIC_UPSTREAM")
	if anthropicUpstream == "" {
		anthropicUpstream = "https://api.anthropic.com"
	}
	return &Proxy{
		collector:         c,
		openAIUpstream:    strings.TrimRight(openAIUpstream, "/"),
		anthropicUpstream: strings.TrimRight(anthropicUpstream, "/"),
		client:            &http.Client{Timeout: 120 * time.Second},
	}
}

func (p *Proxy) Handler(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	r.Body.Close()

	isAnthropic := strings.Contains(r.URL.Path, "/messages") || strings.Contains(r.Header.Get("anthropic-version"), "-")
	systemPrompt, model, promptTokens, outputTokens := extractLLMInfo(body, isAnthropic)
	promptHash := hashPrompt(systemPrompt)
	if promptHash != "" && systemPrompt != "" {
		now := time.Now()
		_ = p.collector.store.UpsertPrompt(&schema.Prompt{
			Hash:            promptHash,
			Content:         systemPrompt,
			FirstSeen:       now,
			LastSeen:        now,
			SessionCount:    1,
			AvgPromptTokens: promptTokens,
		})
	}

	upstream := p.openAIUpstream
	if isAnthropic {
		upstream = p.anthropicUpstream
	}
	upURL := upstream + r.URL.Path
	if r.URL.RawQuery != "" {
		upURL += "?" + r.URL.RawQuery
	}
	req, err := http.NewRequestWithContext(r.Context(), r.Method, upURL, bytes.NewReader(body))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	copyHeaders(r.Header, req.Header)
	resp, err := p.client.Do(req)
	if err != nil {
		p.emitLLMEvent(model, promptHash, body, "", promptTokens, outputTokens, start, err, isAnthropic)
		http.Error(w, err.Error(), http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()
	respBody, _ := io.ReadAll(resp.Body)
	copyHeaders(resp.Header, w.Header())
	w.WriteHeader(resp.StatusCode)

	pt, ot, cost := promptTokens, outputTokens, 0.0
	if resp.StatusCode < 400 {
		w.Write(respBody)
		pt, ot, cost = parseUsageFromResponse(respBody, model, isAnthropic, promptTokens, outputTokens)
	} else {
		w.Write(respBody)
	}

	outputText := extractOutputFromResponse(respBody, isAnthropic)
	promptText := string(body)
	eff := ScoreEfficiency(promptText, outputText)
	errMsg := ""
	if resp.StatusCode >= 400 {
		errMsg = resp.Status
	}
	p.collector.Ingest(schema.Event{
		ID:              store.NewEventID(),
		SessionID:       p.collector.cfg.SessionID,
		Timestamp:       time.Now(),
		Source:          detectSource(r),
		Type:            schema.EventLLMCall,
		Model:           model,
		PromptHash:      promptHash,
		PromptTokens:    pt,
		OutputTokens:    ot,
		EfficiencyScore: eff,
		CostUSD:         cost,
		DurationMS:      int(time.Since(start).Milliseconds()),
		ToolInput:       promptText,
		ToolOutput:      outputText,
		Error:           errMsg,
	})
}

func (p *Proxy) emitLLMEvent(model, promptHash string, reqBody []byte, out string, pt, ot int, start time.Time, err error, isAnthropic bool) {
	errMsg := ""
	if err != nil {
		errMsg = err.Error()
	}
	_ = isAnthropic
	p.collector.Ingest(schema.Event{
		ID:           store.NewEventID(),
		SessionID:    p.collector.cfg.SessionID,
		Timestamp:    time.Now(),
		Source:       "proxy",
		Type:         schema.EventLLMCall,
		Model:        model,
		PromptHash:   promptHash,
		PromptTokens: pt,
		OutputTokens: ot,
		DurationMS:   int(time.Since(start).Milliseconds()),
		ToolInput:    string(reqBody),
		ToolOutput:   out,
		Error:        errMsg,
	})
}

func extractLLMInfo(body []byte, anthropic bool) (systemPrompt, model string, promptTokens, outputTokens int) {
	if anthropic {
		return extractAnthropicPromptInfo(body)
	}
	return extractOpenAIPromptInfo(body)
}

func extractOpenAIPromptInfo(body []byte) (systemPrompt, model string, promptTokens, outputTokens int) {
	var req map[string]any
	if err := json.Unmarshal(body, &req); err != nil {
		return "", "", 0, 0
	}
	if m, ok := req["model"].(string); ok {
		model = m
	}
	msgs, ok := req["messages"].([]any)
	if !ok {
		return "", model, 0, 0
	}
	var parts []string
	for _, m := range msgs {
		msg, ok := m.(map[string]any)
		if !ok {
			continue
		}
		if role, _ := msg["role"].(string); role == "system" {
			if content, ok := msg["content"].(string); ok {
				parts = append(parts, content)
			}
		}
	}
	return strings.Join(parts, "\n"), model, 0, 0
}

func extractAnthropicPromptInfo(body []byte) (systemPrompt, model string, promptTokens, outputTokens int) {
	var req map[string]any
	if err := json.Unmarshal(body, &req); err != nil {
		return "", "", 0, 0
	}
	if m, ok := req["model"].(string); ok {
		model = m
	}
	switch sys := req["system"].(type) {
	case string:
		systemPrompt = sys
	case []any:
		for _, part := range sys {
			if block, ok := part.(map[string]any); ok {
				if text, ok := block["text"].(string); ok {
					systemPrompt += text + "\n"
				}
			}
		}
	}
	return strings.TrimSpace(systemPrompt), model, 0, 0
}

func parseUsageFromResponse(body []byte, model string, anthropic bool, fallbackPT, fallbackOT int) (promptTokens, outputTokens int, cost float64) {
	if anthropic {
		return parseAnthropicUsage(body, model)
	}
	return parseUsage(body, model)
}

func parseAnthropicUsage(body []byte, model string) (promptTokens, outputTokens int, cost float64) {
	var resp map[string]any
	if err := json.Unmarshal(body, &resp); err != nil {
		return 0, 0, 0
	}
	if usage, ok := resp["usage"].(map[string]any); ok {
		if v, ok := usage["input_tokens"].(float64); ok {
			promptTokens = int(v)
		}
		if v, ok := usage["output_tokens"].(float64); ok {
			outputTokens = int(v)
		}
	}
	cost = estimateCost(model, promptTokens, outputTokens)
	return
}

func extractOutputFromResponse(body []byte, anthropic bool) string {
	if anthropic {
		return extractAnthropicOutputText(body)
	}
	return extractOutputText(body)
}

func extractAnthropicOutputText(body []byte) string {
	var resp map[string]any
	if err := json.Unmarshal(body, &resp); err != nil {
		return string(body)
	}
	content, ok := resp["content"].([]any)
	if !ok {
		return ""
	}
	var parts []string
	for _, c := range content {
		block, ok := c.(map[string]any)
		if !ok {
			continue
		}
		if text, ok := block["text"].(string); ok {
			parts = append(parts, text)
		}
	}
	return strings.Join(parts, "\n")
}

func hashPrompt(content string) string {
	if content == "" {
		return ""
	}
	sum := sha256.Sum256([]byte(content))
	return hex.EncodeToString(sum[:])
}

func parseUsage(body []byte, model string) (promptTokens, outputTokens int, cost float64) {
	var resp map[string]any
	if err := json.Unmarshal(body, &resp); err != nil {
		return 0, 0, 0
	}
	usage, ok := resp["usage"].(map[string]any)
	if !ok {
		return 0, 0, 0
	}
	if v, ok := usage["prompt_tokens"].(float64); ok {
		promptTokens = int(v)
	}
	if v, ok := usage["completion_tokens"].(float64); ok {
		outputTokens = int(v)
	}
	cost = estimateCost(model, promptTokens, outputTokens)
	return
}

func estimateCost(model string, pt, ot int) float64 {
	inRate, outRate := 0.000003, 0.000015
	if strings.Contains(model, "gpt-4") || strings.Contains(model, "claude-3") {
		inRate, outRate = 0.00001, 0.00003
	}
	return float64(pt)*inRate + float64(ot)*outRate
}

func extractOutputText(body []byte) string {
	var resp map[string]any
	if err := json.Unmarshal(body, &resp); err != nil {
		return string(body)
	}
	choices, ok := resp["choices"].([]any)
	if !ok || len(choices) == 0 {
		return ""
	}
	choice, ok := choices[0].(map[string]any)
	if !ok {
		return ""
	}
	msg, ok := choice["message"].(map[string]any)
	if !ok {
		return ""
	}
	content, _ := msg["content"].(string)
	return content
}

func copyHeaders(from, to http.Header) {
	for k, vals := range from {
		if strings.EqualFold(k, "Host") {
			continue
		}
		for _, v := range vals {
			to.Add(k, v)
		}
	}
}

func detectSource(r *http.Request) string {
	ua := r.Header.Get("User-Agent")
	if strings.Contains(strings.ToLower(ua), "claude") {
		return "claude"
	}
	if strings.Contains(strings.ToLower(ua), "cursor") {
		return "cursor"
	}
	return "openai_sdk"
}

func (p *Proxy) Listen(addr string) error {
	mux := http.NewServeMux()
	mux.HandleFunc("/", p.Handler)
	log.Info().Str("addr", addr).Msg("LLM proxy listening")
	return http.ListenAndServe(addr, mux)
}
