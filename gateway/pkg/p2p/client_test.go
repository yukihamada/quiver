package p2p

import (
	"context"
	"testing"
	"time"
)

func TestClientCreation(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	client, err := NewClient(ctx, "/ip4/127.0.0.1/udp/0/quic-v1", nil)
	if err != nil {
		t.Fatal(err)
	}
	defer client.Close()

	if client.host.ID() == "" {
		t.Error("Expected non-empty peer ID")
	}
}

func TestFindProviders(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	client, err := NewClient(ctx, "/ip4/127.0.0.1/udp/0/quic-v1", nil)
	if err != nil {
		t.Fatal(err)
	}
	defer client.Close()

	// Should return empty list when no providers
	providers, err := client.FindProviders()
	if err != nil {
		t.Fatal(err)
	}

	if len(providers) != 0 {
		t.Error("Expected no providers initially")
	}
}
