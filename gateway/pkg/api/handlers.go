package api

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"math/rand"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/quiver/gateway/pkg/loadbalancer"
	"github.com/quiver/gateway/pkg/p2p"
	"github.com/quiver/gateway/pkg/ratelimit"
	"github.com/sirupsen/logrus"
)

type Handler struct {
	p2pClient    *p2p.Client
	limiter      *ratelimit.Limiter
	canaryRate   float64
	logger       *logrus.Logger
	loadBalancer *loadbalancer.LoadBalancer
}

var canaryPrompts = []string{
	"What is the capital of France?",
	"Calculate 2 + 2",
	"Who wrote Romeo and Juliet?",
}

var canaryAnswers = map[string]string{
	"What is the capital of France?": "Paris",
	"Calculate 2 + 2":                "4",
	"Who wrote Romeo and Juliet?":    "William Shakespeare",
}

func NewHandler(p2pClient *p2p.Client, limiter *ratelimit.Limiter, canaryRate float64) *Handler {
	logger := logrus.New()
	logger.SetFormatter(&logrus.JSONFormatter{})

	lb := loadbalancer.NewLoadBalancer()
	
	// Start cleanup goroutine
	go func() {
		ticker := time.NewTicker(5 * time.Minute)
		defer ticker.Stop()
		for range ticker.C {
			lb.Cleanup()
		}
	}()

	return &Handler{
		p2pClient:    p2pClient,
		limiter:      limiter,
		canaryRate:   canaryRate,
		logger:       logger,
		loadBalancer: lb,
	}
}

func (h *Handler) Generate(c *gin.Context) {
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

	isCanary := false
	originalPrompt := req.Prompt
	if rand.Float64() < h.canaryRate {
		req.Prompt = canaryPrompts[rand.Intn(len(canaryPrompts))]
		isCanary = true
	}

	providers, err := h.p2pClient.FindProviders()
	if err != nil || len(providers) == 0 {
		c.JSON(http.StatusServiceUnavailable, ErrorResponse{Error: "No providers available"})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 30*time.Second)
	defer cancel()

	// Use load balancer to select provider
	selectedProvider := h.loadBalancer.SelectProvider(providers)
	if selectedProvider == "" {
		selectedProvider = providers[0].ID
	}
	
	// Try selected provider first
	for i, provider := range providers {
		// Reorder to try selected provider first
		if i == 0 && provider.ID != selectedProvider {
			for j, p := range providers {
				if p.ID == selectedProvider {
					providers[0], providers[j] = providers[j], providers[0]
					break
				}
			}
		}
		
		startTime := time.Now()
		streamReq := &p2p.StreamRequest{
			Prompt:    req.Prompt,
			Model:     req.Model,
			MaxTokens: 256,
		}

		resp, err := h.p2pClient.CallProvider(ctx, provider.ID, streamReq)
		responseTime := time.Since(startTime).Seconds()
		
		// Update load balancer metrics
		h.loadBalancer.UpdateProvider(provider.ID, responseTime, err == nil)
		
		if err != nil {
			h.logger.WithError(err).Error("Provider call failed")
			continue
		}

		if isCanary {
			expected := canaryAnswers[req.Prompt]
			passed := strings.Contains(strings.ToLower(resp.Completion), strings.ToLower(expected))

			if receipt, ok := resp.Receipt.(map[string]interface{}); ok {
				if canary, ok := receipt["canary"].(map[string]interface{}); ok {
					canary["id"] = hashString(req.Prompt)
					canary["passed"] = passed
				}
			}

			resp.Completion = "Canary response hidden"
			h.logger.WithFields(logrus.Fields{
				"canary_prompt": req.Prompt,
				"passed":        passed,
			}).Info("Canary check")
		}

		h.logger.WithFields(logrus.Fields{
			"prompt_hash": hashString(originalPrompt),
			"model":       req.Model,
			"provider":    provider.ID.String(),
			"is_canary":   isCanary,
		}).Info("Request processed")

		c.JSON(http.StatusOK, GenerateResponse{
			Completion: resp.Completion,
			Receipt:    resp.Receipt,
		})
		return
	}

	c.JSON(http.StatusServiceUnavailable, ErrorResponse{Error: "All providers failed"})
}

func (h *Handler) Health(c *gin.Context) {
	healthyProviders := h.loadBalancer.GetHealthyProviders()
	
	c.JSON(http.StatusOK, gin.H{
		"status": "healthy",
		"service": "gateway",
		"timestamp": time.Now().Unix(),
		"loadbalancer": gin.H{
			"healthy_providers": len(healthyProviders),
			"providers": healthyProviders,
		},
	})
}

func hashString(s string) string {
	h := sha256.Sum256([]byte(s))
	return hex.EncodeToString(h[:])
}
