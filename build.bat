@echo off
echo Building Enterprise Sales Bot...
go build -o bin/sales_bot.exe ./cmd/sales_bot
if %errorlevel% neq 0 (
    echo Build failed!
    exit /b %errorlevel%
)
echo Build successful.
