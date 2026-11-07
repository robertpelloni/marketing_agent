# DEPLOY — Deployment & Environment Setup Instructions

## Production Server Environment
- **Target Host**: Hetzner VPS (`5.161.250.43`) running Ubuntu 24.04 LTS
- **Service Name**: `marketing-agent.service` (systemd managed)
- **Database**: PostgreSQL (`sales_bot` database with pgvector extension)
- **Web Server & Reverse Proxy**: Nginx with SSL certificates provisioned via Let's Encrypt
- **Port Allocations**:
  - Web Dashboard: `8087` (local staging), proxied via Nginx for production domains
  - TormentNexus Sidecar: `4300` / `7778`

## Build & Release Protocol
1. **Compile Binary**:
   ```bash
   ./build.sh
   ```
   This will run internal package tests, extract the version from `VERSION.md`, and output `bin/marketing_agent`.

2. **Deploy to Production VPS**:
   ```bash
   scp bin/marketing_agent root@5.161.250.43:/opt/marketing_agent/bin/
   ssh root@5.161.250.43 "systemctl restart marketing-agent"
   ```

3. **Verify Service Health**:
   ```bash
   curl http://5.161.250.43:8087/health
   curl http://5.161.250.43:8087/health/detailed
   ```
