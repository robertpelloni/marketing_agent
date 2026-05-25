package deploy

import (
	"fmt"
	"log"
	"os/exec"

	"github.com/robertpelloni/enterprise_sales_bot/internal/gitcheck"
)

// Deployer handles self-service deployment operations.
type Deployer struct{}

// NewDeployer creates a new Deployer instance.
func NewDeployer() *Deployer {
	return &Deployer{}
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
func (d *Deployer) ExecuteBuild() error {
	log.Println("Deployment: Initiating build...")
	// We call 'go build' directly for autonomy, mirroring build.bat logic.
	// We use .exe for consistency with start.bat.
	cmd := exec.Command("go", "build", "-v", "-o", "bin/sales_bot.exe", "./cmd/sales_bot")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("build failed: %v, output: %s", err, string(output))
	}
	log.Println("Deployment: Build successful.")
	return nil
}
