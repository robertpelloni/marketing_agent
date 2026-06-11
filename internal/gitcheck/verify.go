package gitcheck

import (
	"fmt"
	"log/slog"
	"os/exec"
	"strings"
)

func SyncRemote() error {
	slog.Info("Git Sync: Fetching all remotes")
	if err := runGit("fetch", "--all", "--tags"); err != nil {
		return err
	}

	// Check for upstream remote
	out, err := exec.Command("git", "remote").Output()
	if err == nil && strings.Contains(string(out), "upstream") {
		slog.Info("Git Sync: Upstream remote detected, merging changes")
		if err := runGit("merge", "upstream/main"); err != nil {
			slog.Warn("Git Sync: Upstream merge failed", "error", err)
		}
	}

	return runGit("pull", "origin", "main")
}

func UpdateSubmodules() error {
	slog.Info("Git Sync: Updating submodules recursively")
	return runGit("submodule", "update", "--init", "--recursive")
}

func IsClean() (bool, error) {
	out, err := exec.Command("git", "status", "--porcelain").Output()
	if err != nil {
		return false, err
	}
	return len(strings.TrimSpace(string(out))) == 0, nil
}

func CheckConflicts() (bool, error) {
	out, err := exec.Command("git", "diff", "--name-only", "--diff-filter=U").Output()
	if err != nil {
		return false, err
	}
	return len(strings.TrimSpace(string(out))) > 0, nil
}

func IsSynced(branch string) (bool, error) {
	// Check if local branch is behind origin
	if err := runGit("fetch", "origin"); err != nil {
		return false, err
	}
	out, err := exec.Command("git", "rev-list", "HEAD..origin/"+branch, "--count").Output()
	if err != nil {
		return false, err
	}
	count := strings.TrimSpace(string(out))
	return count == "0", nil
}

func ListFeatureBranches() ([]string, error) {
	out, err := exec.Command("git", "branch", "--list", "autodev/*").Output()
	if err != nil {
		return nil, err
	}
	var branches []string
	for _, b := range strings.Fields(string(out)) {
		branches = append(branches, strings.TrimPrefix(b, "*"))
	}
	return branches, nil
}

func CheckoutAndCommit(branch, message string) error {
	if err := runGit("checkout", "-b", branch); err != nil {
		if err := runGit("checkout", branch); err != nil {
			return err
		}
	}
	if err := runGit("add", "."); err != nil {
		return err
	}
	return runGit("commit", "-m", message)
}

func PushBranch(branch string) error {
	return runGit("push", "origin", branch)
}

func DeleteBranch(branch string) error {
	return runGit("branch", "-D", branch)
}

func DeleteRemoteBranch(branch string) error {
	return runGit("push", "origin", "--delete", branch)
}

func runGit(args ...string) error {
	cmd := exec.Command("git", args...)
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("git %s failed: %v, output: %s", args[0], err, string(out))
	}
	return nil
}

func GenerateSubmoduleInventory() (string, error) {
	out, err := exec.Command("git", "submodule", "status", "--recursive").Output()
	if err != nil {
		return "", err
	}

	lines := strings.Split(strings.TrimSpace(string(out)), "\n")
	inventory := "| Path | Commit | Status |\n|------|--------|--------|\n"
	for _, line := range lines {
		parts := strings.Fields(line)
		if len(parts) >= 2 {
			inventory += fmt.Sprintf("| %s | %s | Active |\n", parts[1], parts[0])
		}
	}
	return inventory, nil
}
