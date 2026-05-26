package pricing

import (
	"embed"
	"encoding/json"
	"os"
	"strings"
	"sync"

	"github.com/gog1withme/AgentOps/cli/internal/paths"
)

//go:embed defaults.json
var defaultPricing embed.FS

type ModelRate struct {
	Pattern        string  `json:"pattern"`
	InputPerToken  float64 `json:"input_per_token"`
	OutputPerToken float64 `json:"output_per_token"`
}

type configFile struct {
	Models []ModelRate `json:"models"`
}

type Pricing struct {
	models        []ModelRate
	unknownMu     sync.Mutex
	unknownModels map[string]struct{}
}

var (
	globalMu sync.RWMutex
	global   *Pricing
)

func Load() (*Pricing, error) {
	defaults, err := loadDefaults()
	if err != nil {
		return nil, err
	}
	userPath := paths.PricingPath()
	if data, err := os.ReadFile(userPath); err == nil {
		var user configFile
		if err := json.Unmarshal(data, &user); err == nil && len(user.Models) > 0 {
			defaults = user.Models
		}
	}
	return &Pricing{
		models:        defaults,
		unknownModels: make(map[string]struct{}),
	}, nil
}

func Global() *Pricing {
	globalMu.RLock()
	if global != nil {
		p := global
		globalMu.RUnlock()
		return p
	}
	globalMu.RUnlock()

	globalMu.Lock()
	defer globalMu.Unlock()
	if global == nil {
		p, err := Load()
		if err != nil {
			global = fallbackPricing()
		} else {
			global = p
		}
	}
	return global
}

func Reload() error {
	p, err := Load()
	if err != nil {
		return err
	}
	globalMu.Lock()
	global = p
	globalMu.Unlock()
	return nil
}

func loadDefaults() ([]ModelRate, error) {
	data, err := defaultPricing.ReadFile("defaults.json")
	if err != nil {
		return nil, err
	}
	var cfg configFile
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}
	return cfg.Models, nil
}

func fallbackPricing() *Pricing {
	return &Pricing{
		models: []ModelRate{
			{Pattern: "default", InputPerToken: 0.000003, OutputPerToken: 0.000015},
		},
		unknownModels: make(map[string]struct{}),
	}
}

func (p *Pricing) ModelCount() int {
	return len(p.models)
}

func (p *Pricing) UnknownModels() []string {
	p.unknownMu.Lock()
	defer p.unknownMu.Unlock()
	out := make([]string, 0, len(p.unknownModels))
	for m := range p.unknownModels {
		out = append(out, m)
	}
	return out
}

func (p *Pricing) EstimateCost(model string, promptTokens, outputTokens int) float64 {
	rate, matched := p.matchRate(model)
	if !matched {
		p.unknownMu.Lock()
		if model != "" {
			p.unknownModels[model] = struct{}{}
		}
		p.unknownMu.Unlock()
	}
	if rate == nil {
		rate = p.matchRateDefault()
	}
	if rate == nil {
		return float64(promptTokens)*0.000003 + float64(outputTokens)*0.000015
	}
	return float64(promptTokens)*rate.InputPerToken + float64(outputTokens)*rate.OutputPerToken
}

func (p *Pricing) matchRate(model string) (*ModelRate, bool) {
	modelLower := strings.ToLower(model)
	var fallback *ModelRate
	for i := range p.models {
		r := &p.models[i]
		pattern := strings.ToLower(r.Pattern)
		if pattern == "default" {
			fallback = r
			continue
		}
		if modelLower == pattern || strings.Contains(modelLower, pattern) {
			return r, true
		}
	}
	return fallback, false
}

func (p *Pricing) matchRateDefault() *ModelRate {
	for i := range p.models {
		if strings.EqualFold(p.models[i].Pattern, "default") {
			return &p.models[i]
		}
	}
	return nil
}
