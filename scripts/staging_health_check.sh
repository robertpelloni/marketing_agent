#!/usr/bin/env bash
# scripts/staging_health_check.sh
# Verifies staging or production endpoints for the marketing agent.

set -euo pipefail

BASE_URL="${STAGING_URL:-http://localhost:8087}"
TIMEOUT_SEC="${HEALTH_TIMEOUT:-5}"

echo "=========================================="
echo "Marketing Agent Staging Health Check"
echo "Target Base URL: ${BASE_URL}"
echo "Timeout: ${TIMEOUT_SEC}s"
echo "=========================================="

FAILED=0

check_endpoint() {
    local method="$1"
    local endpoint="$2"
    local expected_code="$3"
    local url="${BASE_URL}${endpoint}"

    echo -n "Checking [${method}] ${endpoint} ... "
    local http_code
    http_code=$(curl -s -o /dev/null -w "%{http_code}" --max-time "${TIMEOUT_SEC}" -X "${method}" "${url}" || echo "000")

    if [ "${http_code}" = "${expected_code}" ]; then
        echo "OK (HTTP ${http_code})"
    else
        echo "FAILED (Expected HTTP ${expected_code}, got ${http_code})"
        FAILED=$((FAILED + 1))
    fi
}

# 1. Health Ping
check_endpoint "GET" "/health" "200"

# 2. Public Outreach Stats
check_endpoint "GET" "/api/v1/outreach/stats" "200"

# 3. Devvit Reddit Social Content API (v1 and alias)
check_endpoint "GET" "/api/v1/social/reddit" "200"
check_endpoint "GET" "/api/social/reddit" "200"

# 4. CSV Outreach Export
check_endpoint "GET" "/api/v1/outreach/export" "200"

echo "------------------------------------------"
if [ "${FAILED}" -eq 0 ]; then
    echo "All staging health checks PASSED successfully."
    exit 0
else
    echo "${FAILED} health check(s) FAILED."
    exit 1
fi
