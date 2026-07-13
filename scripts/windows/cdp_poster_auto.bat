@echo off
REM cdp_poster_auto.bat — Headless scheduled posting to Twitter + Reddit
REM Run by Windows Task Scheduler every 4 hours
REM Uses persistent Chrome profile so login happens once

cd /d "%~dp0"

set CHROME_PROFILE=%LOCALAPPDATA%\TormentNexusCDP

REM Kill existing Chrome debugging instances
taskkill /F /IM chrome.exe 2>nul
timeout /t 2 /nobreak >nul

REM Launch Chrome with persistent profile + remote debugging (minimized)
start /min "" "C:\Program Files\Google\Chrome\Application\chrome.exe" --remote-debugging-port=9222 --headless=new --no-sandbox --disable-gpu --user-data-dir="%CHROME_PROFILE%"

REM Wait for Chrome to fully load
timeout /t 5 /nobreak >nul

REM Post to both platforms
cdp_poster.exe -platform both 2>&1 >> cdp_poster.log

echo [%date% %time%] CDP poster cycle complete >> cdp_poster.log

REM Cleanup
taskkill /F /IM chrome.exe 2>nul
