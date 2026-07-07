@echo off
echo Starting Marketing Agent...
if not exist "bin/marketing_agent.exe" (
    echo Binary not found. Running build...
    call build.bat
)
bin\marketing_agent.exe %*
