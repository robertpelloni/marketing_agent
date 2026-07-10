package web

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
	"time"
)

// ContainerInfo represents the state of a tenant's TormentNexus container.
type ContainerInfo struct {
	State     string `json:"state"`
	Uptime    string `json:"uptime"`
	MountPath string `json:"mount_path"`
}

// GetContainerStatus returns the status of a company's container.
func GetContainerStatus(ctx context.Context, companyID int64) (*ContainerInfo, error) {
	containerName := fmt.Sprintf("tormentnexus_company_%d", companyID)
	
	// Query state
	// #nosec G204
	cmd := exec.CommandContext(ctx, "docker", "inspect", "--format", "{{.State.Status}} {{.State.StartedAt}}", containerName)
	output, err := cmd.CombinedOutput()
	if err != nil {
		if strings.Contains(string(output), "No such object") || strings.Contains(err.Error(), "executable file not found") {
			return &ContainerInfo{
				State:     "not_created",
				Uptime:    "N/A",
				MountPath: fmt.Sprintf("/var/lib/tormentnexus/company_%d", companyID),
			}, nil
		}
		return nil, fmt.Errorf("inspect failed: %w (output: %s)", err, string(output))
	}

	parts := strings.Fields(string(output))
	if len(parts) == 0 {
		return &ContainerInfo{
			State:     "unknown",
			Uptime:    "N/A",
			MountPath: fmt.Sprintf("/var/lib/tormentnexus/company_%d", companyID),
		}, nil
	}

	state := parts[0]
	uptime := "N/A"
	if len(parts) > 1 && state == "running" {
		if t, err := time.Parse(time.RFC3339Nano, parts[1]); err == nil {
			uptime = time.Since(t).Round(time.Second).String()
		}
	}

	return &ContainerInfo{
		State:     state,
		Uptime:    uptime,
		MountPath: fmt.Sprintf("/var/lib/tormentnexus/company_%d", companyID),
	}, nil
}

// StartContainer creates or starts the company's container.
func StartContainer(ctx context.Context, companyID int64) error {
	containerName := fmt.Sprintf("tormentnexus_company_%d", companyID)
	mountDir := fmt.Sprintf("/var/lib/tormentnexus/company_%d", companyID)

	// Ensure mount directory exists on the host
	// #nosec G204
	_ = exec.CommandContext(ctx, "mkdir", "-p", mountDir).Run()

	// Check if already exists
	info, err := GetContainerStatus(ctx, companyID)
	if err != nil {
		return err
	}

	if info.State == "running" {
		return nil
	}

	if info.State == "exited" || info.State == "paused" {
		// #nosec G204
		cmd := exec.CommandContext(ctx, "docker", "start", containerName)
		if output, err := cmd.CombinedOutput(); err != nil {
			return fmt.Errorf("failed to start existing container: %w (%s)", err, string(output))
		}
		return nil
	}

	// Create and run new container
	// #nosec G204
	cmd := exec.CommandContext(ctx, "docker", "run", "-d",
		"--name", containerName,
		"-v", fmt.Sprintf("%s:/data", mountDir),
		"alpine", "sleep", "infinity")
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to run container: %w (%s)", err, string(output))
	}

	return nil
}

// StopContainer stops the container.
func StopContainer(ctx context.Context, companyID int64) error {
	containerName := fmt.Sprintf("tormentnexus_company_%d", companyID)
	// #nosec G204
	cmd := exec.CommandContext(ctx, "docker", "stop", containerName)
	if output, err := cmd.CombinedOutput(); err != nil {
		if strings.Contains(string(output), "No such container") {
			return nil
		}
		return fmt.Errorf("failed to stop container: %w (%s)", err, string(output))
	}
	return nil
}

// RestartContainer restarts the container.
func RestartContainer(ctx context.Context, companyID int64) error {
	containerName := fmt.Sprintf("tormentnexus_company_%d", companyID)
	// #nosec G204
	cmd := exec.CommandContext(ctx, "docker", "restart", containerName)
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to restart container: %w (%s)", err, string(output))
	}
	return nil
}
