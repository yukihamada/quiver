package storage

import (
	"encoding/json"
	"testing"
)

func TestReceiptStorage(t *testing.T) {
	store := NewStore()

	receipt := &SignedReceipt{
		Receipt: Receipt{
			Version:    "1.0.0",
			ProviderPK: "test-key",
			Model:      "test-model",
			ReceiptID:  "test-id-123",
			Epoch:      19723,
			Seq:        1,
		},
		Signature: "test-signature",
	}

	// Store receipt
	err := store.Store(receipt)
	if err != nil {
		t.Fatal(err)
	}

	// Retrieve by ID
	retrieved, err := store.GetByID("test-id-123")
	if err != nil {
		t.Fatal(err)
	}

	if retrieved == nil {
		t.Fatal("Receipt not found")
	}

	if retrieved.Receipt.ReceiptID != receipt.Receipt.ReceiptID {
		t.Error("Retrieved receipt doesn't match")
	}

	// Retrieve by epoch
	epochReceipts, err := store.GetByEpoch(19723)
	if err != nil {
		t.Fatal(err)
	}

	if len(epochReceipts) != 1 {
		t.Errorf("Expected 1 receipt, got %d", len(epochReceipts))
	}
}

func TestProofStorage(t *testing.T) {
	store := NewStore()

	receiptID := "test-receipt-123"
	proof := []string{"hash1", "hash2", "hash3"}

	store.StoreProof(receiptID, proof)

	retrievedProof, exists := store.GetProof(receiptID)
	if !exists {
		t.Fatal("Proof not found")
	}

	if len(retrievedProof) != len(proof) {
		t.Errorf("Expected %d proof elements, got %d", len(proof), len(retrievedProof))
	}

	for i, p := range proof {
		if retrievedProof[i] != p {
			t.Errorf("Proof element %d mismatch", i)
		}
	}
}

func TestExportState(t *testing.T) {
	store := NewStore()

	// Add some data
	receipt := &SignedReceipt{
		Receipt: Receipt{
			ReceiptID: "export-test",
			Epoch:     19723,
		},
		Signature: "sig",
	}

	if err := store.Store(receipt); err != nil {
		t.Fatal(err)
	}
	store.StoreProof("export-test", []string{"proof1"})

	// Export state
	state, err := store.ExportState()
	if err != nil {
		t.Fatal(err)
	}

	// Verify it's valid JSON
	var decoded map[string]interface{}
	if err := json.Unmarshal(state, &decoded); err != nil {
		t.Fatal("Invalid JSON export")
	}

	if _, ok := decoded["receipts"]; !ok {
		t.Error("Missing receipts in export")
	}

	if _, ok := decoded["proofs"]; !ok {
		t.Error("Missing proofs in export")
	}
}
