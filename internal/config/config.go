package config

import (
	"os"
	"time"
)

// Config holds the application configuration.
type Config struct {
	DatabaseURL         string
	GitHubToken         string
	GitHubRepository    string
	GitHubWebhookSecret string
	CRMBaseURL          string
	CRMAPIKey           string
	HermesAPIURL        string
	HermesAPIKey        string
	HermesModel         string
	DeploySyncInterval  time.Duration
	Port                string
	Environment         string
}

// Load loads the configuration from environment variables.
func Load() *Config {
	syncInterval := 1 * time.Hour
	if intervalStr := os.Getenv("DEPLOY_SYNC_INTERVAL"); intervalStr != "" {
		if d, err := time.ParseDuration(intervalStr); err == nil {
			syncInterval = d
		}
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	env := os.Getenv("ENVIRONMENT")
	if env == "" {
		env = "development"
	}

	hermesModel := os.Getenv("HERMES_MODEL")
	if hermesModel == "" {
		hermesModel = "free-llm"
	}

	return &Config{
		DatabaseURL:         getEnv("DATABASE_URL", "postgres://postgres:postgres@localhost:5432/sales_bot?sslmode=disable"),
		GitHubToken:         os.Getenv("GITHUB_TOKEN"),
		GitHubRepository:    os.Getenv("GITHUB_REPOSITORY"),
		GitHubWebhookSecret: os.Getenv("GITHUB_WEBHOOK_SECRET"),
		CRMBaseURL:          os.Getenv("CRM_BASE_URL"),
		CRMAPIKey:           os.Getenv("CRM_API_KEY"),
		HermesAPIURL:        os.Getenv("HERMES_API_URL"),
		HermesAPIKey:        os.Getenv("HERMES_API_KEY"),
		HermesModel:         hermesModel,
		DeploySyncInterval:  syncInterval,
		Port:                port,
		Environment:         env,
	}
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}
