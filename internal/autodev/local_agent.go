package autodev

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/robertpelloni/enterprise_sales_bot/internal/llm"
)

// LocalAgent is an agent that executes tasks and verifies them using local tools.
type LocalAgent struct {
	llm llm.LLMProvider
}

func NewLocalAgent(llmProvider llm.LLMProvider) *LocalAgent {
	return &LocalAgent{llm: llmProvider}
}

func (a *LocalAgent) ProposeSolution(ctx context.Context, task Task) (string, error) {
	log.Printf("LocalAgent: Analyzing task: %s", task.Description)

	if a.llm != nil {
		prompt := llm.Prompt{
			System: "You are an autonomous Go developer. Generate a solution proposal in the format: FILE: <path>\nCONTENT:\n<code>",
			User:   fmt.Sprintf("Implement the following task: %s", task.Description),
		}
		proposal, err := a.llm.Generate(ctx, prompt)
		if err == nil {
			return proposal, nil
		}
		log.Printf("LocalAgent: LLM generation failed, falling back to mock: %v", err)
	}

	if strings.Contains(strings.ToLower(task.Description), "sales-feature") {
		return fmt.Sprintf("FILE: internal/sales/feature.go\nCONTENT:\npackage sales\n\n// Autonomous Feature: %s\nfunc ExecuteSalesFeature() {\n\tprintln(\"Executing autonomous sales logic\")\n}\n", task.Description), nil
	}

	return fmt.Sprintf("Implementation for: %s", task.Description), nil
}

func (a *LocalAgent) ApplyChanges(ctx context.Context, proposal string) error {
	log.Printf("LocalAgent: Applying changes via proposal parsing...")

	wd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get working directory: %w", err)
	}

	lines := strings.Split(proposal, "\n")
	var currentFile string
	var content strings.Builder
	writing := false

	for _, line := range lines {
		if strings.HasPrefix(line, "FILE: ") {
			// Write previous file if exists
			if currentFile != "" && content.Len() > 0 {
				if err := a.safeWriteFile(wd, currentFile, content.String()); err != nil {
					return err
				}
			}
			currentFile = strings.TrimSpace(strings.TrimPrefix(line, "FILE: "))
			content.Reset()
			writing = false
		} else if strings.HasPrefix(line, "CONTENT:") {
			writing = true
		} else if writing {
			content.WriteString(line + "\n")
		}
	}

	if currentFile != "" && content.Len() > 0 {
		return a.safeWriteFile(wd, currentFile, content.String())
	}

	return nil
}

func (a *LocalAgent) safeWriteFile(wd, relPath, content string) error {
	absPath := filepath.Join(wd, relPath)

	// Ensure the path is within the working directory (Security: Path Traversal)
	if !strings.HasPrefix(absPath, wd) {
		return fmt.Errorf("security: blocked attempt to write outside repository: %s", relPath)
	}

	// Ensure directory exists
	dir := filepath.Dir(absPath)
	// #nosec G301 -- Autonomous agent needs to create directories for its generated code
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory %s: %w", dir, err)
	}

	// #nosec G306 -- Generated source code is intended to be world-readable
	if err := os.WriteFile(absPath, []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to write %s: %w", absPath, err)
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
