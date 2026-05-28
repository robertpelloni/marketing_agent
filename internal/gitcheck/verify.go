package gitcheck

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// IsClean checks if the git working directory is clean using 'git status --porcelain'.
// It returns true if there are no staged or unstaged changes.
func IsClean() (bool, error) {
	// In test environments, we allow dirty state for orchestrator unit tests
	if os.Getenv("GO_TEST_MODE") == "true" {
		return true, nil
	}
	cmd := exec.Command("git", "status", "--porcelain")
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		return false, err
	}
	return out.Len() == 0, nil
}

// IsSynced checks if the current branch is synchronized with the remote 'origin/target'.
// It performs a 'git fetch' first to ensure local awareness of remote state.
func IsSynced(target string) (bool, error) {
	// Fetch first to ensure we have latest info
	if err := exec.Command("git", "fetch", "origin").Run(); err != nil {
		return false, fmt.Errorf("failed to fetch from origin: %w", err)
	}

	cmd := exec.Command("git", "rev-list", "--left-right", "--count", "HEAD...origin/"+target)
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		return false, fmt.Errorf("failed to compare branches: %v", err)
	}

	// Output format: "ahead\tbehind"
	counts := strings.Fields(out.String())
	if len(counts) != 2 {
		return false, fmt.Errorf("unexpected rev-list output: %s", out.String())
	}

	// For integrity, we mostly care if we are behind (second number > 0)
	behind := counts[1]
	if behind != "0" {
		return false, nil
	}

	return true, nil
}

// SyncRemote fetches and merges changes from 'origin/main'.
// It uses '--no-edit' to ensure the operation remains autonomous and non-interactive.
func SyncRemote() error {
	fetchCmd := exec.Command("git", "fetch", "origin", "main")
	if err := fetchCmd.Run(); err != nil {
		return fmt.Errorf("failed to fetch from origin: %v", err)
	}

	mergeCmd := exec.Command("git", "merge", "origin/main", "-m", "chore: autonomous sync with origin/main", "--no-edit")
	if err := mergeCmd.Run(); err != nil {
		return fmt.Errorf("failed to merge from origin/main: %v", err)
	}

	return nil
}

// UpdateSubmodules updates all git submodules recursively within the repository.
// It initializes submodules if they haven't been already.
func UpdateSubmodules() error {
	cmd := exec.Command("git", "submodule", "update", "--init", "--recursive")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to update submodules: %v", err)
	}
	return nil
}

// CheckoutAndCommit creates or resets a branch and commits all current working directory changes.
// This is typically used by the autonomous agent to persist its self-directed updates.
func CheckoutAndCommit(branch string, message string) error {
	checkoutCmd := exec.Command("git", "checkout", "-B", branch)
	if err := checkoutCmd.Run(); err != nil {
		return fmt.Errorf("failed to checkout branch %s: %v", branch, err)
	}

	addCmd := exec.Command("git", "add", ".")
	if err := addCmd.Run(); err != nil {
		return fmt.Errorf("failed to stage changes: %v", err)
	}

	commitCmd := exec.Command("git", "commit", "-m", message)
	if err := commitCmd.Run(); err != nil {
		return fmt.Errorf("failed to commit changes: %v", err)
	}

	return nil
}

// PushBranch pushes the specified local branch to the 'origin' remote.
func PushBranch(branch string) error {
	cmd := exec.Command("git", "push", "origin", branch)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to push branch %s: %v", branch, err)
	}
	return nil
}

// ListFeatureBranches returns a list of all local branches prefixed with 'autodev/'.
// These represent the active feature branches managed by the autonomous orchestrator.
func ListFeatureBranches() ([]string, error) {
	cmd := exec.Command("git", "for-each-ref", "--format=%(refname:short)", "refs/heads/")
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		return nil, err
	}

	var branches []string
	scanner := bufio.NewScanner(&out)
	for scanner.Scan() {
		branch := strings.TrimSpace(scanner.Text())
		// Only reconcile branches explicitly created by the autonomous agent
		if strings.HasPrefix(branch, "autodev/") {
			branches = append(branches, branch)
		}
	}
	return branches, nil
}

// CheckConflicts checks the repository for any unmerged paths (merge conflicts).
// It returns true if any conflict markers are detected in the index.
func CheckConflicts() (bool, error) {
	cmd := exec.Command("git", "diff", "--name-only", "--diff-filter=U")
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		return false, err
	}
	return out.Len() > 0, nil
}
