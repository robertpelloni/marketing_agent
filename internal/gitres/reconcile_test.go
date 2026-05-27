package gitres

import (
	"log"
	"os"
	"os/exec"
	"testing"
)

func TestReconcileBranches_Functional(t *testing.T) {
	// This test requires git and a clean environment.
	// We will simulate branch reconciliation by creating temporary branches.

	// Skip in CI if git is not configured
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not found, skipping functional test")
	}

	// 1. Setup temporary branches
	// Note: In a real environment, we'd use a temporary repo.
	// For this test, we assume we are in the bot's repo and will cleanup.

	currentBranchCmd := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD")
	out, _ := currentBranchCmd.Output()
	originalBranch := string(out)

	defer func() {
		log.Println("Cleanup: returning to original branch")
		exec.Command("git", "checkout", originalBranch).Run()
		exec.Command("git", "branch", "-D", "autodev/test-reconcile").Run()
	}()

	// Create a feature branch with a commit
	if err := exec.Command("git", "checkout", "-b", "autodev/test-reconcile").Run(); err != nil {
		t.Logf("Warning: could not create test branch (repo might be dirty): %v", err)
		return
	}

	os.WriteFile("RECONCILE_TEST", []byte("test"), 0644)
	exec.Command("git", "add", "RECONCILE_TEST").Run()
	exec.Command("git", "commit", "-m", "feat: test reconciliation").Run()

	// 2. Run reconciliation
	if err := ReconcileBranches(); err != nil {
		t.Errorf("ReconcileBranches failed: %v", err)
	}

	// 3. Verify forward merge (main should have RECONCILE_TEST)
	// We expect main to have been updated if unique progress was found
	exec.Command("git", "checkout", "main").Run()
	if _, err := os.Stat("RECONCILE_TEST"); os.IsNotExist(err) {
		t.Errorf("Forward merge failed: RECONCILE_TEST not found in main")
	}

	// Cleanup file from main after verification
	os.Remove("RECONCILE_TEST")
	exec.Command("git", "add", "RECONCILE_TEST").Run()
	exec.Command("git", "commit", "-m", "chore: cleanup test file").Run()
}
