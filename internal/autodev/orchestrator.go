package autodev

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/robertpelloni/enterprise_sales_bot/internal/db"
	"github.com/robertpelloni/enterprise_sales_bot/internal/deploy"
	"github.com/robertpelloni/enterprise_sales_bot/internal/gitcheck"
	"github.com/robertpelloni/enterprise_sales_bot/internal/gitres"
)

type Orchestrator struct {
	db        *db.DB
	manager   *TaskManager
	agent     Agent
	prManager gitcheck.PRManager
	tracker   deploy.CITracker
}

func NewOrchestrator(database *db.DB, manager *TaskManager, agent Agent, prManager gitcheck.PRManager, tracker deploy.CITracker) *Orchestrator {
	return &Orchestrator{db: database, manager: manager, agent: agent, prManager: prManager, tracker: tracker}
}

func (o *Orchestrator) Run(ctx context.Context, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	log.Println("Autonomous development orchestrator started...")
	for {
		select {
		case <-ctx.Done(): return
		case <-ticker.C:
			o.ExecuteStep(ctx)
			o.checkPRs(ctx)
		}
	}
}

func (o *Orchestrator) ExecuteStep(ctx context.Context) {
	if os.Getenv("SKIP_AUTODEV_SYNC") != "true" {
		_ = gitcheck.SyncRemote()
		_ = gitcheck.UpdateSubmodules()
		_ = gitres.ReconcileBranches()
	}

	clean, _ := gitcheck.IsClean()
	if !clean { return }

	tasks, err := o.manager.GetReadyTasks(ctx)
	if err != nil || len(tasks) == 0 { return }

	log.Printf("Autodev: %d ready tasks found. Executing concurrently...", len(tasks))

	var wg sync.WaitGroup
	for _, t := range tasks {
		wg.Add(1)
		go func(task Task) {
			defer wg.Done()
			o.processTask(ctx, task)
		}(t)
	}
	wg.Wait()
}

func (o *Orchestrator) processTask(ctx context.Context, task Task) {
	log.Printf("Autodev: Processing task: %s", task.Description)
	proposal, err := o.agent.ProposeSolution(ctx, task)
	if err != nil { return }

	if err := o.agent.ApplyChanges(ctx, proposal); err != nil { return }

	if err := o.agent.Verify(ctx); err != nil {
		log.Printf("Autodev: Verification failed for %s. Rolling back...", task.Description)
		_ = gitcheck.ResetHard()
		return
	}

	_ = o.manager.MarkCompleted(ctx, task.Description)
	o.finalizeCycle(ctx, &task)

	safeDescription := strings.Map(func(r rune) rune {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') { return r }
		return '-'
	}, task.Description)
	branchName := fmt.Sprintf("autodev/%s", safeDescription)

	if os.Getenv("SKIP_AUTODEV_SYNC") != "true" {
		if err := gitcheck.CheckoutAndCommit(branchName, "Autonomous Update: "+task.Description); err == nil {
			_ = gitcheck.PushBranch(branchName)
		}
	}

	pr, err := o.prManager.CreatePullRequest(ctx, branchName, "Autonomous Update: "+task.Description, proposal)
	if err == nil && o.db != nil {
		_ = o.db.CreatePullRequest(ctx, pr, task.Description)
	}
}

func (o *Orchestrator) checkPRs(ctx context.Context) {
	prs, err := o.db.ListActivePullRequests(ctx)
	if err != nil { return }
	for _, pr := range prs {
		status, _ := o.prManager.GetPRStatus(ctx, pr.ID)
		if status == gitcheck.PRStatusOpen {
			comments, _ := o.prManager.GetPRComments(ctx, pr.ID)
			if len(comments) > 0 {
				_ = o.manager.AddTask(ctx, Task{
					Description: fmt.Sprintf("Feedback PR %s: %s", pr.ID, comments[0]),
					Category:    "Refinement",
				})
			}
			ciStatus, _ := o.tracker.GetLatestStatus(ctx, pr.Branch)
			if ciStatus == deploy.CIStatusSuccess {
				if err := o.prManager.MergePullRequest(ctx, pr.ID); err == nil {
					_ = o.db.UpdatePRStatus(ctx, pr.ID, gitcheck.PRStatusMerged)
					_ = gitcheck.DeleteBranch(pr.Branch)
				}
			}
		}
	}
}

func (o *Orchestrator) finalizeCycle(ctx context.Context, task *Task) {
	version, _ := os.ReadFile("VERSION")
	vStr := strings.Split(strings.TrimSpace(string(version)), "+")[0]
	newV := fmt.Sprintf("%s+%d", vStr, time.Now().Unix())
	_ = os.WriteFile("VERSION", []byte(newV), 0644)
	_ = os.WriteFile("VERSION.md", []byte(newV), 0644)
}
