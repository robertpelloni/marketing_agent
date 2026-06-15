package autodev

import (
	"context"
	"fmt"
	"log"
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

	log.Println("Autonomous development orchestrator started...")

	for {
		select {
		case <-ctx.Done():
			log.Println("Autonomous development orchestrator stopping: Draining in-flight work...")
			return
		case <-ticker.C:
			o.ExecuteStep(ctx)
			o.checkPRs(ctx)
		}
	}
}

func (o *Orchestrator) checkPRs(ctx context.Context) {
	log.Println("Autodev: Checking status of active autonomous PRs...")
	prs, err := o.db.ListActivePullRequests(ctx)
	if err != nil {
		log.Printf("Autodev: Error listing active PRs: %v", err)
		return
	}

	for _, pr := range prs {
		status, err := o.prManager.GetPRStatus(ctx, pr.ID)
		if err != nil {
			log.Printf("Autodev: Error checking PR %s: %v", pr.ID, err)
			continue
		}

		if status == gitcheck.PRStatusOpen {
			// Check for PR comments to facilitate future self-correction
			comments, _ := o.prManager.GetPRComments(ctx, pr.ID)
			if len(comments) > 0 {
				log.Printf("Autodev: PR %s has %d comments. Reviewing for feedback...", pr.ID, len(comments))
				for _, c := range comments {
					log.Printf("Autodev: PR %s Comment: %s", pr.ID, c)
				}
			}

			// Gate merge on CI Success and Staging Validation (from unified pipeline)
			ciStatus, err := o.tracker.GetLatestStatus(ctx, pr.Branch)
			if err != nil {
				log.Printf("Autodev: Error checking CI status for branch %s: %v", pr.Branch, err)
				continue
			}

			if ciStatus != deploy.CIStatusSuccess {
				log.Printf("Autodev: PR %s CI/Staging status is %s, waiting for successful validation...", pr.ID, ciStatus)
				continue
			}

			log.Printf("Autodev: PR %s validation successful, attempting autonomous merge...", pr.ID)
			if err := o.prManager.MergePullRequest(ctx, pr.ID); err != nil {
				log.Printf("Autodev: Merge failed for PR %s: %v", pr.ID, err)
			} else {
				log.Printf("Autodev: Successfully merged PR %s. Initiating branch cleanup...", pr.ID)
				if err := o.db.UpdatePRStatus(ctx, pr.ID, gitcheck.PRStatusMerged); err != nil {
					log.Printf("Autodev: Error updating PR status to Merged for %s: %v", pr.ID, err)
				}
				o.cleanupPRBranch(pr.Branch)
			}
		} else if status == gitcheck.PRStatusClosed || status == gitcheck.PRStatusFailed {
			log.Printf("Autodev: PR %s reached terminal state: %s. Initiating branch cleanup...", pr.ID, status)
			if err := o.db.UpdatePRStatus(ctx, pr.ID, status); err != nil {
				log.Printf("Autodev: Error updating terminal PR status for %s: %v", pr.ID, err)
			}
			o.cleanupPRBranch(pr.Branch)
		}
	}
}

func (o *Orchestrator) cleanupPRBranch(branch string) {
	if err := gitcheck.DeleteBranch(branch); err != nil {
		log.Printf("Autodev: Local branch deletion failed for %s: %v", branch, err)
	}
	if err := gitcheck.DeleteRemoteBranch(branch); err != nil {
		log.Printf("Autodev: Remote branch deletion failed for %s: %v", branch, err)
	}
}

// ExecuteStep manually triggers a single step of the orchestrator loop (exported for testing).
func (o *Orchestrator) ExecuteStep(ctx context.Context) {
	// 0. Executive Protocol: Sync & Update
	if os.Getenv("SKIP_AUTODEV_SYNC") != "true" {
		log.Println("Autodev: Executing Executive Sync Protocol...")
		if err := gitcheck.SyncRemote(); err != nil {
			log.Printf("Autodev: Remote sync failed: %v", err)
		}
		if err := gitcheck.UpdateSubmodules(); err != nil {
			log.Printf("Autodev: Submodule update failed: %v", err)
		}
		log.Println("Autodev: Executing Intelligent Branch Reconciliation...")
		if err := gitres.ReconcileBranches(); err != nil {
			log.Printf("Autodev: Branch reconciliation failed: %v", err)
		}
	} else {
		log.Println("Autodev: Skipping Executive Sync Protocol (SKIP_AUTODEV_SYNC=true)")
	}

	// 1. Verify repository state
	clean, err := gitcheck.IsClean()
	if err != nil {
		log.Printf("Autodev: Error checking repo state: %v", err)
		return
	}
	if !clean {
		log.Println("Autodev: Working directory not clean, skipping cycle.")
		return
	}

	// 2. Fetch next task
	task, err := o.manager.GetNextTask(ctx)
	if err != nil {
		log.Printf("Autodev: Error fetching next task: %v", err)
		return
	}
	if task == nil {
		return // No tasks available
	}

	log.Printf("Autodev: Processing task: %s", task.Description)

	// 3. Propose solution
	proposal, err := o.agent.ProposeSolution(ctx, *task)
	if err != nil {
		log.Printf("Autodev: Error proposing solution: %v", err)
		return
	}

	// 4. Apply changes
	err = o.agent.ApplyChanges(ctx, proposal)
	if err != nil {
		log.Printf("Autodev: Error applying changes: %v", err)
		return
	}

	// 5. Verify
	err = o.agent.Verify(ctx)
	if err != nil {
		log.Printf("Autodev: Verification failed for task '%s': %v", task.Description, err)
		// Here we would ideally rollback or attempt fix
		return
	}

	// 6. Mark completed & Bump Version (Metadata update before commit)
	err = o.manager.MarkCompleted(ctx, task.Description)
	if err != nil {
		log.Printf("Autodev: Error marking task as completed: %v", err)
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
		log.Printf("Autodev: Committing changes to branch: %s", branchName)
		if err := gitcheck.CheckoutAndCommit(branchName, fmt.Sprintf("Autonomous Update: %s", task.Description)); err != nil {
			log.Printf("Autodev: Error committing changes: %v", err)
			return
		}

		log.Printf("Autodev: Pushing feature branch: %s", branchName)
		if err := gitcheck.PushBranch(branchName); err != nil {
			log.Printf("Autodev: Error pushing branch: %v", err)
			// We proceed as CreatePullRequest might still work if the branch exists
		}
	} else {
		log.Println("Autodev: Skipping git commit/push (SKIP_AUTODEV_SYNC=true)")
	}

	log.Printf("Autodev: Creating Pull Request for branch: %s", branchName)
	pr, err := o.prManager.CreatePullRequest(ctx, branchName, fmt.Sprintf("Autonomous Update: %s", task.Description), proposal)
	if err != nil {
		log.Printf("Autodev: Error creating PR: %v", err)
	} else {
		log.Printf("Autodev: PR created: %s", pr.URL)
		if o.db != nil {
			if err := o.db.CreatePullRequest(ctx, pr, task.Description); err != nil {
				log.Printf("Autodev: Error persisting PR record: %v", err)
			}
		}
	}

	log.Printf("Autodev: Task completed successfully: %s", task.Description)
}

func (o *Orchestrator) finalizeCycle(ctx context.Context, task *Task) {
	log.Println("Autodev: Finalizing cycle with version governance...")

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
	if err := os.WriteFile("VERSION", []byte(newV), 0644); err != nil { // #nosec G306 G703
		log.Printf("Autodev Warning: Failed to write VERSION: %v", err)
	}
	// #nosec G306 -- Version files are intended to be world-readable in this architecture
	if err := os.WriteFile("VERSION.md", []byte(newV), 0644); err != nil { // #nosec G306 G703
		log.Printf("Autodev Warning: Failed to write VERSION.md: %v", err)
	}

	// 2. Append to CHANGELOG.md
	changelogEntry := fmt.Sprintf("\n## [%s] - %s\n- %s\n", newV, time.Now().Format("2006-01-02"), task.Description)
	// #nosec G302 -- CHANGELOG is intentionally world-readable
	f, err := os.OpenFile("CHANGELOG.md", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err == nil {
		if _, err := f.WriteString(changelogEntry); err != nil {
			log.Printf("Autodev: Error writing to CHANGELOG: %v", err)
		}
		// #nosec G104 -- Explicit close is handled, ignore errors during autonomous loop cleanup
		_ = f.Close()
	}

	log.Printf("Autodev: Cycle finalized. New version: %s", newV)
}
