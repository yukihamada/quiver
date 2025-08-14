package mock

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"math/big"
	"sync"
	"time"
)

type Chain struct {
	roots       map[uint64]string
	claims      map[string]*Claim
	balances    map[string]*big.Int
	commitCount uint64
	mu          sync.RWMutex
}

type Claim struct {
	ID        string
	Epoch     uint64
	Amount    *big.Int
	Timestamp time.Time
	Processed bool
}

func NewChain() *Chain {
	return &Chain{
		roots:    make(map[uint64]string),
		claims:   make(map[string]*Claim),
		balances: make(map[string]*big.Int),
	}
}

func (c *Chain) CommitRoot(epoch uint64, root string) (string, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if _, exists := c.roots[epoch]; exists {
		return "", fmt.Errorf("epoch %d already committed", epoch)
	}

	c.roots[epoch] = root
	c.commitCount++

	txHash := generateTxHash(fmt.Sprintf("commit_%d_%s", epoch, root))
	return txHash, nil
}

func (c *Chain) VerifyClaim(receiptID string, proof []string, epoch uint64) (bool, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	root, exists := c.roots[epoch]
	if !exists {
		return false, fmt.Errorf("epoch %d not found", epoch)
	}

	// Simplified verification - in real implementation would verify Merkle proof
	proofHash := generateTxHash(fmt.Sprintf("%s_%v_%s", receiptID, proof, root))
	return len(proofHash) > 0, nil
}

func (c *Chain) ProcessClaim(claimID string, amount *big.Int) (string, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if claim, exists := c.claims[claimID]; exists && claim.Processed {
		return "", fmt.Errorf("claim %s already processed", claimID)
	}

	claim := &Claim{
		ID:        claimID,
		Amount:    amount,
		Timestamp: time.Now(),
		Processed: true,
	}

	c.claims[claimID] = claim

	// Update balance
	provider := "default_provider"
	if balance, exists := c.balances[provider]; exists {
		c.balances[provider] = new(big.Int).Add(balance, amount)
	} else {
		c.balances[provider] = new(big.Int).Set(amount)
	}

	txHash := generateTxHash(fmt.Sprintf("claim_%s_%s", claimID, amount.String()))
	return txHash, nil
}

func (c *Chain) GetBalance(address string) (*big.Int, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	balance, exists := c.balances[address]
	if !exists {
		return big.NewInt(0), nil
	}

	return new(big.Int).Set(balance), nil
}

func (c *Chain) GetRoot(epoch uint64) (string, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	root, exists := c.roots[epoch]
	if !exists {
		return "", fmt.Errorf("epoch %d not found", epoch)
	}

	return root, nil
}

func (c *Chain) IsClaimProcessed(claimID string) (bool, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	claim, exists := c.claims[claimID]
	if !exists {
		return false, nil
	}

	return claim.Processed, nil
}

func generateTxHash(data string) string {
	hash := sha256.Sum256([]byte(data))
	return "0x" + hex.EncodeToString(hash[:])
}
