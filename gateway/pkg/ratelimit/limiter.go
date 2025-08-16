package ratelimit

import (
	"sync"
	"time"

	"golang.org/x/time/rate"
)

type Limiter struct {
	limiters map[string]*rate.Limiter
	mu       sync.RWMutex
	rps      int
}

func NewLimiter(requestsPerSecond int) *Limiter {
	return &Limiter{
		limiters: make(map[string]*rate.Limiter),
		rps:      requestsPerSecond,
	}
}

func (l *Limiter) GetLimiter(token string) *rate.Limiter {
	l.mu.RLock()
	limiter, exists := l.limiters[token]
	l.mu.RUnlock()

	if !exists {
		l.mu.Lock()
		limiter = rate.NewLimiter(rate.Limit(l.rps), l.rps*2)
		l.limiters[token] = limiter
		l.mu.Unlock()
	}

	return limiter
}

func (l *Limiter) CleanupOldLimiters() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		l.mu.Lock()
		for token, limiter := range l.limiters {
			if limiter.Tokens() == float64(l.rps*2) {
				delete(l.limiters, token)
			}
		}
		l.mu.Unlock()
	}
}

// Allow checks if a request from the given token is allowed
func (l *Limiter) Allow(token string) bool {
	limiter := l.GetLimiter(token)
	return limiter.Allow()
}
