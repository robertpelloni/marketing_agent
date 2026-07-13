# TormentNexus — Ultimate Battle Plan

## 🔴 CRITICAL (fix now)

### 1. Stripe Webhook Fix

**Problem:** `stripe-go 81.4.0 expects API version 2025-02-24.acacia` but Stripe sends `2026-06-24.dahlia`
**Fix:** Add `ignoreAPIVersionMismatch: true` to webhook construction
**Impact:** Revenue flows again

### 2. TormentNexus TTY Crash

**Problem:** 8,088 restarts from "could not open /dev/tty"
**Fix:** Remove TTY dependency in TN Go binary, or use `script -q -c` wrapper
**Impact:** Admin panel + container provisioning come back online

## 🟡 LAUNCH (this week)

### 3. Product Hunt — Tuesday 12am PT

Everything written in `launch_materials/product_hunt.md`
Just needs: screenshots of dashboard + GitHub README paste

### 4. Show HN — Tuesday 8am ET

Everything written in `launch_materials/show_hn.md`
Post immediately after PH goes live

### 5. Awesome List PRs

7 repos listed in `launch_materials/awesome_lists_and_directories.md`
Each PR = permanent backlink = free traffic forever

### 6. AI Directory Submissions

10 directories listed in launch materials
Copy-paste the template, submit one per day

## 🟢 GROWTH (ongoing)

### 7. Comparison Pages (SEO goldmine)

- "TormentNexus vs Cursor"
- "TormentNexus vs Claude Code"
- "TormentNexus vs Windsurf"
- "TormentNexus vs Continue.dev"
These rank #1 for comparison searches = high-intent traffic

### 8. YouTube Short

60-second demo: open TormentNexus dashboard, show multi-agent swarm, progressive routing
Post to YouTube Shorts + TikTok + Instagram Reels

### 9. Apollo Paid Key

$49/mo unlocks 10x better enrichment
1,411 deals waiting for emails right now

### 10. Star Campaign

- Add star badge to both websites (done on HN, need TN)
- Add "⭐ Star us" to every blog post footer
- Ask existing contacts in follow-up email

## 💰 REVENUE

### 11. Stripe Fix #2

Accept checkout metadata (seats) and auto-provision containers
Already coded — just needs webhook fix first

### 12. Free Trial → Paid Email Sequence

Auto-email trial users at day 7, 14, 28 with upgrade prompts
Already have email infrastructure

## 🏗️ INFRASTRUCTURE

### 13. Monitoring Dashboard

Single page showing all service health
Could use the existing admin panel

### 14. PostgreSQL Backups

Daily pg_dump to S3/backup location
Current DB is irreplaceable

### 15. Disk Alert

75% and growing — need cleanup cron job
