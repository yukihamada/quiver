package p2p

import (
	"crypto/rand"
	"io/ioutil"
	"os"

	"github.com/libp2p/go-libp2p/core/crypto"
)

// LoadOrGenerateKey loads a private key from file or generates a new one
func LoadOrGenerateKey(keyPath string) (crypto.PrivKey, crypto.PubKey, error) {
	if keyPath == "" {
		// Generate ephemeral key
		return crypto.GenerateKeyPairWithReader(crypto.Ed25519, -1, rand.Reader)
	}

	// Check if key file exists
	if _, err := os.Stat(keyPath); err == nil {
		// Load existing key
		keyBytes, err := ioutil.ReadFile(keyPath)
		if err != nil {
			return nil, nil, err
		}

		priv, err := crypto.UnmarshalPrivateKey(keyBytes)
		if err != nil {
			return nil, nil, err
		}

		return priv, priv.GetPublic(), nil
	}

	// Generate new key
	priv, pub, err := crypto.GenerateKeyPairWithReader(crypto.Ed25519, -1, rand.Reader)
	if err != nil {
		return nil, nil, err
	}

	// Save key to file
	keyBytes, err := crypto.MarshalPrivateKey(priv)
	if err != nil {
		return nil, nil, err
	}

	if err := ioutil.WriteFile(keyPath, keyBytes, 0600); err != nil {
		return nil, nil, err
	}

	return priv, pub, nil
}