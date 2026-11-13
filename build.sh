#!/bin/bash
# Headless-Safe Unix Build Script
# Builds the whole cmd/marketing_agent package (build tags select
# systray_unix.go on Linux, systray.go on Windows) instead of naming
# individual files, so new files in the package are picked up automatically.
#
# NOTE: this script intentionally does NOT run `go test ./internal/gitcheck/...`.
# Those tests run `git reset` against the working checkout and would wipe
# uncommitted work. Run the test suite separately, in a clean worktree.

set -e

VERSION="$(cat VERSION 2>/dev/null || echo dev)"
echo "Building Marketing Agent v${VERSION}..."

mkdir -p bin
go build -trimpath -ldflags "-X main.version=${VERSION}" -o bin/marketing_agent ./cmd/marketing_agent

echo "Build successful — bin/marketing_agent (v${VERSION})"
