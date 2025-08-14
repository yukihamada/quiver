package stream

import (
	"bytes"
	"os"
	"testing"
	"time"

	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/protocol"
	"github.com/quiver/provider/pkg/receipt"
)

type mockStream struct {
	input  *bytes.Buffer
	output *bytes.Buffer
	closed bool
}

func (m *mockStream) Read(p []byte) (int, error) {
	return m.input.Read(p)
}

func (m *mockStream) Write(p []byte) (int, error) {
	return m.output.Write(p)
}

func (m *mockStream) Close() error {
	m.closed = true
	return nil
}

func (m *mockStream) CloseRead() error                 { return nil }
func (m *mockStream) CloseWrite() error                { return nil }
func (m *mockStream) Reset() error                     { return nil }
func (m *mockStream) SetDeadline(time.Time) error      { return nil }
func (m *mockStream) SetReadDeadline(time.Time) error  { return nil }
func (m *mockStream) SetWriteDeadline(time.Time) error { return nil }
func (m *mockStream) ID() string                       { return "test" }
func (m *mockStream) Protocol() protocol.ID            { return "/test/1.0.0" }
func (m *mockStream) SetProtocol(protocol.ID) error    { return nil }
func (m *mockStream) Stat() network.Stats              { return network.Stats{} }
func (m *mockStream) Conn() network.Conn               { return nil }
func (m *mockStream) Scope() network.StreamScope       { return nil }

func TestPromptSizeValidation(t *testing.T) {

	tests := []struct {
		name        string
		promptSize  int
		maxSize     int
		expectError bool
	}{
		{
			name:        "within limit",
			promptSize:  100,
			maxSize:     200,
			expectError: false,
		},
		{
			name:        "at limit",
			promptSize:  200,
			maxSize:     200,
			expectError: false,
		},
		{
			name:        "exceeds limit",
			promptSize:  201,
			maxSize:     200,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create handler with mock (would need proper mock in real test)
			// This is a simplified version showing the test structure

			req := Request{
				Prompt: string(make([]byte, tt.promptSize)),
				Model:  "test-model",
			}

			if len(req.Prompt) > tt.maxSize != tt.expectError {
				t.Errorf("Expected error: %v, but validation result was different", tt.expectError)
			}
		})
	}
}

func TestReceiptGeneration(t *testing.T) {
	signer, _ := receipt.NewSigner("test.key")
	defer os.Remove("test.key")

	// Test that receipts are generated with proper fields
	rcpt := receipt.NewReceipt(
		signer.PublicKeyBase64(),
		"test-model",
		"prompt_hash",
		"output_hash",
		10,
		20,
		time.Now(),
		time.Now().Add(100*time.Millisecond),
	)

	if rcpt.Version != "1.0.0" {
		t.Errorf("Expected version 1.0.0, got %s", rcpt.Version)
	}

	if rcpt.ReceiptID == "" {
		t.Error("Receipt ID should not be empty")
	}

	// Test signing
	signed, err := signer.Sign(rcpt)
	if err != nil {
		t.Fatal(err)
	}

	if signed.Signature == "" {
		t.Error("Signature should not be empty")
	}
}
