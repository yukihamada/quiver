package blockchain

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
)

// Config holds blockchain configuration
type Config struct {
	RPCEndpoint      string
	ContractAddress  string
	TokenAddress     string
	PrivateKey       string
	ChainID          int64
	GasLimit         uint64
	MaxGasPrice      *big.Int
	ConfirmationWait time.Duration
}

// DefaultPolygonConfig returns default configuration for Polygon
func DefaultPolygonConfig() *Config {
	return &Config{
		RPCEndpoint:      "https://polygon-rpc.com",
		ChainID:          137, // Polygon mainnet
		GasLimit:         3000000,
		MaxGasPrice:      big.NewInt(500_000_000_000), // 500 Gwei
		ConfirmationWait: 30 * time.Second,
	}
}

// DefaultPolygonMumbaiConfig returns configuration for Polygon Mumbai testnet
func DefaultPolygonMumbaiConfig() *Config {
	return &Config{
		RPCEndpoint:      "https://rpc-mumbai.maticvigil.com",
		ChainID:          80001, // Mumbai testnet
		GasLimit:         3000000,
		MaxGasPrice:      big.NewInt(50_000_000_000), // 50 Gwei
		ConfirmationWait: 15 * time.Second,
	}
}

// Client interfaces with the blockchain
type Client struct {
	ethClient       *ethclient.Client
	privateKey      *ecdsa.PrivateKey
	publicAddress   common.Address
	config          *Config
	settlementAddr  common.Address
	tokenAddr       common.Address
	transactOpts    *bind.TransactOpts
}

// NewClient creates a new blockchain client
func NewClient(config *Config) (*Client, error) {
	// Connect to Ethereum node
	ethClient, err := ethclient.Dial(config.RPCEndpoint)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to blockchain: %w", err)
	}

	// Load private key
	privateKey, err := crypto.HexToECDSA(config.PrivateKey)
	if err != nil {
		return nil, fmt.Errorf("failed to load private key: %w", err)
	}

	publicKey := privateKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		return nil, fmt.Errorf("failed to cast public key to ECDSA")
	}

	fromAddress := crypto.PubkeyToAddress(*publicKeyECDSA)

	// Create transaction options
	transactOpts, err := bind.NewKeyedTransactorWithChainID(privateKey, big.NewInt(config.ChainID))
	if err != nil {
		return nil, fmt.Errorf("failed to create transactor: %w", err)
	}
	transactOpts.GasLimit = config.GasLimit
	transactOpts.GasPrice = config.MaxGasPrice

	return &Client{
		ethClient:      ethClient,
		privateKey:     privateKey,
		publicAddress:  fromAddress,
		config:         config,
		settlementAddr: common.HexToAddress(config.ContractAddress),
		tokenAddr:      common.HexToAddress(config.TokenAddress),
		transactOpts:   transactOpts,
	}, nil
}

// GetBalance returns the ETH balance of the client address
func (c *Client) GetBalance(ctx context.Context) (*big.Int, error) {
	return c.ethClient.BalanceAt(ctx, c.publicAddress, nil)
}

// GetTokenBalance returns the QUIV token balance
func (c *Client) GetTokenBalance(ctx context.Context, address common.Address) (*big.Int, error) {
	// This would call the balanceOf method on the token contract
	// Simplified for now
	return big.NewInt(0), nil
}

// OpenChannel opens a payment channel
func (c *Client) OpenChannel(ctx context.Context, provider common.Address, amount *big.Int) (*types.Transaction, error) {
	// This would interact with the settlement contract to open a channel
	// Simplified implementation
	nonce, err := c.ethClient.PendingNonceAt(ctx, c.publicAddress)
	if err != nil {
		return nil, err
	}

	gasPrice, err := c.ethClient.SuggestGasPrice(ctx)
	if err != nil {
		return nil, err
	}

	tx := types.NewTransaction(
		nonce,
		c.settlementAddr,
		big.NewInt(0), // Value in ETH
		c.config.GasLimit,
		gasPrice,
		nil, // Data would contain the encoded function call
	)

	chainID := big.NewInt(c.config.ChainID)
	signedTx, err := types.SignTx(tx, types.NewEIP155Signer(chainID), c.privateKey)
	if err != nil {
		return nil, err
	}

	err = c.ethClient.SendTransaction(ctx, signedTx)
	if err != nil {
		return nil, err
	}

	return signedTx, nil
}

// SubmitBatch submits a batch of receipts for settlement
func (c *Client) SubmitBatch(ctx context.Context, batch ReceiptBatch) (*types.Transaction, error) {
	// This would encode and submit the batch to the settlement contract
	// Simplified implementation
	return c.OpenChannel(ctx, c.publicAddress, big.NewInt(0))
}

// WaitForConfirmation waits for a transaction to be confirmed
func (c *Client) WaitForConfirmation(ctx context.Context, tx *types.Transaction) (*types.Receipt, error) {
	ticker := time.NewTicker(3 * time.Second)
	defer ticker.Stop()

	timeout := time.After(c.config.ConfirmationWait)

	for {
		select {
		case <-ticker.C:
			receipt, err := c.ethClient.TransactionReceipt(ctx, tx.Hash())
			if err == nil {
				return receipt, nil
			}
		case <-timeout:
			return nil, fmt.Errorf("transaction confirmation timeout")
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}
}

// EstimateGas estimates gas for a transaction
func (c *Client) EstimateGas(ctx context.Context, to common.Address, data []byte) (uint64, error) {
	msg := ethereum.CallMsg{
		From:  c.publicAddress,
		To:    &to,
		Data:  data,
		Value: big.NewInt(0),
	}
	return c.ethClient.EstimateGas(ctx, msg)
}

// Close closes the client connection
func (c *Client) Close() {
	c.ethClient.Close()
}