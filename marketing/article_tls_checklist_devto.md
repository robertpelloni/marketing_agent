---
title: Harden Your Self-Hosted AI - A Practical Checklist for TLS, Auth, and Network Isolation
published: true
tags: ai, security, selfhosted, devops
canonical_url: https://hypernexus.site/blog/harden-your-self-hosted-ai-a-practical-checklist-for-tls-auth-and-network-isolation.html
---

Your self-hosted AI is probably not as secure as you think. Open ports, plaintext HTTP on localhost, no auth on tool endpoints - we see it in almost every audit.

Here is the 4-point checklist we run at HyperNexus for every deployment.

## 1. TLS Everywhere - Even on Localhost

Terminate TLS at Nginx with TLSv1.2+, HSTS, and redirect all port 80 to 443. Even internal Ollama / API traffic gets TLS or isolated loopback with firewall rules. No http:// in production.

Check: `curl -v http://your-host/` must 301 to `https://`. Certs from Lets Encrypt, auto-renewed.

## 2. Network Isolation

Separate VLANs / Docker networks for AI services, DB, and web. Only Nginx exposes 80/443. Model servers (Ollama 11434), Postgres 5432, and internal APIs bind to 127.0.0.1 or private net - never 0.0.0.0.

Check: `ss -tlnp` shows no AI backend on public interface.

## 3. Auth on Every Endpoint

No anonymous tool calls, memory reads, or model requests. Bearer tokens per tenant, RBAC with least privilege, SSO (no local passwords) for team access. Log every auth failure.

Check: unauthenticated `/api/v1/` request returns 401, not data.

## 4. Audit Logging

Append-only logs for every tool call, memory access, and model request. Ship to isolated storage. HyperNexus does this by default - who called what, when, with which prompt hash.

## How HyperNexus Ships This

Nginx hardened headers (HSTS, X-Frame-Options, nosniff), per-tenant tokens, isolated data dirs under /var/lib/hypernexus/tenants, and waterfall LLM failover so you never expose keys by falling back to insecure endpoints.

Originally published at https://hypernexus.site/blog/harden-your-self-hosted-ai-a-practical-checklist-for-tls-auth-and-network-isolation.html