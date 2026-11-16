# SCRIPTS_INDEX — scripts/ organization (2026-11-16)

Canonical runners stay at top level or in `marketing_bot/`. Nothing is
deleted; superseded files move to `scripts/archive/<yyyy-mm>/` via
`git mv` (history preserved, reversible).

## Canonical

| Path | Purpose |
|------|---------|
| `scripts/marketing_bot/` | Canonical multi-platform social runner (see its README/AGENTS) |
| `scripts/patch-nginx-hypernexus.py` | Live hypernexus.site nginx patch (already applied; tracked for reproducibility) |
| `scripts/check-site-qa.mjs` | Durable site-QA regression check |
| `scripts/dashboard-smoke.cjs` | Dashboard smoke check |
| `scripts/build_all.mjs` | Build helper |

## Archived 2026-11-16 (unreferenced; providers retired or superseded)

- Apollo/Hunter (retired 2026-11-16, GitHub-only enrichment):
  `apollo-hunter-prospect.sh`, `apollo_cdp_scraper.py`,
  `apollo_console_export.js`, `apollo_export.py`, `apollo_export_all.py`,
  `apollo_hunter_combined.py`, `apollo_prospect_builder.py`,
  `import_apollo_contacts.py`, `import_apollo_v2.py`
- Legacy social posters (superseded by `marketing_bot/`):
  `auto-post.py`, `post-to-twitter.py`, `post-to-twitter-api.py`,
  `twitter_reply_bot.py`, `cdp-crosspost.py`, `cdp-crosspost-edge.py`,
  `cross-post-articles.py`, `crosspost_blog.py`, `sync-all-platforms.py`,
  `post-to-hashnode.py`, `post-to-hashnode-v2.py`

All 20 moves are in `scripts/archive/2026-11/`.

## Unreferenced but kept in place (operator review)

These have no in-repo references but may be run manually or from the
server, so they were NOT moved: `backup_db.sh`, `backup_websites.sh`,
`gmail_sender.py`, `gmail_oauth_setup.py`, `send_intro_emails.py`,
`send_newsletter.py`, `send_newsletter_batch.py`, `discord_bot.py`,
`enrich-catalog.py`, `enrich-mimo.py`, `gen-catalog-page.py`,
`generate_showhn.py`, `purchasing_scraper.py`, `catalog-status.sh`,
`growth-dashboard.sh`, `referral-status.sh`, `sync_repo.sh`,
`tormentnexus-dev.sh`, `test_checkout.sh`, `test_checkout_live.sh`,
`test-billing-deployment.sh`, `update_templates.sql`, and similar
one-offs. Confirm each is dead before archiving.
