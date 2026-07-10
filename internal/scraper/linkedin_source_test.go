package scraper

import (
	"context"
	"testing"
)

// --- LinkedInSource construction tests ---

func TestLinkedInSource_Discover_Simulation(t *testing.T) {
	t.Setenv("LINKEDIN_USERNAME", "")
	t.Setenv("LINKEDIN_PASSWORD", "")

	source := &LinkedInSource{}
	companies, err := source.Discover(context.Background(), []string{"AI", "ML"})

	if err != nil {
		t.Fatalf("unexpected error during simulation: %v", err)
	}

	if len(companies) == 0 {
		t.Errorf("expected simulated companies, got none")
	}
}

func TestLinkedInSource_HeadlessBrowserInit_NoSandbox(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping browser test in short mode")
	}
    // We just verify that with bad credentials, it falls back to simulation safely without panicking the app.
	source := &LinkedInSource{
		Username: "invalid_user_123",
		Password: "invalid_password",
	}

	// This should fail to login and fallback to simulation
	companies, err := source.Discover(context.Background(), []string{"AI", "ML"})

	if err != nil {
		t.Fatalf("unexpected error during scrape fallback: %v", err)
	}

	if len(companies) == 0 {
		t.Errorf("expected simulated companies on fallback, got none")
	}
}

func TestLinkedInSource_HealthCheck(t *testing.T) {
	t.Run("without credentials", func(t *testing.T) {
		source := &LinkedInSource{}
		err := source.HealthCheck(context.Background())
		if err == nil {
			t.Error("expected error for health check without credentials")
		}
	})

	t.Run("with credentials", func(t *testing.T) {
		source := &LinkedInSource{
			Username: "user",
			Password: "password",
		}
		err := source.HealthCheck(context.Background())
		if err != nil {
			t.Errorf("unexpected error for health check with credentials: %v", err)
		}
	})
}

func TestLinkedInSource_SetCredentials(t *testing.T) {
	source := &LinkedInSource{}
	source.SetCredentials("user", "pass")
	if source.Username != "user" || source.Password != "pass" {
		t.Errorf("credentials not set correctly")
	}
}

func TestLinkedInSource_SetTargetTitles(t *testing.T) {
	source := &LinkedInSource{}
	titles := []string{"CEO", "CTO"}
	source.SetTargetTitles(titles)
	if len(source.TargetTitles) != 2 {
		t.Errorf("titles not set correctly")
	}
}
