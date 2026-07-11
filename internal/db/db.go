package db

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"time"
	"log/slog"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/lib/pq"
)

// DB handles database interactions.
type DB struct {
	Conn *sql.DB
	SecretKey string
}
func NewDB(dataSourceName string) (*DB, error) {
	conn, err := sql.Open("postgres", dataSourceName)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Configure connection pooling for production resilience
	conn.SetMaxOpenConns(25)
	conn.SetMaxIdleConns(25)
	conn.SetConnMaxLifetime(5 * time.Minute)

	if err := conn.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	db := &DB{
		Conn:      conn,
		SecretKey: os.Getenv("SECRET_KEY"),
	}
	// Run migrations on startup
	if err := db.RunMigrations(context.Background()); err != nil {
		return nil, fmt.Errorf("database migrations failed: %w", err)
	}
	return db, nil
}

// Close closes the database connection.
func (db *DB) Close() error {
	return db.Conn.Close()
}

// RunMigrations applies all database schema migrations.
func (db *DB) RunMigrations(ctx context.Context) error {
	// Obtain advisory lock on a single connection to serialize concurrent migrations across packages
	if conn, err := db.Conn.Conn(ctx); err == nil {
		defer conn.Close()
		_, _ = conn.ExecContext(ctx, "SELECT pg_advisory_lock(837492)")
		defer func() {
			_, _ = conn.ExecContext(ctx, "SELECT pg_advisory_unlock(837492)")
		}()
	}

	driver, err := postgres.WithInstance(db.Conn, &postgres.Config{})
	if err != nil {
		return fmt.Errorf("could not create postgres driver for migrations: %w", err)
	}

	var m *migrate.Migrate
	if _, err := os.Stat("migrations"); err == nil {
		m, err = migrate.NewWithDatabaseInstance("file://migrations", "postgres", driver)
		if err != nil {
			return fmt.Errorf("failed to initialize migrations from file://migrations: %w", err)
		}
	} else if _, err := os.Stat("../../migrations"); err == nil {
		m, err = migrate.NewWithDatabaseInstance("file://../../migrations", "postgres", driver)
		if err != nil {
			return fmt.Errorf("failed to initialize migrations from file://../../migrations: %w", err)
		}
	} else {
		slog.Warn("Migrations directory not found in relative paths. Skipping golang-migrate schema runner.")
		return nil
	}

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("could not run up migrations: %w", err)
	}

	// Ensure 'Pending_Approval' is in lead_state enum (executed outside migration transaction)
	_, _ = db.Conn.ExecContext(ctx, "ALTER TYPE lead_state ADD VALUE IF NOT EXISTS 'Pending_Approval'")

	// Seed default templates after migrations run
	if err := db.seedTemplates(ctx); err != nil {
		return fmt.Errorf("seed templates: %w", err)
	}

	slog.Info("Database migrations applied successfully.")
	return nil
}

// seedTemplates inserts default outreach templates if they don't exist.
func (db *DB) seedTemplates(ctx context.Context) error {
	templates := []Template{
		{
			ID:      "intro-email",
			Name:    "Introductory Email",
			Channel: "email",
			Subject: "HyperNexus for {{company}} — Quick Question",
			Body: `Hi {{contact}},

I noticed {{company}} is building some really interesting stuff with {{tech_stack}}. The work your team is doing caught my attention.

I'm reaching out because we've built HyperNexus (hypernexus.site) — the enterprise-ready cloud-hosted version of TormentNexus. It coordinates multi-agent LLM workflows with progressive MCP tool routing (only loading the 3 most relevant tools dynamically to prevent context bloat), local-first memory (14K+ persisted memories surviving restarts with sqlite-vec semantic search), and cross-harness tool signature parity (Claude Code, Cursor, Copilot, Windsurf). We maintain our stable fork at github.com/HyperNexusSoft/HyperNexus.

Best,
HyperNexus Team`,
		},
		{
			ID:      "github-hook",
			Name:    "GitHub Comment Hook",
			Channel: "github",
			Subject: "",
			Body: `Hey @{{github_handle}}, I saw your work on {{repo}} — really impressive stuff! We've been tackling similar coordination challenges with HyperNexus (hypernexus.site), the enterprise-grade cloud version of TormentNexus. We maintain our stable fork at github.com/HyperNexusSoft/HyperNexus.`,
		},
		{
			ID:      "followup-email",
			Name:    "Follow-up Email",
			Channel: "email",
			Subject: "Re: HyperNexus for {{company}} — Thoughts?",
			Body: `Hi {{contact}},

Just wanted to follow up on my previous note about HyperNexus (hypernexus.site).

It provides progressive MCP tool routing, local-first dual-tier memory (sqlite-vec semantic search), and cross-harness tool parity to maximize developer velocity when coordinating multi-agent systems. It is built as a stable fork of TormentNexus at github.com/HyperNexusSoft/HyperNexus.

Best,
HyperNexus Team`,
		},
		{
			ID:      "linkedin-connect",
			Name:    "LinkedIn Connection Request",
			Channel: "linkedin",
			Subject: "",
			Body: `Hi {{contact}}, I came across your profile while researching teams working on {{tech_stack}} at {{company}}. Your background in {{role}} is impressive. I'd love to connect.`,
		},
		{
			ID:      "breakup-email",
			Name:    "Breakup Email",
			Channel: "email",
			Subject: "Should I close your file?",
			Body: `Hi {{contact}},

I've reached out a few times about HyperNexus (hypernexus.site).

I'm guessing this isn't a priority right now, or you're swamped with other initiatives.

If you would like to explore the platform in the future, the open-source fork is always available at github.com/HyperNexusSoft/HyperNexus.

Best,
HyperNexus Team`,
		},
	}

	for _, tmpl := range templates {
		query := `INSERT INTO templates (id, name, subject, body, channel, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, NOW(), NOW())
			ON CONFLICT (id) DO NOTHING`
		_, err := db.Conn.ExecContext(ctx, query,
			tmpl.ID, tmpl.Name, tmpl.Subject, tmpl.Body, tmpl.Channel)
		if err != nil {
			return fmt.Errorf("insert template %s: %w", tmpl.ID, err)
		}
	}

	// Create template_metrics table if it doesn't exist
	query := `CREATE TABLE IF NOT EXISTS template_metrics (
		template_id TEXT PRIMARY KEY,
		impressions INTEGER DEFAULT 0,
		successes INTEGER DEFAULT 0,
		updated_at TIMESTAMP NOT NULL DEFAULT NOW()
	);`
	_, err := db.Conn.ExecContext(ctx, query)
	if err != nil {
		return fmt.Errorf("create template_metrics table: %w", err)
	}

	return nil
}

// SetSecretKey sets the encryption key for encrypting secrets at rest.
func (db *DB) SetSecretKey(key string) {
    db.SecretKey = key
}
