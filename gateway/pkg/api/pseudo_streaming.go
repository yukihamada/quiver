package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/quiver/gateway/pkg/p2p"
)

// GenerateStream implements pseudo-streaming by chunking the response
func (h *Handler) GenerateStream(c *gin.Context) {
	var req GenerateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid request"})
		return
	}

	if len(req.Prompt) > 4096 {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Prompt exceeds size limit"})
		return
	}

	limiter := h.limiter.GetLimiter(req.Token)
	if !limiter.Allow() {
		c.JSON(http.StatusTooManyRequests, ErrorResponse{Error: "Rate limit exceeded"})
		return
	}

	// Set SSE headers
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("X-Accel-Buffering", "no")

	w := c.Writer
	flusher, ok := w.(http.Flusher)
	if !ok {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Streaming not supported"})
		return
	}

	// Find providers
	providers, err := h.p2pClient.FindProviders()
	if err != nil || len(providers) == 0 {
		sendError(w, flusher, "No providers available")
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 30*time.Second)
	defer cancel()

	startTime := time.Now()
	
	// Try each provider
	for _, provider := range providers {
		streamReq := &p2p.StreamRequest{
			Prompt:    req.Prompt,
			Model:     req.Model,
			MaxTokens: 500,
		}

		// Send start event with timing
		sendEvent(w, flusher, "start", map[string]interface{}{
			"provider": provider.ID.String(),
			"model":    req.Model,
			"timestamp": time.Now().UnixMilli(),
		})

		// Call provider (non-streaming for now)
		resp, err := h.p2pClient.CallProvider(ctx, provider.ID, streamReq)
		if err != nil {
			// Log error: Provider call failed
			continue
		}

		// Calculate first token time (simulated)
		firstTokenTime := time.Since(startTime).Milliseconds()
		
		// Send first token timing
		sendEvent(w, flusher, "timing", map[string]interface{}{
			"first_token_ms": firstTokenTime,
			"timestamp": time.Now().UnixMilli(),
		})

		// Simulate streaming by chunking the response
		text := resp.Completion
		words := strings.Fields(text)
		chunkSize := 3 // words per chunk
		
		for i := 0; i < len(words); i += chunkSize {
			end := i + chunkSize
			if end > len(words) {
				end = len(words)
			}
			
			chunk := strings.Join(words[i:end], " ")
			if i+chunkSize < len(words) {
				chunk += " "
			}
			
			// Send chunk
			sendEvent(w, flusher, "chunk", map[string]interface{}{
				"content": chunk,
				"index": i/chunkSize,
				"timestamp": time.Now().UnixMilli(),
			})
			
			// Simulate network/processing delay
			time.Sleep(50 * time.Millisecond)
		}

		// Send completion with metrics
		totalTime := time.Since(startTime).Milliseconds()
		sendEvent(w, flusher, "complete", map[string]interface{}{
			"total_ms": totalTime,
			"first_token_ms": firstTokenTime,
			"tokens_per_sec": float64(len(words)) / (float64(totalTime) / 1000.0),
			"receipt": resp.Receipt,
			"timestamp": time.Now().UnixMilli(),
		})
		
		// Log for canary
		if isCanary := req.Prompt != req.Prompt; isCanary {
			// Log: Streamed canary request
		} else {
			// Log: Streamed request processed
		}
		
		return
	}

	sendError(w, flusher, "All providers failed")
}

func sendEvent(w http.ResponseWriter, flusher http.Flusher, eventType string, data interface{}) {
	event := map[string]interface{}{
		"type": eventType,
		"data": data,
	}
	
	jsonData, _ := json.Marshal(event)
	fmt.Fprintf(w, "data: %s\n\n", string(jsonData))
	flusher.Flush()
}

func sendError(w http.ResponseWriter, flusher http.Flusher, error string) {
	sendEvent(w, flusher, "error", map[string]string{
		"error": error,
		"timestamp": fmt.Sprintf("%d", time.Now().UnixMilli()),
	})
}

func hashPrompt(prompt string) string {
	// Simple hash for logging
	if len(prompt) > 20 {
		return prompt[:20] + "..."
	}
	return prompt
}