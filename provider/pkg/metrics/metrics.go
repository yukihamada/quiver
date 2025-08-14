package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// Provider metrics
	RequestsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "provider_requests_total",
		Help: "Total number of requests processed",
	}, []string{"model", "status"})

	RequestDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "provider_request_duration_seconds",
		Help:    "Request processing duration in seconds",
		Buckets: prometheus.DefBuckets,
	}, []string{"model"})

	TokensProcessed = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "provider_tokens_processed_total",
		Help: "Total number of tokens processed",
	}, []string{"type"}) // type: "input" or "output"

	ActiveStreams = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "provider_active_streams",
		Help: "Number of active P2P streams",
	})

	RateLimitHits = promauto.NewCounter(prometheus.CounterOpts{
		Name: "provider_rate_limit_hits_total",
		Help: "Total number of rate limit hits",
	})

	ReceiptSignatures = promauto.NewCounter(prometheus.CounterOpts{
		Name: "provider_receipt_signatures_total",
		Help: "Total number of receipt signatures created",
	})
)