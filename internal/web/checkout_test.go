package web

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"gitlab.com/robertpelloni/marketing_agent/internal/billing"
	"gitlab.com/robertpelloni/marketing_agent/internal/db"
)

type spyCheckoutBillingClient struct {
	billing.MockBillingClient
	calledTrial  bool
	companyID    int64
	seats        int
	founding     bool
	email        string
	successURL   string
	cancelURL    string
	calledLegacy bool
}

func (s *spyCheckoutBillingClient) CreateHyperNexusTrialCheckout(_ context.Context, companyID int64, seats int, founding bool, customerEmail, successURL, cancelURL string) (string, error) {
	s.calledTrial = true
	s.companyID = companyID
	s.seats = seats
	s.founding = founding
	s.email = customerEmail
	s.successURL = successURL
	s.cancelURL = cancelURL
	return "https://checkout.stripe.com/c/pay/cs_test_gate2", nil
}

func (s *spyCheckoutBillingClient) CreateCheckoutSessionWithPrice(_ context.Context, _ int64, _ billing.Tier, _, _ string, _ int, _ int64, _ string, _ int64) (string, error) {
	s.calledLegacy = true
	return "https://checkout.stripe.com/c/pay/cs_should_never_happen", nil
}

func (s *spyCheckoutBillingClient) CreateInvoice(_ context.Context, _ db.Deal, _ db.Company) (string, error) {
	return "inv_mock", nil
}

func TestHandleCreateCheckout_RejectsForgedPriceAmount(t *testing.T) {
	spy := &spyCheckoutBillingClient{}
	srv := &Server{billingClient: spy}

	payloads := []string{
		`{"tier":"HYPERNEXUS_PROFESSIONAL_LICENSE","seats":1,"price_amount":1}`,
		`{"tier":"HYPERNEXUS_PROFESSIONAL_LICENSE","seats":1,"price_amount":0}`,
		`{"tier":"HYPERNEXUS_PROFESSIONAL_LICENSE","seats":1,"first_month_amount":1}`,
	}

	for _, body := range payloads {
		req := httptest.NewRequest(http.MethodPost, "/api/v1/billing/checkout", bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")
		rr := httptest.NewRecorder()

		srv.handleCreateCheckout(rr, req)

		if rr.Code != http.StatusBadRequest {
			t.Fatalf("for payload %s: got status %d, want 400 (body=%s)", body, rr.Code, rr.Body.String())
		}
		if spy.calledTrial || spy.calledLegacy {
			t.Fatalf("billing client should NOT be called when forged amount is supplied (%s)", body)
		}
		if !strings.Contains(rr.Body.String(), "price_amount") {
			t.Errorf("expected error mentioning price_amount, got %s", rr.Body.String())
		}
	}
}

func TestHandleCreateCheckout_StandardAndFoundingPlanFields(t *testing.T) {
	spy := &spyCheckoutBillingClient{}
	srv := &Server{billingClient: spy}

	// Case 1: Standard trial, seats omitted -> clamped to 1
	req1 := httptest.NewRequest(http.MethodPost, "/api/v1/billing/checkout", bytes.NewBufferString(`{
		"tier": "HYPERNEXUS_PROFESSIONAL_LICENSE",
		"plan": "trial",
		"success_url": "https://hypernexus.site/success.html",
		"cancel_url": "https://hypernexus.site/#pricing"
	}`))
	rr1 := httptest.NewRecorder()
	srv.handleCreateCheckout(rr1, req1)

	if rr1.Code != http.StatusOK {
		t.Fatalf("standard trial status = %d, body = %s", rr1.Code, rr1.Body.String())
	}
	if !spy.calledTrial || spy.seats != 1 || spy.founding {
		t.Fatalf("unexpected spy state for standard trial: %+v", spy)
	}
	var out map[string]string
	_ = json.Unmarshal(rr1.Body.Bytes(), &out)
	if out["url"] != "https://checkout.stripe.com/c/pay/cs_test_gate2" {
		t.Errorf("unexpected url %q", out["url"])
	}

	// Case 2: Founding option + seats > 5 -> clamped to 5, founding=true
	spy2 := &spyCheckoutBillingClient{}
	srv2 := &Server{billingClient: spy2}
	req2 := httptest.NewRequest(http.MethodPost, "/api/v1/billing/checkout", bytes.NewBufferString(`{
		"tier": "HYPERNEXUS_PROFESSIONAL_LICENSE",
		"plan": "founding",
		"seats": 12,
		"email": "cto@acme.io",
		"success_url": "https://hypernexus.site/success.html",
		"cancel_url": "https://hypernexus.site/checkout.html"
	}`))
	rr2 := httptest.NewRecorder()
	srv2.handleCreateCheckout(rr2, req2)

	if rr2.Code != http.StatusOK {
		t.Fatalf("founding trial status = %d, body = %s", rr2.Code, rr2.Body.String())
	}
	if !spy2.calledTrial || spy2.seats != 5 || !spy2.founding || spy2.email != "cto@acme.io" {
		t.Fatalf("unexpected spy state for founding trial: %+v", spy2)
	}

	// Case 3: Untrusted redirect host rejected with 400
	req3 := httptest.NewRequest(http.MethodPost, "/api/v1/billing/checkout", bytes.NewBufferString(`{
		"tier": "HYPERNEXUS_PROFESSIONAL_LICENSE",
		"success_url": "https://evil.example.com/steal"
	}`))
	rr3 := httptest.NewRecorder()
	srv2.handleCreateCheckout(rr3, req3)
	if rr3.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for untrusted success_url, got %d", rr3.Code)
	}
}
