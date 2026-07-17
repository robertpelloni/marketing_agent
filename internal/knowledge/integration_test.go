package knowledge

import (
	"context"
	"os"
	"testing"

	"gitlab.com/robertpelloni/marketing_agent/internal/db"
)

func setupKnowledgeTestDB(t *testing.T) *db.DB {
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		t.Skip("DATABASE_URL not set, skipping integration test")
	}

	database, err := db.NewDB(dbURL)
	if err != nil {
		t.Fatalf("Failed to connect to database: %v", err)
	}

	ctx := context.Background()
	if err := database.RunMigrations(ctx); err != nil {
		_ = database.Close()
		t.Fatalf("Failed to run migrations: %v", err)
	}

	return database
}

func TestKnowledgeVault_Integration_StoreAndRetrieve(t *testing.T) {
	database := setupKnowledgeTestDB(t)
	defer func() { _ = database.Close() }()

	ctx := context.Background()
	vault := NewMemoryVault(database)

	// Create a memory node
	node := db.MemoryNode{
		Type:    "IntegrationTest",
		Content: "GraphRAG is highly effective for LLM routing",
		Metadata: string(`{"source": "test"}`),
	}

	id, err := vault.StoreMemory(ctx, node)
	if err != nil {
		t.Fatalf("Failed to store memory: %v", err)
	}

	if id == 0 {
		t.Errorf("Expected non-zero ID for stored memory")
	}

	// Retrieve context using partial string match
	results, err := vault.RetrieveContext(ctx, "GraphRAG", 5)
	if err != nil {
		t.Fatalf("Failed to retrieve context: %v", err)
	}

	found := false
	for _, res := range results {
		if res.ID == id && res.Type == "IntegrationTest" {
			found = true
			break
		}
	}

	if !found {
		t.Errorf("Stored memory node was not retrieved")
	}
}
