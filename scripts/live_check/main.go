package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/robertpelloni/enterprise_sales_bot/internal/config"
	"github.com/robertpelloni/enterprise_sales_bot/internal/crm"
	"github.com/stripe/stripe-go/v81/client"
)

func main() {
	cfg := config.Load()
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	fmt.Println("Starting Production Live Check...")

	// 1. Verify CRM
	var crmClient crm.CRMClient
	switch cfg.CRMProvider {
	case "hubspot":
		crmClient = crm.NewHubSpotCRMClient(cfg.CRMAPIKey)
	case "salesforce":
		crmClient = crm.NewSalesforceCRMClient(cfg.CRMBaseURL, cfg.CRMAPIKey, cfg.SalesforceClientID, cfg.SalesforceClientSecret, cfg.SalesforceAuthURL)
	default:
		crmClient = crm.NewRestCRMClient(cfg.CRMBaseURL, cfg.CRMAPIKey)
	}

	fmt.Printf("Checking CRM Connectivity (%s)... ", cfg.CRMProvider)
	_, err := crmClient.GetLeadUpdates(ctx)
	if err != nil {
		fmt.Printf("FAILED: %v\n", err)
	} else {
		fmt.Println("OK")
	}

	// 2. Verify Stripe
	stripeKey := os.Getenv("STRIPE_SECRET_KEY")
	if stripeKey != "" {
		fmt.Print("Checking Stripe Connectivity... ")
		sc := &client.API{}
		sc.Init(stripeKey, nil)
		_, err := sc.Balance.Get(nil)
		if err != nil {
			fmt.Printf("FAILED: %v\n", err)
		} else {
			fmt.Println("OK")
		}
	} else {
		fmt.Println("Stripe: Skipped (STRIPE_SECRET_KEY not set)")
	}

	// 3. Verify GitHub
	if cfg.GitHubToken != "" {
		fmt.Print("Checking GitHub API Connectivity... ")
		req, _ := http.NewRequestWithContext(ctx, "GET", "https://api.github.com/user", nil)
		req.Header.Set("Authorization", "token "+cfg.GitHubToken)
		resp, err := http.DefaultClient.Do(req)
		if err != nil || resp.StatusCode != http.StatusOK {
			fmt.Printf("FAILED (status %d)\n", resp.StatusCode)
		} else {
			fmt.Println("OK")
			resp.Body.Close()
		}
	} else {
		fmt.Println("GitHub: Skipped (GITHUB_TOKEN not set)")
	}

	fmt.Println("Live Check Complete.")
}
