# TODO — marketing_agent

Immediate, granular tasks. Long-term direction lives in ROADMAP.md.

## Now
- [x] Merge backup/wip-leader-sync-20261107 -> main, push (2026-11-07; main == origin/main)
- [x] Confirm leader-sync FF finished before next merge (done 2026-11-13: github/main, GitLab main, and feature/v0.7.1 branch all merged and pushed)
- [ ] Live real-charge test: 1.99 first month then 5-mo recurring (user running the charge).
- [x] Decide one-time $1.99 vs recurring (recurring chosen 2026-11-08: 1.99 first month, 5-mo after).
- [ ] Manual posts: LinkedIn (6) / X (tweets in POSTING_GUIDE) / Hashnode / Medium; tick sync checklist

## Next
- [x] Integration workflow: fixed (2026-11-13/14: dockerignore, golangci Go 1.25, 43+ lint issues, migrations 000013/000014, enrichment query, staging port; all workflows green)
- [ ] Marketing bot secrets: DEVTO_API_KEY, BLUESKY_TN/HN_PASSWORD, HERMES_API_KEY
- [ ] Archive legacy scripts/ runners superseded by scripts/marketing_bot/ (verify unreferenced first):
  auto_marketing_bot.py, auto_marketing_bot_v2.py, autonomous_marketing.py,
  auto_post.py, auto-post-all-platforms.py, apollo_bulk_export.js/.py dupes
- [x] Decide fate of unreferenced media variants (2026-11-13: v3/v4/psychedelic/poster_optimized archived to hypernexus_site/archive/media/; poster.jpg kept, it is referenced)
- [ ] Resubmit sitemap in Search Console after lastmod corrections

## Later
- [x] signup.html: archived to hypernexus_site/archive/pages/ as dormant fallback (2026-11-13; was bypassed and unlinked)
- [x] index_a.html / index_b.html stale variants: archived with indexB.html and orphan home_router.html (2026-11-13)
- [x] pricing.html $5/mo toggle vs $1.99 trial button: reconciled (2026-11-13, 34303c35)
## Autopilot directive 2026-11-12 dashboard plus cleanup queue
- [x] Single-page dashboard: outreach merged into / (#outreach, 301 from legacy route); /demo kept as major subproject view and linked from nav; tooltips + high-value-first sections (v0.7.3, 2026-11-13)
- [x] Nondestructive scripts archive: move legacy runners to archive with README (2026-11-13, scripts/archive/LEGACY_RUNNERS.md)
- [x] Full run verified on Hetzner only (user directive: no local servers; prod deployed + journal watched each loop; live-Apollo test now skips on 401)
- [ ] Align checkout.html copy with the decided $1.99-first-month-then-$5/mo model (merge took the GitHub-side $1.99 first-month page; pricing.html kept the trial copy — reconcile both surfaces)
- [x] Fix or skip-guard TestApolloSource_HealthCheck so the suite is green without live credentials (2026-11-13: 401 now skips like 403)

- [x] Classify Apollo 401/403 as failure (not empty) for honest source telemetry (2026-11-13, 7c914640, verified in prod metrics: Apollo failures=100)
