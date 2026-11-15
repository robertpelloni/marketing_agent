package billing

import (
	"context"
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
