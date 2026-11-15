package billing

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/stripe/stripe-go/v81"
)

func TestClampHyperNexusTrialSeats(t *testing.T) {
	cases := []struct {
		in   int
		want int
	}{
		{-5, 1},
		{0, 1},
		{1, 1},
		{3, 3},
		{5, 5},
		{6, 5},
		{100, 5},
	}
	for _, tc := range cases {
		if got := ClampHyperNexusTrialSeats(tc.in); got != tc.want {
			t.Errorf("ClampHyperNexusTrialSeats(%d) = %d, want %d", tc.in, got, tc.want)
		}
	}
}

func TestBuildHyperNexusTrialSessionParams_StandardTrial(t *testing.T) {
	params, clampedSeats := BuildHyperNexusTrialSessionParams(
		42,
		3,
		false,
		"buyer@example.com",
		"https://hypernexus.site/success.html",
		"https://hypernexus.site/pricing.html",
	)
	if clampedSeats != 3 {
		t.Fatalf("clampedSeats = %d, want 3", clampedSeats)
	}
	if stripe.StringValue(params.Mode) != string(stripe.CheckoutSessionModeSubscription) {
		t.Fatalf("Mode = %q, want subscription", stripe.StringValue(params.Mode))
	}
	if len(params.Discounts) != 0 {
		t.Fatalf("expected 0 discounts/coupons, got %d", len(params.Discounts))
	}
	if params.SubscriptionData == nil || stripe.Int64Value(params.SubscriptionData.TrialPeriodDays) != 14 {
		t.Fatalf("expected TrialPeriodDays=14, got %+v", params.SubscriptionData)
	}
	if stripe.StringValue(params.CustomerEmail) != "buyer@example.com" {
		t.Fatalf("CustomerEmail = %q, want buyer@example.com", stripe.StringValue(params.CustomerEmail))
	}

	if len(params.LineItems) != 2 {
		t.Fatalf("expected 2 line items (trial + seat sub), got %d", len(params.LineItems))
	}

	// Line item 0: $1.99 one-time per-team trial fee, quantity 1, non-adjustable, non-recurring
	li0 := params.LineItems[0]
	if stripe.Int64Value(li0.Quantity) != 1 {
		t.Errorf("trial fee quantity = %d, want 1", stripe.Int64Value(li0.Quantity))
	}
	if li0.AdjustableQuantity != nil {
		t.Errorf("trial fee must not have AdjustableQuantity")
	}
	if li0.PriceData == nil || stripe.Int64Value(li0.PriceData.UnitAmount) != 199 {
		t.Errorf("trial fee unit_amount = %v, want 199", li0.PriceData)
	}
	if li0.PriceData.Recurring != nil {
		t.Errorf("trial fee must be one-time (Recurring == nil)")
	}

	// Line item 1: $5.00/seat/month recurring, quantity 3, adjustable 1..5
	li1 := params.LineItems[1]
	if stripe.Int64Value(li1.Quantity) != 3 {
		t.Errorf("seat subscription quantity = %d, want 3", stripe.Int64Value(li1.Quantity))
	}
	if li1.PriceData == nil || stripe.Int64Value(li1.PriceData.UnitAmount) != 500 {
		t.Errorf("seat subscription unit_amount = %v, want 500", li1.PriceData)
	}
	if li1.PriceData.Recurring == nil || stripe.StringValue(li1.PriceData.Recurring.Interval) != "month" {
		t.Errorf("seat subscription must be monthly recurring, got %+v", li1.PriceData.Recurring)
	}
	if li1.AdjustableQuantity == nil || !stripe.BoolValue(li1.AdjustableQuantity.Enabled) ||
		stripe.Int64Value(li1.AdjustableQuantity.Minimum) != 1 ||
		stripe.Int64Value(li1.AdjustableQuantity.Maximum) != 5 {
		t.Errorf("unexpected AdjustableQuantity: %+v", li1.AdjustableQuantity)
	}

	if params.Metadata["founding_lock"] != "false" || params.Metadata["plan"] != "trial" {
		t.Errorf("unexpected session metadata: %+v", params.Metadata)
	}
}

func TestBuildHyperNexusTrialSessionParams_FoundingOptionAndSeatClamp(t *testing.T) {
	params, clampedSeats := BuildHyperNexusTrialSessionParams(
		7,
		99, // clamped to 5
		true,
		"",
		"",
		"",
	)
	if clampedSeats != 5 {
		t.Fatalf("clampedSeats = %d, want 5", clampedSeats)
	}
	if len(params.LineItems) != 3 {
		t.Fatalf("expected 3 line items (trial + seat sub + $200 founding lock), got %d", len(params.LineItems))
	}
	if stripe.Int64Value(params.LineItems[0].Quantity) != 1 || stripe.Int64Value(params.LineItems[0].PriceData.UnitAmount) != 199 {
		t.Errorf("trial fee must remain $1.99 x 1 even at 5 seats")
	}
	if stripe.Int64Value(params.LineItems[1].Quantity) != 5 || stripe.Int64Value(params.LineItems[1].PriceData.UnitAmount) != 500 {
		t.Errorf("seat line item expected $5.00 x 5")
	}
	li2 := params.LineItems[2]
	if stripe.Int64Value(li2.Quantity) != 1 || stripe.Int64Value(li2.PriceData.UnitAmount) != 20000 {
		t.Errorf("founding fee expected $200.00 (20000 cents) x 1, got qty=%d amt=%d",
			stripe.Int64Value(li2.Quantity), stripe.Int64Value(li2.PriceData.UnitAmount))
	}
	if li2.PriceData.Recurring != nil {
		t.Errorf("founding fee must be one-time (Recurring == nil)")
	}
	if params.Metadata["founding_lock"] != "true" || params.Metadata["rate_lock_months"] != "12" {
		t.Errorf("expected founding_lock=true & rate_lock_months=12, got %+v", params.Metadata)
	}
}

func TestCreateCheckoutSessionWithPrice_RejectsHyperNexusProClientAmount(t *testing.T) {
	c := NewStripeBillingClient("sk_test_dummy", "whsec_dummy", nil, "", "", nil)
	_, err := c.CreateCheckoutSessionWithPrice(
		context.Background(),
		1,
		TierHyperNexusPro,
		"https://hypernexus.site/success.html",
		"https://hypernexus.site/pricing.html",
		1,
		1, // forged $0.01
		"month",
		0,
	)
	if err == nil {
		t.Fatal("expected CreateCheckoutSessionWithPrice to reject client-supplied price_amount for TierHyperNexusPro")
	}
}

func TestStripeBillingClient_WireLevelHyperNexusTrialCheckout_AllEntryFlows(t *testing.T) {
	var capturedForms []url.Values
	var capturedAuth string

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/v1/checkout/sessions" {
			t.Errorf("unexpected Stripe API request: %s %s", r.Method, r.URL.Path)
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		capturedAuth = r.Header.Get("Authorization")
		raw, _ := io.ReadAll(r.Body)
		vals, err := url.ParseQuery(string(raw))
		if err != nil {
			t.Errorf("failed parsing stripe form body: %v", err)
		}
		capturedForms = append(capturedForms, vals)

		w.Header().Set("Content-Type", "application/json")
		_, _ = fmt.Fprintf(w, `{"id":"cs_test_flow_%d","object":"checkout.session","url":"https://checkout.stripe.com/c/pay/cs_test_flow_%d"}`, len(capturedForms), len(capturedForms))
	}))
	defer ts.Close()

	backend := stripe.GetBackendWithConfig(stripe.APIBackend, &stripe.BackendConfig{
		URL:               stripe.String(ts.URL),
		HTTPClient:        ts.Client(),
		MaxNetworkRetries: stripe.Int64(0),
	})

	client := NewStripeBillingClient("sk_test_wire_verify", "whsec_dummy", nil, "", "", nil)
	client.APIBackend = backend

	flows := []struct {
		name      string
		seats     int
		wantSeats string
		founding  bool
		email     string
	}{
		{name: "homepage_default_trial", seats: 1, wantSeats: "1", founding: false, email: ""},
		{name: "pricing_3_seats_trial", seats: 3, wantSeats: "3", founding: false, email: ""},
		{name: "checkout_1_seat_with_email", seats: 1, wantSeats: "1", founding: false, email: "cto@acme.io"},
		{name: "founding_clamped_9_to_5_seats", seats: 9, wantSeats: "5", founding: true, email: "founder@acme.io"},
	}

	for i, fl := range flows {
		gotURL, err := client.CreateHyperNexusTrialCheckout(
			context.Background(),
			101,
			fl.seats,
			fl.founding,
			fl.email,
			"https://hypernexus.site/success.html",
			"https://hypernexus.site/pricing.html",
		)
		if err != nil {
			t.Fatalf("%s: unexpected error: %v", fl.name, err)
		}
		wantURL := fmt.Sprintf("https://checkout.stripe.com/c/pay/cs_test_flow_%d", i+1)
		if gotURL != wantURL {
			t.Fatalf("%s: got URL %q, want %q", fl.name, gotURL, wantURL)
		}

		form := capturedForms[i]
		if form.Get("mode") != "subscription" {
			t.Errorf("%s: mode = %q, want subscription", fl.name, form.Get("mode"))
		}
		if form.Get("subscription_data[trial_period_days]") != "14" {
			t.Errorf("%s: trial_period_days = %q, want 14", fl.name, form.Get("subscription_data[trial_period_days]"))
		}
		if form.Get("discounts[0][coupon]") != "" {
			t.Errorf("%s: unexpected coupon in wire payload: %q", fl.name, form.Get("discounts[0][coupon]"))
		}
		// Line 0: $1.99 one-time team trial today (qty 1, non-recurring, non-adjustable)
		if form.Get("line_items[0][price_data][unit_amount]") != "199" || form.Get("line_items[0][quantity]") != "1" {
			t.Errorf("%s: line_items[0] expected 199 x 1, got amt=%q qty=%q",
				fl.name, form.Get("line_items[0][price_data][unit_amount]"), form.Get("line_items[0][quantity]"))
		}
		if form.Get("line_items[0][price_data][recurring][interval]") != "" {
			t.Errorf("%s: line_items[0] must be one-time, got recurring=%q", fl.name, form.Get("line_items[0][price_data][recurring][interval]"))
		}
		// Line 1: $5.00/seat/mo recurring, clamped seats, adjustable 1..5
		if form.Get("line_items[1][price_data][unit_amount]") != "500" ||
			form.Get("line_items[1][price_data][recurring][interval]") != "month" ||
			form.Get("line_items[1][quantity]") != fl.wantSeats {
			t.Errorf("%s: line_items[1] mismatch: amt=%q interval=%q qty=%q (want %s)",
				fl.name,
				form.Get("line_items[1][price_data][unit_amount]"),
				form.Get("line_items[1][price_data][recurring][interval]"),
				form.Get("line_items[1][quantity]"),
				fl.wantSeats,
			)
		}
		if form.Get("line_items[1][adjustable_quantity][minimum]") != "1" || form.Get("line_items[1][adjustable_quantity][maximum]") != "5" {
			t.Errorf("%s: adjustable_quantity bounds expected 1..5", fl.name)
		}

		if fl.founding {
			if form.Get("line_items[2][price_data][unit_amount]") != "20000" || form.Get("line_items[2][quantity]") != "1" {
				t.Errorf("%s: founding line_items[2] expected 20000 x 1, got amt=%q qty=%q",
					fl.name, form.Get("line_items[2][price_data][unit_amount]"), form.Get("line_items[2][quantity]"))
			}
			if form.Get("line_items[2][price_data][recurring][interval]") != "" {
				t.Errorf("%s: founding fee must be one-time", fl.name)
			}
		} else if form.Get("line_items[2][price_data][unit_amount]") != "" {
			t.Errorf("%s: unexpected line_items[2] on non-founding flow", fl.name)
		}

		if form.Get("customer_email") != fl.email {
			t.Errorf("%s: customer_email = %q, want %q", fl.name, form.Get("customer_email"), fl.email)
		}
	}

	if !strings.Contains(capturedAuth, "sk_test_wire_verify") {
		t.Errorf("expected Bearer auth header on Stripe request, got %q", capturedAuth)
	}
}

func TestStripeBillingClient_HealthCheck_ExpiredVsWriteOnlyVsValid(t *testing.T) {
	mode := "expired"
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch mode {
		case "expired":
			w.WriteHeader(http.StatusUnauthorized)
			_, _ = w.Write([]byte(`{"error":{"code":"api_key_expired","message":"Expired API Key provided: rk_live_secret_knwwrr","type":"invalid_request_error"}}`))
		case "write_only":
			w.WriteHeader(http.StatusForbidden)
			_, _ = w.Write([]byte(`{"error":{"message":"The provided key does not have the required permissions for this endpoint.","type":"invalid_request_error"}}`))
		default:
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"object":"list","data":[],"has_more":false}`))
		}
	}))
	defer ts.Close()

	backend := stripe.GetBackendWithConfig(stripe.APIBackend, &stripe.BackendConfig{
		URL:               stripe.String(ts.URL),
		HTTPClient:        ts.Client(),
		MaxNetworkRetries: stripe.Int64(0),
	})

	// 1. Expired restricted live key -> error reported, zero secret characters leaked
	c1 := NewStripeBillingClient("rk_live_secret_knwwrr", "", nil, "", "", nil)
	c1.APIBackend = backend
	if c1.KeyMode() != "Stripe (live)" {
		t.Fatalf("KeyMode() = %q, want Stripe (live)", c1.KeyMode())
	}
	err := c1.HealthCheck(context.Background())
	if err == nil || !strings.Contains(err.Error(), "api_key_expired") {
		t.Fatalf("expected api_key_expired health error, got %v", err)
	}
	if strings.Contains(err.Error(), "rk_live_") || strings.Contains(err.Error(), "knwwrr") {
		t.Fatalf("HealthCheck leaked secret key fragment: %v", err)
	}

	// 2. Write-only restricted key (403 on list) -> authenticated & healthy
	mode = "write_only"
	c2 := NewStripeBillingClient("rk_test_write_only", "", nil, "", "", nil)
	c2.APIBackend = backend
	if err := c2.HealthCheck(context.Background()); err != nil {
		t.Fatalf("expected write-only restricted key to pass HealthCheck, got %v", err)
	}

	// 3. Full key (200 OK) -> healthy
	mode = "ok"
	c3 := NewStripeBillingClient("sk_test_ok", "", nil, "", "", nil)
	c3.APIBackend = backend
	if err := c3.HealthCheck(context.Background()); err != nil {
		t.Fatalf("expected active key to pass HealthCheck, got %v", err)
	}
}
