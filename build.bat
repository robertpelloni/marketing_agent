@echo off
echo Running Merge Integrity Tests...
go test ./internal/gitcheck/...
if %errorlevel% neq 0 (
    echo Integrity check failed! Please resolve conflicts or sync with main.
    exit /b %errorlevel%
)

echo Building Enterprise Sales Bot...
go build -o bin/sales_bot.exe ./cmd/sales_bot
if %errorlevel% neq 0 (
    echo Build failed!
    exit /b %errorlevel%
)
echo Build successful.
