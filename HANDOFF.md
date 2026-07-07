# Handoff — 2026-07-06

## Completed This Session

### Repository Sync

- **Fetched all** from origin (8 new commits on jules-chore-replace-mocks branch)
- **Merged** `jules-chore-replace-mocks` → `main` (GraphRAG, telemetry WebSocket, test coverage, webhook IP allowlisting)
- Resolved merge conflicts in `internal/web/server.go` (kept: tooltip + channel status from both, deployment dashboard + telemetry from incoming, billing routes from stash)

### Stripe Billing System (New)

- **Database migration**: `subscriptions`, `subscription_price_history`, `billing_events` tables with grandfathering support
- **`internal/billing/billing.go`**: Rewritten with full subscription lifecycle:
  - `CreateCheckoutSession` → Stripe Checkout URL
  - `GetSubscription` → Stripe + local grandfathering data
  - `CancelSubscription` / `UpdateSubscriptionSeats`
  - `HandleWebhook` → processes 5 event types (checkout.completed, invoice.paid/failed, subscription.updated/deleted)
  - Auto-grandfathering: detects price increases, freezes old rate
- **`internal/billing/storedb.go`**: `DBAdapter` implementing `SubscriptionStore` interface
- **Config**: Added `StripeWebhookSecret`, `StripePriceCommunity/Professional/Enterprise`
- **Web server routes**: `/api/v1/webhook/stripe`, `/api/v1/billing/checkout|subscription|cancel|portal`
- **hypernexus.site**: $5 sale pricing (slashed $100 + SALE ENDING SOON), Stripe checkout integration on pricing buttons

### Site Changes

- **tormentnexus.site**: Upgrade banner wrapped in opaque card with 3D mouse tilt, shine animation, red border glow. Audio (heartbeat/scream) removed. Nose cavity moved between eyes+mouth, angular, drawn on top layer.
- **hypernexus.site**: $5 rainbow card glow boosted (4s cycle, 80px bloom). Shine animation removed from paragraph, now only on title. Stripe checkout JS added.

### Version: 0.5.3 → 0.6.0

- CHANGELOG.md updated with all changes
- VERSION bumped
- TODO.md marked billing as completed

## Pending / Next

### Unmerged Branches

- `origin/jules-crm-field-mapping-12193946835217908533`: No unique commits vs main (already merged)
- `origin/dashboard-redesign-and-social-marketing-1909221195229587742`: No unique commits

### To Deploy

- **Backend**: Needs `STRIPE_API_KEY`, `STRIPE_WEBHOOK_SECRET`, `STRIPE_PRICE_*` env vars on VPS
- Build: `go build -v -o bin/marketing_agent ./cmd/marketing_agent`
- Restart the running binary (currently on port 8080)

### Unresolved Issues

- `db.DB.ListSocialPosts` method doesn't exist — social posts dashboard section broken
- `.memory/branches/main/log.md` has CRLF warnings — line ending mismatch
