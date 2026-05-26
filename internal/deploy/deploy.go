package deploy

import (
	"context"
	"fmt"
	"log"
	"os/exec"
	"time"

	"github.com/robertpelloni/enterprise_sales_bot/internal/gitcheck"
)

// Deployer handles self-service deployment operations.
type Deployer struct {
	tracker CITracker
}

// NewDeployer creates a new Deployer instance.
func NewDeployer(tracker CITracker) *Deployer {
	return &Deployer{
		tracker: tracker,
	}
}

// ExecuteSync performs a remote sync and submodule update.
func (d *Deployer) ExecuteSync() error {
	log.Println("Deployment: Initiating repository sync...")
	if err := gitcheck.SyncRemote(); err != nil {
		return fmt.Errorf("sync failed: %w", err)
	}
	if err := gitcheck.UpdateSubmodules(); err != nil {
		return fmt.Errorf("submodule update failed: %w", err)
	}
	log.Println("Deployment: Sync successful.")
	return nil
}

// ExecuteBuild triggers the project build process.
// MonitorDeployment periodically checks the health of the deployment.
func (d *Deployer) MonitorDeployment(ctx context.Context, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	log.Printf("Deployment monitor started (interval: %v)...", interval)

	for {
		select {
		case <-ctx.Done():
			log.Println("Deployment monitor stopping...")
			return
		case <-ticker.C:
			health, err := d.tracker.GetSystemHealth(ctx)
			if err != nil {
				log.Printf("Deployment Monitor: Health check failed: %v", err)
				continue
			}
			log.Printf("Deployment Monitor: System Health: %s", health)
		}
	}
}

func (d *Deployer) ExecuteBuild() error {
	log.Println("Deployment: Initiating build...")
	// We call 'go build' directly for autonomy, mirroring build.bat logic.
	// We use bin/sales_bot for standardized cross-platform consistency.
	cmd := exec.Command("go", "build", "-v", "-o", "bin/sales_bot", "./cmd/sales_bot")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("build failed: %v, output: %s", err, string(output))
	}
	log.Println("Deployment: Build successful.")
	return nil
}

// Run starts the background synchronization worker.
func (d *Deployer) Run(ctx context.Context, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	log.Printf("Deployment worker started (interval: %v)...", interval)

	for {
		select {
		case <-ctx.Done():
			log.Println("Deployment worker stopping...")
			return
		case <-ticker.C:
			if err := d.ExecuteSync(); err != nil {
				log.Printf("Deployment background sync failed: %v", err)
			}
		}
	}
}
