package blockchain

import (
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/common"
)

// ReceiptBatch represents a batch of receipts for on-chain settlement
type ReceiptBatch struct {
	BatchID      string
	Epoch        uint64
	MerkleRoot   [32]byte
	TotalAmount  *big.Int
	ReceiptCount uint64
	Timestamp    time.Time
	Receipts     []Receipt
}

// Receipt represents an individual receipt
type Receipt struct {
	ReceiptID    string
	ProviderAddr common.Address
	ConsumerAddr common.Address
	PromptHash   [32]byte
	OutputHash   [32]byte
	TokensIn     uint64
	TokensOut    uint64
	Amount       *big.Int
	Timestamp    time.Time
	Signature    []byte
}

// Channel represents a payment channel
type Channel struct {
	ChannelID    [32]byte
	Provider     common.Address
	Consumer     common.Address
	Deposit      *big.Int
	Balance      *big.Int
	Nonce        uint64
	LastActivity time.Time
	IsOpen       bool
}

// SettlementProof contains data needed to settle a batch
type SettlementProof struct {
	BatchID      string
	ChannelID    [32]byte
	Amount       *big.Int
	Nonce        uint64
	MerkleProof  [][32]byte
	ProviderSig  []byte
	ConsumerSig  []byte
}

// TransactionStatus represents the status of a blockchain transaction
type TransactionStatus struct {
	TxHash        common.Hash
	Status        string // pending, confirmed, failed
	Confirmations uint64
	GasUsed       uint64
	BlockNumber   uint64
	Timestamp     time.Time
}

// ProviderStats tracks provider statistics
type ProviderStats struct {
	Address          common.Address
	TotalReceipts    uint64
	TotalTokensIn    uint64
	TotalTokensOut   uint64
	TotalEarned      *big.Int
	LastSettlement   time.Time
	ActiveChannels   uint64
	ReputationScore  uint64
}

// EpochInfo contains information about a settlement epoch
type EpochInfo struct {
	Epoch         uint64
	StartTime     time.Time
	EndTime       time.Time
	MerkleRoot    [32]byte
	TotalReceipts uint64
	TotalAmount   *big.Int
	Finalized     bool
	TxHash        common.Hash
}