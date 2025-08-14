package receipt

import (
	"crypto/ed25519"
	"crypto/rand"
	"encoding/base64"
	"os"
	"testing"
	"time"
)

func TestEd25519SignVerify(t *testing.T) {
	keyPath := "test.key"
	defer os.Remove(keyPath)

	signer, err := NewSigner(keyPath)
	if err != nil {
		t.Fatal(err)
	}

	receipt := NewReceipt(
		signer.PublicKeyBase64(),
		"test-model",
		"prompt-hash",
		"output-hash",
		10, 20,
		time.Now(), time.Now().Add(100*time.Millisecond),
	)

	signed, err := signer.Sign(receipt)
	if err != nil {
		t.Fatal(err)
	}

	valid, err := VerifySignature(signed, signer.PublicKeyBase64())
	if err != nil {
		t.Fatal(err)
	}

	if !valid {
		t.Error("Valid signature failed verification")
	}

	// Test invalid signature
	signed.Signature = base64.StdEncoding.EncodeToString([]byte("invalid"))
	valid, err = VerifySignature(signed, signer.PublicKeyBase64())
	if err != nil {
		t.Fatal(err)
	}

	if valid {
		t.Error("Invalid signature passed verification")
	}
}

func TestSignerKeyPersistence(t *testing.T) {
	keyPath := "test-persist.key"
	defer os.Remove(keyPath)

	// Create new key
	signer1, err := NewSigner(keyPath)
	if err != nil {
		t.Fatal(err)
	}
	pubKey1 := signer1.PublicKeyBase64()

	// Load existing key
	signer2, err := NewSigner(keyPath)
	if err != nil {
		t.Fatal(err)
	}
	pubKey2 := signer2.PublicKeyBase64()

	if pubKey1 != pubKey2 {
		t.Error("Key not persisted correctly")
	}
}

func TestSignatureConsistency(t *testing.T) {
	_, privKey, _ := ed25519.GenerateKey(rand.Reader)
	signer := &Signer{
		privateKey: privKey,
		publicKey:  privKey.Public().(ed25519.PublicKey),
	}

	receipt := NewReceipt(
		signer.PublicKeyBase64(),
		"model",
		"hash1",
		"hash2",
		5, 10,
		time.Now(), time.Now(),
	)

	// Sign multiple times
	signatures := make([]string, 10)
	for i := 0; i < 10; i++ {
		signed, err := signer.Sign(receipt)
		if err != nil {
			t.Fatal(err)
		}
		signatures[i] = signed.Signature
	}

	// All signatures should be identical
	for i := 1; i < len(signatures); i++ {
		if signatures[i] != signatures[0] {
			t.Error("Signatures not deterministic")
		}
	}
}
