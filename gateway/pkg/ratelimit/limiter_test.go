package ratelimit

import (
	"fmt"
	"sync"
	"testing"
	"time"
)

func TestRateLimiterBoundary(t *testing.T) {
	limiter := NewLimiter(10) // 10 requests per second
	token := "test-token"

	rl := limiter.GetLimiter(token)

	// Should allow burst up to 20
	allowed := 0
	for i := 0; i < 25; i++ {
		if rl.Allow() {
			allowed++
		}
	}

	if allowed != 20 {
		t.Errorf("Expected 20 allowed requests in burst, got %d", allowed)
	}

	// Wait and try again
	time.Sleep(100 * time.Millisecond)
	if !rl.Allow() {
		t.Error("Should allow request after waiting")
	}
}

func TestRateLimiterConcurrency(t *testing.T) {
	limiter := NewLimiter(100)
	token := "concurrent-token"

	var wg sync.WaitGroup
	allowed := 0
	var mu sync.Mutex

	// 50 concurrent goroutines
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			rl := limiter.GetLimiter(token)
			if rl.Allow() {
				mu.Lock()
				allowed++
				mu.Unlock()
			}
		}()
	}

	wg.Wait()

	// Should allow up to burst size (200)
	if allowed == 0 || allowed > 200 {
		t.Errorf("Unexpected allowed count: %d", allowed)
	}
}

func TestRateLimiterMultipleTokens(t *testing.T) {
	limiter := NewLimiter(5)

	tokens := []string{"token1", "token2", "token3"}

	for _, token := range tokens {
		rl := limiter.GetLimiter(token)

		// Each token should have independent limits
		allowed := 0
		for i := 0; i < 15; i++ {
			if rl.Allow() {
				allowed++
			}
		}

		if allowed != 10 { // burst = 2 * rate
			t.Errorf("Token %s: expected 10 allowed, got %d", token, allowed)
		}
	}
}

func TestRateLimiterZeroRate(t *testing.T) {
	limiter := NewLimiter(0)
	rl := limiter.GetLimiter("zero-token")

	if rl.Allow() {
		t.Error("Zero rate should not allow any requests")
	}
}

func TestLimiterCleanup(t *testing.T) {
	limiter := NewLimiter(1)

	// Create limiters
	for i := 0; i < 10; i++ {
		token := fmt.Sprintf("cleanup-token-%d", i)
		limiter.GetLimiter(token)
	}

	// Check initial count
	limiter.mu.RLock()
	initialCount := len(limiter.limiters)
	limiter.mu.RUnlock()

	if initialCount != 10 {
		t.Errorf("Expected 10 limiters, got %d", initialCount)
	}
}
