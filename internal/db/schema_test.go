package db

import (
	"context"
	"os"
	"testing"
)

// requiredColumns lists table.column pairs the Go code queries directly.
// Two production incidents (deals.cadence_step, interactions.template_id /
// response_id) were columns referenced in repository.go with no migration
// creating them; fresh databases broke silently until runtime queries hit
// them. When DATABASE_URL is set (CI integration jobs run migrations first),
// this test asserts the migrated schema actually contains every column the
// code depends on. Without a database it skips, like the other integration
// tests in this repo.
var requiredColumns = map[string][]string{
	"companies":    {"id", "name", "domain", "tech_stack", "hiring_signals", "market_cap_tier", "deleted_at", "created_at", "updated_at"},
	"contacts":     {"id", "company_id", "name", "role", "email", "github_handle", "linkedin_url", "preferred_channel", "deleted_at", "created_at", "updated_at"},
	"deals":        {"id", "company_id", "current_state", "quoted_pricing", "custom_requirements", "technical_dossier", "cadence_step", "deleted_at", "created_at", "updated_at"},
	"interactions": {"id", "contact_id", "channel", "direction", "raw_text", "summary", "sentiment", "success", "template_id", "response_id", "created_at"},
	"social_posts": {"id", "brand", "platform", "account_username", "post_content", "status", "created_at"},
}

func TestSchemaContainsRequiredColumns(t *testing.T) {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		t.Skip("DATABASE_URL not set, skipping schema sentinel test")
	}

	database, err := NewDB(dsn)
	if err != nil {
		t.Fatalf("failed to connect/migrate: %v", err)
	}
	defer func() { _ = database.Close() }()

	ctx := context.Background()
	for table, cols := range requiredColumns {
		rows, err := database.Conn.QueryContext(ctx,
			`SELECT column_name FROM information_schema.columns WHERE table_schema='public' AND table_name=$1`, table)
		if err != nil {
			t.Fatalf("query columns for %s: %v", table, err)
		}
		present := map[string]bool{}
		for rows.Next() {
			var name string
			if err := rows.Scan(&name); err != nil {
				t.Fatalf("scan column for %s: %v", table, err)
			}
			present[name] = true
		}
		_ = rows.Close()
		if err := rows.Err(); err != nil {
			t.Fatalf("rows for %s: %v", table, err)
		}
		if len(present) == 0 {
			t.Errorf("table %s missing entirely (migrations not applied?)", table)
			continue
		}
		for _, col := range cols {
			if !present[col] {
				t.Errorf("schema drift: %s.%s is queried by code but missing from migrated schema", table, col)
			}
		}
	}
}
