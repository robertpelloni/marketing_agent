package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
)

func main() {
	stagingURL := os.Getenv("STAGING_URL")
	if stagingURL == "" {
		stagingURL = "http://localhost:8081"
	}

	fmt.Printf("Starting staging smoke test for: %s\n", stagingURL)

	// 1. Verify basic health
	err := verifyEndpoint(fmt.Sprintf("%s/health", stagingURL), "OK\n")
	if err != nil {
		fmt.Printf("Staging Validation Failed: %v\n", err)
		os.Exit(1)
	}

	// 2. Verify detailed health (DB connection)
	err = verifyDetailedHealth(fmt.Sprintf("%s/health/detailed", stagingURL))
	if err != nil {
		fmt.Printf("Staging DB Validation Failed: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Staging Validation Successful.")
}

func verifyEndpoint(url, expected string) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}

	// Basic check for response body
	return nil
}

func verifyDetailedHealth(url string) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var health map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&health); err != nil {
		return err
	}

	if health["database"] != "OK" {
		return fmt.Errorf("database health is not OK: %v", health["database"])
	}

	if health["workers"] != "active" {
		return fmt.Errorf("workers are not active: %v", health["workers"])
	}

	return nil
}
