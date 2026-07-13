@echo off
REM cdp_poster.bat — Posts to Twitter + Reddit via local Chrome CDP
REM Usage: cdp_poster [twitter|reddit|both]

REM Close any existing Chrome with remote debugging
taskkill /F /IM chrome.exe 2>nul

REM Launch Chrome with remote debugging port
start "" "C:\Program Files\Google\Chrome\Application\chrome.exe" --remote-debugging-port=9222 --new-window

REM Wait for Chrome to start
timeout /t 3 /nobreak >nul

REM Run the poster
cdp_poster.exe -platform "%~1" %2 %3 %4
