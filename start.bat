@echo off
setlocal
REM Standing Hetzner-only policy: runtime execution lives on 5.161.250.43.
REM Allow non-server one-shot flags (-reconcile, -inventory) or explicit --allow-local override.
if /I "%~1"=="-reconcile" goto run_cli
if /I "%~1"=="-inventory" goto run_cli
if /I "%~1"=="--allow-local" (
    shift
    goto run_local
)

echo [Hetzner-Only Policy] Refusing local server launch on workstation.
echo Production runtime runs exclusively on Hetzner (5.161.250.43: systemctl status marketing-agent).
echo Pass --allow-local as the first argument only if explicitly authorized to run locally.
exit /b 1

:run_cli
if not exist "bin\marketing_agent.exe" (
    echo Binary not found. Running build...
    call build.bat
    if errorlevel 1 exit /b 1
)
bin\marketing_agent.exe %*
exit /b %errorlevel%

:run_local
echo Starting Marketing Agent (--allow-local override)...
if not exist "bin\marketing_agent.exe" (
    echo Binary not found. Running build...
    call build.bat
    if errorlevel 1 exit /b 1
)
bin\marketing_agent.exe %1 %2 %3 %4 %5 %6 %7 %8 %9
