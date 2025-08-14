package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/quiver/aggregator/pkg/epoch"
	"github.com/quiver/aggregator/pkg/merkle"
	"github.com/quiver/aggregator/pkg/storage"
)

func TestCommitEndpoint(t *testing.T) {
	gin.SetMode(gin.TestMode)

	store := storage.NewStore()
	epochManager := epoch.NewManager()
	handler := NewHandler(store, epochManager)

	router := gin.New()
	router.POST("/commit", handler.Commit)

	receipts := []*storage.SignedReceipt{
		{
			Receipt: storage.Receipt{
				Version:   "1.0.0",
				ReceiptID: "test1",
				Epoch:     19723,
				Seq:       1,
				TokensIn:  5,
				TokensOut: 10,
			},
			Signature: "sig1",
		},
		{
			Receipt: storage.Receipt{
				Version:   "1.0.0",
				ReceiptID: "test2",
				Epoch:     19723,
				Seq:       2,
				TokensIn:  3,
				TokensOut: 7,
			},
			Signature: "sig2",
		},
	}

	req := CommitRequest{
		Receipts: receipts,
		Epoch:    19723,
	}

	body, _ := json.Marshal(req)
	request := httptest.NewRequest("POST", "/commit", bytes.NewReader(body))
	request.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, request)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var resp CommitResponse
	json.Unmarshal(w.Body.Bytes(), &resp)

	if resp.Epoch != 19723 {
		t.Errorf("Expected epoch 19723, got %d", resp.Epoch)
	}

	if resp.ReceiptCount != 2 {
		t.Errorf("Expected 2 receipts, got %d", resp.ReceiptCount)
	}

	if resp.MerkleRoot == "" {
		t.Error("Expected non-empty merkle root")
	}
}

func TestClaimEndpoint(t *testing.T) {
	gin.SetMode(gin.TestMode)

	store := storage.NewStore()
	epochManager := epoch.NewManager()
	handler := NewHandler(store, epochManager)

	router := gin.New()
	router.POST("/claim", handler.Claim)

	// Setup test data
	receipt := &storage.SignedReceipt{
		Receipt: storage.Receipt{
			ReceiptID: "claim-test",
			TokensIn:  10,
			TokensOut: 20,
		},
		Signature: "sig",
	}
	store.Store(receipt)

	// Create a merkle tree with the receipt
	tree := merkle.NewTree()
	canonical, _ := canonicalizeJSON(receipt.Receipt)
	tree.AddLeaf(canonical)
	tree.Build()

	// Get the proof for this receipt
	proof, _ := tree.Proof(0)

	// Finalize epoch with actual root
	epochManager.FinalizeEpoch(19723, tree.Root(), 1)

	req := ClaimRequest{
		ReceiptID:   "claim-test",
		MerkleProof: proof,
		Epoch:       19723,
	}

	body, _ := json.Marshal(req)
	request := httptest.NewRequest("POST", "/claim", bytes.NewReader(body))
	request.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, request)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var resp ClaimResponse
	json.Unmarshal(w.Body.Bytes(), &resp)

	if resp.Amount == "0" {
		t.Error("Expected non-zero amount")
	}

	if resp.TxHash == "" {
		t.Error("Expected non-empty tx hash")
	}
}
