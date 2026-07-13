@echo off
set SCHTASK_CMD=schtasks /create /tn "TormentNexus-CDP-Poster" /tr "wscript.exe C:\Users\hyper\workspace\marketing_agent\cdp_poster_silent.vbs" /sc hourly /mo 2 /st 00:00 /f

REM Delete old task
schtasks /delete /tn "TormentNexus-CDP-Poster" /f 2>nul

REM Create new task with VBS wrapper (no visible windows)
%SCHTASK_CMD%

echo Task updated — now runs completely silent every 2 hours.
echo No more console popups.
pause
