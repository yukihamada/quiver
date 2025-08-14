package mock

import (
	"fmt"
	"math/big"
	"testing"
)

func TestChainCommitRoot(t *testing.T) {
	chain := NewChain()

	txHash, err := chain.CommitRoot(1, "0xabc123")
	if err != nil {
		t.Fatalf("Failed to commit root: %v", err)
	}

	if txHash == "" {
		t.Fatal("Expected non-empty tx hash")
	}

	// Try to commit same epoch again
	_, err = chain.CommitRoot(1, "0xdef456")
	if err == nil {
		t.Fatal("Expected error when committing duplicate epoch")
	}
}

func TestChainProcessClaim(t *testing.T) {
	chain := NewChain()

	amount := big.NewInt(1000)
	txHash, err := chain.ProcessClaim("claim1", amount)
	if err != nil {
		t.Fatalf("Failed to process claim: %v", err)
	}

	if txHash == "" {
		t.Fatal("Expected non-empty tx hash")
	}

	// Check if claim was processed
	processed, err := chain.IsClaimProcessed("claim1")
	if err != nil {
		t.Fatalf("Failed to check claim status: %v", err)
	}

	if !processed {
		t.Fatal("Expected claim to be processed")
	}

	// Try to process same claim again
	_, err = chain.ProcessClaim("claim1", amount)
	if err == nil {
		t.Fatal("Expected error when processing duplicate claim")
	}
}

func TestChainBalance(t *testing.T) {
	chain := NewChain()

	// Process multiple claims
	amounts := []*big.Int{
		big.NewInt(1000),
		big.NewInt(2000),
		big.NewInt(3000),
	}

	for i, amount := range amounts {
		claimID := fmt.Sprintf("claim%d", i)
		_, err := chain.ProcessClaim(claimID, amount)
		if err != nil {
			t.Fatalf("Failed to process claim %s: %v", claimID, err)
		}
	}

	// Check balance
	balance, err := chain.GetBalance("default_provider")
	if err != nil {
		t.Fatalf("Failed to get balance: %v", err)
	}

	expectedBalance := big.NewInt(6000)
	if balance.Cmp(expectedBalance) != 0 {
		t.Fatalf("Expected balance %s, got %s", expectedBalance, balance)
	}
}
