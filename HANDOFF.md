# HANDOFF.md — marketing_agent (`v0.7.11`, 2026-11-17)

## 1. Current State Summary
- **Version:** `0.7.11` (`VERSION` and `VERSION.md` aligned; 2026-11-16/17 fix batch: lane retirements, cadence aborts, discussions engager, LLM stopgap hardening)
- **Remotes & Production:** Local `main`, GitHub (`github/main`), GitLab (`origin/main`), and Hetzner production (`5.161.250.43:/opt/marketing_agent`) synchronized on `v0.7.10` with clean working trees and exact deployed binary parity (`go version -m /opt/marketing_agent/bin/marketing_agent` -> `vcs.revision == git rev-parse HEAD`, `vcs.modified=false`).
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

## LOOP 2026-11-16 Muse Code (Antigravity catch-up)
- Latest Antigravity for this project: fac645a2 'Analyzing Project and PI Session' (enterprise_sales_bot workspace, 11,387 steps, active 2026-11-15): site QA audits (bad links, umami/blogs, inbound bodies, Hetzner QA) against marketing_agent. No marketing_agent-scoped conversation exists in the Antigravity index; HN imported-sessions upstream unavailable (empty fallback).
- Re-ran its bad-link audit on v0.7.10: MDMAtk download links already fixed by other lanes (releases.hypernexus.site live 200). Remaining trial copy normalized on download/account/checkout/compare (5d9fc59a, deployed to /var/www, live verified).
- Funnel update: Outreach_Sent 1,039 (cadence firing), contacts 4,359, Closed_Lost 52, Negotiating 2. DeepSeek still 401; disk 84% watch.

## FIX 2026-11-16 Muse Code (GitHub comment sender)
- GitHubCommentSender: 401/403 on search now switches to anonymous client + marks tokenRejected; posting skipped gracefully while rejected; anonymous 403 rate-limit (10/min) stops search softly with no error so cadence records no failed sends (46980ef2, deployed, health OK). Tests: AnonFallbackOn401, AnonRateLimitStopsSoftly.
- Context: 27 failed outreach rows on 2026-11-16 traced to revoked GITHUB_TOKEN in this sender (not email). Refreshing GITHUB_TOKEN re-enables GitHub comment outreach; until then discovery still works anonymously within quota.
- Note: system_health intermittently 'Health check failed' when GitHub API rate limits bite the unauthenticated tracker; recovers.

## FIX 2026-11-16 Muse Code (Gmail daily limit)
- Outreach failed count growth traced to Gmail 550 5.4.5 daily limit: cadence aborted per cycle now on isDailyLimitError (f6b9191a, test + deployed). Failed stats before fix are quota artifacts, not recipient failures.
- Gmail daily quota exhaustion also proves cadence email volume is substantial; consider pacing/multi-account if volume keeps hitting the cap.

## DIRECTIVES 2026-11-16 (user) Muse Code
- GITHUB_TOKEN refreshed from workspace parent .env (validated ghp_, search 30/min + core quota): saved to local marketing_agent/.env (single line) and prod /opt/marketing_agent/.env (backup env.bak-token-*). Verified post-restart: GitHub Commits 17 successes / 0 failures - authenticated lane healthy. GITHUB_POSTING_DISABLED is set in prod env - check its value before expecting comment posts.
- Apollo/Hunter RETIRED from enrichment chain (b2ede783) + dashboard panel updated (531d5a80). GitHub-only enrichment.
- FreeLLM (aimm-freellm.service, :4000, /opt/aimoneymachine/freellm): up 16h+, 454 models, actively routing/judging. Probes: 1x 90s timeout, 1x 21s success, 1x cached instant; 139 failed provider attempts/24h with self-healing model disables. Verdict: works, fine as fallback, NOT primary-grade reliability/latency. Candidate: point marketing_agent LLM at http://127.0.0.1:4000/v1 (model free-llm) as DeepSeek stopgap - awaiting user nod.
- Bluesky: user can supply credentials (env names already present). X: user will NOT fund credits - X lane stays failing 402 each social cycle unless keys blanked or a disable flag is added.

## LOOP 2026-11-16 Muse Code (catch-up continuation)
- Caught up from repo state (HEAD cedfdbc2 v0.7.10) + HANDOFF/TODO. Prod verified: /health/detailed OK (Stripe live, DB OK), disk 84% (13G free, /opt/marketing_agent only ~500M - watch, no cleanup target there), GITHUB_POSTING_DISABLED=true in prod env (GitHub comment posting intentionally off; discovery/enrichment unaffected).
- X lane fixed in code: TWITTER_DISABLED hub gate + ErrTwitterCreditsDepleted sentinel (402) + poster skip for unregistered platforms (no LLM gen, no simulated rows). Tests: pkg/agents/social_platform_test.go (3 tests, green). Prod env sets TWITTER_DISABLED=true.
- scripts/patch-nginx-hypernexus.py committed; patch already live (verified /features 301, extensionless 200, real 404).
- FreeLLM stopgap wired (prod env only, backup taken): DEEPSEEK_BASE_URL=http://127.0.0.1:4000, DEEPSEEK_MODEL=free-llm, placeholder key. FreeLLM chat verified from prod host before switch. Rollback: restore env backup + restart.

## LOOP 2026-11-16 Muse Code (autopilot directive)
- Git sanitization: main 0/0 vs GitHub+GitLab @ 058d6c39. Submodules removed by design; borg-pointer branches (fix/jules-borg-pin, backup/fdafa92-pre-borg-repin) deliberately unmerged (recorded ROADMAP M8). Stashes from v0.7.0 era left untouched.
- Dashboard: Section 6A extracted to providerPanelList (lane truth: TWITTER_DISABLED / GITHUB_POSTING_DISABLED shown amber); /api/v1/leads wired into Operator Tools 6D (List Leads). Tests: internal/web/dashboard_panel_test.go green. Dashboard remains the single-page hub (Subscription/KPIs/Outreach/Deals/Social/Ops/Tools + Demo Console) - no fragmentation to consolidate. No system tray icon exists for this project (headless Go service under systemd; quit = systemctl stop marketing-agent).
- Scripts: 20 retired Apollo/Hunter + superseded poster scripts git-mv'd to scripts/archive/2026-11/; scripts/SCRIPTS_INDEX.md indexes canonical vs kept-unreferenced.
- Log watch findings (prod, post-deploy): Gmail 550 5.4.5 daily limit hit again (cadence abort contains it; volume pacing is the durable fix); LinkedIn headless (go-rod) times out on the VPS every cycle -> simulation fallback, and its 60s context used to poison the audit-row persist (fixed 058d6c39: fresh 10s persist ctx); Reddit headless stays skipped by design.
- Coordination: Edge CDP 9222 check-in SUBMITTED to Muse canonical thread 8f3e8782 as [muse-wonder-1.0 in muse-cli in C:\Users\hyper\workspace\marketing_agent on laptop]. No Gemini/ChatGPT tabs open - nothing sent there.
- Deployed: 058d6c39 (vcs parity clean, version 0.7.10), /health/detailed all green except system_health Pending (watch).

## LOOP 2026-11-17 Muse Code (autopilot directive, round 2)
- **Gmail OAuth client DISABLED by Google** (found in prod logs): every cadence email fails at token fetch (`disabled_client`). Operator must re-enable/replace the OAuth client in Google Cloud Console; code mitigation shipped (e12df7be): `isProviderAuthError` aborts the cadence cycle with one operator-action line - verified live (1 attempt -> abort, no per-deal failed stats).
- **OOM kill 2026-11-17 02:28 UTC** (718M peak, systemd restart): headless Chromium lanes prime suspect. LinkedIn social lane gated off (`LINKEDIN_DISABLED=true`, hub gate + panel line, 04d8d588 deployed). Cadence linkedin-connect still headless - watch.
- **system_health=Failure root cause:** unscoped Dependabot `uv` dynamic jobs failing on vendored `go/agents/software-agent-sdk`; CI/Integration/Deploy/Docker/CodeQL all green on HEAD. Scoped `.github/dependabot.yml` shipped (force-added; `.github/` is gitignored at .gitignore:80). Confirm Success next cycle.
- **Disk 86% (11G free):** growth is other tenants' /tmp backups (3.3G) + playwright cache (701M) - left untouched, owner decision in TODO.
- Full `go test ./...` green (20 packages, BelowNormal). Deployed 04d8d588, vcs parity clean. CDP check-in SUBMITTED to open Bobbi tab (thread b6ac3348..., note: differs from canonical 8f3e8782 in §1); no Gemini/ChatGPT tabs open.

## LOOP 2026-11-17 Muse Code (autopilot directive, round 3)
- Verified round 2 held: system_health=Success (dependabot scoping effective), Gmail abort contained (single-digit lines/day), no new OOM, service ~19-22MB with headless lanes off.
- LinkedIn retirement completed on all 3 surfaces (c954a78b deployed): social hub gate, cadence linkedin-connect -> Skipped via linkedInLaneDisabled, discovery LinkedInSource.Discover -> immediate nil (was 47s timeout + 3 fabricated companies; test fails on old code exactly that way, passes now). All three verified in one prod startup log.
- New findings: IMAP inbound login invalid (operator: refresh app password - TODO); one FreeLLM blog-gen 120s timeout (stopgap flakiness, watch).
- CDP check-in SUBMITTED (Bobbi tab); no Gemini/ChatGPT tabs. Deployed c954a78b, vcs parity clean.

## LOOP 2026-11-17 Muse Code (autopilot directive, round 4)
- LLM stopgap hardened (b2209e18 deployed): DEEPSEEK_MODEL=free-llm-plain (reasoning free-llm leaked "The user wants..." thinking into copy + timed out on blogs), DEEPSEEK_TIMEOUT_SECONDS=300 (new provider config, default 120, TestDeepSeekTimeoutConfigurable). Verified: long generations succeed (3,286 tokens), direct marketing succeeds, cadence Gmail abort still clean.
- Remaining noise explained: pre-fix PID deadline line, designed Gmail abort, one Bluesky createSession connection timeout (fell back to simulation; posted fine in earlier cycles - watch).
- DEPLOY.md updated with lane flags/stopgap/mail repairs. system_health returned to Success. Operator list unchanged: Gmail OAuth re-enable, IMAP app password, real LLM key rotation, disk 86% (other tenants).

## LOOP 2026-11-17 Muse Code (autopilot round 5, 15m cron 6060af34)
- GitHub Discussions engager was a silent no-op skeleton in prod (search discarded responses, postReply nil-stub). Implemented for real (752f44c2): GraphQL JSON decode + addDiscussionComment mutation + GITHUB_POSTING_DISABLED cycle gate; 4 tests green. Prod stays gated off until operator unsets the flag.
- DISK ESCALATION: Hetzner 91% (6.9G free, was 86%): other tenants (aicrm/ercm/build-leadg) write ~1G backups to /tmp hourly. Owner cleanup decision needed; this project's footprint ~500M. If disk fills, Postgres dies first.
- Cron loop scheduled: job 6060af34 every 15m (permanent, skips while active) re-runs this directive.
- Verified: cadence GitHub step already honors GITHUB_POSTING_DISABLED (github_sender.go:60); codebase TODO scan clean except Reddit API-only (recorded, lane retired by design).

## LOOP 2026-11-17 Muse Code (autopilot round 6)
- Autodev test pollution fixed at source (e112bca0 deployed): finalizeCycle honors GO_TEST_MODE; full `go test ./...` (20 pkgs) now leaves a clean tree (was: phantom VERSION/CHANGELOG bumps under internal/autodev after every run).
- Disk still 91% (6.9G free, stable this cycle). Health all-green; Pending is CI on the new push. Operator items unchanged.

## LOOP 2026-11-17 Muse Code (autopilot round 13)
- Gmail auth circuit breaker shipped (57254341, deployed): 15m fail-fast after provider auth rejection on ALL caller paths; verified live (single trip log 07:22:38 UTC, cadence abort follows, sender quiet after). Box load spiked to 103 earlier (other tenants' Rust build/Ollama - ours 23MB idle); Stripe 'unreachable' blip was that starvation, self-recovered.

## LOOP 2026-11-17 Muse Code (autopilot round 12)
- Dashboard Leads paging controls shipped (c4527c1f, deployed): Prev/Next + range display over paginated /api/v1/leads. Verified: health Success, leads-noauth 303, zero panics post-deploy, JS parse-checked.
- Gmail per-attempt ERROR lines continue at low rate from Comm Manager initiations (sender-level log; cadence cycles abort cleanly). Only fixable by operator re-enabling the OAuth client.

## LOOP 2026-11-17 Muse Code (autopilot round 11)
- Signed-session build verified settled: no panic, NRestarts=0, ~17MB; system_health Pending only while Docker/Deploy workflows run on the docs push (CI/CodeQL/Integration success).
- OOM guard applied: MemoryHigh=1G drop-in persisted (50-MemoryHigh.conf), verified live (ActiveState=active, health green). TODO system_health + OOM items closed; DEPLOY.md documents guard + session secret behavior.
- CDP retry: canonical Muse thread 8f3e8782 open (check-in sent); still NO Gemini/ChatGPT tabs in Edge - nothing to send there. Disk 87% (10G free).

## LOOP 2026-11-17 Muse Code (autopilot round 10)
- Signed sessions shipped (6de5084f, deployed): static `authorized_admin_session`/`company_N` cookies replaced by HMAC-SHA256 `v1.<subject>.<exp>.<sig>` tokens (24h). Secret = SESSION_SECRET or ADMIN_PASSWORD-derived (no new required prod env). All 4 web cookie parses now validated via auth.CompanyID. Live-verified: legacy + forged company cookies -> 303, wrong-password login -> 401, health green. Existing browsers/tenants must log in once more. Full suite 21 pkgs green.
- Leads pagination also shipped this cycle (d9525181): limit default 100/max 500, X-Total-Count 24116, verified 579B page vs 6.7MB.

## LOOP 2026-11-17 Muse Code (autopilot round 9; round 8 = MEMORY.md docs)
- SECURITY FIX (6ffeb777 + 45971bfb, deployed): /api/v1/leads, /api/v1/outreach/export, /api/v1/metrics, /metrics were UNAUTHENTICATED while the service binds *:8084 and nginx proxies /api/v1/ publicly - outreach CSV (emails), full CRM list/create, DB metrics all internet-facing. Now session-gated (303 unauth; dashboard same-origin unaffected). Public funnel (health/quote/checkout/newsletter/stats aggregates) unchanged.
- Leads 500 root-caused via new handler error logging: NULL market_cap_tier (+ NULL arrays) scan failures; COALESCE fix verified live (200, 24k companies, 6.7MB - pagination TODO'd). handleDetailedHealth nil-guards added (was panic on nil deps).
- MCP scope: the box's MCP stack is the separate HyperNexus kernel project (:7778, healthy v1.0.13, 31d uptime) + hypernexus-cloud/dashboard/watchdog services, all running. Exhaustive MCP tool testing belongs to that project, not marketing_agent; this repo's go/ subtree is vendored.
- Verified externally: IP:8084 answers (hence the auth fix mattered); static session cookie value noted as tech debt in TODO.

## LOOP 2026-11-17 Muse Code (autopilot round 7)
- Released v0.7.11 (49643912, deployed, health version=0.7.11): the 2026-11-16/17 fix batch. Worker inventory audited: all 12 goroutine workers in main.go log started (scraper, enricher, researcher, CRM, outreach/discovery, deployer, social poster, blog, discussions, comm manager, IMAP, autodev orchestrator) + deployment monitor. Nothing not-running that should run except operator-blocked lanes.
