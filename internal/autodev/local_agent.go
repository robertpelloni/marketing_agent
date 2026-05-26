package autodev

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
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
	log.Printf("LocalAgent: Applying: %s", proposal)
	// This would typically involve writing to files.
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
