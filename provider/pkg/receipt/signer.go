package receipt

import (
	"crypto/ed25519"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"os"
)

type Signer struct {
	privateKey ed25519.PrivateKey
	publicKey  ed25519.PublicKey
}

func NewSigner(keyPath string) (*Signer, error) {
	var privateKey ed25519.PrivateKey

	if data, err := os.ReadFile(keyPath); err == nil {
		if len(data) == ed25519.PrivateKeySize {
			privateKey = ed25519.PrivateKey(data)
		} else {
			return nil, fmt.Errorf("invalid key size")
		}
	} else {
		publicKey, privKey, err := ed25519.GenerateKey(rand.Reader)
		if err != nil {
			return nil, err
		}
		privateKey = privKey

		if err := os.WriteFile(keyPath, privateKey, 0600); err != nil {
			return nil, err
		}

		return &Signer{
			privateKey: privateKey,
			publicKey:  publicKey,
		}, nil
	}

	publicKey := privateKey.Public().(ed25519.PublicKey)

	return &Signer{
		privateKey: privateKey,
		publicKey:  publicKey,
	}, nil
}

func (s *Signer) Sign(receipt *Receipt) (*SignedReceipt, error) {
	canonical, err := CanonicalizeJSON(receipt)
	if err != nil {
		return nil, err
	}

	signature := ed25519.Sign(s.privateKey, canonical)

	return &SignedReceipt{
		Receipt:   *receipt,
		Signature: base64.StdEncoding.EncodeToString(signature),
	}, nil
}

func (s *Signer) PublicKeyBase64() string {
	return base64.StdEncoding.EncodeToString(s.publicKey)
}

func VerifySignature(receipt *SignedReceipt, publicKeyBase64 string) (bool, error) {
	publicKey, err := base64.StdEncoding.DecodeString(publicKeyBase64)
	if err != nil {
		return false, err
	}

	signature, err := base64.StdEncoding.DecodeString(receipt.Signature)
	if err != nil {
		return false, err
	}

	canonical, err := CanonicalizeJSON(receipt.Receipt)
	if err != nil {
		return false, err
	}

	return ed25519.Verify(ed25519.PublicKey(publicKey), canonical, signature), nil
}
