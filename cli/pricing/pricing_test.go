package pricing

import "testing"

func TestEstimateCostKnownModel(t *testing.T) {
	p := &Pricing{
		models: []ModelRate{
			{Pattern: "gpt-4o-mini", InputPerToken: 0.00000015, OutputPerToken: 0.0000006},
			{Pattern: "default", InputPerToken: 0.000003, OutputPerToken: 0.000015},
		},
		unknownModels: make(map[string]struct{}),
	}
	got := p.EstimateCost("gpt-4o-mini", 1000, 500)
	want := 1000*0.00000015 + 500*0.0000006
	if got != want {
		t.Fatalf("EstimateCost = %v, want %v", got, want)
	}
}

func TestEstimateCostUnknownModelUsesDefault(t *testing.T) {
	p := &Pricing{
		models: []ModelRate{
			{Pattern: "gpt-4o-mini", InputPerToken: 0.00000015, OutputPerToken: 0.0000006},
			{Pattern: "default", InputPerToken: 0.000003, OutputPerToken: 0.000015},
		},
		unknownModels: make(map[string]struct{}),
	}
	got := p.EstimateCost("custom-model-v9", 100, 100)
	want := 100*0.000003 + 100*0.000015
	if got != want {
		t.Fatalf("EstimateCost = %v, want %v", got, want)
	}
	if len(p.UnknownModels()) != 1 {
		t.Fatalf("expected unknown model tracked, got %v", p.UnknownModels())
	}
}

func TestLoadDefaults(t *testing.T) {
	p, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if p.ModelCount() < 5 {
		t.Fatalf("expected bundled pricing models, got %d", p.ModelCount())
	}
}
