package gitres

import (
	"fmt"
	"os/exec"
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
