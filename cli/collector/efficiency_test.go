package collector

import (
	"testing"
)

func TestScoreEfficiency(t *testing.T) {
	prompt := "The quick brown fox jumps over the lazy dog near the river bank."
	output := "Summary: the quick brown fox jumps over the lazy dog today."
	score := ScoreEfficiency(prompt, output)
	if score <= 0 {
		t.Fatalf("expected positive score, got %v", score)
	}
	if score > 100 {
		t.Fatalf("score capped at 100, got %v", score)
	}
}

func TestScoreEfficiencyEmpty(t *testing.T) {
	if ScoreEfficiency("", "out") != 0 {
		t.Fatal("empty prompt should score 0")
	}
}
