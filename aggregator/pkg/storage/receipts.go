package storage

import (
	"encoding/json"
	"sync"
)

type Receipt struct {
	Version    string                 `json:"version"`
	ProviderPK string                 `json:"provider_pk"`
	Model      string                 `json:"model"`
	PromptHash string                 `json:"prompt_hash"`
	OutputHash string                 `json:"output_hash"`
	TokensIn   int                    `json:"tokens_in"`
	TokensOut  int                    `json:"tokens_out"`
	StartISO   string                 `json:"start_iso"`
	EndISO     string                 `json:"end_iso"`
	DurationMs int64                  `json:"duration_ms"`
	Epoch      int64                  `json:"epoch"`
	Seq        int64                  `json:"seq"`
	PrevHash   string                 `json:"prev_hash"`
	Canary     map[string]interface{} `json:"canary"`
	Rate       map[string]interface{} `json:"rate"`
	ReceiptID  string                 `json:"receipt_id"`
}

type SignedReceipt struct {
	Receipt   Receipt `json:"receipt"`
	Signature string  `json:"signature"`
}

type Store struct {
	receipts        map[string]*SignedReceipt
	receiptsByEpoch map[uint64][]*SignedReceipt
	proofs          map[string][]string
	mu              sync.RWMutex
}

func NewStore() *Store {
	return &Store{
		receipts:        make(map[string]*SignedReceipt),
		receiptsByEpoch: make(map[uint64][]*SignedReceipt),
		proofs:          make(map[string][]string),
	}
}

func (s *Store) Store(receipt *SignedReceipt) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.receipts[receipt.Receipt.ReceiptID] = receipt

	epoch := uint64(receipt.Receipt.Epoch)
	s.receiptsByEpoch[epoch] = append(s.receiptsByEpoch[epoch], receipt)

	return nil
}

func (s *Store) GetByID(id string) (*SignedReceipt, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	receipt, exists := s.receipts[id]
	if !exists {
		return nil, nil
	}

	return receipt, nil
}

func (s *Store) GetByEpoch(epoch uint64) ([]*SignedReceipt, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	receipts := s.receiptsByEpoch[epoch]
	result := make([]*SignedReceipt, len(receipts))
	copy(result, receipts)

	return result, nil
}

func (s *Store) Count(epoch uint64) int {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return len(s.receiptsByEpoch[epoch])
}

func (s *Store) StoreProof(receiptID string, proof []string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.proofs[receiptID] = proof
}

func (s *Store) GetProof(receiptID string) ([]string, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	proof, exists := s.proofs[receiptID]
	return proof, exists
}

func (s *Store) ExportState() ([]byte, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	state := map[string]interface{}{
		"receipts": s.receipts,
		"proofs":   s.proofs,
	}

	return json.MarshalIndent(state, "", "  ")
}
