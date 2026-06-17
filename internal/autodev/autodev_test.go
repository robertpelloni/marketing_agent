package autodev

import (
	"context"
	"os"
	"testing"
)

func TestTaskManager_DependencyResolution(t *testing.T) {
	content := `
# TODO
- [ ] Task 1 [id:t1]
- [ ] Task 2 [id:t2] [depends:t1]
- [x] Task 0 [id:t0]
`
	tmpfile, err := os.CreateTemp("", "TODO.md")
	if err != nil { t.Fatal(err) }
	defer os.Remove(tmpfile.Name())

	if err := os.WriteFile(tmpfile.Name(), []byte(content), 0644); err != nil { t.Fatal(err) }

	manager := NewTaskManager(tmpfile.Name())

	// First call should return Task 1 because Task 2 depends on it
	task, err := manager.GetNextTask(context.Background())
	if err != nil { t.Fatalf("Expected no error, got %v", err) }
	if task.ID != "t1" { t.Errorf("Expected 't1', got '%s'", task.ID) }

	// Mark Task 1 as completed
	err = manager.MarkCompleted(context.Background(), "Task 1")
	if err != nil { t.Fatalf("Failed to mark task 1 completed: %v", err) }

	// Next call should return Task 2
	task, err = manager.GetNextTask(context.Background())
	if err != nil { t.Fatalf("Expected no error, got %v", err) }
	if task.ID != "t2" { t.Errorf("Expected 't2', got '%s'", task.ID) }
}
