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
	driver, err := postgres.WithInstance(db.Conn, &postgres.Config{})
	if err != nil {
		return fmt.Errorf("could not create postgres driver for migrations: %w", err)
	}

	m, err := migrate.NewWithDatabaseInstance(
		"file://migrations",
		"postgres", driver)
	if err != nil {
		// If migrations dir does not exist (e.g. running from tests in subdirs), we might skip or warn
		if os.IsNotExist(err) {
			slog.Warn("Migrations directory not found. Skipping golang-migrate runner.")
			return nil
		}
		// Attempt to fallback relative to project root
		m, err = migrate.NewWithDatabaseInstance(
			"file://../../migrations",
			"postgres", driver)
		if err != nil {
			slog.Warn("Migrations directory not found in fallback path either. Skipping golang-migrate runner.")
			return nil
		}
	}

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("could not run up migrations: %w", err)
	}

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

I noticed {{company}} is building some really interesting stuff with {{tech_stack}}. The work your team is doing on {{specific_project}} caught my attention.

I'm reaching out because we've built HyperNexus (hypernexus.site) — the enterprise-ready cloud-hosted version of TormentNexus. HyperNexus coordinates multi-agent LLM workflows, MCP tool routing, and provider failover, backed by a stable open-source fork of TormentNexus at github.com/HyperNexusSoft/HyperNexus. Teams using similar stacks to yours have seen 3-5x improvements in agent coordination efficiency.

Would you be open to a quick 15-minute chat this week to explore if this could help your team?

Best,
[Your Name]`,
		},
		{
			ID:      "github-hook",
			Name:    "GitHub Comment Hook",
			Channel: "github",
			Subject: "",
			Body: `Hey @{{github_handle}}, I saw your work on {{repo}} — really impressive stuff! We've been tackling similar coordination challenges with HyperNexus (hypernexus.site), the enterprise-grade cloud version of TormentNexus. We maintain our stable fork at github.com/HyperNexusSoft/HyperNexus. Would love to get your thoughts if you're open to it.`,
		},
		{
			ID:      "followup-email",
			Name:    "Follow-up Email",
			Channel: "email",
			Subject: "Re: HyperNexus for {{company}} — Thoughts?",
			Body: `Hi {{contact}},

Just wanted to follow up on my previous note about HyperNexus.

I know things get busy, so I'll keep this brief: HyperNexus (hypernexus.site) provides progressive MCP tool routing, dual-tier memory (14K+ persisted memories), and a resilient LLM waterfall that cascades across providers with zero downtime. It is built as a stable fork of TormentNexus at github.com/HyperNexusSoft/HyperNexus.

If you're even remotely curious about improving your agent coordination, I'd love to share a quick demo.

Worth a conversation?

Best,
[Your Name]`,
		},
		{
			ID:      "linkedin-connect",
			Name:    "LinkedIn Connection Request",
			Channel: "linkedin",
			Subject: "",
			Body: `Hi {{contact}}, I came across your profile while researching teams working on {{tech_stack}} at {{company}}. Your background in {{role}} is impressive. I'd love to connect and exchange insights on AI infrastructure.`,
		},
		{
			ID:      "breakup-email",
			Name:    "Breakup Email",
			Channel: "email",
			Subject: "Should I close your file?",
			Body: `Hi {{contact}},

I've reached out a few times about HyperNexus, but haven't heard back.

I'm guessing this isn't a priority right now, or you're swamped with other initiatives. Either way, I don't want to be a pest.

If you'd like me to close your file on this, just reply "close". If I got the timing wrong and you'd still like to chat, hit me with a quick "yes" and we'll find a time.

No hard feelings either way.

Best,
[Your Name]`,
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
