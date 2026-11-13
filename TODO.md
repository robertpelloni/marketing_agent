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
- [ ] Decide fate of unreferenced media variants (hypernexus_v3/v4, psychedelic dupe, poster.jpg)
- [ ] Resubmit sitemap in Search Console after lastmod corrections

## Later
- [ ] signup.html: keep as dormant fallback or remove from nav entirely (currently bypassed, still served)
- [ ] index_a.html / index_b.html stale variants: delete or re-sync with index.html
- [ ] pricing.html $5/mo toggle vs $1.99 trial button: reconcile display model
## Autopilot directive 2026-11-12 dashboard plus cleanup queue
- [ ] Single-page dashboard: merge / demo and dashboard-outreach into / with section tabs tooltips high-value-first
- [ ] Nondestructive scripts archive: move legacy runners to archive with README
- [ ] Local full run: go build plus go test plus log watch and report gaps
