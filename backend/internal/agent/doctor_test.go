package agent

import "testing"

func TestNormalizeOpenAIBaseURL(t *testing.T) {
	tests := map[string]string{
		"https://llm.example.test":     "https://llm.example.test/v1",
		"https://llm.example.test/":    "https://llm.example.test/v1",
		"https://llm.example.test/v1":  "https://llm.example.test/v1",
		"https://llm.example.test/api": "https://llm.example.test/api/v1",
	}

	for input, expected := range tests {
		if got := normalizeOpenAIBaseURL(input); got != expected {
			t.Fatalf("normalizeOpenAIBaseURL(%q) = %q, want %q", input, got, expected)
		}
	}
}
