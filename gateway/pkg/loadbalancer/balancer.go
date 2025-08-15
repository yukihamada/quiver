package loadbalancer

import (
	"math/rand"
	"sync"
	"sync/atomic"
	"time"

	"github.com/libp2p/go-libp2p/core/peer"
)

type LoadBalancer struct {
	mu        sync.RWMutex
	providers []Provider
	counter   uint64
}

type Provider struct {
	ID          peer.ID
	LoadScore   float64
	LastSeen    time.Time
	SuccessRate float64
	ResponseTime float64
}

func NewLoadBalancer() *LoadBalancer {
	return &LoadBalancer{
		providers: make([]Provider, 0),
	}
}

// UpdateProvider updates or adds a provider's metrics
func (lb *LoadBalancer) UpdateProvider(id peer.ID, responseTime float64, success bool) {
	lb.mu.Lock()
	defer lb.mu.Unlock()

	// Find or create provider
	var provider *Provider
	for i := range lb.providers {
		if lb.providers[i].ID == id {
			provider = &lb.providers[i]
			break
		}
	}
	
	if provider == nil {
		lb.providers = append(lb.providers, Provider{
			ID:          id,
			LastSeen:    time.Now(),
			SuccessRate: 1.0,
		})
		provider = &lb.providers[len(lb.providers)-1]
	}

	// Update metrics
	provider.LastSeen = time.Now()
	provider.ResponseTime = (provider.ResponseTime*0.8 + responseTime*0.2) // Exponential moving average
	
	if success {
		provider.SuccessRate = provider.SuccessRate*0.95 + 0.05
	} else {
		provider.SuccessRate = provider.SuccessRate * 0.95
	}
	
	// Calculate load score (lower is better)
	provider.LoadScore = provider.ResponseTime / (provider.SuccessRate + 0.01)
}

// SelectProvider selects the best provider based on load balancing strategy
func (lb *LoadBalancer) SelectProvider(providers []peer.AddrInfo) peer.ID {
	if len(providers) == 0 {
		return ""
	}

	lb.mu.RLock()
	defer lb.mu.RUnlock()

	// Create a map of available providers
	available := make(map[peer.ID]bool)
	for _, p := range providers {
		available[p.ID] = true
	}

	// Filter and sort providers by load score
	var candidates []Provider
	for _, p := range lb.providers {
		if available[p.ID] && time.Since(p.LastSeen) < 5*time.Minute {
			candidates = append(candidates, p)
		}
	}

	// If we have metrics, use them
	if len(candidates) > 0 {
		// Sort by load score (lower is better)
		best := candidates[0]
		for _, c := range candidates[1:] {
			if c.LoadScore < best.LoadScore {
				best = c
			}
		}
		return best.ID
	}

	// Otherwise, use round-robin with new providers
	idx := atomic.AddUint64(&lb.counter, 1) % uint64(len(providers))
	return providers[idx].ID
}

// GetHealthyProviders returns a list of healthy providers
func (lb *LoadBalancer) GetHealthyProviders() []Provider {
	lb.mu.RLock()
	defer lb.mu.RUnlock()

	var healthy []Provider
	cutoff := time.Now().Add(-5 * time.Minute)
	
	for _, p := range lb.providers {
		if p.LastSeen.After(cutoff) && p.SuccessRate > 0.5 {
			healthy = append(healthy, p)
		}
	}
	
	return healthy
}

// Cleanup removes stale providers
func (lb *LoadBalancer) Cleanup() {
	lb.mu.Lock()
	defer lb.mu.Unlock()

	cutoff := time.Now().Add(-10 * time.Minute)
	var active []Provider
	
	for _, p := range lb.providers {
		if p.LastSeen.After(cutoff) {
			active = append(active, p)
		}
	}
	
	lb.providers = active
}

// WeightedRandomSelection selects a provider using weighted random selection
func (lb *LoadBalancer) WeightedRandomSelection(providers []peer.AddrInfo) peer.ID {
	if len(providers) == 0 {
		return ""
	}

	lb.mu.RLock()
	defer lb.mu.RUnlock()

	// Create weights based on inverse load scores
	weights := make([]float64, len(providers))
	totalWeight := 0.0
	
	for i, p := range providers {
		weight := 1.0
		for _, tracked := range lb.providers {
			if tracked.ID == p.ID {
				// Inverse of load score (higher is better)
				weight = 1.0 / (tracked.LoadScore + 0.01)
				break
			}
		}
		weights[i] = weight
		totalWeight += weight
	}

	// Weighted random selection
	r := rand.Float64() * totalWeight
	cumulative := 0.0
	
	for i, w := range weights {
		cumulative += w
		if r <= cumulative {
			return providers[i].ID
		}
	}

	// Fallback
	return providers[len(providers)-1].ID
}