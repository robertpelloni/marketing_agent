package autodev

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"
)

// TaskManager handles parsing of tasks from TODO.md and tracking their progress.
type TaskManager struct {
	todoPath string
}

// NewTaskManager creates a new TaskManager.
func NewTaskManager(todoPath string) *TaskManager {
	return &TaskManager{todoPath: todoPath}
}

// GetNextTask parses TODO.md and returns the highest priority uncompleted task.
func (m *TaskManager) GetNextTask(ctx context.Context) (*Task, error) {
	file, err := os.Open(m.todoPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open TODO.md: %w", err)
	}
	defer file.Close()

	var tasks []Task
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "- [ ]") {
			desc := strings.TrimSpace(strings.TrimPrefix(line, "- [ ]"))
			t := Task{
				Description: desc,
				Completed:   false,
			}
			// Priority parsing: e.g. [HIGH]
			if strings.Contains(desc, "[HIGH]") {
				t.Category = "High"
			} else {
				t.Category = "Normal"
			}
			tasks = append(tasks, t)
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading TODO.md: %w", err)
	}

	if len(tasks) == 0 {
		return nil, nil
	}

	// Simple sort: High first
	for _, t := range tasks {
		if t.Category == "High" {
			return &t, nil
		}
	}

	return &tasks[0], nil
}

// ListAllTasks returns all tasks from TODO.md.
func (m *TaskManager) ListAllTasks(ctx context.Context) ([]Task, error) {
	file, err := os.Open(m.todoPath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var tasks []Task
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "- [ ]") || strings.HasPrefix(line, "- [x]") {
			tasks = append(tasks, Task{
				Description: strings.TrimSpace(line[5:]),
				Completed:   strings.HasPrefix(line, "- [x]"),
			})
		}
	}
	return tasks, nil
}

// MarkCompleted updates TODO.md to mark a task as completed.
func (m *TaskManager) MarkCompleted(ctx context.Context, taskDescription string) error {
	// Simple implementation: read whole file, replace line, write back
	input, err := os.ReadFile(m.todoPath)
	if err != nil {
		return err
	}

	lines := strings.Split(string(input), "\n")
	found := false
	for i, line := range lines {
		if strings.Contains(line, "- [ ]") && strings.Contains(line, taskDescription) {
			lines[i] = strings.Replace(line, "- [ ]", "- [x]", 1)
			found = true
			break
		}
	}

	if !found {
		return fmt.Errorf("task not found in TODO.md: %s", taskDescription)
	}

	output := strings.Join(lines, "\n")
	// #nosec G306 G304 G703 -- TODO file is intentionally world-readable
	err = os.WriteFile(m.todoPath, []byte(output), 0644)
	if err != nil {
		return err
	}

	return nil
}
