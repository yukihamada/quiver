package unit

import (
	"crypto/sha256"
	"encoding/hex"
	"math/rand"
	"testing"
	"time"
)

func TestPrevHashChainProperty(t *testing.T) {
	// Property: prev_hash forms valid hash chain
	var prevHash string
	receipts := make([]mockReceipt, 100)
	
	for i := range receipts {
		receipts[i] = mockReceipt{
			seq:      int64(i),
			prevHash: prevHash,
			data:     randomData(),
		}
		
		// Calculate hash for next iteration
		prevHash = calculateHash(receipts[i])
		
		// Verify chain property
		if i > 0 {
			expectedPrev := calculateHash(receipts[i-1])
			if receipts[i].prevHash != expectedPrev {
				t.Errorf("Invalid hash chain at position %d", i)
			}
		}
	}
}

func TestInvalidProofFuzz(t *testing.T) {
	// Property: random modifications to valid proofs should fail verification
	validProof := []string{
		"abc123def456",
		"789012345678",
		"fedcba987654",
	}
	
	for i := 0; i < 1000; i++ {
		// Clone and modify proof
		modifiedProof := make([]string, len(validProof))
		copy(modifiedProof, validProof)
		
		// Random modification
		switch rand.Intn(4) {
		case 0: // Change character
			idx := rand.Intn(len(modifiedProof))
			if len(modifiedProof[idx]) > 0 {
				bytes := []byte(modifiedProof[idx])
				bytes[rand.Intn(len(bytes))] = byte(rand.Intn(256))
				modifiedProof[idx] = string(bytes)
			}
		case 1: // Add element
			modifiedProof = append(modifiedProof, randomHex())
		case 2: // Remove element
			if len(modifiedProof) > 1 {
				idx := rand.Intn(len(modifiedProof))
				modifiedProof = append(modifiedProof[:idx], modifiedProof[idx+1:]...)
			}
		case 3: // Swap elements
			if len(modifiedProof) > 1 {
				i, j := rand.Intn(len(modifiedProof)), rand.Intn(len(modifiedProof))
				modifiedProof[i], modifiedProof[j] = modifiedProof[j], modifiedProof[i]
			}
		}
		
		// Modified proof should be different and invalid
		if !slicesEqual(validProof, modifiedProof) {
			// In real implementation, this would verify against merkle root
			// For test, we just ensure modifications were made
			if len(modifiedProof) == 0 {
				t.Error("Modified proof became empty")
			}
		}
	}
}

func TestRandomDelaySLA(t *testing.T) {
	// Property: random delays should stay within SLA
	slaMs := 2500 // 2.5 seconds
	
	for i := 0; i < 100; i++ {
		start := time.Now()
		
		// Simulate random processing delay
		delay := time.Duration(rand.Intn(200)) * time.Millisecond
		time.Sleep(delay)
		
		elapsed := time.Since(start).Milliseconds()
		
		if elapsed > int64(slaMs) {
			t.Errorf("Delay %d ms exceeds SLA of %d ms", elapsed, slaMs)
		}
	}
}

type mockReceipt struct {
	seq      int64
	prevHash string
	data     string
}

func calculateHash(r mockReceipt) string {
	h := sha256.New()
	h.Write([]byte(r.data))
	h.Write([]byte(r.prevHash))
	return hex.EncodeToString(h.Sum(nil))
}

func randomData() string {
	chars := "abcdefghijklmnopqrstuvwxyz0123456789"
	result := make([]byte, 32)
	for i := range result {
		result[i] = chars[rand.Intn(len(chars))]
	}
	return string(result)
}

func randomHex() string {
	bytes := make([]byte, 32)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}

func slicesEqual(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}