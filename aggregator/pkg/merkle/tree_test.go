package merkle

import (
	"crypto/sha256"
	"encoding/hex"
	"testing"
)

func TestMerkleRootProofVerify(t *testing.T) {
	tree := NewTree()

	// Add test data
	data := []string{"receipt1", "receipt2", "receipt3", "receipt4"}
	for _, d := range data {
		tree.AddLeaf([]byte(d))
	}

	err := tree.Build()
	if err != nil {
		t.Fatal(err)
	}

	root := tree.Root()

	// Test verification for each original leaf
	for i, d := range data {
		proof, err := tree.Proof(i)
		if err != nil {
			t.Fatal(err)
		}

		// Create a new tree to find sorted position
		verifyTree := NewTree()
		for _, item := range data {
			verifyTree.AddLeaf([]byte(item))
		}
		if err := verifyTree.Build(); err != nil {
			t.Fatal(err)
		}

		// Find sorted position of this leaf
		leafHash := sha256.Sum256([]byte(d))
		sortedPos := -1
		for j, idx := range verifyTree.sortedIndex {
			if idx == i {
				sortedPos = j
				break
			}
		}

		// Manual verification
		current := leafHash[:]
		index := sortedPos
		for _, proofElem := range proof {
			proofBytes, _ := hex.DecodeString(proofElem)

			var combined []byte
			if index%2 == 0 {
				combined = append(current, proofBytes...)
			} else {
				combined = append(proofBytes, current...)
			}

			nextHash := sha256.Sum256(combined)
			current = nextHash[:]
			index /= 2
		}

		if hex.EncodeToString(current) != root {
			t.Errorf("Manual verification failed for leaf %d", i)
		}
	}
}

func TestMerkleTreeDeterminism(t *testing.T) {
	data := []string{"a", "b", "c", "d", "e"}

	// Build tree multiple times
	roots := make([]string, 5)
	for i := 0; i < 5; i++ {
		tree := NewTree()
		for _, d := range data {
			tree.AddLeaf([]byte(d))
		}
		if err := tree.Build(); err != nil {
			t.Fatal(err)
		}
		roots[i] = tree.Root()
	}

	// All roots should be identical
	for i := 1; i < len(roots); i++ {
		if roots[i] != roots[0] {
			t.Errorf("Tree not deterministic: %s != %s", roots[i], roots[0])
		}
	}
}

func TestMerkleOddLeaves(t *testing.T) {
	tree := NewTree()

	// Odd number of leaves
	data := []string{"one", "two", "three", "four", "five"}
	for _, d := range data {
		tree.AddLeaf([]byte(d))
	}

	err := tree.Build()
	if err != nil {
		t.Fatal(err)
	}

	// Find the last item's original index
	lastIndex := 4
	proof, err := tree.Proof(lastIndex)
	if err != nil {
		t.Fatal(err)
	}

	// Just verify proof is not empty for odd tree
	if len(proof) == 0 {
		t.Error("Empty proof for odd tree")
	}
}

func TestEmptyTree(t *testing.T) {
	tree := NewTree()

	err := tree.Build()
	if err == nil {
		t.Error("Empty tree should fail to build")
	}
}

func TestProofOutOfRange(t *testing.T) {
	tree := NewTree()
	tree.AddLeaf([]byte("test"))
	if err := tree.Build(); err != nil {
		t.Fatal(err)
	}

	_, err := tree.Proof(1)
	if err == nil {
		t.Error("Out of range proof should fail")
	}
}

func TestInvalidProof(t *testing.T) {
	// Test with corrupted proof
	leafData := []byte("test")
	validRoot := "abc123"
	invalidProof := []string{"not-hex", "123"}

	if Verify(leafData, invalidProof, validRoot) {
		t.Error("Invalid proof format should fail")
	}
}

func TestSimpleMerkleTree(t *testing.T) {
	tree := NewTree()
	tree.AddLeaf([]byte("A"))
	tree.AddLeaf([]byte("B"))

	err := tree.Build()
	if err != nil {
		t.Fatal(err)
	}

	root := tree.Root()
	if root == "" {
		t.Error("Expected non-empty root")
	}

	// Test proof for first leaf
	proof0, err := tree.Proof(0)
	if err != nil {
		t.Fatal(err)
	}

	if len(proof0) != 1 {
		t.Errorf("Expected 1 proof element, got %d", len(proof0))
	}
}

func BenchmarkMerkleTree(b *testing.B) {
	for n := 0; n < b.N; n++ {
		tree := NewTree()
		for i := 0; i < 1000; i++ {
			data := sha256.Sum256([]byte{byte(i), byte(i >> 8)})
			tree.AddLeaf(data[:])
		}
		_ = tree.Build() // Ignore error in benchmark
	}
}
