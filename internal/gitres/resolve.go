package gitres

import (
	"fmt"
	"log"
	"os/exec"

	"github.com/robertpelloni/enterprise_sales_bot/internal/gitcheck"
)

// ResolveConflict performs a merge of the source branch into the current branch
// using the specified strategy.
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

// AbortMerge aborts an ongoing merge.
func AbortMerge() error {
	cmd := exec.Command("git", "merge", "--abort")
	return cmd.Run()
}

// ReconcileBranches implements the Dual-Direction Intelligent Merge Engine.
func ReconcileBranches() error {
	branches, err := gitcheck.ListFeatureBranches()
	if err != nil {
		return err
	}

	for _, branch := range branches {
		log.Printf("Reconciling branch: %s", branch)

		// 1. Forward Merge: Feature -> Main (if on main)
		// 2. Reverse Merge: Main -> Feature
		// For simplicity in this implementation, we focus on the Reverse Merge to prevent drift.

		// Checkout feature branch
		checkoutCmd := exec.Command("git", "checkout", branch)
		if err := checkoutCmd.Run(); err != nil {
			log.Printf("Failed to checkout %s: %v", branch, err)
			continue
		}

		// Merge main into feature
		if err := ResolveConflict("main", "ours"); err != nil {
			log.Printf("Failed to reconcile %s with main: %v", branch, err)
		} else {
			log.Printf("Successfully reconciled %s with main", branch)
		}
	}

	// Switch back to main
	exec.Command("git", "checkout", "main").Run()
	return nil
}
