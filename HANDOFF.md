# HANDOFF.md — marketing_agent (`v0.7.6`, 2026-11-14)

## 1. Current State Summary
- **Version:** `0.7.6` (`VERSION` and `VERSION.md` aligned)
- **Remotes & Production:** Local `main`, GitHub (`github/main`), GitLab (`origin/main`), and Hetzner production (`5.161.250.43:/opt/marketing_agent`) synchronized on `v0.7.6` with clean working trees and exact deployed binary parity (`go version -m /opt/marketing_agent/bin/marketing_agent` -> `vcs.revision == git rev-parse HEAD`, `vcs.modified=false`).
- **CDP Coordination Thread:** `https://muse.ai/thread/8f3e8782-3a1b-4afe-8520-2c81819e195b` (`HyperNexus main side chat`), prefixed with `[gemini in antigravity in c:\Users\hyper\workspace\marketing_agent on BOBPC]`.

## 2. What Shipped in `v0.7.5` & `v0.7.6`
1. **Gate 2 Server-Side Checkout Alignment (`v0.7.5`, commit `7bd3c18d2a124cfbb8710339b9e2b3911727f62e`):**
   - Preserved 100% of Bobbi's pricing copy commit (`e5c43b8cd10e80498999ac25f3a86bd13db30d09`) across `hypernexus_site/` and `/var/www/hypernexus.site/`.
   - Implemented `CreateHyperNexusTrialCheckout` & `BuildHyperNexusTrialSessionParams` (`internal/billing/billing.go`): single `mode="subscription"` Stripe Checkout session with `$1.99` (`199c`) one-time 14-day team trial today (`quantity: 1`, non-adjustable), `$5.00/seat/mo` (`500c`) recurring starting day 15 (`TrialPeriodDays = 14`, seats clamped `1–5`, adjustable `1..5`), and optional `$200.00` (`20000c`) one-time Founding Team Rate Lock (`12` months, persisted via `SetGrandfatheredRate` on `checkout.session.completed`).
   - Removed the legacy `$3.01` coupon mechanism completely.
   - Closed the `$0.01` forged-amount hole (`HTTP 400 Bad Request` on any client-supplied `price_amount` or `first_month_amount` + return-URL host allowlist in `internal/web/server.go`) and unified all buy buttons on `index.html`, `pricing.html`, `checkout.html`, and `cloud_login.html` to POST plan fields only.
2. **Stripe Key Expiry Observability, Error Redaction & Wire-Level Verification (`v0.7.6`):**
   - Added cached `StripeBillingClient.HealthCheck` + `KeyMode` (`internal/billing/billing.go`) and exposed live `STRIPE_API_KEY` health in `/health/detailed` (`billing` field) and Dashboard Section `6A` (`Stripe (checkout/billing)`).
   - Hardened `handleCreateCheckout` (`internal/web/server.go`) to return `HTTP 503` with a redacted diagnostic on Stripe auth/config errors (`api_key_expired` / `401` / `403`) so secret key prefixes/suffixes are never echoed to public HTTP callers.
   - Added wire-level `httptest.Server` tests (`internal/billing/billing_test.go`, `internal/web/checkout_test.go`) verifying exact `stripe-go/v81` form payloads across all 4 checkout entry flows plus expired-key redaction.
   - Fixed `golangci-lint` `SA4006` in `internal/communication/cadence_test.go:323` so GitHub Actions `lint` passes cleanly.

## 3. Dependabot Audit Notes (2026-11-14)
- `github/dependabot/go_modules/go_modules-20443bdb86` (`749d85d0`): **Skipped intentionally** — bumps `go.mod` toolchain directive from `go 1.25.0` to `go 1.26.0`, which conflicts with `.github/workflows/ci.yml` (`go-version: '1.25'`) and `golangci-lint`. Revisit only when CI runners upgrade to Go 1.26.
- `github/dependabot/npm_and_yarn/npm_and_yarn-e9c900b0fc` (`61a41a74`): **Deferred intentionally** — bundles major breaking semver jumps (`zod 3 -> 4`, `@hono/node-server 1 -> 2`, `glob 13 -> 14`, `minimatch 9 -> 11`) across 10 extension/monorepo directories without TS migration verification.

## 4. Open Blockers / Next Actions for Next Cycle
1. **Rotate Expired `STRIPE_API_KEY` on Hetzner (`/opt/marketing_agent/.env`):**
   - Live prod `STRIPE_API_KEY` (`rk_live_...knwwrr`, also in `STRIPE_RESTRICTED_KEY`) returns `HTTP 401 {"code":"api_key_expired"}` from `api.stripe.com` (now visible directly at `http://127.0.0.1:8084/health/detailed` -> `"billing": "ERROR: Stripe (live): STRIPE_API_KEY expired (api_key_expired)"`).
   - Once Robert updates `STRIPE_API_KEY` (either `sk_live_`/`sk_test_` or `rk_` with `Checkout Sessions: Write`, `Products: Write`, `Prices: Write`, `Customers: Write`, `Subscriptions: Write`) and runs `systemctl restart marketing-agent`:
     - Verify `curl -sS http://127.0.0.1:8084/health/detailed` reports `"billing": "Stripe (...): Connected"`.
     - Create unpaid Stripe test sessions across the 4 entry flows (`homepage`, `pricing`, `checkout`, `founding`) and post the `cs_...` IDs + Stripe-returned line items & `trial_end` to `https://muse.ai/thread/8f3e8782-3a1b-4afe-8520-2c81819e195b`.
2. **Rotate Other Provider Keys When Ready:** `GITHUB_TOKEN`, `DEEPSEEK_API_KEY`, `HERMES_API_KEY`, `APOLLO_API_KEY`, `HUNTER_API_KEY` in `/opt/marketing_agent/.env`.
