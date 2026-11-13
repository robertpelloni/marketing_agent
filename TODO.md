# TODO — marketing_agent

Immediate, granular tasks. Long-term direction lives in ROADMAP.md.

## Now
- [x] Merge backup/wip-leader-sync-20261107 -> main, push (2026-11-07; main == origin/main)
- [ ] Confirm leader-sync FF finished before next merge (another agent active today)
- [ ] Live real-charge test: 1.99 first month then 5-mo recurring (user running the charge).
- [x] Decide one-time $1.99 vs recurring (recurring chosen 2026-11-08: 1.99 first month, 5-mo after).
- [ ] Manual posts: LinkedIn (6) / X (tweets in POSTING_GUIDE) / Hashnode / Medium; tick sync checklist

## Next
- [ ] Integration workflow: fix or quarantine (red since 2026-07-11)
- [ ] Marketing bot secrets: DEVTO_API_KEY, BLUESKY_TN/HN_PASSWORD, HERMES_API_KEY
- [ ] Archive legacy scripts/ runners superseded by scripts/marketing_bot/ (verify unreferenced first):
  auto_marketing_bot.py, auto_marketing_bot_v2.py, autonomous_marketing.py,
  auto_post.py, auto-post-all-platforms.py, apollo_bulk_export.js/.py dupes
- [x] Decide fate of unreferenced media variants (2026-11-13: v3/v4/psychedelic/poster_optimized archived to hypernexus_site/archive/media/; poster.jpg kept, it is referenced)
- [ ] Resubmit sitemap in Search Console after lastmod corrections

## Later
- [x] signup.html: archived to hypernexus_site/archive/pages/ as dormant fallback (2026-11-13; was bypassed and unlinked)
- [x] index_a.html / index_b.html stale variants: archived with indexB.html and orphan home_router.html (2026-11-13)
- [ ] pricing.html $5/mo toggle vs $1.99 trial button: reconcile display model
## Autopilot directive 2026-11-12 dashboard plus cleanup queue
- [x] Single-page dashboard: outreach merged into / (#outreach, 301 from legacy route); /demo kept as major subproject view and linked from nav; tooltips + high-value-first sections (v0.7.3, 2026-11-13)
- [x] Nondestructive scripts archive: move legacy runners to archive with README (2026-11-13, scripts/archive/LEGACY_RUNNERS.md)
- [x] Local full run: go build plus go test plus log watch and report gaps (2026-11-13; only live-Apollo health check fails, needs valid API key)
- [ ] Align checkout.html copy with the decided $1.99-first-month-then-$5/mo model (merge took the GitHub-side $1.99 first-month page; pricing.html kept the trial copy — reconcile both surfaces)
- [x] Fix or skip-guard TestApolloSource_HealthCheck so the suite is green without live credentials (2026-11-13: 401 now skips like 403)
