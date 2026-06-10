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

	slog.Info("Autonomous development orchestrator started...")

	for {
		select {
		case <-ctx.Done():
			slog.Info("Autonomous development orchestrator stopping: Draining in-flight work...")
			return
		case <-ticker.C:
			o.ExecuteStep(ctx)
			o.checkPRs(ctx)
		}
	}
}

func (o *Orchestrator) checkPRs(ctx context.Context) {
	slog.Info("Autodev: Checking status of active autonomous PRs...")
	prs, err := o.db.ListActivePullRequests(ctx)
	if err != nil {
		slog.Error("Autodev Error listing active PRs", "error", err)
		return
	}

	for _, pr := range prs {
		status, err := o.prManager.GetPRStatus(ctx, pr.ID)
		if err != nil {
			slog.Error("Autodev Error checking PR", "pr_ID", pr.ID, "error", err)
			continue
		}

		if status == gitcheck.PRStatusOpen {
			// Check for PR comments to facilitate future self-correction
			comments, _ := o.prManager.GetPRComments(ctx, pr.ID)
			if len(comments) > 0 {
				slog.Info("Autodev PR  has  comments. Reviewing for feedback...", "pr_ID", pr.ID, "len(comments", len(comments))
				for _, c := range comments {
					slog.Info("Autodev PR  Comment", "pr_ID", pr.ID, "c", c)
				}
			}

			// Gate merge on CI Success and Staging Validation (from unified pipeline)
			ciStatus, err := o.tracker.GetLatestStatus(ctx, pr.Branch)
			if err != nil {
				slog.Error("Autodev Error checking CI status for branch", "pr_Branch", pr.Branch, "error", err)
				continue
			}

			if ciStatus != deploy.CIStatusSuccess {
				slog.Info("Autodev PR  CI/Staging status is , waiting for successful validation...", "pr_ID", pr.ID, "ciStatus", ciStatus)
				continue
			}

			slog.Info("Autodev PR  validation successful, attempting autonomous merge...", "pr_ID", pr.ID)
			if err := o.prManager.MergePullRequest(ctx, pr.ID); err != nil {
				slog.Error("Autodev Merge failed for PR", "pr_ID", pr.ID, "error", err)
			} else {
				slog.Info("Autodev Successfully merged PR . Initiating branch cleanup...", "pr_ID", pr.ID)
				if err := o.db.UpdatePRStatus(ctx, pr.ID, gitcheck.PRStatusMerged); err != nil {
					slog.Error("Autodev Error updating PR status to Merged for", "pr_ID", pr.ID, "error", err)
				}
				o.cleanupPRBranch(pr.Branch)
			}
		} else if status == gitcheck.PRStatusClosed || status == gitcheck.PRStatusFailed {
			slog.Info("Autodev PR  reached terminal state . Initiating branch cleanup...", "pr_ID", pr.ID, "status", status)
			if err := o.db.UpdatePRStatus(ctx, pr.ID, status); err != nil {
				slog.Error("Autodev Error updating terminal PR status for", "pr_ID", pr.ID, "error", err)
			}
			o.cleanupPRBranch(pr.Branch)
		}
	}
}

func (o *Orchestrator) cleanupPRBranch(branch string) {
	if err := gitcheck.DeleteBranch(branch); err != nil {
		slog.Error("Autodev Local branch deletion failed for", "branch", branch, "error", err)
	}
	if err := gitcheck.DeleteRemoteBranch(branch); err != nil {
		slog.Error("Autodev Remote branch deletion failed for", "branch", branch, "error", err)
	}
}

// ExecuteStep manually triggers a single step of the orchestrator loop (exported for testing).
func (o *Orchestrator) ExecuteStep(ctx context.Context) {
	// 0. Executive Protocol: Sync & Update
	if os.Getenv("SKIP_AUTODEV_SYNC") != "true" {
		slog.Info("Autodev: Executing Executive Sync Protocol...")
		if err := gitcheck.SyncRemote(); err != nil {
			slog.Error("Autodev Remote sync failed", "error", err)
		}
		if err := gitcheck.UpdateSubmodules(); err != nil {
			slog.Error("Autodev Submodule update failed", "error", err)
		}
		slog.Info("Autodev: Executing Intelligent Branch Reconciliation...")
		if err := gitres.ReconcileBranches(); err != nil {
			slog.Error("Autodev Branch reconciliation failed", "error", err)
		}
	} else {
		slog.Info("Autodev: Skipping Executive Sync Protocol (SKIP_AUTODEV_SYNC=true)")
	}

	// 1. Verify repository state
	clean, err := gitcheck.IsClean()
	if err != nil {
		slog.Error("Autodev Error checking repo state", "error", err)
		return
	}
	if !clean {
		slog.Info("Autodev: Working directory not clean, skipping cycle.")
		return
	}

	// 2. Fetch next task
	task, err := o.manager.GetNextTask(ctx)
	if err != nil {
		slog.Error("Autodev Error fetching next task", "error", err)
		return
	}
	if task == nil {
		return // No tasks available
	}

	slog.Info("Autodev Processing task", "task_Description", task.Description)

	// 3. Propose solution
	proposal, err := o.agent.ProposeSolution(ctx, *task)
	if err != nil {
		slog.Error("Autodev Error proposing solution", "error", err)
		return
	}

	// 4. Apply changes
	err = o.agent.ApplyChanges(ctx, proposal)
	if err != nil {
		slog.Error("Autodev Error applying changes", "error", err)
		return
	}

	// 5. Verify
	err = o.agent.Verify(ctx)
	if err != nil {
		slog.Error("Autodev Verification failed for task ''", "task_Description", task.Description, "error", err)
		// Here we would ideally rollback or attempt fix
		return
	}

	// 6. Mark completed & Bump Version (Metadata update before commit)
	err = o.manager.MarkCompleted(ctx, task.Description)
	if err != nil {
		slog.Error("Autodev Error marking task as completed", "error", err)
		return
	}
	o.finalizeCycle(ctx, task)

	// 4a. Create unique feature branch, push, and PR
	// Sanitize branch name: replace spaces and special characters
	safeDescription := strings.Map(func(r rune) rune {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') {
			return r
		}
		return '-'
	}, task.Description)
	branchName := fmt.Sprintf("autodev/%s", safeDescription)

	if os.Getenv("SKIP_AUTODEV_SYNC") != "true" {
		slog.Info("Autodev Committing changes to branch", "branchName", branchName)
		if err := gitcheck.CheckoutAndCommit(branchName, fmt.Sprintf("Autonomous Update: %s", task.Description)); err != nil {
			slog.Error("Autodev Error committing changes", "error", err)
			return
		}

		slog.Info("Autodev Pushing feature branch", "branchName", branchName)
		if err := gitcheck.PushBranch(branchName); err != nil {
			slog.Error("Autodev Error pushing branch", "error", err)
			// We proceed as CreatePullRequest might still work if the branch exists
		}
	} else {
		slog.Info("Autodev: Skipping git commit/push (SKIP_AUTODEV_SYNC=true)")
	}

	slog.Info("Autodev Creating Pull Request for branch", "branchName", branchName)
	pr, err := o.prManager.CreatePullRequest(ctx, branchName, fmt.Sprintf("Autonomous Update: %s", task.Description), proposal)
	if err != nil {
		slog.Error("Autodev Error creating PR", "error", err)
	} else {
		slog.Info("Autodev PR created", "pr_URL", pr.URL)
		if o.db != nil {
			if err := o.db.CreatePullRequest(ctx, pr, task.Description); err != nil {
				slog.Error("Autodev Error persisting PR record", "error", err)
			}
		}
	}

	slog.Info("Autodev Task completed successfully", "task_Description", task.Description)
}

func (o *Orchestrator) finalizeCycle(ctx context.Context, task *Task) {
	slog.Info("Autodev: Finalizing cycle with version governance...")

	// 1. Update VERSION file (bump build number)
	version, _ := os.ReadFile("VERSION")
	vStr := strings.TrimSpace(string(version))
	if vStr == "" {
		vStr = "0.0.0"
	}

	// Extract base version (remove previous build metadata if present)
	baseVersion := strings.Split(vStr, "+")[0]
	newV := fmt.Sprintf("%s+%d", baseVersion, time.Now().Unix())

	// #nosec G306 -- Version files are intended to be world-readable in this architecture
	if err := os.WriteFile("VERSION", []byte(newV), 0644); err != nil {
		slog.Error("Autodev Warning Failed to write VERSION", "error", err)
	}
	// #nosec G306 -- Version files are intended to be world-readable in this architecture
	if err := os.WriteFile("VERSION.md", []byte(newV), 0644); err != nil {
		slog.Error("Autodev Warning Failed to write VERSION.md", "error", err)
	}

	// 2. Append to CHANGELOG.md
	changelogEntry := fmt.Sprintf("\n## [%s] - %s\n- %s\n", newV, time.Now().Format("2006-01-02"), task.Description)
	// #nosec G302 -- CHANGELOG is intentionally world-readable
	f, err := os.OpenFile("CHANGELOG.md", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err == nil {
		if _, err := f.WriteString(changelogEntry); err != nil {
			slog.Error("Autodev Error writing to CHANGELOG", "error", err)
		}
		// #nosec G104 -- Explicit close is handled, ignore errors during autonomous loop cleanup
		_ = f.Close()
	}

	slog.Info("Autodev Cycle finalized. New version", "newV", newV)
}
