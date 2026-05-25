package autodev

import (
	"context"
	"log"
)

// MockAgent is a simulated development agent for initial integration.
type MockAgent struct{}

func (m *MockAgent) ProposeSolution(ctx context.Context, task Task) (string, error) {
	log.Printf("MockAgent: Proposing solution for: %s", task.Description)
	return "Simulated changes", nil
}

func (m *MockAgent) ApplyChanges(ctx context.Context, proposal string) error {
	log.Printf("MockAgent: Applying changes: %s", proposal)
	return nil
}

func (m *MockAgent) Verify(ctx context.Context) error {
	log.Println("MockAgent: Verifying changes...")
	return nil
}
