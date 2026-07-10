package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"
)

func main() {
	fmt.Println("=== Staging Webhook & Promotion Verification Suite ===")

	// Staging endpoint details
	stagingURL := os.Getenv("STAGING_URL")
	if stagingURL == "" {
		stagingURL = "http://localhost:8087"
	}

	fmt.Printf("Targeting staging environment: %s\n", stagingURL)

	// Step 1: Simulate Checkout Session Completed event
	checkoutPayload := map[string]interface{}{
		"id": "evt_test_checkout_completed",
		"type": "checkout.session.completed",
		"data": map[string]interface{}{
			"object": map[string]interface{}{
				"id": "cs_test_session_12345",
				"metadata": map[string]string{
					"company_id": "1", // Target Company in DB seed
					"tier":       "professional",
				},
				"subscription": "sub_test_active_123",
				"customer_details": map[string]string{
					"email": "lead@test-company.com",
					"name":  "Staging Lead Owner",
				},
			},
		},
	}

	bodyBytes, err := json.Marshal(checkoutPayload)
	if err != nil {
		fmt.Printf("Failed to marshal checkout payload: %v\n", err)
		os.Exit(1)
	}

	client := &http.Client{Timeout: 10 * time.Second}
	req, err := http.NewRequestWithContext(context.Background(), http.MethodPost, stagingURL+"/api/v1/webhook/stripe", bytes.NewBuffer(bodyBytes))
	if err != nil {
		fmt.Printf("Failed to create webhook request: %v\n", err)
		os.Exit(1)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Stripe-Signature", "t=123,v1=mock_signature") // Bypass header check in dev mode

	fmt.Println("Sending mock Checkout Completed webhook to staging...")
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("Webhook call failed: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		fmt.Printf("Warning: webhook returned non-200 status: %d\n", resp.StatusCode)
	} else {
		fmt.Println("✓ Checkout webhook successfully routed to staging.")
	}

	// Step 2: Trigger Invoice Paid event to simulate successful conversion
	invoicePayload := map[string]interface{}{
		"id": "evt_test_invoice_paid",
		"type": "invoice.paid",
		"data": map[string]interface{}{
			"object": map[string]interface{}{
				"id": "in_test_invoice_123",
				"subscription": "sub_test_active_123",
				"period_start": time.Now().Unix(),
				"period_end": time.Now().Add(30 * 24 * time.Hour).Unix(),
			},
		},
	}

	invoiceBytes, err := json.Marshal(invoicePayload)
	if err != nil {
		fmt.Printf("Failed to marshal invoice payload: %v\n", err)
		os.Exit(1)
	}

	req2, err := http.NewRequestWithContext(context.Background(), http.MethodPost, stagingURL+"/api/v1/webhook/stripe", bytes.NewBuffer(invoiceBytes))
	if err != nil {
		fmt.Printf("Failed to create invoice webhook request: %v\n", err)
		os.Exit(1)
	}
	req2.Header.Set("Content-Type", "application/json")
	req2.Header.Set("Stripe-Signature", "t=123,v1=mock_signature")

	fmt.Println("Sending mock Invoice Paid webhook to staging...")
	resp2, err := client.Do(req2)
	if err != nil {
		fmt.Printf("Invoice webhook call failed: %v\n", err)
		os.Exit(1)
	}
	defer resp2.Body.Close()

	if resp2.StatusCode != http.StatusOK {
		fmt.Printf("Warning: invoice webhook returned non-200 status: %d\n", resp2.StatusCode)
	} else {
		fmt.Println("✓ Invoice Paid webhook successfully routed to staging.")
	}

	fmt.Println("\nWebhook verification requests sent successfully.")
}
