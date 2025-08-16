package main

import (
	"context"
	"encoding/json"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/libp2p/go-libp2p/core/protocol"
	"github.com/quiver/provider/internal/config"
	"github.com/quiver/provider/pkg/llm"
	"github.com/quiver/provider/pkg/p2p"
	"github.com/quiver/provider/pkg/receipt"
	"github.com/quiver/provider/pkg/stream"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sirupsen/logrus"
)

const protocolID = "/quiver/inference/1.0.0"
const dhtTopic = "quiver.providers"

var startTime = time.Now()

func main() {
	cfg := config.DefaultConfig()

	// Setup logger
	logger := logrus.New()
	logger.SetLevel(logrus.InfoLevel)
	if os.Getenv("DEBUG") == "true" {
		logger.SetLevel(logrus.DebugLevel)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	signer, err := receipt.NewSigner(cfg.PrivateKeyPath)
	if err != nil {
		logger.Fatal("Failed to create signer:", err)
	}

	host, err := p2p.NewHost(ctx, cfg.ListenAddr, cfg.DHTBootstrapPeers)
	if err != nil {
		logger.Fatal("Failed to create P2P host:", err)
	}
	defer host.Close()

	// Initialize connection manager
	connMgr, err := p2p.NewConnectionManager(host.GetHost(), cfg.DHTBootstrapPeers, logger)
	if err != nil {
		logger.Fatal("Failed to create connection manager:", err)
	}
	connMgr.Start()
	defer connMgr.Stop()

	llmClient := llm.NewClient(cfg.OllamaURL)

	handler := stream.NewHandler(
		llmClient,
		signer,
		cfg.MaxPromptBytes,
		cfg.TokensPerSecond,
	)

	host.SetStreamHandler(protocol.ID(protocolID), handler.HandleStream)

	// Start health check and metrics endpoints
	go func() {
		mux := http.NewServeMux()
		mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{
				"status": "healthy",
				"service": "provider",
				"peer_id": host.ID().String(),
				"uptime_seconds": time.Since(startTime).Seconds(),
				"connections": connMgr.GetConnectionStats(),
			})
		})
		mux.Handle("/metrics", promhttp.Handler())
		port := os.Getenv("PROVIDER_METRICS_PORT")
		if port == "" {
			port = "8090"
		}
		logger.Infof("Health check and metrics endpoints started on :%s", port)
		if err := http.ListenAndServe(":"+port, mux); err != nil {
			logger.Errorf("Health/metrics server failed: %v", err)
		}
	}()

	go func() {
		for {
			if err := host.Advertise(dhtTopic); err != nil {
				logger.Warnf("Failed to advertise: %v", err)
			}
			time.Sleep(5 * time.Minute)
		}
	}()

	// Periodic connection status logging
	go func() {
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()
		
		for range ticker.C {
			if !connMgr.IsConnected() {
				logger.Warn("No P2P connections! Attempting to reconnect...")
				connMgr.ForceReconnect()
			} else {
				stats := connMgr.GetConnectionStats()
				logger.Infof("P2P Status: %d connected peers", stats["connected_peers"])
			}
		}
	}()

	logger.Info("Provider started")
	logger.Infof("PeerID: %s", host.ID())
	logger.Info("Addresses:")
	for _, addr := range host.Addrs() {
		logger.Infof("  %s/p2p/%s", addr, host.ID())
	}
	logger.Infof("Bootstrap peers: %v", cfg.DHTBootstrapPeers)
	logger.Infof("Public Key: %s", signer.PublicKeyBase64())

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	<-sigCh

	logger.Info("Shutting down...")
}
