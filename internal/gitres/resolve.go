package gitres

import (
	"fmt"
	"log/slog"
	"os/exec"
	"strings"
)

func ReconcileBranches() error {
	// 1. Get all local branches under the specific owner/namespace
	out, err := exec.Command("git", "branch", "--list", "autodev/*").Output()
	if err != nil {
		return err
	}

	branches := strings.Fields(string(out))
	for _, branch := range branches {
		branch = strings.TrimPrefix(branch, "*")
		slog.Info("Intelligent Merge: Reconciling branch", "branch", branch)

		// 2. Forward Merge: Attempt to merge feature into main
		if err := runGit("checkout", "main"); err != nil {
			return err
		}
		slog.Info("Intelligent Merge: Attempting Forward Merge", "branch", branch, "target", "main")
		if err := runGit("merge", branch); err != nil {
			slog.Warn("Intelligent Merge: Forward merge failed, resolving conflicts", "branch", branch, "error", err)
			if err := resolveConflicts(); err != nil {
				return err
			}
		}
		slog.Info("Intelligent Merge: Successfully merged into main", "branch", branch)

		// 3. Reverse Merge: Catch up feature branch with updated main
		if err := runGit("checkout", branch); err != nil {
			continue
		}
		slog.Info("Intelligent Merge: Attempting Reverse Merge", "source", "main", "branch", branch)
		if err := runGit("merge", "main"); err != nil {
			slog.Warn("Intelligent Merge: Reverse merge failed", "branch", branch, "error", err)
			_ = AbortMerge()
			continue
		}
		slog.Info("Intelligent Merge: Successfully reconciled with main", "branch", branch)
	}

	if err := runGit("checkout", "main"); err != nil {
		return err
	}

	slog.Info("Intelligent Merge: Finalizing reconciliation by pushing main to origin")
	if err := runGit("push", "origin", "main"); err != nil {
		slog.Warn("Intelligent Merge: Final push failed", "error", err)
	}

	return nil
}

func resolveConflicts() error {
	out, _ := exec.Command("git", "diff", "--name-only", "--diff-filter=U").Output()
	conflicts := strings.TrimSpace(string(out))
	if conflicts != "" {
		slog.Info("Intelligent Merge: Conflicts detected", "files", conflicts)
		return runGit("checkout", "--ours", ".")
	}
	return nil
}

func AbortMerge() error {
	return runGit("merge", "--abort")
}

func runGit(args ...string) error {
	cmd := exec.Command("git", args...)
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("git %s failed: %v, output: %s", args[0], err, string(out))
	}
	return nil
}
