package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/robertpelloni/enterprise_sales_bot/internal/autodev"
	"github.com/robertpelloni/enterprise_sales_bot/internal/communication"
	"github.com/robertpelloni/enterprise_sales_bot/internal/db"
	"github.com/robertpelloni/enterprise_sales_bot/internal/enrichment"
	"github.com/robertpelloni/enterprise_sales_bot/internal/researcher"
	"github.com/robertpelloni/enterprise_sales_bot/internal/scraper"
	"github.com/robertpelloni/enterprise_sales_bot/internal/web"

	_ "github.com/lib/pq" // PostgreSQL driver
)

func main() {
	log.Println("Starting Enterprise Sales Bot...")

	// 1. Initialize Database
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		// Default for local development if not provided
		dbURL = "postgres://postgres:postgres@localhost:5432/sales_bot?sslmode=disable"
	}

	database, err := db.NewDB(dbURL)
	if err != nil {
		log.Fatalf("Could not connect to database: %v", err)
	}
	defer database.Close()

	// 2. Setup Scraper
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sources := []scraper.LeadSource{
		&scraper.MockJobBoardSource{},
	}
	s := scraper.NewScraper(database, sources)

	// Run scraper in background
	keywords := []string{"AI Engineer", "LLM Orchestration", "Agentic Workflows"}
	go s.Run(ctx, 1*time.Hour, keywords)

	// 2b. Setup Enricher
	enrichmentSources := []enrichment.EnrichmentSource{
		&enrichment.MockApolloSource{},
	}
	e := enrichment.NewEnricher(database, enrichmentSources)

	// Run enricher in background
	go e.Run(ctx, 1*time.Hour)

	// 2c. Setup Researcher
	crawlers := []researcher.Crawler{
		&researcher.GitHubCrawler{},
		&researcher.BlogCrawler{},
	}
	processor := &researcher.DefaultDossierProcessor{}
	r := researcher.NewResearcher(database, crawlers, processor)

	// Run researcher in background
	go r.Run(ctx, 1*time.Hour)

	// 2d. Setup Communication Manager
	classifier := &communication.MockIntentClassifier{}
	responder := &communication.MockResponseGenerator{}
	commManager := communication.NewManager(database, classifier, responder)

	// Initializing but not yet running a loop as it's event-driven or requires a poller
	_ = commManager

	// 3. Initialize Autonomous Development
	taskManager := autodev.NewTaskManager("TODO.md")
	agent := &autodev.MockAgent{}
	orchestrator := autodev.NewOrchestrator(taskManager, agent)

	// Run autodev worker in background (every 1 hour)
	go orchestrator.Run(ctx, 1*time.Hour)

	// 4. Start Web Server
	webServer := web.NewServer(database)
	go func() {
		if err := webServer.ListenAndServe(":8080"); err != nil {
			log.Printf("Web server error: %v", err)
		}
	}()

	// Wait for termination signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	log.Println("Shutting down...")
}
