package autodev

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"strings"
	"time"

	"github.com/robertpelloni/enterprise_sales_bot/internal/db"
	"github.com/robertpelloni/enterprise_sales_bot/internal/deploy"
	"github.com/robertpelloni/enterprise_sales_bot/internal/gitcheck"
	"github.com/robertpelloni/enterprise_sales_bot/internal/gitres"
)

// Orchestrator coordinates the autonomous development lifecycle.
type Orchestrator struct {
	db        *db.DB
	manager   *TaskManager
	agent     Agent
	prManager gitcheck.PRManager
	tracker   deploy.CITracker
}

// NewOrchestrator creates a new Orchestrator instance.
func NewOrchestrator(database *db.DB, manager *TaskManager, agent Agent, prManager gitcheck.PRManager, tracker deploy.CITracker) *Orchestrator {
	return &Orchestrator{
		db:        database,
		manager:   manager,
		agent:     agent,
		prManager: prManager,
		tracker:   tracker,
	}
}

// Run starts the autonomous development loop.
func (o *Orchestrator) Run(ctx context.Context, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	slog.Info("Autonomous development orchestrator started", "interval", interval)

	for {
		select {
		case <-ctx.Done():
			slog.Info("Autonomous development orchestrator stopping: Draining in-flight work")
			return
		case <-ticker.C:
			o.ExecuteStep(ctx)
			o.checkPRs(ctx)
		}
	}
}

func (o *Orchestrator) checkPRs(ctx context.Context) {
	slog.Info("Autodev: Checking status of active autonomous PRs")
	prs, err := o.db.ListActivePullRequests(ctx)
	if err != nil {
		slog.Error("Autodev: Error listing active PRs", "error", err)
		return
	}

	for _, pr := range prs {
		status, comments, err := o.prManager.GetPRStatus(ctx, pr.ID)
		if err != nil {
			slog.Error("Autodev: Error checking PR status", "pr_id", pr.ID, "error", err)
			continue
		}

		if status == gitcheck.PRStatusOpen {
			if len(comments) > 0 {
				slog.Info("Autodev: PR has comments. Reviewing for feedback", "pr_id", pr.ID, "comments_count", len(comments))
				for _, c := range comments {
					slog.Info("Autodev: PR comment", "pr_id", pr.ID, "text", c)
				}
			}

			// Gate merge on CI Success and Staging Validation
			ciStatus, err := o.tracker.GetLatestStatus(ctx, pr.Branch)
			if err != nil {
				slog.Error("Autodev: Error checking CI status", "branch", pr.Branch, "error", err)
				continue
			}

			if ciStatus != deploy.CIStatusSuccess {
				slog.Info("Autodev: PR waiting for validation", "pr_id", pr.ID, "ci_status", ciStatus)
				continue
			}

			slog.Info("Autodev: PR validation successful, attempting autonomous merge", "pr_id", pr.ID)
			if err := o.prManager.MergePR(ctx, pr.ID); err != nil {
				slog.Error("Autodev: Merge failed for PR", "pr_id", pr.ID, "error", err)
			} else {
				slog.Info("Autodev: Successfully merged PR. Initiating branch cleanup", "pr_id", pr.ID)
				if err := o.db.UpdatePRStatus(ctx, pr.ID, gitcheck.PRStatusMerged); err != nil {
					slog.Error("Autodev: Error updating PR status", "pr_id", pr.ID, "status", "Merged", "error", err)
				}
				o.cleanupPRBranch(ctx, pr.Branch)
			}
		} else if status == gitcheck.PRStatusMerged || status == gitcheck.PRStatusClosed {
			slog.Info("Autodev: PR reached terminal state. Initiating branch cleanup", "pr_id", pr.ID, "status", status)
			if err := o.db.UpdatePRStatus(ctx, pr.ID, status); err != nil {
				slog.Error("Autodev: Error updating terminal PR status", "pr_id", pr.ID, "error", err)
			}
			o.cleanupPRBranch(ctx, pr.Branch)
		}
	}
}

func (o *Orchestrator) cleanupPRBranch(ctx context.Context, branch string) {
	if os.Getenv("SKIP_AUTODEV_SYNC") == "true" {
		return
	}
	if err := gitcheck.DeleteBranch(branch); err != nil {
		slog.Warn("Autodev: Local branch deletion failed", "branch", branch, "error", err)
	}
	if err := o.prManager.DeleteRemoteBranch(ctx, branch); err != nil {
		slog.Warn("Autodev: Remote branch deletion failed", "branch", branch, "error", err)
	}
}

// ExecuteStep manually triggers a single step of the orchestrator loop.
func (o *Orchestrator) ExecuteStep(ctx context.Context) {
	// 0. Executive Protocol: Sync & Update
	if os.Getenv("SKIP_AUTODEV_SYNC") != "true" {
		slog.Info("Autodev: Executing Executive Sync Protocol")
		if err := gitcheck.SyncRemote(); err != nil {
			slog.Error("Autodev: Remote sync failed", "error", err)
		}
		if err := gitcheck.UpdateSubmodules(); err != nil {
			slog.Error("Autodev: Submodule update failed", "error", err)
		}
		slog.Info("Autodev: Executing Intelligent Branch Reconciliation")
		if err := gitres.ReconcileBranches(); err != nil {
			slog.Error("Autodev: Branch reconciliation failed", "error", err)
		}
	} else {
		slog.Info("Autodev: Skipping Executive Sync Protocol (SKIP_AUTODEV_SYNC=true)")
	}

	// 1. Verify repository state
	if os.Getenv("GO_TEST_MODE") != "true" {
		clean, err := gitcheck.IsClean()
		if err != nil {
			slog.Error("Autodev: Error checking repo state", "error", err)
			return
		}
		if !clean {
			slog.Info("Autodev: Working directory not clean, skipping cycle")
			return
		}
	}

	// 2. Fetch next task
	task, err := o.manager.GetNextTask(ctx)
	if err != nil {
		slog.Error("Autodev: Error fetching next task", "error", err)
		return
	}
	if task == nil {
		return
	}

	slog.Info("Autodev: Processing task", "description", task.Description)

	// 3. Propose solution
	proposal, err := o.agent.ProposeSolution(ctx, *task)
	if err != nil {
		slog.Error("Autodev: Error proposing solution", "error", err)
		return
	}

	// 4. Apply changes
	err = o.agent.ApplyChanges(ctx, proposal)
	if err != nil {
		slog.Error("Autodev: Error applying changes", "error", err)
		return
	}

	// 5. Verify
	err = o.agent.Verify(ctx)
	if err != nil {
		slog.Error("Autodev: Verification failed for task", "description", task.Description, "error", err)
		return
	}

	// 6. Mark completed
	err = o.manager.MarkCompleted(ctx, task.Description)
	if err != nil {
		slog.Error("Autodev: Error marking task as completed", "error", err)
		return
	}
	o.finalizeCycle(ctx, task)

	// 4a. Create unique feature branch, push, and PR
	safeDescription := strings.Map(func(r rune) rune {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') {
			return r
		}
		return '-'
	}, task.Description)
	branchName := fmt.Sprintf("autodev/%s", safeDescription)

	if os.Getenv("SKIP_AUTODEV_SYNC") != "true" {
		slog.Info("Autodev: Committing changes to branch", "branch", branchName)
		if err := gitcheck.CheckoutAndCommit(branchName, fmt.Sprintf("Autonomous Update: %s", task.Description)); err != nil {
			slog.Error("Autodev: Error committing changes", "error", err)
			return
		}
		slog.Info("Autodev: Pushing feature branch", "branch", branchName)
		if err := gitcheck.PushBranch(branchName); err != nil {
			slog.Error("Autodev: Error pushing branch", "error", err)
		}
	} else {
		slog.Info("Autodev: Skipping git commit/push (SKIP_AUTODEV_SYNC=true)")
	}

	slog.Info("Autodev: Creating Pull Request for branch", "branch", branchName)
	pr, err := o.prManager.CreatePR(ctx, branchName, "Autonomous Update: "+task.Description, proposal)
	if err != nil {
		slog.Error("Autodev: Error creating PR", "error", err)
	} else {
		slog.Info("Autodev: PR created", "url", pr.URL)
		if o.db != nil {
			if err := o.db.CreatePullRequest(ctx, pr, task.Description); err != nil {
				slog.Error("Autodev: Error persisting PR record", "error", err)
			}
		}
	}

	slog.Info("Autodev: Task completed successfully", "description", task.Description)
}

func (o *Orchestrator) finalizeCycle(ctx context.Context, task *Task) {
	slog.Info("Autodev: Finalizing cycle with version governance")

	version, _ := os.ReadFile("VERSION")
	vStr := strings.TrimSpace(string(version))
	if vStr == "" {
		vStr = "0.0.0"
	}

	baseVersion := strings.Split(vStr, "+")[0]
	newV := fmt.Sprintf("%s+%d", baseVersion, time.Now().Unix())

	if err := os.WriteFile("VERSION", []byte(newV), 0644); err != nil {
		slog.Warn("Autodev: Failed to write VERSION", "error", err)
	}
	if err := os.WriteFile("VERSION.md", []byte(newV), 0644); err != nil {
		slog.Warn("Autodev: Failed to write VERSION.md", "error", err)
	}

	changelogEntry := fmt.Sprintf("\n## [%s] - %s\n- %s\n", newV, time.Now().Format("2006-01-02"), task.Description)
	f, err := os.OpenFile("CHANGELOG.md", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err == nil {
		if _, err := f.WriteString(changelogEntry); err != nil {
			slog.Error("Autodev: Error writing to CHANGELOG", "error", err)
		}
		_ = f.Close()
	}

	slog.Info("Autodev: Cycle finalized", "new_version", newV)
}
