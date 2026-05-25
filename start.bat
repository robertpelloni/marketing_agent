@echo off
echo Starting Enterprise Sales Bot...
if not exist "bin/sales_bot.exe" (
    echo Binary not found. Running build...
    call build.bat
)
bin\sales_bot.exe %*
