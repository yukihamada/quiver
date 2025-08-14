package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// Gateway metrics
	RequestsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "gateway_requests_total",
		Help: "Total number of requests processed",
	}, []string{"model", "status"})

	RequestDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "gateway_request_duration_seconds",
		Help:    "Request processing duration in seconds",
		Buckets: prometheus.DefBuckets,
	}, []string{"model"})

	ProviderCalls = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "gateway_provider_calls_total",
		Help: "Total number of calls to providers",
	}, []string{"provider_id", "status"})

	RateLimitHits = promauto.NewCounter(prometheus.CounterOpts{
		Name: "gateway_rate_limit_hits_total",
		Help: "Total number of rate limit hits",
	})

	CanaryChecks = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "gateway_canary_checks_total",
		Help: "Total number of canary checks",
	}, []string{"result"}) // result: "pass" or "fail"

	ActiveProviders = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "gateway_active_providers",
		Help: "Number of active providers discovered",
	})
)