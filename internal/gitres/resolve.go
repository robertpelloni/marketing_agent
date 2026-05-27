package gitres

import (
	"fmt"
	"log"
	"os/exec"

	"github.com/robertpelloni/enterprise_sales_bot/internal/gitcheck"
)

// ResolveConflict performs a 'git merge' of the source branch into the current branch.
// It accepts an optional strategy (e.g., 'ours', 'theirs') to automatically resolve conflicts.
func ResolveConflict(source string, strategy string) error {
	args := []string{"merge", source}
	if strategy != "" {
		args = append(args, "-X", strategy)
	}

	cmd := exec.Command("git", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("merge failed: %v, output: %s", err, string(output))
	}
	return nil
}

// AbortMerge resets the current merge state using 'git merge --abort'.
// This is used for recovery when an autonomous merge fails to resolve cleanly.
func AbortMerge() error {
	cmd := exec.Command("git", "merge", "--abort")
	return cmd.Run()
}

// ReconcileBranches implements the Dual-Direction Intelligent Merge Engine.
// It iterates through all 'autodev/' branches, attempting to forward-merge them into 'main'
// and reverse-merge 'main' back into each feature branch to prevent drift.
func ReconcileBranches() error {
	branches, err := gitcheck.ListFeatureBranches()
	if err != nil {
		return err
	}

	for _, branch := range branches {
		log.Printf("Intelligent Merge: Reconciling branch: %s", branch)

		// 1. Forward Merge: Feature -> Main
		log.Printf("Intelligent Merge: Attempting Forward Merge (%s -> main)...", branch)
		if out, err := exec.Command("git", "checkout", "main").CombinedOutput(); err != nil {
			return fmt.Errorf("failed to checkout main: %v, output: %s", err, string(out))
		}
		if err := ResolveConflict(branch, "theirs"); err != nil {
			log.Printf("Intelligent Merge: Forward merge failed for %s: %v", branch, err)
		} else {
			log.Printf("Intelligent Merge: Successfully merged %s into main", branch)
		}

		// 2. Reverse Merge: Main -> Feature
		log.Printf("Intelligent Merge: Attempting Reverse Merge (main -> %s)...", branch)
		if out, err := exec.Command("git", "checkout", branch).CombinedOutput(); err != nil {
			return fmt.Errorf("failed to checkout feature branch %s: %v, output: %s", branch, err, string(out))
		}
		// We use standard merge to catch drift and ensure metadata (TODO, VERSION) is preserved correctly
		if err := ResolveConflict("main", ""); err != nil {
			log.Printf("Intelligent Merge: Reverse merge failed for %s: %v", branch, err)
		} else {
			log.Printf("Intelligent Merge: Successfully reconciled %s with main", branch)
		}
	}

	// Switch back to main
	if out, err := exec.Command("git", "checkout", "main").CombinedOutput(); err != nil {
		return fmt.Errorf("failed to return to main: %v, output: %s", err, string(out))
	}
	return nil
}
