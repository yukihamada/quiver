package receipt

import (
	"testing"
	"time"
)

func TestReceiptHashDeterminism(t *testing.T) {
	// Test that identical receipts produce identical hashes
	receipt1 := &Receipt{
		Version:    "1.0.0",
		ProviderPK: "test-pk",
		Model:      "model1",
		PromptHash: "prompt-hash",
		OutputHash: "output-hash",
		TokensIn:   10,
		TokensOut:  20,
		StartISO:   "2024-01-01T00:00:00Z",
		EndISO:     "2024-01-01T00:00:00.1Z",
		DurationMs: 100,
		Epoch:      19723,
		Seq:        1,
		PrevHash:   "fixed-prev-hash",
		Canary: Canary{
			ID:     "",
			Passed: true,
		},
		Rate: RateInfo{
			Throttle:  false,
			Truncated: false,
		},
		ReceiptID: "",
	}

	receipt2 := &Receipt{
		Version:    "1.0.0",
		ProviderPK: "test-pk",
		Model:      "model1",
		PromptHash: "prompt-hash",
		OutputHash: "output-hash",
		TokensIn:   10,
		TokensOut:  20,
		StartISO:   "2024-01-01T00:00:00Z",
		EndISO:     "2024-01-01T00:00:00.1Z",
		DurationMs: 100,
		Epoch:      19723,
		Seq:        1,
		PrevHash:   "fixed-prev-hash",
		Canary: Canary{
			ID:     "",
			Passed: true,
		},
		Rate: RateInfo{
			Throttle:  false,
			Truncated: false,
		},
		ReceiptID: "",
	}

	canonical1, err := CanonicalizeJSON(receipt1)
	if err != nil {
		t.Fatal(err)
	}

	canonical2, err := CanonicalizeJSON(receipt2)
	if err != nil {
		t.Fatal(err)
	}

	hash1 := HashData(canonical1)
	hash2 := HashData(canonical2)

	if hash1 != hash2 {
		t.Errorf("Same input produced different hashes: %s != %s", hash1, hash2)
	}

	id1 := GenerateReceiptID(hash1)
	id2 := GenerateReceiptID(hash2)

	if id1 != id2 {
		t.Errorf("Same hash produced different IDs: %s != %s", id1, id2)
	}
}

func TestCanonicalizeJSON(t *testing.T) {
	tests := []struct {
		name  string
		input map[string]interface{}
	}{
		{
			name: "nested objects",
			input: map[string]interface{}{
				"z": "last",
				"a": map[string]interface{}{
					"nested": true,
					"order":  "should not matter",
				},
				"m": 123,
			},
		},
		{
			name: "arrays and types",
			input: map[string]interface{}{
				"strings": []interface{}{"a", "b", "c"},
				"numbers": []interface{}{3, 1, 2},
				"bool":    true,
				"null":    nil,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			canonical1, err := CanonicalizeJSON(tt.input)
			if err != nil {
				t.Fatal(err)
			}

			canonical2, err := CanonicalizeJSON(tt.input)
			if err != nil {
				t.Fatal(err)
			}

			if string(canonical1) != string(canonical2) {
				t.Error("Canonicalization not deterministic")
			}
		})
	}
}

func TestNewReceiptFields(t *testing.T) {
	start := time.Now()
	end := start.Add(100 * time.Millisecond)

	receipt := NewReceipt("provider-key", "test-model", "prompt-hash", "output-hash", 10, 20, start, end)

	if receipt.Version != "1.0.0" {
		t.Errorf("Expected version 1.0.0, got %s", receipt.Version)
	}

	if receipt.Epoch != start.UTC().Unix()/86400 {
		t.Errorf("Epoch calculation incorrect")
	}

	if receipt.DurationMs != end.Sub(start).Milliseconds() {
		t.Errorf("Duration calculation incorrect")
	}

	if receipt.ReceiptID == "" {
		t.Error("Receipt ID not generated")
	}
}
