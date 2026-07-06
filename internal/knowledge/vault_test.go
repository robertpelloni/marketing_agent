package knowledge

import (
	"context"
	"testing"

	"github.com/robertpelloni/marketing_agent/internal/db"
)

func TestMemoryVault_NoDB(t *testing.T) {
	vault := NewMemoryVault(&db.DB{Conn: nil})

	_, err := vault.StoreMemory(context.Background(), db.MemoryNode{Content: "Test"})
	if err == nil {
		t.Error("Expected error on nil DB, got nil")
	}

	_, err = vault.RetrieveContext(context.Background(), "query", 5)
	if err == nil {
		t.Error("Expected error on nil DB, got nil")
	}
}
