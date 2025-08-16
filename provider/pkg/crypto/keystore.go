package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"os"

	"golang.org/x/crypto/pbkdf2"
)

const (
	saltSize   = 32
	keySize    = 32
	iterations = 100000
)

// EncryptedKey represents an encrypted private key
type EncryptedKey struct {
	Algorithm  string `json:"algorithm"`
	Salt       string `json:"salt"`
	IV         string `json:"iv"`
	Ciphertext string `json:"ciphertext"`
	Iterations int    `json:"iterations"`
}

// Keystore handles secure key storage
type Keystore struct {
	filepath string
}

// NewKeystore creates a new keystore
func NewKeystore(filepath string) *Keystore {
	return &Keystore{
		filepath: filepath,
	}
}

// StoreKey encrypts and stores a private key
func (ks *Keystore) StoreKey(privateKey []byte, passphrase string) error {
	// Generate salt
	salt := make([]byte, saltSize)
	if _, err := io.ReadFull(rand.Reader, salt); err != nil {
		return fmt.Errorf("failed to generate salt: %w", err)
	}

	// Derive key from passphrase
	key := pbkdf2.Key([]byte(passphrase), salt, iterations, keySize, sha256.New)

	// Create cipher
	block, err := aes.NewCipher(key)
	if err != nil {
		return fmt.Errorf("failed to create cipher: %w", err)
	}

	// Generate IV
	iv := make([]byte, aes.BlockSize)
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return fmt.Errorf("failed to generate IV: %w", err)
	}

	// Encrypt the key
	stream := cipher.NewCFBEncrypter(block, iv)
	ciphertext := make([]byte, len(privateKey))
	stream.XORKeyStream(ciphertext, privateKey)

	// Create encrypted key structure
	encKey := EncryptedKey{
		Algorithm:  "AES-256-CFB",
		Salt:       base64.StdEncoding.EncodeToString(salt),
		IV:         base64.StdEncoding.EncodeToString(iv),
		Ciphertext: base64.StdEncoding.EncodeToString(ciphertext),
		Iterations: iterations,
	}

	// Save to file
	data, err := json.MarshalIndent(encKey, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal encrypted key: %w", err)
	}

	// Write with restricted permissions
	if err := os.WriteFile(ks.filepath, data, 0600); err != nil {
		return fmt.Errorf("failed to write keystore: %w", err)
	}

	return nil
}

// LoadKey decrypts and loads a private key
func (ks *Keystore) LoadKey(passphrase string) ([]byte, error) {
	// Read encrypted key
	data, err := os.ReadFile(ks.filepath)
	if err != nil {
		return nil, fmt.Errorf("failed to read keystore: %w", err)
	}

	var encKey EncryptedKey
	if err := json.Unmarshal(data, &encKey); err != nil {
		return nil, fmt.Errorf("failed to unmarshal encrypted key: %w", err)
	}

	// Decode base64
	salt, err := base64.StdEncoding.DecodeString(encKey.Salt)
	if err != nil {
		return nil, fmt.Errorf("failed to decode salt: %w", err)
	}

	iv, err := base64.StdEncoding.DecodeString(encKey.IV)
	if err != nil {
		return nil, fmt.Errorf("failed to decode IV: %w", err)
	}

	ciphertext, err := base64.StdEncoding.DecodeString(encKey.Ciphertext)
	if err != nil {
		return nil, fmt.Errorf("failed to decode ciphertext: %w", err)
	}

	// Derive key from passphrase
	key := pbkdf2.Key([]byte(passphrase), salt, encKey.Iterations, keySize, sha256.New)

	// Create cipher
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("failed to create cipher: %w", err)
	}

	// Decrypt
	stream := cipher.NewCFBDecrypter(block, iv)
	plaintext := make([]byte, len(ciphertext))
	stream.XORKeyStream(plaintext, ciphertext)

	return plaintext, nil
}

// Exists checks if keystore file exists
func (ks *Keystore) Exists() bool {
	_, err := os.Stat(ks.filepath)
	return err == nil
}

// GenerateAndStore generates a new key and stores it encrypted
func (ks *Keystore) GenerateAndStore(keyType string, passphrase string) error {
	var privateKey []byte
	var err error

	switch keyType {
	case "ed25519":
		privateKey, err = GenerateEd25519Key()
	case "secp256k1":
		privateKey, err = GenerateSecp256k1Key()
	default:
		return fmt.Errorf("unsupported key type: %s", keyType)
	}

	if err != nil {
		return fmt.Errorf("failed to generate key: %w", err)
	}

	return ks.StoreKey(privateKey, passphrase)
}

// GenerateEd25519Key generates a new Ed25519 private key
func GenerateEd25519Key() ([]byte, error) {
	// Implementation would use crypto/ed25519
	// Placeholder for now
	key := make([]byte, 32)
	_, err := io.ReadFull(rand.Reader, key)
	return key, err
}

// GenerateSecp256k1Key generates a new secp256k1 private key
func GenerateSecp256k1Key() ([]byte, error) {
	// Implementation would use appropriate crypto library
	// Placeholder for now
	key := make([]byte, 32)
	_, err := io.ReadFull(rand.Reader, key)
	return key, err
}