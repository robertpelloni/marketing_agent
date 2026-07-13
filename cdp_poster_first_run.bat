@echo off
REM cdp_poster_first_run.bat — Run ONCE visibly to log into Twitter + Reddit
REM After this, cdp_poster_auto.bat works fully headless with saved cookies

cd /d "%~dp0"

set CHROME_PROFILE=%LOCALAPPDATA%\TormentNexusCDP

echo ============================================================
echo  FIRST RUN SETUP — Login once, then automation runs headless
echo ============================================================
echo.
echo This will open Chrome. Please:
echo   1. Log into Twitter (x.com) manually in the Chrome window
echo   2. Log into Reddit (reddit.com) manually in the Chrome window
echo   3. Close Chrome when done — cookies will be saved
echo.
echo After this one-time setup, the scheduled task runs fully automatically.
echo.
pause

REM Launch Chrome VISIBLY with persistent profile
start "" "C:\Program Files\Google\Chrome\Application\chrome.exe" --remote-debugging-port=9222 --user-data-dir="%CHROME_PROFILE%" https://x.com/login https://old.reddit.com/login

echo.
echo Chrome opened. Log in to both sites, then come back here.
echo.
pause

REM Kill Chrome so the scheduled task can start fresh next time
taskkill /F /IM chrome.exe 2>nul

echo.
echo ✅ First-run setup complete! Chrome profile saved.
echo You can now install the scheduled task with: install_scheduled_task.bat
pause
