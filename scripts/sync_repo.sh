#!/bin/bash
# Autonomous Sync & Merge Protocol

echo "Starting Autonomous Sync & Merge Protocol..."

# 1. Upstream Tracking
echo "Fetching all tags and branches..."
git fetch --all --tags

# 2. Recursive Submodule Sanitization
echo "Updating submodules recursively..."
git submodule update --init --recursive

# 3. Intelligent Branch Reconciliation
CURRENT_BRANCH=$(git rev-parse --abbrev-ref HEAD)

if [ "$CURRENT_BRANCH" == "main" ]; then
    echo "On main branch. Syncing with remote..."
    git merge origin/main --no-edit
else
    echo "On feature branch: $CURRENT_BRANCH. Merging main back to feature..."
    git merge main -m "chore: sync feature branch with main" --no-edit
fi

# 4. Workspace Cleanup & Build
echo "Validating build..."
go build -v -o bin/sales_bot.exe ./cmd/sales_bot

echo "Sync complete."
