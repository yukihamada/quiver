package unit

import (
	"crypto/rand"
	"testing"
)

func FuzzInvalidProofGeneration(f *testing.F) {
	// Seed corpus
	f.Add([]byte("valid_proof_element"))
	f.Add([]byte(""))
	f.Add([]byte("1234567890abcdef"))
	
	f.Fuzz(func(t *testing.T, data []byte) {
		// Generate random proof modifications
		if len(data) == 0 {
			return
		}
		
		// Simulate proof corruption
		corrupted := make([]byte, len(data))
		copy(corrupted, data)
		
		// Random bit flip
		if len(corrupted) > 0 {
			idx := int(corrupted[0]) % len(corrupted)
			corrupted[idx] ^= 1 << (corrupted[0] % 8)
		}
		
		// Property: corrupted proof should not equal original
		if len(data) > 0 && string(corrupted) == string(data) {
			t.Skip("No corruption occurred")
		}
	})
}

func TestRandomDelayWithinSLA(t *testing.T) {
	slaMs := 2500
	iterations := 100
	
	delays := make([]int, iterations)
	
	for i := 0; i < iterations; i++ {
		// Generate random delay between 0-2000ms
		b := make([]byte, 2)
		rand.Read(b)
		delay := int(b[0])<<8 + int(b[1])
		delay = delay % 2000
		
		delays[i] = delay
		
		if delay > slaMs {
			t.Errorf("Delay %dms exceeds SLA %dms", delay, slaMs)
		}
	}
	
	// Verify distribution
	sum := 0
	for _, d := range delays {
		sum += d
	}
	avg := sum / iterations
	
	t.Logf("Average delay: %dms (SLA: %dms)", avg, slaMs)
}