package collector

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gog1withme/AgentOps/cli/store"
	"github.com/gog1withme/AgentOps/schema"
)

type MCPProxy struct {
	collector     *Collector
	servers       map[string]string
	client        *http.Client
	latencyMu     sync.Mutex
	latencyWindow map[string][]int
}

func NewMCPProxy(c *Collector, servers map[string]string) *MCPProxy {
	return &MCPProxy{
		collector:     c,
		servers:       servers,
		client:        &http.Client{Timeout: 60 * time.Second},
		latencyWindow: make(map[string][]int),
	}
}

func (m *MCPProxy) Handler(serverName string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		upstream, ok := m.servers[serverName]
		if !ok {
			http.Error(w, "unknown mcp server", http.StatusNotFound)
			return
		}
		body, _ := io.ReadAll(r.Body)
		r.Body.Close()

		target := upstream
		prefix := "/mcp/" + serverName
		if suffix := strings.TrimPrefix(r.URL.Path, prefix); suffix != "" && suffix != "/" {
			target = strings.TrimRight(upstream, "/") + suffix
		}
		if r.URL.RawQuery != "" {
			target += "?" + r.URL.RawQuery
		}

		method := parseMCPMethod(body)
		req, err := http.NewRequestWithContext(r.Context(), r.Method, target, bytes.NewReader(body))
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		req.Header = r.Header.Clone()
		resp, err := m.client.Do(req)
		latency := int(time.Since(start).Milliseconds())
		errMsg := ""
		respSize := 0
		if err != nil {
			errMsg = err.Error()
			m.recordCall(serverName, upstream, method, latency, errMsg, respSize)
			http.Error(w, err.Error(), http.StatusBadGateway)
			return
		}
		defer resp.Body.Close()
		respBody, _ := io.ReadAll(resp.Body)
		respSize = len(respBody)
		for k, vals := range resp.Header {
			for _, v := range vals {
				w.Header().Add(k, v)
			}
		}
		w.WriteHeader(resp.StatusCode)
		w.Write(respBody)
		if resp.StatusCode >= 400 {
			errMsg = resp.Status
		}
		m.recordCall(serverName, upstream, method, latency, errMsg, respSize)
	}
}

func (m *MCPProxy) recordCall(name, url, method string, latencyMS int, errMsg string, respSize int) {
	now := time.Now()
	errCount := 0
	if errMsg != "" {
		errCount = 1
	}
	servers, _ := m.collector.store.ListMCPServers()
	var oldAvg float64
	for _, s := range servers {
		if s.Name == name {
			oldAvg = s.AvgLatencyMS
			break
		}
	}
	newAvg := 0.9*oldAvg + 0.1*float64(latencyMS)
	m.latencyMu.Lock()
	m.latencyWindow[name] = appendLatencyWindow(m.latencyWindow[name], latencyMS)
	p95 := computeP95(m.latencyWindow[name])
	m.latencyMu.Unlock()
	_ = m.collector.store.UpsertMCPServer(&schema.MCPServer{
		Name:             name,
		URL:              url,
		FirstSeen:        now,
		LastSeen:         now,
		TotalCalls:       1,
		ErrorCount:       errCount,
		AvgLatencyMS:     newAvg,
		P95LatencyMS:     p95,
		AvgResponseBytes: respSize,
	})
	if method == "" {
		method = "mcp"
	}
	m.collector.Ingest(schema.Event{
		ID:           store.NewEventID(),
		SessionID:    m.collector.cfg.SessionID,
		Timestamp:    now,
		Source:       "mcp",
		Type:         schema.EventMCPCall,
		MCPServer:    name,
		MCPLatencyMS: latencyMS,
		ToolInput:    method,
		Error:        errMsg,
		Metadata: map[string]string{
			"response_bytes": mcpItoa(respSize),
		},
	})
}

type mcpRPC struct {
	Method string `json:"method"`
}

func parseMCPMethod(body []byte) string {
	if len(body) == 0 {
		return "mcp"
	}
	var rpc mcpRPC
	if err := json.Unmarshal(body, &rpc); err != nil {
		return "mcp"
	}
	if rpc.Method == "" {
		return "mcp"
	}
	return rpc.Method
}

func mcpItoa(n int) string {
	if n == 0 {
		return "0"
	}
	var b [20]byte
	i := len(b)
	for n > 0 {
		i--
		b[i] = byte('0' + n%10)
		n /= 10
	}
	return string(b[i:])
}

// RegisterMCPRoutes registers MCP proxy handlers on a mux (legacy helper).
func RegisterMCPRoutes(mux *http.ServeMux, proxy *MCPProxy) {
	for name := range proxy.servers {
		n := name
		path := "/mcp/" + strings.ReplaceAll(n, " ", "_")
		mux.HandleFunc(path, proxy.Handler(n))
	}
}
