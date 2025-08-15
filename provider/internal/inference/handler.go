package inference

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// OllamaHandler handles inference requests using Ollama
type OllamaHandler struct {
	ollamaURL string
	client    *http.Client
}

// NewOllamaHandler creates a new Ollama inference handler
func NewOllamaHandler(ollamaURL string) *OllamaHandler {
	if ollamaURL == "" {
		ollamaURL = "http://localhost:11434"
	}
	
	return &OllamaHandler{
		ollamaURL: ollamaURL,
		client: &http.Client{
			Timeout: 120 * time.Second,
		},
	}
}

// InferenceRequest represents an inference request
type InferenceRequest struct {
	Prompt    string `json:"prompt"`
	Model     string `json:"model"`
	MaxTokens int    `json:"max_tokens,omitempty"`
	Stream    bool   `json:"stream,omitempty"`
}

// InferenceResponse represents an inference response
type InferenceResponse struct {
	Completion string      `json:"completion"`
	Model      string      `json:"model"`
	Error      string      `json:"error,omitempty"`
	Receipt    interface{} `json:"receipt,omitempty"`
}

// OllamaRequest represents a request to Ollama API
type OllamaRequest struct {
	Model  string `json:"model"`
	Prompt string `json:"prompt"`
	Stream bool   `json:"stream"`
}

// OllamaResponse represents a response from Ollama API
type OllamaResponse struct {
	Model              string `json:"model"`
	CreatedAt          string `json:"created_at"`
	Response           string `json:"response"`
	Done               bool   `json:"done"`
	TotalDuration      int64  `json:"total_duration,omitempty"`
	LoadDuration       int64  `json:"load_duration,omitempty"`
	PromptEvalCount    int    `json:"prompt_eval_count,omitempty"`
	PromptEvalDuration int64  `json:"prompt_eval_duration,omitempty"`
	EvalCount          int    `json:"eval_count,omitempty"`
	EvalDuration       int64  `json:"eval_duration,omitempty"`
}

// Generate performs inference using Ollama
func (h *OllamaHandler) Generate(ctx context.Context, req *InferenceRequest) (*InferenceResponse, error) {
	// Check if Ollama is available
	if !h.isOllamaAvailable() {
		return nil, fmt.Errorf("Ollama is not available")
	}

	// Prepare Ollama request
	ollamaReq := OllamaRequest{
		Model:  req.Model,
		Prompt: req.Prompt,
		Stream: false, // Non-streaming for now
	}

	reqBody, err := json.Marshal(ollamaReq)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Send request to Ollama
	httpReq, err := http.NewRequestWithContext(ctx, "POST", h.ollamaURL+"/api/generate", bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := h.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to send request to Ollama: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("Ollama returned status %d: %s", resp.StatusCode, string(body))
	}

	// Read response
	var ollamaResp OllamaResponse
	if err := json.NewDecoder(resp.Body).Decode(&ollamaResp); err != nil {
		return nil, fmt.Errorf("failed to decode Ollama response: %w", err)
	}

	// Create inference response
	return &InferenceResponse{
		Completion: ollamaResp.Response,
		Model:      ollamaResp.Model,
		Receipt: map[string]interface{}{
			"eval_count":    ollamaResp.EvalCount,
			"eval_duration": ollamaResp.EvalDuration,
			"timestamp":     time.Now().Unix(),
		},
	}, nil
}

// GenerateStream performs streaming inference using Ollama
func (h *OllamaHandler) GenerateStream(ctx context.Context, req *InferenceRequest, onChunk func(string) error) error {
	// Check if Ollama is available
	if !h.isOllamaAvailable() {
		return fmt.Errorf("Ollama is not available")
	}

	// Prepare Ollama request
	ollamaReq := OllamaRequest{
		Model:  req.Model,
		Prompt: req.Prompt,
		Stream: true,
	}

	reqBody, err := json.Marshal(ollamaReq)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	// Send request to Ollama
	httpReq, err := http.NewRequestWithContext(ctx, "POST", h.ollamaURL+"/api/generate", bytes.NewBuffer(reqBody))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := h.client.Do(httpReq)
	if err != nil {
		return fmt.Errorf("failed to send request to Ollama: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("Ollama returned status %d: %s", resp.StatusCode, string(body))
	}

	// Read streaming response
	decoder := json.NewDecoder(resp.Body)
	for {
		var chunk OllamaResponse
		if err := decoder.Decode(&chunk); err != nil {
			if err == io.EOF {
				break
			}
			return fmt.Errorf("failed to decode chunk: %w", err)
		}

		if chunk.Response != "" {
			if err := onChunk(chunk.Response); err != nil {
				return err
			}
		}

		if chunk.Done {
			break
		}
	}

	return nil
}

// isOllamaAvailable checks if Ollama is running and accessible
func (h *OllamaHandler) isOllamaAvailable() bool {
	resp, err := h.client.Get(h.ollamaURL + "/api/tags")
	if err != nil {
		return false
	}
	defer resp.Body.Close()
	return resp.StatusCode == http.StatusOK
}

// ListModels returns available models from Ollama
func (h *OllamaHandler) ListModels() ([]string, error) {
	resp, err := h.client.Get(h.ollamaURL + "/api/tags")
	if err != nil {
		return nil, fmt.Errorf("failed to get models: %w", err)
	}
	defer resp.Body.Close()

	var result struct {
		Models []struct {
			Name string `json:"name"`
		} `json:"models"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode models response: %w", err)
	}

	models := make([]string, len(result.Models))
	for i, m := range result.Models {
		models[i] = m.Name
	}

	return models, nil
}