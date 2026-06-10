# Session Handoff: Hermes LLM Integration (v0.4.8)

## Overview
The TormentNexus Autonomous Sales Bot now has a real LLM brain. The `MockLLMProvider` has been replaced by `HermesLLMProvider`, which routes all LLM calls through a local Hermes Agent gateway running in WSL. This is the Phase 7 foundation — the first real integration replacing a mock component.

## What Changed (v0.4.8)

### 1. Hermes LLM Provider (`internal/llm/hermes.go`)
- `HermesLLMProvider` implements the `LLMProvider` interface using OpenAI-compatible `/v1/chat/completions` calls to the Hermes API server.
- `HermesConfig` struct holds `BaseURL`, `APIKey`, `Model` for connection configuration.
- `HealthCheck()` method verifies Hermes connectivity at startup and on-demand.
- Automatic fallback: if `HERMES_API_URL`/`HERMES_API_KEY` are not set, the bot falls back to `MockLLMProvider`.

### 2. LLM-Backed Intent Classification
- When Hermes is available, the bot uses `LLMIntentClassifier` instead of `MockIntentClassifier`.
- Real intent classification via LLM replaces keyword-matching heuristics.

### 3. Config Extensions (`internal/config/config.go`)
- New fields: `HermesAPIURL`, `HermesAPIKey`, `HermesModel`.
- Environment variables: `HERMES_API_URL`, `HERMES_API_KEY`, `HERMES_MODEL` (default: `free-llm`).

### 4. Dashboard Health Integration (`internal/web/server.go`)
- `web.NewServer()` now accepts `llm.LLMProvider` for health reporting.
- Dashboard shows LLM provider status: green "Hermes: Connected" or grey "Mock".
- `/health/detailed` JSON endpoint includes `llm_provider` field.

### 5. AGENTS.md Product Reference
- Added comprehensive "THE PRODUCT: TormentNexus AI Hypervisor" section (19KB) as the authoritative product reference for the sales bot.

## Hermes Setup (WSL)
- Hermes Agent v0.15.1 running in WSL as a systemd gateway service.
- API server configured with `API_SERVER_HOST=0.0.0.0` for cross-WSL/Windows access.
- API server key: `sales-bot-bridge-key-2026` (set in `~/.hermes/.env`).
- WSL IP: `172.21.116.32`, API port: `8642`.
- Model: `free-llm` (routes through litellm proxy at `172.21.112.1:4000`).

## Verification
- `go build ./...` — CLEAN
- `go vet ./...` — CLEAN
- `go test ./internal/...` — ALL PASS (16 packages)
- Integration test: Hermes health check + LLM generation → "Paris" response (15s, 54K tokens)

## Environment Variables for Production
```
HERMES_API_URL=http://172.21.116.32:8642
HERMES_API_KEY=sales-bot-bridge-key-2026
HERMES_MODEL=free-llm
```

## Next Steps
- **Phase 7.1:** Replace `MockApolloSource` with real Apollo.io API enrichment.
- **Phase 7.2:** Implement SMTP email sender using Hermes's himalaya skill or direct SMTP.
- **Phase 7.3:** Implement IMAP email polling for inbound ingestion.
- **Phase 6.4:** Implement structured JSON logging (slog).
- **Phase 8:** Use Hermes subagent delegation for AutoDev code generation (replace `LocalAgent`).
