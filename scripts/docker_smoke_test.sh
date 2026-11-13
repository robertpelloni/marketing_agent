#!/usr/bin/env bash
# scripts/docker_smoke_test.sh
# Validates syntax and configuration of Docker Compose files.

set -euo pipefail

echo "=========================================="
echo "Running Docker Compose Smoke Tests..."
echo "=========================================="

COMPOSE_FILES=(
    "docker-compose.staging.yml"
    "docker-compose.yml"
    "deploy/docker-compose.yml"
)

FAILED=0

for file in "${COMPOSE_FILES[@]}"; do
    if [ ! -f "$file" ]; then
        echo "FAIL: Compose file $file does not exist"
        FAILED=$((FAILED + 1))
        continue
    fi

    echo -n "Validating $file ... "
    if docker compose -f "$file" config -q 2>/dev/null; then
        echo "OK"
    else
        echo "FAILED"
        FAILED=$((FAILED + 1))
    fi
done

echo "------------------------------------------"
if [ "$FAILED" -eq 0 ]; then
    echo "All Docker Compose configurations are valid."
    exit 0
else
    echo "$FAILED Docker Compose file(s) failed validation."
    exit 1
fi
