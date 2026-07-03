@echo off
echo Updating submodules...
git submodule update --init --recursive
if %errorlevel% neq 0 (
    echo Submodule update failed!
    exit /b %errorlevel%
)

echo Running Merge Integrity Tests...
go test ./internal/gitcheck/...
if %errorlevel% neq 0 (
    echo Integrity check failed! Please resolve conflicts or sync with main.
    exit /b %errorlevel%
)

echo Building Marketing Agent...
go build -o bin/marketing_agent ./cmd/marketing_agent
if %errorlevel% neq 0 (
    echo Build failed!
    exit /b %errorlevel%
)
echo Build successful.
