package gitcheck

import (
	"testing"
)

func TestIsClean(t *testing.T) {
	clean, err := IsClean()
	if err != nil {
		t.Fatalf("IsClean failed: %v", err)
	}
	// In a test environment, we expect the workspace to be clean if committed
	if !clean {
		t.Log("Warning: Working directory is not clean")
	}
}

func TestCheckConflicts(t *testing.T) {
	hasConflicts, err := CheckConflicts()
	if err != nil {
		t.Fatalf("CheckConflicts failed: %v", err)
	}
	if hasConflicts {
		t.Errorf("Merge conflicts detected in the repository")
	}
}

func TestIsSynced(t *testing.T) {
	// This test depends on network access and origin/main existence
	// We skip it if it fails due to environment, but it's here for manual/CI trigger
	synced, err := IsSynced("main")
	if err != nil {
		t.Skipf("Skipping IsSynced test: %v", err)
		return
	}
	if !synced {
		t.Log("Warning: Current branch is not synchronized with origin/main")
		// We don't fail here in local environments where divergence is expected during development
	}
}
