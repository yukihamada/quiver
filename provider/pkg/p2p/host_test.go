package p2p

import (
	"context"
	"testing"
	"time"

	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/protocol"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHostCreation(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	host, err := NewHost(ctx, "/ip4/127.0.0.1/udp/0/quic-v1", nil)
	if err != nil {
		t.Fatal(err)
	}
	defer host.Close()

	if host.ID() == "" {
		t.Error("Expected non-empty peer ID")
	}

	if len(host.Addrs()) == 0 {
		t.Error("Expected at least one address")
	}
}

func TestAdvertise(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	host, err := NewHost(ctx, "/ip4/127.0.0.1/udp/0/quic-v1", nil)
	if err != nil {
		t.Fatal(err)
	}
	defer host.Close()

	// Should not error
	err = host.Advertise("test-topic")
	if err != nil {
		t.Errorf("Advertise failed: %v", err)
	}
}

func TestQUICCommunication(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Create two hosts
	host1, err := NewHost(ctx, "/ip4/127.0.0.1/udp/0/quic-v1", nil)
	require.NoError(t, err)
	defer host1.Close()

	host2, err := NewHost(ctx, "/ip4/127.0.0.1/udp/0/quic-v1", nil)
	require.NoError(t, err)
	defer host2.Close()

	// Set up stream handler
	const testProtocol = "/test/quic/1.0.0"
	receivedMsg := make(chan string, 1)

	host2.SetStreamHandler(protocol.ID(testProtocol), func(s network.Stream) {
		defer s.Close()
		buf := make([]byte, 1024)
		n, err := s.Read(buf)
		if err == nil {
			receivedMsg <- string(buf[:n])
		}
	})

	// Connect hosts
	host2Info := peer.AddrInfo{
		ID:    host2.ID(),
		Addrs: host2.Addrs(),
	}

	err = host1.host.Connect(ctx, host2Info)
	assert.NoError(t, err)

	// Send message
	stream, err := host1.host.NewStream(ctx, host2.ID(), protocol.ID(testProtocol))
	require.NoError(t, err)
	defer stream.Close()

	testMsg := "QUIC test message"
	_, err = stream.Write([]byte(testMsg))
	assert.NoError(t, err)

	// Verify receipt
	select {
	case msg := <-receivedMsg:
		assert.Equal(t, testMsg, msg)
	case <-time.After(5 * time.Second):
		t.Fatal("Timeout waiting for message")
	}
}

func TestRelaySetup(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Create relay node
	relayHost, err := SetupRelayNode(ctx, "/ip4/127.0.0.1/tcp/0", "")
	require.NoError(t, err)
	defer relayHost.Close()

	// Verify relay is accessible
	assert.NotEmpty(t, relayHost.ID())
	assert.NotEmpty(t, relayHost.Addrs())
}
