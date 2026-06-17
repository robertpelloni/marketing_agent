package autodev

import (
	"context"
	"fmt"
	"os"
	"regexp"
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

	entry := fmt.Sprintf("- [ ] **%s** — %s", task.Category, task.Description)
	if len(task.DependsOn) > 0 {
		entry += fmt.Sprintf(" [depends:%s]", strings.Join(task.DependsOn, ","))
	}
	entry += "\n"

	// #nosec G302 -- TODO file is intentionally world-readable
	f, err := os.OpenFile(m.todoPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil { return err }
	defer f.Close()

	if _, err := f.WriteString(entry); err != nil { return err }
	return nil
}

// GetNextTask parses TODO.md and returns the highest priority uncompleted task whose dependencies are met.
func (m *TaskManager) GetNextTask(ctx context.Context) (*Task, error) {
	tasks, err := m.ListAllTasks(ctx)
	if err != nil { return nil, err }

	completedTasks := make(map[string]bool)
	for _, t := range tasks {
		if t.Completed {
			completedTasks[t.ID] = true
		}
	}

	for _, t := range tasks {
		if t.Completed { continue }

		depsMet := true
		for _, dep := range t.DependsOn {
			if !completedTasks[dep] {
				depsMet = false
				break
			}
		}

		if depsMet {
			return &t, nil
		}
	}

	return nil, nil
}

// MarkCompleted updates the state of a task in TODO.md.
func (m *TaskManager) MarkCompleted(ctx context.Context, description string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	data, err := os.ReadFile(m.todoPath)
	if err != nil { return err }

	lines := strings.Split(string(data), "\n")
	found := false
	for i, line := range lines {
		if strings.Contains(line, "- [ ]") && strings.Contains(line, description) {
			lines[i] = strings.Replace(line, "- [ ]", "- [x]", 1)
			found = true
			break
		}
	}

	if !found { return fmt.Errorf("task not found: %s", description) }

	// #nosec G306 -- TODO file is intended to be world-readable
	return os.WriteFile(m.todoPath, []byte(strings.Join(lines, "\n")), 0644)
}

// ListAllTasks returns all tasks from TODO.md, including dependency parsing.
func (m *TaskManager) ListAllTasks(ctx context.Context) ([]Task, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	var tasks []Task
	data, err := os.ReadFile(m.todoPath)
	if err != nil { return nil, err }

	depRegex := regexp.MustCompile(`\[depends:([^\]]+)\]`)
	idRegex := regexp.MustCompile(`\[id:([^\]]+)\]`)

	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		if strings.Contains(line, "- [ ]") || strings.Contains(line, "- [x]") {
			completed := strings.Contains(line, "- [x]")

			// Extract ID if present, otherwise use description as ID
			idMatch := idRegex.FindStringSubmatch(line)
			id := ""
			if len(idMatch) > 1 {
				id = idMatch[1]
			} else {
				id = strings.TrimSpace(strings.TrimPrefix(line, "- [ ]"))
				id = strings.TrimSpace(strings.TrimPrefix(id, "- [x]"))
			}

			// Extract dependencies
			depMatch := depRegex.FindStringSubmatch(line)
			var deps []string
			if len(depMatch) > 1 {
				deps = strings.Split(depMatch[1], ",")
			}

			desc := strings.TrimSpace(strings.TrimPrefix(line, "- [ ]"))
			desc = strings.TrimSpace(strings.TrimPrefix(desc, "x]"))

			tasks = append(tasks, Task{
				ID:          id,
				Description: desc,
				Completed:   completed,
				DependsOn:   deps,
			})
		}
	}
	return tasks, nil
}

// GetReadyTasks returns all uncompleted tasks whose dependencies are met.
func (m *TaskManager) GetReadyTasks(ctx context.Context) ([]Task, error) {
	tasks, err := m.ListAllTasks(ctx)
	if err != nil { return nil, err }

	completedTasks := make(map[string]bool)
	for _, t := range tasks {
		if t.Completed { completedTasks[t.ID] = true }
	}

	var ready []Task
	for _, t := range tasks {
		if t.Completed { continue }
		depsMet := true
		for _, dep := range t.DependsOn {
			if !completedTasks[dep] { depsMet = false; break }
		}
		if depsMet { ready = append(ready, t) }
	}
	return ready, nil
}
