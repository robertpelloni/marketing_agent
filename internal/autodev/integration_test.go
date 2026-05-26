package autodev

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/robertpelloni/enterprise_sales_bot/internal/db"
	"github.com/robertpelloni/enterprise_sales_bot/internal/deploy"
	"github.com/robertpelloni/enterprise_sales_bot/internal/gitcheck"
)

func TestAutonomousLoop_Integration(t *testing.T) {
	// Integration test for the orchestrator loop
	// Note: This requires a running PostgreSQL instance for db.DB

	connStr := os.Getenv("DATABASE_URL")
	if connStr == "" {
		t.Skip("DATABASE_URL not set, skipping integration test")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	database, err := db.NewDB(connStr)
	if err != nil {
		t.Fatalf("Failed to connect to database: %v", err)
	}

	// Create a temporary TODO.md
	content := "- [ ] Task Integration Test"
	tmpTodo, err := os.CreateTemp("", "TODO_integration.md")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpTodo.Name())
	os.WriteFile(tmpTodo.Name(), []byte(content), 0644)

	manager := NewTaskManager(tmpTodo.Name())
	agent := &MockAgent{}
	prManager := &gitcheck.MockPRManager{}
	tracker := &deploy.MockCITracker{}

	orchestrator := NewOrchestrator(database, manager, agent, prManager, tracker)

	// We skip the sync protocol to avoid git errors in CI environment
	os.Setenv("SKIP_AUTODEV_SYNC", "true")
	defer os.Unsetenv("SKIP_AUTODEV_SYNC")

	// Run one cycle manually
	orchestrator.executeStep(ctx)

	// Verify task is marked completed in the file
	newContent, _ := os.ReadFile(tmpTodo.Name())
	if !contains(string(newContent), "[x] Task Integration Test") {
		t.Errorf("Task was not marked completed in TODO.md. Content: %s", string(newContent))
	}
}

func TestMultiAgentWorkflow_Integration(t *testing.T) {
	connStr := os.Getenv("DATABASE_URL")
	if connStr == "" {
		t.Skip("DATABASE_URL not set, skipping multi-agent integration test")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	database, err := db.NewDB(connStr)
	if err != nil {
		t.Fatalf("Failed to connect to database: %v", err)
	}

	// Prepare multi-step TODO
	content := "- [ ] Step 1: Scrape Leads\n- [ ] Step 2: Enrich Contacts"
	tmpTodo, _ := os.CreateTemp("", "TODO_multi.md")
	defer os.Remove(tmpTodo.Name())
	os.WriteFile(tmpTodo.Name(), []byte(content), 0644)

	manager := NewTaskManager(tmpTodo.Name())
	agent := &MockAgent{}
	prManager := &gitcheck.MockPRManager{}
	tracker := &deploy.MockCITracker{}
	orchestrator := NewOrchestrator(database, manager, agent, prManager, tracker)

	os.Setenv("SKIP_AUTODEV_SYNC", "true")
	defer os.Unsetenv("SKIP_AUTODEV_SYNC")

	// Execute two steps
	orchestrator.executeStep(ctx)
	orchestrator.executeStep(ctx)

	// Verify both completed
	newContent, _ := os.ReadFile(tmpTodo.Name())
	if !contains(string(newContent), "[x] Step 1") || !contains(string(newContent), "[x] Step 2") {
		t.Errorf("Workflow steps not completed. Content: %s", string(newContent))
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s[:len(substr)] == substr || contains(s[1:], substr))
}
