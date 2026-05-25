package deploy

import (
	"context"
	"log"
)

// CIStatus represents the outcome of a CI pipeline run.
type CIStatus string

const (
	CIStatusSuccess CIStatus = "Success"
	CIStatusFailure CIStatus = "Failure"
	CIStatusPending CIStatus = "Pending"
	CIStatusUnknown CIStatus = "Unknown"
)

// CITracker defines an interface for monitoring the status of CI jobs.
type CITracker interface {
	GetLatestStatus(ctx context.Context, branch string) (CIStatus, error)
}

// MockCITracker provides a simulated CI tracking implementation.
type MockCITracker struct{}

func (m *MockCITracker) GetLatestStatus(ctx context.Context, branch string) (CIStatus, error) {
	log.Printf("MockCITracker: Checking status for branch: %s", branch)
	// In a real implementation, this would query the GitHub Actions API.
	return CIStatusSuccess, nil
}
