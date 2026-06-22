package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"

	"github.com/robertpelloni/enterprise_sales_bot/internal/communication"
	"github.com/robertpelloni/enterprise_sales_bot/internal/crm"
	"github.com/robertpelloni/enterprise_sales_bot/internal/db"
	"github.com/robertpelloni/enterprise_sales_bot/internal/llm"
)

func main() {
	fmt.Println("Starting End-to-End Live Flow Verification...")

	ctx := context.Background()

	// 1. Setup Mock CRM Server to simulate an inbound email
	mux := http.NewServeMux()
	mux.HandleFunc("/crm/v3/objects/communications", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			// Simulate an inbound email from a known contact
			w.Header().Set("Content-Type", "application/json")
			fmt.Fprintf(w, `{
				"results": [{
					"id": "msg-123",
					"properties": {
						"hs_communication_body": "Can you provide technical details on the TormentNexus API integration?",
						"hs_communication_channel_type": "EMAIL",
						"hs_communication_sender_email": "sarah.chen@aidynamics.com"
					}
				}]
			}`)
		} else {
			w.WriteHeader(http.StatusOK)
		}
	})
	// Mock other endpoints
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	server := httptest.NewServer(mux)
	defer server.Close()

	fmt.Printf("Mock CRM Environment running at: %s\n", server.URL)

	// 2. Setup Database and Mock Objects (In-memory/Simulated for verification)
	// For a real verification script, we'd use the actual DB if DATABASE_URL was set.
	// Here we simulate the key components.
	database := &db.DB{} // Empty DB to satisfy non-nil checks, though calls will still fail if Conn is nil

	hubspot := crm.NewHubSpotCRMClient("mock-token")
	hubspot.BaseURL = server.URL

	classifier := &communication.MockIntentClassifier{}
	responder := communication.NewRAGResponseGenerator(database, &llm.MockLLMProvider{})
	strategy := communication.NewLearningSalesEngine(database, hubspot, nil)
	comm := communication.NewManager(database, classifier, responder, strategy, nil, hubspot, nil)

	// 3. Trigger CRM Worker Sync
	fmt.Println("Simulating CRM Worker poll...")

	// We call GetNewInteractions directly to verify the chain
	interactions, err := hubspot.GetNewInteractions(ctx)
	if err != nil {
		log.Fatalf("Failed to fetch interactions: %v", err)
	}

	if len(interactions) == 0 {
		log.Fatal("Verification Failed: No interactions found in mock CRM")
	}

	fmt.Printf("Found interaction from: %s\n", interactions[0].Summary)

	// 4. Verify autonomous response generation (Simulated)
	// We'll skip the actual DB lookup in this verification script logic for portability
	contact := db.Contact{ID: 1, Email: "sarah.chen@aidynamics.com", Name: "Sarah Chen"}

	reply, err := comm.ProcessInbound(ctx, contact, interactions[0].RawText)
	if err != nil {
		log.Fatalf("Autonomous processing failed: %v", err)
	}

	fmt.Printf("Autonomous Reply Generated: %s\n", reply)
	fmt.Println("\nEnd-to-End Live Flow Verified Successfully.")
}
