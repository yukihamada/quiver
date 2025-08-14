package llm

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestDeterministicGeneration(t *testing.T) {
	// Mock Ollama server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if _, err := w.Write([]byte(`{
			"model": "test-model",
			"response": "Deterministic response",
			"total_duration": 50000000,
			"prompt_eval_count": 5,
			"eval_count": 3
		}`)); err != nil {
			t.Logf("Failed to write response: %v", err)
		}
	}))
	defer server.Close()

	client := NewClient(server.URL)

	// Make same request multiple times
	hashes := make(map[string]bool)

	for i := 0; i < 5; i++ {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		resp, promptHash, outputHash, err := client.Generate(ctx, "test prompt", "test-model")

		if err != nil {
			t.Fatal(err)
		}

		if resp.Response != "Deterministic response" {
			t.Error("Response not deterministic")
		}

		// Hashes should be consistent
		key := promptHash + ":" + outputHash
		hashes[key] = true
	}

	if len(hashes) != 1 {
		t.Error("Hashes not deterministic across requests")
	}
}

func TestHashConsistency(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{
			input:    "test",
			expected: "9f86d081884c7d659a2feaa0c55ad015a3bf4f1b2b0b822cd15d6c15b0f00a08",
		},
		{
			input:    "",
			expected: "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855",
		},
	}

	for _, tt := range tests {
		got := hashString(tt.input)
		if got != tt.expected {
			t.Errorf("hashString(%q) = %s, want %s", tt.input, got, tt.expected)
		}
	}
}
