@echo off
echo Updating submodules (if configured)...
git submodule update --init --recursive 2>nul
if %errorlevel% neq 0 (
    echo Submodule update skipped (no submodules configured).
)

echo Running Merge Integrity Tests...
go test ./internal/gitcheck/...
if %errorlevel% neq 0 (
    echo Integrity check failed! Please resolve conflicts or sync with main.
    exit /b %errorlevel%
)

echo Building Marketing Agent...
go build -o bin/marketing_agent.exe ./cmd/marketing_agent
if %errorlevel% neq 0 (
    echo Build failed!
    exit /b %errorlevel%
)
echo Build successful — bin/marketing_agent.exe
