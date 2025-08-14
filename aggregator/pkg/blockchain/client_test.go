package blockchain

import (
	"math/big"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/assert"
)

func TestBlockchainConfig(t *testing.T) {
	// Test Polygon config
	config := DefaultPolygonConfig()
	assert.Equal(t, int64(137), config.ChainID)
	assert.Equal(t, "https://polygon-rpc.com", config.RPCEndpoint)
	assert.Equal(t, uint64(3000000), config.GasLimit)

	// Test Mumbai config
	mumbaiConfig := DefaultPolygonMumbaiConfig()
	assert.Equal(t, int64(80001), mumbaiConfig.ChainID)
	assert.Contains(t, mumbaiConfig.RPCEndpoint, "mumbai")
}

func TestReceiptBatchCreation(t *testing.T) {
	batch := ReceiptBatch{
		BatchID:      "test-batch-001",
		Epoch:        12345,
		MerkleRoot:   [32]byte{1, 2, 3},
		TotalAmount:  big.NewInt(1000000),
		ReceiptCount: 100,
		Timestamp:    time.Now(),
	}

	assert.Equal(t, "test-batch-001", batch.BatchID)
	assert.Equal(t, uint64(12345), batch.Epoch)
	assert.Equal(t, uint64(100), batch.ReceiptCount)
	assert.Equal(t, big.NewInt(1000000), batch.TotalAmount)
}

func TestChannelStructure(t *testing.T) {
	channel := Channel{
		ChannelID:    [32]byte{1, 2, 3},
		Provider:     common.HexToAddress("0x1234567890123456789012345678901234567890"),
		Consumer:     common.HexToAddress("0x0987654321098765432109876543210987654321"),
		Deposit:      big.NewInt(1000000),
		Balance:      big.NewInt(500000),
		Nonce:        42,
		LastActivity: time.Now(),
		IsOpen:       true,
	}

	assert.True(t, channel.IsOpen)
	assert.Equal(t, uint64(42), channel.Nonce)
	assert.Equal(t, 0, channel.Deposit.Cmp(big.NewInt(1000000)))
}

func TestSettlementProof(t *testing.T) {
	proof := SettlementProof{
		BatchID:   "batch-001",
		ChannelID: [32]byte{1, 2, 3},
		Amount:    big.NewInt(100000),
		Nonce:     10,
		MerkleProof: [][32]byte{
			{4, 5, 6},
			{7, 8, 9},
		},
		ProviderSig: []byte("provider-signature"),
		ConsumerSig: []byte("consumer-signature"),
	}

	assert.Equal(t, "batch-001", proof.BatchID)
	assert.Equal(t, uint64(10), proof.Nonce)
	assert.Len(t, proof.MerkleProof, 2)
}

// Integration test - requires actual blockchain connection
func TestBlockchainClientIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	// This would require actual Mumbai testnet connection
	config := DefaultPolygonMumbaiConfig()
	config.PrivateKey = "test-private-key" // Would need real key for actual test
	config.ContractAddress = "0x0000000000000000000000000000000000000000"
	config.TokenAddress = "0x0000000000000000000000000000000000000000"

	// Test client creation (would fail without valid key)
	_, err := NewClient(config)
	assert.Error(t, err) // Expected to fail with test key
}

func TestProviderStats(t *testing.T) {
	stats := ProviderStats{
		Address:         common.HexToAddress("0x1234567890123456789012345678901234567890"),
		TotalReceipts:   1000,
		TotalTokensIn:   50000,
		TotalTokensOut:  450000,
		TotalEarned:     big.NewInt(5000000),
		LastSettlement:  time.Now(),
		ActiveChannels:  5,
		ReputationScore: 9500, // Out of 10000
	}

	assert.Equal(t, uint64(1000), stats.TotalReceipts)
	assert.Equal(t, uint64(500000), stats.TotalTokensIn+stats.TotalTokensOut)
	assert.Equal(t, uint64(9500), stats.ReputationScore)
}

func TestEpochInfo(t *testing.T) {
	epoch := EpochInfo{
		Epoch:         19953,
		StartTime:     time.Now().Add(-24 * time.Hour),
		EndTime:       time.Now(),
		MerkleRoot:    [32]byte{1, 2, 3, 4, 5},
		TotalReceipts: 10000,
		TotalAmount:   big.NewInt(100000000),
		Finalized:     true,
		TxHash:        common.HexToHash("0x1234567890abcdef"),
	}

	assert.True(t, epoch.Finalized)
	assert.Equal(t, uint64(19953), epoch.Epoch)
	assert.Equal(t, uint64(10000), epoch.TotalReceipts)
}