package config

import (
	"bufio"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
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

	// Safety
	DryRun bool

	// Lead Discovery
	HunterAPIKey string
	ApolloAPIKey string

	// CRM Field Mapping
	CRMDealNameProp     string
	CRMDealAmountProp   string
	CRMDealStageProp    string
	CRMDealDescProp     string
	CRMDealRouteProp    string
	CRMContactEmailProp string
	CRMContactRoleProp  string
	CRMAccountWebProp   string

	// Email - SMTP
	SMTPHost     string
	SMTPPort     int
	SMTPUsername string
	SMTPPassword string
	SMTPFrom     string
	SMTPFromName string

	// Email - IMAP
	IMAPHost         string
	IMAPPort         int
	IMAPUsername     string
	IMAPPassword     string
	IMAPFolder       string
	IMAPPollInterval time.Duration
}

// Load loads the configuration from environment variables and .env file.
func Load() *Config {
	// Try to load .env file from current directory or executable directory
	loadDotEnv()

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

	imapPollInterval := 30 * time.Minute
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

		// Safety
		DryRun: os.Getenv("DRY_RUN") == "true",

		// Lead Discovery
		HunterAPIKey: os.Getenv("HUNTER_API_KEY"),
		ApolloAPIKey: os.Getenv("APOLLO_API_KEY"),

		// CRM Field Mapping
		CRMDealNameProp:     os.Getenv("CRM_DEAL_NAME_PROP"),
		CRMDealAmountProp:   os.Getenv("CRM_DEAL_AMOUNT_PROP"),
		CRMDealStageProp:    os.Getenv("CRM_DEAL_STAGE_PROP"),
		CRMDealDescProp:     os.Getenv("CRM_DEAL_DESC_PROP"),
		CRMDealRouteProp:    os.Getenv("CRM_DEAL_ROUTE_PROP"),
		CRMContactEmailProp: os.Getenv("CRM_CONTACT_EMAIL_PROP"),
		CRMContactRoleProp:  os.Getenv("CRM_CONTACT_ROLE_PROP"),
		CRMAccountWebProp:   os.Getenv("CRM_ACCOUNT_WEB_PROP"),

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

// loadDotEnv reads a .env file and sets environment variables.
// It looks in the current working directory first, then the executable's directory.
// Existing environment variables are NOT overwritten.
func loadDotEnv() {
	paths := []string{".env"}

	// Also check next to the executable
	if exe, err := os.Executable(); err == nil {
		paths = append(paths, filepath.Join(filepath.Dir(exe), ".env"))
	}

	for _, p := range paths {
		file, err := os.Open(p)
		if err != nil {
			continue // .env is optional
		}
		defer file.Close()

		log.Printf("Config: Loading environment from %s", p)
		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())
			if line == "" || strings.HasPrefix(line, "#") {
				continue
			}

			parts := strings.SplitN(line, "=", 2)
			if len(parts) != 2 {
				continue
			}

			key := strings.TrimSpace(parts[0])
			value := strings.TrimSpace(parts[1])

			// Remove surrounding quotes
			if len(value) >= 2 && (value[0] == '"' && value[len(value)-1] == '"') ||
				(value[0] == '\'' && value[len(value)-1] == '\'') {
				value = value[1 : len(value)-1]
			}

			// Don't overwrite existing env vars
			if os.Getenv(key) == "" {
				os.Setenv(key, value)
			}
		}
		return // only load the first .env found
	}
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}
