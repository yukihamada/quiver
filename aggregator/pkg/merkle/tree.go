package merkle

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sort"
)

type Tree struct {
	leaves      [][]byte
	sortedIndex []int
	nodes       [][]byte
	root        []byte
}

func NewTree() *Tree {
	return &Tree{
		leaves: make([][]byte, 0),
	}
}

func (t *Tree) AddLeaf(data []byte) {
	leaf := sha256.Sum256(data)
	t.leaves = append(t.leaves, leaf[:])
}

func (t *Tree) Build() error {
	if len(t.leaves) == 0 {
		return fmt.Errorf("no leaves to build tree")
	}

	// Create sorted index for deterministic ordering
	t.sortedIndex = make([]int, len(t.leaves))
	for i := range t.sortedIndex {
		t.sortedIndex[i] = i
	}

	sort.Slice(t.sortedIndex, func(i, j int) bool {
		return hex.EncodeToString(t.leaves[t.sortedIndex[i]]) < hex.EncodeToString(t.leaves[t.sortedIndex[j]])
	})

	// Build tree with sorted leaves
	sortedLeaves := make([][]byte, len(t.leaves))
	for i, idx := range t.sortedIndex {
		sortedLeaves[i] = t.leaves[idx]
	}

	level := make([][]byte, len(sortedLeaves))
	copy(level, sortedLeaves)

	t.nodes = make([][]byte, 0)
	t.nodes = append(t.nodes, sortedLeaves...)

	for len(level) > 1 {
		nextLevel := make([][]byte, 0)

		for i := 0; i < len(level); i += 2 {
			var combined []byte
			if i+1 < len(level) {
				combined = append(combined, level[i]...)
				combined = append(combined, level[i+1]...)
			} else {
				combined = append(combined, level[i]...)
				combined = append(combined, level[i]...)
			}

			hash := sha256.Sum256(combined)
			nextLevel = append(nextLevel, hash[:])
			t.nodes = append(t.nodes, hash[:])
		}

		level = nextLevel
	}

	t.root = level[0]
	return nil
}

func (t *Tree) Root() string {
	if t.root == nil {
		return ""
	}
	return hex.EncodeToString(t.root)
}

func (t *Tree) Proof(leafIndex int) ([]string, error) {
	if leafIndex >= len(t.leaves) {
		return nil, fmt.Errorf("leaf index out of range")
	}

	// Find sorted position
	sortedPos := -1
	for i, idx := range t.sortedIndex {
		if idx == leafIndex {
			sortedPos = i
			break
		}
	}

	if sortedPos == -1 {
		return nil, fmt.Errorf("leaf not found in sorted index")
	}

	proof := make([]string, 0)
	index := sortedPos
	levelStart := 0
	levelSize := len(t.leaves)

	for levelSize > 1 {
		siblingIndex := index ^ 1
		if siblingIndex < levelSize {
			proof = append(proof, hex.EncodeToString(t.nodes[levelStart+siblingIndex]))
		} else {
			proof = append(proof, hex.EncodeToString(t.nodes[levelStart+index]))
		}

		index /= 2
		levelStart += levelSize
		levelSize = (levelSize + 1) / 2
	}

	return proof, nil
}

func Verify(leafData []byte, proof []string, rootHex string) bool {
	hash := sha256.Sum256(leafData)
	current := hash[:]

	for _, proofElement := range proof {
		proofBytes, err := hex.DecodeString(proofElement)
		if err != nil {
			return false
		}

		var combined []byte
		if hex.EncodeToString(current) < proofElement {
			combined = append(current, proofBytes...)
		} else {
			combined = append(proofBytes, current...)
		}

		nextHash := sha256.Sum256(combined)
		current = nextHash[:]
	}

	return hex.EncodeToString(current) == rootHex
}
