package config

import (
	"os"
	"strconv"
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

	// Lead Discovery
	HunterAPIKey string

	// Email - SMTP
	SMTPHost     string
	SMTPPort     int
	SMTPUsername string
	SMTPPassword string
	SMTPFrom     string
	SMTPFromName string

	// Email - IMAP
	IMAPHost     string
	IMAPPort     int
	IMAPUsername string
	IMAPPassword string
	IMAPFolder   string
	IMAPPollInterval time.Duration
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

	smtpPort := 587
	if p := os.Getenv("SMTP_PORT"); p != "" {
		if n, err := strconv.Atoi(p); err == nil {
			smtpPort = n
		}
	}

	imapPort := 993
	if p := os.Getenv("IMAP_PORT"); p != "" {
		if n, err := strconv.Atoi(p); err == nil {
			imapPort = n
		}
	}

	imapPollInterval := 3 * time.Minute
	if d := os.Getenv("IMAP_POLL_INTERVAL"); d != "" {
		if parsed, err := time.ParseDuration(d); err == nil {
			imapPollInterval = parsed
		}
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

		// Lead Discovery
		HunterAPIKey: os.Getenv("HUNTER_API_KEY"),

		// SMTP
		SMTPHost:     os.Getenv("SMTP_HOST"),
		SMTPPort:     smtpPort,
		SMTPUsername: os.Getenv("SMTP_USERNAME"),
		SMTPPassword: os.Getenv("SMTP_PASSWORD"),
		SMTPFrom:     os.Getenv("SMTP_FROM"),
		SMTPFromName: getEnv("SMTP_FROM_NAME", "TormentNexus Sales"),

		// IMAP
		IMAPHost:         os.Getenv("IMAP_HOST"),
		IMAPPort:         imapPort,
		IMAPUsername:     os.Getenv("IMAP_USERNAME"),
		IMAPPassword:     os.Getenv("IMAP_PASSWORD"),
		IMAPFolder:       getEnv("IMAP_FOLDER", "INBOX"),
		IMAPPollInterval: imapPollInterval,
	}
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}
