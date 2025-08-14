package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// Aggregator metrics
	CommitsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "aggregator_commits_total",
		Help: "Total number of epoch commits",
	}, []string{"status"})

	ReceiptsPerCommit = promauto.NewHistogram(prometheus.HistogramOpts{
		Name:    "aggregator_receipts_per_commit",
		Help:    "Number of receipts per commit",
		Buckets: prometheus.ExponentialBuckets(1, 2, 10),
	})

	ClaimsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "aggregator_claims_total",
		Help: "Total number of claims processed",
	}, []string{"valid"})

	EpochsActive = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "aggregator_epochs_active",
		Help: "Number of active epochs",
	})

	MerkleTreeBuildTime = promauto.NewHistogram(prometheus.HistogramOpts{
		Name:    "aggregator_merkle_tree_build_seconds",
		Help:    "Time taken to build Merkle tree in seconds",
		Buckets: prometheus.DefBuckets,
	})

	StorageSize = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "aggregator_storage_size_bytes",
		Help: "Total size of stored receipts in bytes",
	})
)