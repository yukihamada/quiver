package stream

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/libp2p/go-libp2p/core/network"
	"github.com/quiver/provider/pkg/llm"
	"github.com/quiver/provider/pkg/metrics"
	"github.com/quiver/provider/pkg/receipt"
	"github.com/sirupsen/logrus"
	"golang.org/x/time/rate"
)

type Request struct {
	Prompt    string `json:"prompt"`
	Model     string `json:"model"`
	MaxTokens int    `json:"max_tokens"`
}

type Response struct {
	Completion string                 `json:"completion"`
	Receipt    *receipt.SignedReceipt `json:"receipt"`
	Error      string                 `json:"error,omitempty"`
}

type Handler struct {
	llmClient      *llm.Client
	signer         *receipt.Signer
	maxPromptBytes int
	limiter        *rate.Limiter
	logger         *logrus.Logger
	
	// Protected by mu
	mu       sync.Mutex
	prevHash string
	sequence int64
}

func NewHandler(llmClient *llm.Client, signer *receipt.Signer, maxPromptBytes int, tokensPerSecond int) *Handler {
	logger := logrus.New()
	logger.SetFormatter(&logrus.JSONFormatter{})

	return &Handler{
		llmClient:      llmClient,
		signer:         signer,
		maxPromptBytes: maxPromptBytes,
		limiter:        rate.NewLimiter(rate.Limit(tokensPerSecond), tokensPerSecond*2),
		logger:         logger,
		prevHash:       "",
		sequence:       0,
	}
}

func (h *Handler) HandleStream(s network.Stream) {
	defer s.Close()
	metrics.ActiveStreams.Inc()
	defer metrics.ActiveStreams.Dec()

	var req Request
	decoder := json.NewDecoder(s)
	if err := decoder.Decode(&req); err != nil {
		h.sendError(s, "invalid request")
		return
	}

	if len(req.Prompt) > h.maxPromptBytes {
		h.sendError(s, "prompt exceeds size limit")
		return
	}

	// Use a longer timeout for LLM generation
	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()

	if err := h.limiter.Wait(ctx); err != nil {
		metrics.RateLimitHits.Inc()
		h.sendError(s, "rate limit exceeded")
		metrics.RequestsTotal.WithLabelValues(req.Model, "rate_limited").Inc()
		return
	}

	start := time.Now()

	llmResp, promptHash, outputHash, err := h.llmClient.Generate(ctx, req.Prompt, req.Model)
	if err != nil {
		h.sendError(s, fmt.Sprintf("llm error: %v", err))
		metrics.RequestsTotal.WithLabelValues(req.Model, "error").Inc()
		return
	}

	end := time.Now()
	duration := end.Sub(start).Seconds()
	metrics.RequestDuration.WithLabelValues(req.Model).Observe(duration)

	// Protect sequence counter and prevHash with mutex
	h.mu.Lock()
	h.sequence++
	currentSeq := h.sequence
	currentPrevHash := h.prevHash
	h.mu.Unlock()

	rcpt := receipt.NewReceipt(
		h.signer.PublicKeyBase64(),
		req.Model,
		promptHash,
		outputHash,
		llmResp.PromptEvalCount,
		llmResp.EvalCount,
		start,
		end,
	)
	rcpt.PrevHash = currentPrevHash
	rcpt.Seq = currentSeq

	canonical, err := receipt.CanonicalizeJSON(rcpt)
	if err != nil {
		h.sendError(s, "failed to canonicalize receipt")
		return
	}
	
	// Update prevHash under lock
	h.mu.Lock()
	h.prevHash = receipt.HashData(canonical)
	h.mu.Unlock()

	signedReceipt, err := h.signer.Sign(rcpt)
	if err != nil {
		h.sendError(s, "failed to sign receipt")
		metrics.RequestsTotal.WithLabelValues(req.Model, "sign_error").Inc()
		return
	}
	metrics.ReceiptSignatures.Inc()

	h.logger.WithFields(logrus.Fields{
		"prompt_hash": promptHash,
		"output_hash": outputHash,
		"tokens_in":   llmResp.PromptEvalCount,
		"tokens_out":  llmResp.EvalCount,
		"duration_ms": end.Sub(start).Milliseconds(),
		"receipt_id":  rcpt.ReceiptID,
	}).Info("request processed")

	resp := Response{
		Completion: llmResp.Response,
		Receipt:    signedReceipt,
	}

	encoder := json.NewEncoder(s)
	if err := encoder.Encode(resp); err != nil {
		h.logger.WithError(err).Error("failed to encode response")
		metrics.RequestsTotal.WithLabelValues(req.Model, "encode_error").Inc()
	} else {
		metrics.RequestsTotal.WithLabelValues(req.Model, "success").Inc()
		metrics.TokensProcessed.WithLabelValues("input").Add(float64(llmResp.PromptEvalCount))
		metrics.TokensProcessed.WithLabelValues("output").Add(float64(llmResp.EvalCount))
	}
}

func (h *Handler) sendError(s network.Stream, msg string) {
	resp := Response{Error: msg}
	encoder := json.NewEncoder(s)
	if err := encoder.Encode(resp); err != nil {
		h.logger.WithError(err).Error("failed to encode error response")
	}
}
