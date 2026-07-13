@echo off
REM cdp_poster_silent.bat — No console windows, Chrome stays alive
cd /d "C:\Users\hyper\workspace\marketing_agent"

REM Check if Chrome is already running with debugging port
curl -s http://127.0.0.1:9222/json/version >nul 2>&1
if errorlevel 1 (
    REM Chrome not running — launch it minimized, headless
    start /min "" "C:\Program Files\Google\Chrome\Application\chrome.exe" --remote-debugging-port=9222 --headless=new --no-sandbox --disable-gpu --user-data-dir="%LOCALAPPDATA%\TormentNexusCDP"
    REM Wait for Chrome to be ready
    timeout /t 4 /nobreak >nul
)

REM Post to both platforms
cdp_poster.exe -platform both >> cdp_poster.log 2>&1

REM Never kill Chrome — keep it alive for next cycle
