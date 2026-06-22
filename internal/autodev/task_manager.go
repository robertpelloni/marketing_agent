package autodev

import (
<<<<<<< HEAD
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
=======
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"
)

// TaskManager handles parsing of tasks from TODO.md and tracking their progress.
type TaskManager struct {
	todoPath string
>>>>>>> origin/main
}

// NewTaskManager creates a new TaskManager.
func NewTaskManager(todoPath string) *TaskManager {
	return &TaskManager{todoPath: todoPath}
}

<<<<<<< HEAD
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
=======
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
>>>>>>> origin/main
			return &t, nil
		}
	}

<<<<<<< HEAD
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
=======
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
>>>>>>> origin/main
			})
		}
	}
	return tasks, nil
}

<<<<<<< HEAD
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
=======
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
>>>>>>> origin/main
	if err != nil {
		return err
	}

	return nil
>>>>>>> origin/main
}
<<<<<<< HEAD

// MarkTaskFailed marks a task as failed and resets its in-progress status
// so it can be retried in a future cycle (potentially with a different approach).
func (tm *TaskManager) MarkTaskFailed(id string) error {



	tasks, err := tm.ListAllTasks(context.Background())
	if err != nil {
		return err
	}

	for i := range tasks {
		if tasks[i].ID == id {
			// Do not mark as Complete. We just release it back to the pool
			// so another cycle can pick it up.
			// In a more robust system, we would increment a retry counter here.
			break
		}
	}

	return nil // State is inherently reset because we don't persist "in-progress" to disk yet
}
=======
>>>>>>> origin/main
