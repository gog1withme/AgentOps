package collector

import (
	"strings"
	"unicode"
)

func ScoreEfficiency(promptText, outputText string) float64 {
	if promptText == "" || outputText == "" {
		return 0
	}
	chunks := splitChunks(promptText)
	if len(chunks) == 0 {
		return 0
	}
	referenced := 0
	outLower := strings.ToLower(outputText)
	for _, chunk := range chunks {
		if chunkReferenced(outLower, chunk) {
			referenced++
		}
	}
	score := float64(referenced) / float64(len(chunks)) * 100
	if score > 100 {
		return 100
	}
	if score < 0 {
		return 0
	}
	return score
}

func splitChunks(text string) []string {
	parts := strings.Split(text, "\n\n")
	if len(parts) == 1 {
		parts = strings.Split(text, "\n")
	}
	var chunks []string
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if len(p) > 20 {
			chunks = append(chunks, p)
		}
	}
	if len(chunks) == 0 && len(strings.TrimSpace(text)) > 0 {
		chunks = append(chunks, strings.TrimSpace(text))
	}
	return chunks
}

func chunkReferenced(outputLower, chunk string) bool {
	words := tokenize(chunk)
	if len(words) < 6 {
		return strings.Contains(outputLower, strings.ToLower(strings.TrimSpace(chunk)))
	}
	for i := 0; i <= len(words)-6; i++ {
		window := strings.Join(words[i:i+6], " ")
		if strings.Contains(outputLower, strings.ToLower(window)) {
			return true
		}
	}
	// fallback: any 4-word window
	for i := 0; i <= len(words)-4; i++ {
		window := strings.Join(words[i:i+4], " ")
		if strings.Contains(outputLower, strings.ToLower(window)) {
			return true
		}
	}
	return false
}

func tokenize(s string) []string {
	return strings.FieldsFunc(s, func(r rune) bool {
		return unicode.IsSpace(r) || unicode.IsPunct(r)
	})
}
