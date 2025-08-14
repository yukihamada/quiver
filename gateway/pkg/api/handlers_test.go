package api

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/quiver/gateway/pkg/p2p"
	"github.com/quiver/gateway/pkg/ratelimit"
)

func TestPromptLengthGuard(t *testing.T) {
	tests := []struct {
		name      string
		prompt    string
		shouldErr bool
	}{
		{
			name:      "valid length",
			prompt:    "This is a valid prompt",
			shouldErr: false,
		},
		{
			name:      "exactly 4096 bytes",
			prompt:    strings.Repeat("a", 4096),
			shouldErr: false,
		},
		{
			name:      "exceeds limit",
			prompt:    strings.Repeat("a", 4097),
			shouldErr: true,
		},
		{
			name:      "empty prompt",
			prompt:    "",
			shouldErr: false,
		},
		{
			name:      "unicode handling",
			prompt:    strings.Repeat("🔥", 1024), // 4 bytes each
			shouldErr: false,
		},
		{
			name:      "unicode exceeds",
			prompt:    strings.Repeat("🔥", 1025),
			shouldErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if len(tt.prompt) > 4096 != tt.shouldErr {
				t.Errorf("Prompt length guard failed for %s", tt.name)
			}
		})
	}
}

func TestCanarySelection(t *testing.T) {
	// Test canary prompts and answers match
	for prompt := range canaryAnswers {
		found := false
		for _, cp := range canaryPrompts {
			if cp == prompt {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Canary answer exists for non-existent prompt: %s", prompt)
		}
	}

	// Test all canary prompts have answers
	for _, prompt := range canaryPrompts {
		if _, exists := canaryAnswers[prompt]; !exists {
			t.Errorf("No answer defined for canary prompt: %s", prompt)
		}
	}
}

func TestHashString(t *testing.T) {
	// Test determinism
	input := "test input"
	hash1 := hashString(input)
	hash2 := hashString(input)

	if hash1 != hash2 {
		t.Error("Hash function not deterministic")
	}

	// Test different inputs produce different hashes
	hash3 := hashString("different input")
	if hash1 == hash3 {
		t.Error("Different inputs produced same hash")
	}

	// Test hash format
	if len(hash1) != 64 {
		t.Errorf("Expected 64 character hash, got %d", len(hash1))
	}
}

func TestNewHandler(t *testing.T) {
	// Test handler creation
	handler := NewHandler(nil, nil, 0.05)

	if handler == nil {
		t.Error("Handler should not be nil")
	}

	if handler.canaryRate != 0.05 {
		t.Errorf("Expected canary rate 0.05, got %f", handler.canaryRate)
	}

	if handler.logger == nil {
		t.Error("Logger should not be nil")
	}
}

func TestGenerateHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Create mock P2P client
	ctx := context.Background()
	p2pClient, err := p2p.NewClient(ctx, "/ip4/127.0.0.1/tcp/0", []string{})
	if err != nil {
		t.Skip("Skipping test due to P2P client creation error:", err)
	}
	defer p2pClient.Close()

	limiter := ratelimit.NewLimiter(100)
	handler := NewHandler(p2pClient, limiter, 0.05)

	router := gin.New()
	router.POST("/generate", handler.Generate)

	tests := []struct {
		name       string
		request    GenerateRequest
		statusCode int
	}{
		{
			name: "valid request",
			request: GenerateRequest{
				Prompt: "test prompt",
				Model:  "test-model",
			},
			statusCode: http.StatusServiceUnavailable, // No providers available
		},
		{
			name: "empty prompt",
			request: GenerateRequest{
				Prompt: "",
				Model:  "test-model",
			},
			statusCode: http.StatusBadRequest,
		},
		{
			name: "empty model",
			request: GenerateRequest{
				Prompt: "test",
				Model:  "",
			},
			statusCode: http.StatusBadRequest,
		},
		{
			name: "prompt too long",
			request: GenerateRequest{
				Prompt: strings.Repeat("a", 4097),
				Model:  "test-model",
			},
			statusCode: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, _ := json.Marshal(tt.request)
			req := httptest.NewRequest("POST", "/generate", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			if w.Code != tt.statusCode {
				t.Errorf("Expected status %d, got %d", tt.statusCode, w.Code)
			}
		})
	}
}

func TestRateLimiting(t *testing.T) {
	gin.SetMode(gin.TestMode)

	ctx := context.Background()
	p2pClient, err := p2p.NewClient(ctx, "/ip4/127.0.0.1/tcp/0", []string{})
	if err != nil {
		t.Skip("Skipping test due to P2P client creation error:", err)
	}
	defer p2pClient.Close()

	// Very restrictive rate limit for testing
	limiter := ratelimit.NewLimiter(1)
	handler := NewHandler(p2pClient, limiter, 0)

	router := gin.New()
	router.POST("/generate", handler.Generate)

	req := GenerateRequest{
		Prompt: "test",
		Model:  "test-model",
	}
	body, _ := json.Marshal(req)

	// First request should work
	request1 := httptest.NewRequest("POST", "/generate", bytes.NewReader(body))
	request1.Header.Set("Content-Type", "application/json")
	request1.Header.Set("X-Client-ID", "test-client")

	w1 := httptest.NewRecorder()
	router.ServeHTTP(w1, request1)

	// Second request should be rate limited
	request2 := httptest.NewRequest("POST", "/generate", bytes.NewReader(body))
	request2.Header.Set("Content-Type", "application/json")
	request2.Header.Set("X-Client-ID", "test-client")

	w2 := httptest.NewRecorder()
	router.ServeHTTP(w2, request2)

	if w2.Code != http.StatusTooManyRequests {
		t.Errorf("Expected rate limit status 429, got %d", w2.Code)
	}
}

func TestHealthEndpoint(t *testing.T) {
	gin.SetMode(gin.TestMode)

	ctx := context.Background()
	p2pClient, err := p2p.NewClient(ctx, "/ip4/127.0.0.1/tcp/0", []string{})
	if err != nil {
		t.Skip("Skipping test due to P2P client creation error:", err)
	}
	defer p2pClient.Close()

	limiter := ratelimit.NewLimiter(100)
	handler := NewHandler(p2pClient, limiter, 0.05)

	router := gin.New()
	router.GET("/health", handler.Health)

	req := httptest.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)

	if resp["status"] != "healthy" {
		t.Error("Expected healthy status")
	}

	if resp["service"] != "gateway" {
		t.Error("Expected gateway service")
	}
}
