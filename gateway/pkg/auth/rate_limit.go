package auth

import (
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
)

// PlanLimits defines rate limits for different subscription plans
type PlanLimits struct {
	RequestsPerSecond int
	RequestsPerMonth  int
	MaxTokensPerReq   int
	BurstSize         int
}

var PlanLimitMap = map[string]PlanLimits{
	"free": {
		RequestsPerSecond: 1,
		RequestsPerMonth:  10000,
		MaxTokensPerReq:   1000,
		BurstSize:         5,
	},
	"starter": {
		RequestsPerSecond: 10,
		RequestsPerMonth:  100000,
		MaxTokensPerReq:   2000,
		BurstSize:         20,
	},
	"pro": {
		RequestsPerSecond: 50,
		RequestsPerMonth:  1000000,
		MaxTokensPerReq:   4000,
		BurstSize:         100,
	},
	"enterprise": {
		RequestsPerSecond: 500,
		RequestsPerMonth:  -1, // Unlimited
		MaxTokensPerReq:   8000,
		BurstSize:         1000,
	},
}

type RateLimiter struct {
	limiters map[string]*rate.Limiter
	mu       sync.RWMutex
}

func NewRateLimiter() *RateLimiter {
	return &RateLimiter{
		limiters: make(map[string]*rate.Limiter),
	}
}

// RateLimitMiddleware enforces rate limits based on user plan
func (rl *RateLimiter) RateLimitMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := c.Get("user_id")
		if !exists {
			c.Next()
			return
		}

		plan, _ := c.Get("plan")
		planStr, _ := plan.(string)
		if planStr == "" {
			planStr = "free"
		}

		limits, ok := PlanLimitMap[planStr]
		if !ok {
			limits = PlanLimitMap["free"]
		}

		limiter := rl.getLimiter(userID.(string), limits)
		if !limiter.Allow() {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error": "Rate limit exceeded",
				"retry_after": "1s",
			})
			c.Abort()
			return
		}

		// Set max tokens in context for request validation
		c.Set("max_tokens_allowed", limits.MaxTokensPerReq)
		c.Next()
	}
}

func (rl *RateLimiter) getLimiter(userID string, limits PlanLimits) *rate.Limiter {
	rl.mu.RLock()
	limiter, exists := rl.limiters[userID]
	rl.mu.RUnlock()

	if !exists {
		rl.mu.Lock()
		limiter = rate.NewLimiter(rate.Limit(limits.RequestsPerSecond), limits.BurstSize)
		rl.limiters[userID] = limiter
		rl.mu.Unlock()
	}

	return limiter
}