package autodev

import (
	"context"
	"log/slog"
)

// MockAgent is a simulated development agent for initial integration.
type MockAgent struct{}

func (m *MockAgent) ProposeSolution(ctx context.Context, task Task) (string, error) {
	slog.Info("MockAgent Proposing solution for", "task_Description", task.Description)
	return "Simulated changes", nil
}

func (m *MockAgent) ApplyChanges(ctx context.Context, proposal string) error {
	slog.Info("MockAgent Applying changes", "proposal", proposal)
	return nil
}

func (m *MockAgent) Verify(ctx context.Context) error {
	slog.Info("MockAgent: Verifying changes...")
	return nil
}
