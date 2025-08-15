package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
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
)

const protocolID = "/quiver/inference/1.0.0"
const dhtTopic = "quiver.providers"

var startTime = time.Now()

func main() {
	cfg := config.DefaultConfig()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	signer, err := receipt.NewSigner(cfg.PrivateKeyPath)
	if err != nil {
		log.Fatal("Failed to create signer:", err)
	}

	host, err := p2p.NewHost(ctx, cfg.ListenAddr, cfg.DHTBootstrapPeers)
	if err != nil {
		log.Fatal("Failed to create P2P host:", err)
	}
	defer host.Close()

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
			})
		})
		mux.Handle("/metrics", promhttp.Handler())
		port := os.Getenv("PROVIDER_METRICS_PORT")
		if port == "" {
			port = "8090"
		}
		log.Printf("Health check and metrics endpoints started on :%s\n", port)
		if err := http.ListenAndServe(":"+port, mux); err != nil {
			log.Printf("Health/metrics server failed: %v", err)
		}
	}()

	go func() {
		for {
			if err := host.Advertise(dhtTopic); err != nil {
				log.Printf("Failed to advertise: %v", err)
			}
			time.Sleep(5 * time.Minute)
		}
	}()

	fmt.Printf("Provider started\n")
	fmt.Printf("PeerID: %s\n", host.ID())
	fmt.Printf("Addresses:\n")
	for _, addr := range host.Addrs() {
		fmt.Printf("  %s/p2p/%s\n", addr, host.ID())
	}
	fmt.Printf("Bootstrap peers: %v\n", cfg.DHTBootstrapPeers)
	fmt.Printf("Public Key: %s\n", signer.PublicKeyBase64())

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	<-sigCh

	fmt.Println("Shutting down...")
}
