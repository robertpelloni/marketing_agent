# Legacy runners archived 2026-11-13

Superseded by the canonical runner `scripts/marketing_bot/` (see AGENTS.md).
Moved here nondestructively with `git mv`; history is preserved. Verified
unreferenced by AGENTS.md, README.md, `scripts/marketing_bot/`, and
`cmd/marketing_agent/` before the move.

- `auto_marketing_bot.py`, `auto_marketing_bot_v2.py` — early standalone posters
- `autonomous_marketing.py` — early autonomous loop
- `auto-post-all-platforms.py` — multi-platform one-off (auto_post.py was already gone)
- `apollo_bulk_export.js`, `apollo_bulk_export.py` — duplicate Apollo export one-offs

Do not run these; use `python scripts/marketing_bot/run.py`.
