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
		if err := ResolveConflict("main", "ours"); err != nil {
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
