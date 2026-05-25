package autodev

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/robertpelloni/enterprise_sales_bot/internal/gitcheck"
)

// Orchestrator coordinates the autonomous development lifecycle.
type Orchestrator struct {
	manager   *TaskManager
	agent     Agent
	prManager gitcheck.PRManager
	activePRs map[string]*gitcheck.PullRequest
}

// NewOrchestrator creates a new Orchestrator instance.
func NewOrchestrator(manager *TaskManager, agent Agent, prManager gitcheck.PRManager) *Orchestrator {
	return &Orchestrator{
		manager:   manager,
		agent:     agent,
		prManager: prManager,
		activePRs: make(map[string]*gitcheck.PullRequest),
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
			o.executeStep(ctx)
			o.checkPRs(ctx)
		}
	}
}

func (o *Orchestrator) checkPRs(ctx context.Context) {
	log.Println("Autodev: Checking status of active autonomous PRs...")
	for id := range o.activePRs {
		status, err := o.prManager.GetPRStatus(ctx, id)
		if err != nil {
			log.Printf("Autodev: Error checking PR %s: %v", id, err)
			continue
		}

		if status == gitcheck.PRStatusOpen {
			// In a real scenario, we would check CITracker here
			log.Printf("Autodev: PR %s is open, attempting autonomous merge...", id)
			if err := o.prManager.MergePullRequest(ctx, id); err != nil {
				log.Printf("Autodev: Merge failed for PR %s: %v", id, err)
			} else {
				log.Printf("Autodev: Successfully merged PR %s", id)
				delete(o.activePRs, id)
			}
		}
	}
}

func (o *Orchestrator) executeStep(ctx context.Context) {
	// 0. Executive Protocol: Sync & Update
	log.Println("Autodev: Executing Executive Sync Protocol...")
	if err := gitcheck.SyncRemote(); err != nil {
		log.Printf("Autodev: Remote sync failed: %v", err)
	}
	if err := gitcheck.UpdateSubmodules(); err != nil {
		log.Printf("Autodev: Submodule update failed: %v", err)
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

	// 4a. Create unique feature branch and PR
	branchName := fmt.Sprintf("autodev/%s", task.Description) // simplified branch name
	log.Printf("Autodev: Creating Pull Request for branch: %s", branchName)
	pr, err := o.prManager.CreatePullRequest(ctx, branchName, fmt.Sprintf("Autonomous Update: %s", task.Description), proposal)
	if err != nil {
		log.Printf("Autodev: Error creating PR: %v", err)
	} else {
		log.Printf("Autodev: PR created: %s", pr.URL)
		o.activePRs[pr.ID] = pr
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
}
