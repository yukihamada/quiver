package config

import (
	"os"
	"strings"
	"time"
)

type Config struct {
	Port              string
	P2PListenAddr     string
	DHTBootstrapPeers []string
	RequestTimeout    time.Duration
	RateLimitPerToken int
	CanaryRate        float64
}

func DefaultConfig() *Config {
	cfg := &Config{
		Port:              "8080",
		P2PListenAddr:     "/ip4/0.0.0.0/tcp/4002",
		DHTBootstrapPeers: []string{
			"/ip4/127.0.0.1/tcp/4001/p2p/12D3KooWLfChFxVatDEJocxtMgdT8yqRAePZJG26h6WHt6kuCNUW",
		},
		RequestTimeout:    60 * time.Second,
		RateLimitPerToken: 10,
		CanaryRate:        0.05,
	}
	
	// Read from environment
	if bootstrap := os.Getenv("QUIVER_BOOTSTRAP"); bootstrap != "" {
		cfg.DHTBootstrapPeers = strings.Split(bootstrap, ",")
	}
	
	if port := os.Getenv("QUIVER_GATEWAY_PORT"); port != "" {
		cfg.Port = port
	}
	
	return cfg
}
