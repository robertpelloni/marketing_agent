package db

import (
	"context"
	"testing"
	"os"
)

func TestSecretsEncryption(t *testing.T) {
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		t.Skip("DATABASE_URL not set, skipping integration test")
	}

	database, err := NewDB(dbURL)
	if err != nil {
		t.Fatalf("Failed to connect to database: %v", err)
	}
	defer database.Close()

	ctx := context.Background()

	// Set test key (32 bytes)
	database.SetSecretKey("01234567890123456789012345678901")

	// Store secret
	err = database.StoreSecret(ctx, "test_api_key", "sk_test_12345")
	if err != nil {
		t.Fatalf("Failed to store secret: %v", err)
	}

	// Retrieve secret
	retrieved, err := database.GetSecret(ctx, "test_api_key")
	if err != nil {
		t.Fatalf("Failed to retrieve secret: %v", err)
	}

	if retrieved != "sk_test_12345" {
		t.Errorf("Expected 'sk_test_12345', got '%s'", retrieved)
	}
}
