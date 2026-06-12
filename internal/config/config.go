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
	CRMProvider         string
	SalesforceAuthURL    string
	SalesforceClientID   string
	SalesforceClientSecret string
	SMTPHost            string
	SMTPPort            string
	SMTPUser            string
	SMTPPass            string
	SMTPFrom            string
	DeploySyncInterval  time.Duration
	Port                string
	Environment         string
	CRMDealNameProp     string
	CRMDealStageProp    string
	CRMDealAmountProp   string
	CRMDealDossierProp  string
	CRMContactEmailProp string
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

	return &Config{
		DatabaseURL:         getEnv("DATABASE_URL", "postgres://postgres:postgres@localhost:5432/sales_bot?sslmode=disable"),
		GitHubToken:         os.Getenv("GITHUB_TOKEN"),
		GitHubRepository:    os.Getenv("GITHUB_REPOSITORY"),
		GitHubWebhookSecret: os.Getenv("GITHUB_WEBHOOK_SECRET"),
		CRMBaseURL:          os.Getenv("CRM_BASE_URL"),
		CRMAPIKey:           os.Getenv("CRM_API_KEY"),
		CRMProvider:         getEnv("CRM_PROVIDER", "generic"),
		SalesforceAuthURL:    os.Getenv("SALESFORCE_AUTH_URL"),
		SalesforceClientID:   os.Getenv("SALESFORCE_CLIENT_ID"),
		SalesforceClientSecret: os.Getenv("SALESFORCE_CLIENT_SECRET"),
		SMTPHost:            os.Getenv("SMTP_HOST"),
		SMTPPort:            os.Getenv("SMTP_PORT"),
		SMTPUser:            os.Getenv("SMTP_USER"),
		SMTPPass:            os.Getenv("SMTP_PASS"),
		SMTPFrom:            os.Getenv("SMTP_FROM"),
		DeploySyncInterval:  syncInterval,
		Port:                port,
		Environment:         env,
		CRMDealNameProp:     os.Getenv("CRM_DEAL_NAME_PROP"),
		CRMDealStageProp:    os.Getenv("CRM_DEAL_STAGE_PROP"),
		CRMDealAmountProp:   os.Getenv("CRM_DEAL_AMOUNT_PROP"),
		CRMDealDossierProp:  os.Getenv("CRM_DEAL_DOSSIER_PROP"),
		CRMContactEmailProp: os.Getenv("CRM_CONTACT_EMAIL_PROP"),
	}
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}
