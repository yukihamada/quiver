package api

import "github.com/quiver/aggregator/pkg/storage"

type CommitRequest struct {
	Receipts []*storage.SignedReceipt `json:"receipts" binding:"required"`
	Epoch    uint64                   `json:"epoch" binding:"required"`
}

type CommitResponse struct {
	MerkleRoot   string `json:"merkle_root"`
	Epoch        uint64 `json:"epoch"`
	ReceiptCount int    `json:"receipt_count"`
}

type ClaimRequest struct {
	ReceiptID   string   `json:"receipt_id" binding:"required"`
	MerkleProof []string `json:"merkle_proof" binding:"required"`
	Epoch       uint64   `json:"epoch" binding:"required"`
}

type ClaimResponse struct {
	Valid  bool   `json:"valid"`
	Amount string `json:"amount"`
	TxHash string `json:"tx_hash"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}
