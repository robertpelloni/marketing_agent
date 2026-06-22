@echo off
REM XENOCIDE Watchdog — checks all services, logs issues
REM Usage: double-click or run from command line
REM To run every 5 min: schtasks /create /tn "XENOCIDE Watchdog" /tr "C:\path\to\xenocide_watchdog.bat" /sc minute /mo 5 /f

set LOG=%TEMP%\xenocide_watchdog.log
set DATE=%DATE% %TIME%

echo [%DATE%] === XENOCIDE WATCHDOG === >> %LOG%

REM 1. LiteLLM Proxy (port 4000)
powershell -Command "try { $r = Invoke-WebRequest -Uri 'http://localhost:4000/health' -Method GET -Headers @{'Authorization'='Bearer sk-litellm'} -UseBasicParsing -TimeoutSec 3; if ($r.StatusCode -eq 200) { exit 0 } else { exit 1 } } catch { exit 1 }" > nul 2>&1
if %errorlevel% equ 0 (
    echo [%DATE%] OK: LiteLLM proxy >> %LOG%
) else (
    echo [%DATE%] DOWN: LiteLLM proxy on port 4000 >> %LOG%
)

REM 2. Local Bot (port 8085)
powershell -Command "try { $r = Invoke-WebRequest -Uri 'http://localhost:8085/health' -UseBasicParsing -TimeoutSec 3; if ($r.StatusCode -eq 200) { exit 0 } else { exit 1 } } catch { exit 1 }" > nul 2>&1
if %errorlevel% equ 0 (
    echo [%DATE%] OK: Local bot >> %LOG%
) else (
    echo [%DATE%] DOWN: Local bot on port 8085 >> %LOG%
)

REM 3. LM Studio (port 1234)
powershell -Command "try { $r = Invoke-WebRequest -Uri 'http://localhost:1234/v1/models' -UseBasicParsing -TimeoutSec 3; if ($r.StatusCode -eq 200) { exit 0 } else { exit 1 } } catch { exit 1 }" > nul 2>&1
if %errorlevel% equ 0 (
    echo [%DATE%] OK: LM Studio >> %LOG%
) else (
    echo [%DATE%] DOWN: LM Studio on port 1234 >> %LOG%
)

REM 4. tormentnexus.site
powershell -Command "try { $r = Invoke-WebRequest -Uri 'https://tormentnexus.site/' -UseBasicParsing -TimeoutSec 5; if ($r.StatusCode -eq 200) { exit 0 } else { exit 1 } } catch { exit 1 }" > nul 2>&1
if %errorlevel% equ 0 (
    echo [%DATE%] OK: tormentnexus.site >> %LOG%
) else (
    echo [%DATE%] DOWN: tormentnexus.site >> %LOG%
)

REM 5. hypernexus.site
powershell -Command "try { $r = Invoke-WebRequest -Uri 'https://hypernexus.site/' -UseBasicParsing -TimeoutSec 5; if ($r.StatusCode -eq 200) { exit 0 } else { exit 1 } } catch { exit 1 }" > nul 2>&1
if %errorlevel% equ 0 (
    echo [%DATE%] OK: hypernexus.site >> %LOG%
) else (
    echo [%DATE%] DOWN: hypernexus.site >> %LOG%
)

echo [%DATE%] === CYCLE COMPLETE === >> %LOG%

REM Show last 10 lines
echo.
echo === XENOCIDE WATCHDOG ===
echo Log: %LOG%
echo.
powershell -Command "Get-Content '%LOG%' -Tail 10"
echo.
