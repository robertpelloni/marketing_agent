# HANDOFF.md — marketing_agent (`v0.7.9`, 2026-11-15)

## 1. Current State Summary
- **Version:** `0.7.9` (`VERSION` and `VERSION.md` aligned)
- **Remotes & Production:** Local `main`, GitHub (`github/main`), GitLab (`origin/main`), and Hetzner production (`5.161.250.43:/opt/marketing_agent`) synchronized on `v0.7.9` with clean working trees and exact deployed binary parity (`go version -m /opt/marketing_agent/bin/marketing_agent` -> `vcs.revision == git rev-parse HEAD`, `vcs.modified=false`).
- **Live Billing Status (Gate 2 Complete):** Live `STRIPE_API_KEY` rotated in `/opt/marketing_agent/.env` (`chmod 600`, LF-normalized); `http://127.0.0.1:8084/health/detailed` reports `"billing": "Stripe (live): Connected"`. All 4 live unpaid `cs_live_...` sessions (`homepage`, `pricing 3 seats`, `checkout 1 seat + email`, `founding 9->5 clamped seats`) passed both Antigravity wire/API checks and Bobbi's independent live browser walk on Muse.
- **CDP Coordination Thread:** `https://muse.ai/thread/8f3e8782-3a1b-4afe-8520-2c81819e195b` (`HyperNexus main side chat`), prefixed with `[gemini in antigravity in c:\Users\hyper\workspace\marketing_agent on BOBPC]`.

## 2. What Shipped in `v0.7.7` – `v0.7.9`
1. **Inbound IMAP MIME Body Extraction, RFC 3501 UID Deduplication & DSN Bounce Processing (`v0.7.9` — `internal/communication/imap_receiver.go`, `internal/db/db.go`, `repository.go`):**
   - Fixed `EmailReceiver` fetching `imap.FetchBody` (`BODYSTRUCTURE` metadata, which left `email.Body == ""` on 100% of inbound emails) by switching to `BODY.PEEK[]` (`&imap.BodySectionName{Peek: true}`) + `ParseRawMIMEBody` (`multipart/*`, `quoted-printable`, `base64`, HTML fallback, RFC 2047 subject decoding).
   - Eliminated the 30-minute duplicate re-ingestion loop caused by RFC 3501 `UID <lastUID+1>:*` wrap-around and service-restart `lastUID` resets via strict `uid > r.lastUID` filtering + persistent `processed_inbound_emails` tracking (`Message-ID` / UID).
   - Fixed dead bounce handling (`mailer-daemon@googlemail.com` previously failed `GetContactByEmail` before `isBounce` could run): `ExtractPermanentBounceRecipients` parses permanent `5xx` DSNs (`Final-Recipient`, `Address not found`, `Message blocked`) while ignoring temporary `(Delay)` deferrals, clearing `contacts.email` (`NULLIF($1, '')` to preserve `contacts_email_key` uniqueness) and closing bounced deals as `Closed_Lost`.
2. **Stale `[Your Name]` DB Template Auto-Upgrade, Outbound Sanitizer & Polite-Decline Opt-Outs (`v0.7.9` — `internal/db/db.go`, `internal/communication/responder.go`, `manager.go`):**
   - Upgraded `seedTemplates()` from `ON CONFLICT (id) DO NOTHING` to auto-refresh legacy Postgres templates containing `[Your Name]`, `[My Name]`, or `LLM waterfall`, replacing all 3 stale production email templates with clean `HyperNexus Team` / `hypernexus.site` copy.
   - Added `SanitizeOutboundText` as defense-in-depth across `Generate`, `GenerateFromTemplate`, and `Manager.ProcessInbound`, plus `StripQuotedReply` and `isPoliteDecline` so polite non-interest / hobby-project replies transition cleanly to `Closed_Lost` without follow-up spam.
3. **OpenAI-Compatible `/v1` LLM URL Normalization (`v0.7.9` — `internal/llm/deepseek.go`, `hermes.go`, `cmd/marketing_agent/main.go`):**
   - Normalized `BaseURL` across `DeepSeekLLMProvider` and `HermesLLMProvider` so both root and `/v1` gateway URLs hit `/v1/models` and `/v1/chat/completions` accurately (reporting true `401 (invalid_key)` status instead of `404`), and wired `LLMIntentClassifier` when `DEEPSEEK_API_KEY` is set.
4. **Bot / Foundation-Org Hygiene & Enrichment FIFO Rotation (`v0.7.7` – `v0.7.9`):**
   - Filtered AI coding agents, CI/dependency bots (`-bot`, ` Bot`, `*[bot]*`, `action@github.com`, `actions@github.com`), malformed numeric/unlinked TLDs, and foundation-model/self orgs (`anthropic`, `anthropics`, `openai`, `modelcontextprotocol`, `hypernexussoft`, `robertpelloni`); added `TouchDealUpdatedAt` + GitHub rate-limit circuit tripping so the `17,900+` `Discovered` backlog rotates cleanly.

## 3. Dependabot Audit Notes
- `github/dependabot/go_modules/go_modules-20443bdb86` (`749d85d0`): **Skipped intentionally** — bumps `go.mod` toolchain directive from `go 1.25.0` to `go 1.26.0`, conflicting with `.github/workflows/ci.yml` (`go-version: '1.25'`).
- `github/dependabot/npm_and_yarn/npm_and_yarn-e9c900b0fc` (`61a41a74`): **Deferred intentionally** — bundles breaking major upgrades (`zod 3 -> 4`, `@hono/node-server 1 -> 2`).

## 4. Open Items for Operator / Next Cycle
1. **Optional Real-Card Charge Check:** Complete a real `$1.99` trial checkout on `https://hypernexus.site/checkout.html` whenever Robert chooses to verify live settlement + webhook tenant provisioning end-to-end.
2. **Provider API Keys (`/opt/marketing_agent/.env`):** Current `DEEPSEEK_API_KEY` / `HERMES_API_KEY` (`tp-s0j...` on `token-plan-sgp.xiaomimimo.com`) returns `HTTP 401 (invalid_key)`. Rotating a live LLM key (`DEEPSEEK_API_KEY` or `HERMES_API_KEY`) and fresh `GITHUB_TOKEN` will unlock live LLM replies to engaged prospects (e.g. `lancey3@gmail.com`) and 5,000 req/hr GitHub enrichment.
