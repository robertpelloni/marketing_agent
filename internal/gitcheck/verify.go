package gitcheck

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"
)

// IsClean checks if the git working directory is clean.
func IsClean() (bool, error) {
	cmd := exec.Command("git", "status", "--porcelain")
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		return false, err
	}
	return out.Len() == 0, nil
}

// IsSynced checks if the current branch is synchronized with the target branch.
func IsSynced(target string) (bool, error) {
	// Fetch first to ensure we have latest info
	exec.Command("git", "fetch", "origin").Run()

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

// CheckConflicts checks if there are any unmerged paths (conflicts).
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
