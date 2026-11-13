package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"gitlab.com/robertpelloni/marketing_agent/internal/db"
)

func main() {
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		slog.Error("DATABASE_URL environment variable is required")
		os.Exit(1)
	}

	slog.Info("Running database migrations...")
	database, err := db.NewDB(dbURL)
	if err != nil {
		slog.Error(fmt.Sprintf("Failed to initialize database and run migrations: %v", err))
		os.Exit(1)
	}
	defer func() { _ = database.Close() }()

	if err := database.RunMigrations(context.Background()); err != nil {
		slog.Error(fmt.Sprintf("Migration execution failed: %v", err))
		os.Exit(1)
	}

	slog.Info("Database migrations completed successfully.")
}
