package autodev

import (
	"context"
	"fmt"
	"os"
	"strings"
	"sync"
)

// TaskManager handles the ingestion and tracking of autonomous development tasks.
type TaskManager struct {
	todoPath string
	mu       sync.Mutex
}

// NewTaskManager creates a new TaskManager.
func NewTaskManager(todoPath string) *TaskManager {
	return &TaskManager{todoPath: todoPath}
}

// AddTask appends a new task to the TODO list.
func (m *TaskManager) AddTask(ctx context.Context, task Task) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	entry := fmt.Sprintf("- [ ] **%s** — %s\n", task.Category, task.Description)
	// #nosec G302 -- TODO file is intentionally world-readable
	f, err := os.OpenFile(m.todoPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer f.Close()

	if _, err := f.WriteString(entry); err != nil {
		return err
	}
	return nil
}

// GetNextTask parses TODO.md and returns the highest priority uncompleted task.
func (m *TaskManager) GetNextTask(ctx context.Context) (*Task, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	file, err := os.Open(m.todoPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open TODO.md: %w", err)
	}
	defer file.Close()

	// Simple parser for Markdown checkboxes: "- [ ] task"
	data, _ := os.ReadFile(m.todoPath)
	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "- [ ]") {
			desc := strings.TrimSpace(strings.TrimPrefix(line, "- [ ]"))
			return &Task{Description: desc, Completed: false}, nil
		}
	}

	return nil, nil
}

// MarkCompleted updates the state of a task in TODO.md.
func (m *TaskManager) MarkCompleted(ctx context.Context, description string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	data, err := os.ReadFile(m.todoPath)
	if err != nil {
		return err
	}

	lines := strings.Split(string(data), "\n")
	found := false
	for i, line := range lines {
		if strings.Contains(line, "- [ ]") && strings.Contains(line, description) {
			lines[i] = strings.Replace(line, "- [ ]", "- [x]", 1)
			found = true
			break
		}
	}

	if !found {
		return fmt.Errorf("task not found: %s", description)
	}

	// #nosec G306 -- TODO file is intended to be world-readable
	return os.WriteFile(m.todoPath, []byte(strings.Join(lines, "\n")), 0644)
}

// ListAllTasks returns all tasks from TODO.md.
func (m *TaskManager) ListAllTasks(ctx context.Context) ([]Task, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	var tasks []Task
	data, err := os.ReadFile(m.todoPath)
	if err != nil {
		return nil, err
	}

	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		if strings.Contains(line, "- [ ]") || strings.Contains(line, "- [x]") {
			completed := strings.Contains(line, "- [x]")
			desc := strings.TrimSpace(strings.TrimPrefix(line, "- [ ]"))
			desc = strings.TrimSpace(strings.TrimPrefix(desc, "x]"))
			tasks = append(tasks, Task{Description: desc, Completed: completed})
		}
	}
	return tasks, nil
}
