@echo off
REM Install scheduled task — posts to Twitter + Reddit every 2 hours
REM Run as Administrator

schtasks /delete /tn "TormentNexus-CDP-Poster" /f 2>nul
schtasks /create /tn "TormentNexus-CDP-Poster" /tr "C:\Users\hyper\workspace\marketing_agent\cdp_poster_auto.bat" /sc hourly /mo 2 /st 00:00 /f

echo.
echo ============================================================
echo  Scheduled task: TormentNexus-CDP-Poster
echo  Frequency: Every 2 hours = 12 posts/day
echo  Subreddits: 10 rotating AI/LLM subs (one per cycle)
echo  Safety: Each sub sees ~1 post every 20 hours
echo ============================================================
echo.
schtasks /query /tn "TormentNexus-CDP-Poster" /fo list | findstr "TaskName Schedule"
echo.
echo Run cdp_poster_first_run.bat ONCE to login, then this runs auto.
pause
