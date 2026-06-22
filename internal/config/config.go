package config

import (
<<<<<<< HEAD
	"os"
=======
	"bufio"
	"encoding/json"
<<<<<<< HEAD
=======
	"fmt"
>>>>>>> origin/main
	"log/slog"
	"os"
	"path/filepath"
	"strconv"
	"strings"
>>>>>>> origin/main
	"time"
<<<<<<< HEAD
	"fmt"
=======
>>>>>>> origin/main
)

// Config holds the application configuration.
type Config struct {
<<<<<<< HEAD
	DatabaseURL		string
	GitHubToken		string
	GitHubRepository	string
	GitHubWebhookSecret	string
	CRMBaseURL		string
	CRMAPIKey		string
	HermesAPIURL		string
	HermesAPIKey		string
	HermesModel		string
	DeploySyncInterval	time.Duration
	Port			string
	Environment		string

	// Safety
	DryRun	bool

	// Lead Discovery
	HunterAPIKey	string
	ApolloAPIKey	string

	// Email - SMTP
	SMTPHost	string
	SMTPPort	int
	SMTPUsername	string
	SMTPPassword	string
	SMTPFrom	string
	SMTPFromName	string

	// Email - IMAP
	IMAPHost		string
	IMAPPort		int
	IMAPUsername		string
	IMAPPassword		string
	IMAPFolder		string
		IMAPPollInterval	time.Duration

	// Webhooks
	OutboundWebhookURL    string
	OutboundWebhookSecret string
=======
	DatabaseURL         string
	GitHubToken         string
	GitHubRepository    string
	GitHubWebhookSecret string
	CRMBaseURL          string
	CRMAPIKey           string
<<<<<<< HEAD
	CRMProvider         string
	SalesforceAuthURL    string
	SalesforceClientID   string
	SalesforceClientSecret string
	SMTPHost            string
	SMTPPort            string
	SMTPUser            string
	SMTPPass            string
	SMTPFrom            string
	IMAPHost            string
	IMAPPort            string
	IMAPUser            string
	IMAPPass            string
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
=======
	HermesAPIURL        string
	HermesAPIKey        string
	HermesModel         string
	DeploySyncInterval  time.Duration
	Port                string
	Environment         string

	// Safety
	DryRun bool

	// Lead Discovery
	HunterAPIKey             string
	ApolloAPIKey             string
	TwitterBearerToken       string
	TwitterAPIKey            string
	TwitterAPIKeySecret      string
	TwitterAccessToken       string
	TwitterAccessTokenSecret string
	LinkedInClientID         string
	LinkedInClientSecret     string
	LinkedInAccessToken      string

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
>>>>>>> origin/main

	// CRM Field Mappings
	SalesforceStageMapping        map[string]string
	HubSpotStageMapping           map[string]string
	SalesforceReverseStageMapping map[string]string
	HubSpotReverseStageMapping    map[string]string
}

// Load loads the configuration from environment variables and .env file.
func Load() *Config {
	// Try to load .env file from current directory or executable directory
	loadDotEnv()

>>>>>>> origin/main
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

<<<<<<< HEAD
=======
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

>>>>>>> origin/main
	return &Config{
<<<<<<< HEAD
		DatabaseURL:		getEnv("DATABASE_URL", "postgres://postgres:postgres@localhost:5432/sales_bot?sslmode=disable"),
		GitHubToken:		os.Getenv("GITHUB_TOKEN"),
		GitHubRepository:	os.Getenv("GITHUB_REPOSITORY"),
		GitHubWebhookSecret:	os.Getenv("GITHUB_WEBHOOK_SECRET"),
		CRMBaseURL:		os.Getenv("CRM_BASE_URL"),
		CRMAPIKey:		os.Getenv("CRM_API_KEY"),
		HermesAPIURL:		os.Getenv("HERMES_API_URL"),
		HermesAPIKey:		os.Getenv("HERMES_API_KEY"),
		HermesModel:		hermesModel,
		DeploySyncInterval:	syncInterval,
		Port:			port,
		Environment:		env,

		// Safety
		DryRun:	os.Getenv("DRY_RUN") == "true",

		// Lead Discovery
		HunterAPIKey:	os.Getenv("HUNTER_API_KEY"),
		ApolloAPIKey:	os.Getenv("APOLLO_API_KEY"),

		// SMTP
		SMTPHost:	os.Getenv("SMTP_HOST"),
		SMTPPort:	smtpPort,
		SMTPUsername:	os.Getenv("SMTP_USERNAME"),
		SMTPPassword:	os.Getenv("SMTP_PASSWORD"),
		SMTPFrom:	os.Getenv("SMTP_FROM"),
		SMTPFromName:	getEnv("SMTP_FROM_NAME", "TormentNexus Sales"),

		// IMAP
		IMAPHost:		os.Getenv("IMAP_HOST"),
		IMAPPort:		imapPort,
		IMAPUsername:		os.Getenv("IMAP_USERNAME"),
		IMAPPassword:		os.Getenv("IMAP_PASSWORD"),
		IMAPFolder:		getEnv("IMAP_FOLDER", "INBOX"),
				IMAPPollInterval:	imapPollInterval,

		// Webhooks
		OutboundWebhookURL:    os.Getenv("OUTBOUND_WEBHOOK_URL"),
		OutboundWebhookSecret: os.Getenv("OUTBOUND_WEBHOOK_SECRET"),
=======
		DatabaseURL:         getEnv("DATABASE_URL", "postgres://postgres:postgres@localhost:5432/sales_bot?sslmode=disable"),
		GitHubToken:         os.Getenv("GITHUB_TOKEN"),
		GitHubRepository:    os.Getenv("GITHUB_REPOSITORY"),
		GitHubWebhookSecret: os.Getenv("GITHUB_WEBHOOK_SECRET"),
		CRMBaseURL:          os.Getenv("CRM_BASE_URL"),
		CRMAPIKey:           os.Getenv("CRM_API_KEY"),
<<<<<<< HEAD
		CRMProvider:         getEnv("CRM_PROVIDER", "generic"),
		SalesforceAuthURL:    os.Getenv("SALESFORCE_AUTH_URL"),
		SalesforceClientID:   os.Getenv("SALESFORCE_CLIENT_ID"),
		SalesforceClientSecret: os.Getenv("SALESFORCE_CLIENT_SECRET"),
		SMTPHost:            os.Getenv("SMTP_HOST"),
		SMTPPort:            os.Getenv("SMTP_PORT"),
		SMTPUser:            os.Getenv("SMTP_USER"),
		SMTPPass:            os.Getenv("SMTP_PASS"),
		SMTPFrom:            os.Getenv("SMTP_FROM"),
		IMAPHost:            os.Getenv("IMAP_HOST"),
		IMAPPort:            os.Getenv("IMAP_PORT"),
		IMAPUser:            os.Getenv("IMAP_USER"),
		IMAPPass:            os.Getenv("IMAP_PASS"),
		DeploySyncInterval:  syncInterval,
		Port:                port,
		Environment:         env,
		CRMDealNameProp:     os.Getenv("CRM_DEAL_NAME_PROP"),
		CRMDealStageProp:    os.Getenv("CRM_DEAL_STAGE_PROP"),
		CRMDealAmountProp:   os.Getenv("CRM_DEAL_AMOUNT_PROP"),
		CRMDealDossierProp:  os.Getenv("CRM_DEAL_DOSSIER_PROP"),
		CRMContactEmailProp: os.Getenv("CRM_CONTACT_EMAIL_PROP"),
=======
		HermesAPIURL:        os.Getenv("HERMES_API_URL"),
		HermesAPIKey:        os.Getenv("HERMES_API_KEY"),
		HermesModel:         hermesModel,
		DeploySyncInterval:  syncInterval,
		Port:                port,
		Environment:         env,

		// Safety
		DryRun: os.Getenv("DRY_RUN") == "true",

		// Lead Discovery
		HunterAPIKey:             os.Getenv("HUNTER_API_KEY"),
		ApolloAPIKey:             os.Getenv("APOLLO_API_KEY"),
		TwitterBearerToken:       os.Getenv("TWITTER_BEARER_TOKEN"),
		TwitterAPIKey:            os.Getenv("TWITTER_API_KEY"),
		TwitterAPIKeySecret:      os.Getenv("TWITTER_API_KEY_SECRET"),
		TwitterAccessToken:       os.Getenv("TWITTER_ACCESS_TOKEN"),
		TwitterAccessTokenSecret: os.Getenv("TWITTER_ACCESS_TOKEN_SECRET"),
		LinkedInClientID:         os.Getenv("LINKEDIN_CLIENT_ID"),
		LinkedInClientSecret:     os.Getenv("LINKEDIN_CLIENT_SECRET"),
		LinkedInAccessToken:      os.Getenv("LINKEDIN_ACCESS_TOKEN"),

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
>>>>>>> origin/main

		// CRM Field Mappings
		SalesforceStageMapping: parseMapFromEnv("SALESFORCE_STAGE_MAPPING", map[string]string{
			"Discovered":       "Prospecting",
			"Researched":       "Qualification",
			"Outreach_Sent":    "Needs Analysis",
			"Engaged":          "Value Proposition",
			"Negotiating":      "Negotiation/Review",
			"Pending_Approval": "Id. Decision Makers",
			"Closed_Won":       "Closed Won",
			"Closed_Lost":      "Closed Lost",
		}),
		HubSpotStageMapping: parseMapFromEnv("HUBSPOT_STAGE_MAPPING", map[string]string{
			"Discovered":       "appointmentscheduled",
			"Researched":       "qualifiedtobuy",
			"Outreach_Sent":    "presentationscheduled",
			"Engaged":          "decisionmakerboughtin",
			"Negotiating":      "contractsent",
			"Pending_Approval": "contractsent",
			"Closed_Won":       "closedwon",
			"Closed_Lost":      "closedlost",
		}),
		SalesforceReverseStageMapping: parseMapFromEnv("SALESFORCE_REVERSE_STAGE_MAPPING", map[string]string{
			"Prospecting":         "Discovered",
			"Qualification":       "Researched",
			"Needs Analysis":      "Outreach_Sent",
			"Value Proposition":   "Engaged",
			"Negotiation/Review":  "Negotiating",
			"Id. Decision Makers": "Pending_Approval",
			"Closed Won":          "Closed_Won",
			"Closed Lost":         "Closed_Lost",
		}),
		HubSpotReverseStageMapping: parseMapFromEnv("HUBSPOT_REVERSE_STAGE_MAPPING", map[string]string{
			"appointmentscheduled":  "Discovered",
			"qualifiedtobuy":        "Researched",
			"presentationscheduled": "Outreach_Sent",
			"decisionmakerboughtin": "Engaged",
			"contractsent":          "Negotiating",
			"closedwon":             "Closed_Won",
			"closedlost":            "Closed_Lost",
		}),
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
		file, err := os.Open(filepath.Clean(p))
		if err != nil {
<<<<<<< HEAD
			continue	// .env is optional
=======
			continue // .env is optional
>>>>>>> origin/main
		}
		defer file.Close()

		slog.Info(fmt.Sprintf("Config: Loading environment from %s", p))
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
				_ = os.Setenv(key, value)
			}
		}
<<<<<<< HEAD
		return	// only load the first .env found
=======
		return // only load the first .env found
>>>>>>> origin/main
	}
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}
<<<<<<< HEAD
=======

func parseMapFromEnv(key string, defaultMap map[string]string) map[string]string {
	val := os.Getenv(key)
	if val == "" {
		return defaultMap
	}
	var m map[string]string
	err := json.Unmarshal([]byte(val), &m)
	if err != nil {
		slog.Error("Config: Failed to parse JSON, using defaults", "key", key, "error", err)
		return defaultMap
	}
	return m
}
>>>>>>> origin/main
