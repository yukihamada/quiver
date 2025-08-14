package config

import (
	"os"
	"strings"
	"time"
)

type Config struct {
	ListenAddr        string
	OllamaURL         string
	PrivateKeyPath    string
	MaxPromptBytes    int
	TokensPerSecond   int
	RequestTimeout    time.Duration
	DHTBootstrapPeers []string
}

func DefaultConfig() *Config {
	cfg := &Config{
		ListenAddr:        "/ip4/0.0.0.0/tcp/4003",
		OllamaURL:         "http://127.0.0.1:11434",
		PrivateKeyPath:    "provider.key",
		MaxPromptBytes:    4096,
		TokensPerSecond:   10,
		RequestTimeout:    30 * time.Second,
		DHTBootstrapPeers: []string{},
	}
	
	// Read from environment
	if bootstrap := os.Getenv("QUIVER_BOOTSTRAP"); bootstrap != "" {
		cfg.DHTBootstrapPeers = strings.Split(bootstrap, ",")
	}
	
	if ollamaURL := os.Getenv("QUIVER_OLLAMA_URL"); ollamaURL != "" {
		cfg.OllamaURL = ollamaURL
	}
	
	return cfg
}
