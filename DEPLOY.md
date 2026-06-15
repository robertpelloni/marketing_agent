# Deployment Guide

## Prerequisites

- **Go 1.24+**
- **PostgreSQL 13+**
- **Docker & Docker Compose** (for containerized deployment)
- **GitHub PAT** (with `repo` permissions)

## Configuration

The bot is configured entirely via environment variables. Create a `.env` file in the root directory.

### Core Variables

| Variable | Description | Default |
|---|---|---|
| `DATABASE_URL` | PostgreSQL connection string | `postgres://...` |
| `PORT` | Web dashboard port | `8080` |
| `ENVIRONMENT` | deployment environment | `development` |
| `ADMIN_PASSWORD` | Dashboard login password | `admin` |

### Integrations

#### GitHub
- `GITHUB_TOKEN`: Your Personal Access Token.
- `GITHUB_REPOSITORY`: `owner/repo` for CI tracking and AutoDev.
- `GITHUB_BOT_USERNAME`: Username for outreach comments.
- `GITHUB_WEBHOOK_SECRET`: HMAC secret for webhook verification.

#### LLM (Hermes)
- `HERMES_API_URL`: URL to your Hermes Agent gateway.
- `HERMES_API_KEY`: API key for Hermes.
- `HERMES_MODEL`: Model name to use (e.g., `mistral-large`).

#### CRM
- `CRM_PROVIDER`: `salesforce`, `hubspot`, or `generic`.
- `CRM_BASE_URL`: For generic REST CRM.
- `CRM_API_KEY`: For generic REST CRM.
- **Salesforce:** `SALESFORCE_INSTANCE_URL`, `SALESFORCE_ACCESS_TOKEN`.
- **HubSpot:** `HUBSPOT_ACCESS_TOKEN`.

#### CRM Dynamic Field Mapping (Optional)
- `CRM_DEAL_NAME_PROP`: (e.g., `Name`, `dealname`)
- `CRM_DEAL_AMOUNT_PROP`: (e.g., `Amount`, `amount`)
- `CRM_DEAL_STAGE_PROP`: (e.g., `StageName`, `dealstage`)
- `CRM_DEAL_DESC_PROP`: (e.g., `Description`, `description`)
- `CRM_DEAL_ROUTE_PROP`: Custom field for routing metadata.

#### Outreach
- **SMTP:** `SMTP_HOST`, `SMTP_PORT`, `SMTP_USERNAME`, `SMTP_PASSWORD`, `SMTP_FROM`.
- **IMAP:** `IMAP_HOST`, `IMAP_PORT`, `IMAP_USERNAME`, `IMAP_PASSWORD`.
- **LinkedIn:** `LINKEDIN_USERNAME`, `LINKEDIN_PASSWORD`.

## Local Deployment

1. **Clone and Sync:**
   ```bash
   git clone ...
   git submodule update --init --recursive
   ./scripts/sync_repo.sh
   ```

2. **Database:**
   Ensure PostgreSQL is running and apply migrations (if using a manual tool) or let the bot auto-apply them on startup.

3. **Build and Run:**
   ```bash
   go build -o bin/sales_bot ./cmd/sales_bot
   ./bin/sales_bot
   ```

## Docker Deployment

```bash
docker compose up --build -d
```

## Production Hardening

1. **Auth:** Change `ADMIN_PASSWORD` immediately.
2. **Safety:** Set `DRY_RUN=true` initially to verify outreach logs without sending real messages.
3. **Monitoring:** Access `/metrics` for Prometheus data and `/health/detailed` for worker status.
