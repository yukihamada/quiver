package epoch

import (
	"testing"
	"time"
)

func TestCurrentEpoch(t *testing.T) {
	manager := NewManager()

	epoch1 := manager.CurrentEpoch()
	time.Sleep(100 * time.Millisecond)
	epoch2 := manager.CurrentEpoch()

	if epoch1 != epoch2 {
		t.Error("Epoch should not change within same day")
	}

	expectedEpoch := uint64(time.Now().UTC().Unix() / 86400)
	if epoch1 != expectedEpoch {
		t.Errorf("Expected epoch %d, got %d", expectedEpoch, epoch1)
	}
}

func TestGetOrCreateEpoch(t *testing.T) {
	manager := NewManager()

	epoch := uint64(19723)
	info1 := manager.GetOrCreateEpoch(epoch)
	info2 := manager.GetOrCreateEpoch(epoch)

	if info1 != info2 {
		t.Error("Should return same epoch info")
	}

	if info1.Epoch != epoch {
		t.Errorf("Expected epoch %d, got %d", epoch, info1.Epoch)
	}

	if info1.Finalized {
		t.Error("New epoch should not be finalized")
	}
}

func TestFinalizeEpoch(t *testing.T) {
	manager := NewManager()

	epoch := uint64(19723)
	root := "test_root_hash"
	count := 100

	err := manager.FinalizeEpoch(epoch, root, count)
	if err != nil {
		t.Fatal(err)
	}

	info, exists := manager.GetEpochInfo(epoch)
	if !exists {
		t.Fatal("Epoch not found after finalization")
	}

	if !info.Finalized {
		t.Error("Epoch should be finalized")
	}

	if info.Root != root {
		t.Errorf("Expected root %s, got %s", root, info.Root)
	}

	if info.ReceiptCount != count {
		t.Errorf("Expected count %d, got %d", count, info.ReceiptCount)
	}
}
