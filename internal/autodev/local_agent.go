package autodev

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
)

// LocalAgent is an agent that executes tasks and verifies them using local tools.
type LocalAgent struct{}

func (a *LocalAgent) ProposeSolution(ctx context.Context, task Task) (string, error) {
	log.Printf("LocalAgent: Analyzing task: %s", task.Description)
	// In a full implementation, this might call an LLM.
	// For now, it returns a placeholder or uses predefined scripts.
	return fmt.Sprintf("Implementation for: %s", task.Description), nil
}

func (a *LocalAgent) ApplyChanges(ctx context.Context, proposal string) error {
	log.Printf("LocalAgent: Applying changes via proposal parsing...")
	// For simple autonomous updates, we expect the proposal to be in the format:
	// FILE: <filepath>
	// CONTENT: <content>
	lines := strings.Split(proposal, "\n")
	var currentFile string
	var content strings.Builder
	writing := false

	for _, line := range lines {
		if strings.HasPrefix(line, "FILE: ") {
			if currentFile != "" && content.Len() > 0 {
				if err := os.WriteFile(currentFile, []byte(content.String()), 0644); err != nil {
					return fmt.Errorf("failed to write %s: %w", currentFile, err)
				}
			}
			currentFile = strings.TrimPrefix(line, "FILE: ")
			content.Reset()
			writing = false
		} else if strings.HasPrefix(line, "CONTENT:") {
			writing = true
		} else if writing {
			content.WriteString(line + "\n")
		}
	}

	if currentFile != "" && content.Len() > 0 {
		if err := os.WriteFile(currentFile, []byte(content.String()), 0644); err != nil {
			return fmt.Errorf("failed to write %s: %w", currentFile, err)
		}
	}

	return nil
}

func (a *LocalAgent) Verify(ctx context.Context) error {
	log.Println("LocalAgent: Running full verification suite...")

	// 1. Check if it builds
	buildCmd := exec.CommandContext(ctx, "go", "build", "./...")
	if out, err := buildCmd.CombinedOutput(); err != nil {
		return fmt.Errorf("build verification failed: %v, output: %s", err, string(out))
	}
	log.Println("LocalAgent: Build verification passed.")

	// 2. Run unit tests
	// Skip tests if requested (useful for CI/Test environments to avoid recursion/timeouts)
	if os.Getenv("SKIP_AUTODEV_TESTS") == "true" {
		log.Println("LocalAgent: Skipping test verification (SKIP_AUTODEV_TESTS=true)")
		return nil
	}

	testCmd := exec.CommandContext(ctx, "go", "test", "./...")
	if out, err := testCmd.CombinedOutput(); err != nil {
		return fmt.Errorf("test verification failed: %v, output: %s", err, string(out))
	}
	log.Println("LocalAgent: Test verification passed.")

	return nil
}
