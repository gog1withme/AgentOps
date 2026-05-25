package scrubber

import (
	"strings"
	"testing"

	"github.com/gog1withme/AgentOps/schema"
)

func TestScrubberPatterns(t *testing.T) {
	s, err := New("")
	if err != nil {
		t.Fatal(err)
	}
	cases := []struct {
		name  string
		input string
	}{
		{"openai_key", "key=sk-1234567890abcdefghijklmnopqrstuvwxyz"},
		{"anthropic_key", "sk-ant-api03-abcdefghijklmnopqrstuvwxyz"},
		{"aws_key", "AKIAIOSFODNN7EXAMPLE"},
		{"generic_secret", `password="supersecretpassword123"`},
		{"jwt", "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIn0.dozjgNryP4J3jVmNHl0w5N_XgL0n3I9PlFUP0THsR8U"},
		{"private_key", "-----BEGIN RSA PRIVATE KEY-----\nMIIE"},
		{"env_style", "DATABASE_URL=postgres://user:pass@localhost/db"},
		{"github_token", "ghp_1234567890abcdefghijklmnopqrstuvwxyz1234"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			out := s.ScrubString(tc.input)
			if out == tc.input {
				t.Fatalf("expected redaction for %q", tc.input)
			}
			if !strings.Contains(out, "[REDACTED]") {
				t.Fatalf("expected [REDACTED] in %q", out)
			}
		})
	}
}

func TestScrubEventImmutable(t *testing.T) {
	s, _ := New("")
	e := &schema.Event{
		ToolInput: "password=supersecretpassword123",
		Metadata:  map[string]string{"token": "sk-1234567890abcdefghijklmnopqrstuvwxyz"},
	}
	scrubbed := s.ScrubEvent(e)
	if e.ToolInput == scrubbed.ToolInput {
		t.Fatal("original should not be mutated")
	}
}
