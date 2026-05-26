package autodev

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/robertpelloni/enterprise_sales_bot/internal/gitcheck"
	"github.com/robertpelloni/enterprise_sales_bot/internal/gitres"
)

import (
	"github.com/robertpelloni/enterprise_sales_bot/internal/db"
	"github.com/robertpelloni/enterprise_sales_bot/internal/deploy"
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
			log.Println("Autonomous development orchestrator stopping...")
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
			// Gate merge on CI Success
			ciStatus, err := o.tracker.GetLatestStatus(ctx, pr.Branch)
			if err != nil {
				log.Printf("Autodev: Error checking CI status for branch %s: %v", pr.Branch, err)
				continue
			}

			if ciStatus != deploy.CIStatusSuccess {
				log.Printf("Autodev: PR %s CI status is %s, waiting...", pr.ID, ciStatus)
				continue
			}

			log.Printf("Autodev: PR %s CI successful, attempting autonomous merge...", pr.ID)
			if err := o.prManager.MergePullRequest(ctx, pr.ID); err != nil {
				log.Printf("Autodev: Merge failed for PR %s: %v", pr.ID, err)
			} else {
				log.Printf("Autodev: Successfully merged PR %s", pr.ID)
				o.db.UpdatePRStatus(ctx, pr.ID, gitcheck.PRStatusMerged)
			}
		}
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
		o.db.CreatePullRequest(ctx, pr, task.Description)
	}

	// 5. Verify
	err = o.agent.Verify(ctx)
	if err != nil {
		log.Printf("Autodev: Verification failed for task '%s': %v", task.Description, err)
		// Here we would ideally rollback or attempt fix
		return
	}

	// 6. Mark completed
	err = o.manager.MarkCompleted(ctx, task.Description)
	if err != nil {
		log.Printf("Autodev: Error marking task as completed: %v", err)
		return
	}

	log.Printf("Autodev: Task completed successfully: %s", task.Description)
	o.finalizeCycle(ctx, task)
}

func (o *Orchestrator) finalizeCycle(ctx context.Context, task *Task) {
	log.Println("Autodev: Finalizing cycle with version governance...")

	// 1. Update VERSION file (bump build number)
	version, _ := os.ReadFile("VERSION")
	vStr := strings.TrimSpace(string(version))
	if vStr == "" {
		vStr = "0.0.0"
	}
	newV := fmt.Sprintf("%s+%d", vStr, time.Now().Unix())
	os.WriteFile("VERSION", []byte(newV), 0644)
	os.WriteFile("VERSION.md", []byte(newV), 0644)

	// 2. Append to CHANGELOG.md
	changelogEntry := fmt.Sprintf("\n## [%s] - %s\n- %s\n", newV, time.Now().Format("2006-01-02"), task.Description)
	f, err := os.OpenFile("CHANGELOG.md", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err == nil {
		f.WriteString(changelogEntry)
		f.Close()
	}

	log.Printf("Autodev: Cycle finalized. New version: %s", newV)
}
