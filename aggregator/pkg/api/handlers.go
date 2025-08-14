package api

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/quiver/aggregator/pkg/epoch"
	"github.com/quiver/aggregator/pkg/merkle"
	"github.com/quiver/aggregator/pkg/storage"
	"github.com/sirupsen/logrus"
)

const (
	// MaxReceiptSize is the maximum allowed size for a single receipt (10KB)
	MaxReceiptSize = 10 * 1024
)

type Handler struct {
	store        *storage.Store
	epochManager *epoch.Manager
	logger       *logrus.Logger
}

func NewHandler(store *storage.Store, epochManager *epoch.Manager) *Handler {
	logger := logrus.New()
	logger.SetFormatter(&logrus.JSONFormatter{})

	return &Handler{
		store:        store,
		epochManager: epochManager,
		logger:       logger,
	}
}

func (h *Handler) Commit(c *gin.Context) {
	var req CommitRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid request"})
		return
	}

	// Validate and store receipts
	for _, receipt := range req.Receipts {
		// Validate receipt size
		receiptJSON, err := json.Marshal(receipt)
		if err != nil {
			c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid receipt format"})
			return
		}
		if len(receiptJSON) > MaxReceiptSize {
			c.JSON(http.StatusBadRequest, ErrorResponse{Error: fmt.Sprintf("Receipt exceeds maximum size of %d bytes", MaxReceiptSize)})
			return
		}
		
		if err := h.store.Store(receipt); err != nil {
			h.logger.WithError(err).Error("Failed to store receipt")
		}
	}

	// Sort receipts by sequence number
	sort.Slice(req.Receipts, func(i, j int) bool {
		return req.Receipts[i].Receipt.Seq < req.Receipts[j].Receipt.Seq
	})

	// Build Merkle tree
	tree := merkle.NewTree()
	for _, receipt := range req.Receipts {
		canonical, err := canonicalizeJSON(receipt.Receipt)
		if err != nil {
			c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to canonicalize receipt"})
			return
		}
		tree.AddLeaf(canonical)
	}

	if err := tree.Build(); err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to build merkle tree"})
		return
	}

	root := tree.Root()

	// Store proofs
	for i, receipt := range req.Receipts {
		proof, err := tree.Proof(i)
		if err == nil {
			h.store.StoreProof(receipt.Receipt.ReceiptID, proof)
		}
	}

	// Finalize epoch
	if err := h.epochManager.FinalizeEpoch(req.Epoch, root, len(req.Receipts)); err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to finalize epoch"})
		return
	}

	h.logger.WithFields(logrus.Fields{
		"epoch":         req.Epoch,
		"receipt_count": len(req.Receipts),
		"merkle_root":   root,
	}).Info("Epoch committed")

	c.JSON(http.StatusOK, CommitResponse{
		MerkleRoot:   root,
		Epoch:        req.Epoch,
		ReceiptCount: len(req.Receipts),
	})
}

func (h *Handler) Claim(c *gin.Context) {
	var req ClaimRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid request"})
		return
	}

	// Get receipt
	receipt, err := h.store.GetByID(req.ReceiptID)
	if err != nil || receipt == nil {
		c.JSON(http.StatusNotFound, ErrorResponse{Error: "Receipt not found"})
		return
	}

	// Get epoch info
	epochInfo, exists := h.epochManager.GetEpochInfo(req.Epoch)
	if !exists || !epochInfo.Finalized {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Epoch not finalized"})
		return
	}

	// Verify proof
	canonical, err := canonicalizeJSON(receipt.Receipt)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to canonicalize receipt"})
		return
	}
	valid := merkle.Verify(canonical, req.MerkleProof, epochInfo.Root)

	amount := "0"
	if valid {
		// Calculate amount based on tokens
		totalTokens := receipt.Receipt.TokensIn + receipt.Receipt.TokensOut
		amount = calculateAmount(totalTokens)
	}

	h.logger.WithFields(logrus.Fields{
		"receipt_id": req.ReceiptID,
		"epoch":      req.Epoch,
		"valid":      valid,
		"amount":     amount,
	}).Info("Claim processed")

	c.JSON(http.StatusOK, ClaimResponse{
		Valid:  valid,
		Amount: amount,
		TxHash: "0x" + hashString(req.ReceiptID+epochInfo.Root),
	})
}

func (h *Handler) GetState(c *gin.Context) {
	state, err := h.store.ExportState()
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to export state"})
		return
	}

	c.Data(http.StatusOK, "application/json", state)
}

func canonicalizeJSON(v interface{}) ([]byte, error) {
	data, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}

	var m map[string]interface{}
	if err := json.Unmarshal(data, &m); err != nil {
		return nil, err
	}

	return json.Marshal(sortKeys(m))
}

func sortKeys(m map[string]interface{}) map[string]interface{} {
	result := make(map[string]interface{})

	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, k := range keys {
		v := m[k]
		switch val := v.(type) {
		case map[string]interface{}:
			result[k] = sortKeys(val)
		default:
			result[k] = v
		}
	}

	return result
}

func hashString(s string) string {
	h := sha256.Sum256([]byte(s))
	return hex.EncodeToString(h[:])
}

func (h *Handler) Health(c *gin.Context) {
	epochCount := h.epochManager.GetEpochCount()
	c.JSON(http.StatusOK, gin.H{
		"status": "healthy",
		"service": "aggregator",
		"epochs_processed": epochCount,
		"timestamp": time.Now().Unix(),
	})
}

func calculateAmount(tokens int) string {
	// Simple calculation: 1 token = 0.0001 units
	amount := tokens * 100
	return fmt.Sprintf("%d", amount)
}
