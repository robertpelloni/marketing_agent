package autodev

import (
	"context"
	"os"
	"testing"
	"strings"

	"github.com/robertpelloni/enterprise_sales_bot/internal/deploy"
	"github.com/robertpelloni/enterprise_sales_bot/internal/gitcheck"
)

func TestOrchestrator_ExecuteStep_Mock(t *testing.T) {
	// Setup mock environment
	tmpTodo, _ := os.CreateTemp("", "TODO_TEST.md")
	defer os.Remove(tmpTodo.Name())
	if err := os.WriteFile(tmpTodo.Name(), []byte("- [ ] Test integration task"), 0644); err != nil {
		t.Fatalf("Failed to write mock TODO: %v", err)
	}

	manager := NewTaskManager(tmpTodo.Name())
	agent := &MockAgent{}
	prManager := &gitcheck.MockPRManager{}
	tracker := &deploy.MockCITracker{}

	orchestrator := NewOrchestrator(nil, manager, agent, prManager, tracker)

	os.Setenv("SKIP_AUTODEV_SYNC", "true")
	os.Setenv("GO_TEST_MODE", "true")
	os.Setenv("SKIP_AUTODEV_TESTS", "true")
	defer os.Unsetenv("SKIP_AUTODEV_SYNC")
	defer os.Unsetenv("GO_TEST_MODE")
	defer os.Unsetenv("SKIP_AUTODEV_TESTS")

	ctx := context.Background()
	orchestrator.ExecuteStep(ctx)

	content, _ := os.ReadFile(tmpTodo.Name())
	if !strings.Contains(string(content), "- [x] Test integration task") {
		t.Errorf("Task was not marked completed in TODO file. Content: %s", string(content))
	}
}
