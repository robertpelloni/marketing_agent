package autodev

import (
	"context"
	"log"
	"time"

	"github.com/robertpelloni/enterprise_sales_bot/internal/gitcheck"
)

// Orchestrator coordinates the autonomous development lifecycle.
type Orchestrator struct {
	manager *TaskManager
	agent   Agent
}

// NewOrchestrator creates a new Orchestrator instance.
func NewOrchestrator(manager *TaskManager, agent Agent) *Orchestrator {
	return &Orchestrator{
		manager: manager,
		agent:   agent,
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
		}
	}
}

func (o *Orchestrator) executeStep(ctx context.Context) {
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

	// 6. Mark completed
	err = o.manager.MarkCompleted(ctx, task.Description)
	if err != nil {
		log.Printf("Autodev: Error marking task as completed: %v", err)
		return
	}

	log.Printf("Autodev: Task completed successfully: %s", task.Description)
}
