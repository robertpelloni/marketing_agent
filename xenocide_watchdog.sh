#!/bin/bash
# XENOCIDE Watchdog — runs from Windows Task Scheduler every 5 min
LOG_FILE="/tmp/xenocide_watchdog.log"
DATE=$(date "+%Y-%m-%d %H:%M:%S")

log() { echo "[$DATE] $1" >>"$LOG_FILE"; }

# 1. LiteLLM Proxy (port 4000)
if curl -sf -m 3 http://localhost:4000/v1/models >/dev/null 2>&1; then
	log "OK: LiteLLM proxy"
else
	log "DOWN: LiteLLM proxy on port 4000"
fi

# 2. Local Bot (port 8085)
if curl -sf -m 3 http://localhost:8085/health >/dev/null 2>&1; then
	log "OK: Local bot"
else
	log "DOWN: Local bot on port 8085"
fi

# 3. LM Studio (port 1234)
if curl -sf -m 3 http://localhost:1234/v1/models >/dev/null 2>&1; then
	log "OK: LM Studio"
else
	log "DOWN: LM Studio on port 1234"
fi

# 4-5. Websites
curl -sf -m 5 https://tormentnexus.site/ >/dev/null 2>&1 && log "OK: tormentnexus.site" || log "DOWN: tormentnexus.site"
curl -sf -m 5 https://hypernexus.site/ >/dev/null 2>&1 && log "OK: hypernexus.site" || log "DOWN: hypernexus.site"

log "=== CYCLE COMPLETE ==="
