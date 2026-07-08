#!/bin/bash
# Headless-Safe Unix Build Script
# Compiles the application resolving the headless setupSystray tag difference.

set -e

echo "Updating submodules recursively..."
git submodule update --init --recursive 2>/dev/null || echo "No submodules configured."

echo "Running Merge Integrity Tests..."
go test ./internal/gitcheck/...

echo "Building Headless Marketing Agent..."
cd cmd/marketing_agent
go build -o ../../bin/marketing_agent main.go systray_unix.go
cd ../..

echo "Build successful — bin/marketing_agent"
