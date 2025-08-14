package llm

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type GenerateRequest struct {
	Model       string  `json:"model"`
	Prompt      string  `json:"prompt"`
	Temperature float64 `json:"temperature"`
	Seed        int     `json:"seed"`
	Stream      bool    `json:"stream"`
}

type GenerateResponse struct {
	Model           string `json:"model"`
	Response        string `json:"response"`
	TotalDuration   int64  `json:"total_duration"`
	PromptEvalCount int    `json:"prompt_eval_count"`
	EvalCount       int    `json:"eval_count"`
}

type Client struct {
	baseURL    string
	httpClient *http.Client
}

func NewClient(baseURL string) *Client {
	return &Client{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 60 * time.Second,
		},
	}
}

func (c *Client) Generate(ctx context.Context, prompt, model string) (*GenerateResponse, string, string, error) {
	promptHash := hashString(prompt)

	req := GenerateRequest{
		Model:       model,
		Prompt:      prompt,
		Temperature: 0,
		Seed:        42,
		Stream:      false,
	}

	body, err := json.Marshal(req)
	if err != nil {
		return nil, "", "", err
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+"/api/generate", bytes.NewReader(body))
	if err != nil {
		return nil, "", "", err
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, "", "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, "", "", fmt.Errorf("ollama error %d: %s", resp.StatusCode, string(body))
	}

	var genResp GenerateResponse
	if err := json.NewDecoder(resp.Body).Decode(&genResp); err != nil {
		return nil, "", "", err
	}

	outputHash := hashString(genResp.Response)

	return &genResp, promptHash, outputHash, nil
}

func hashString(s string) string {
	h := sha256.Sum256([]byte(s))
	return hex.EncodeToString(h[:])
}
