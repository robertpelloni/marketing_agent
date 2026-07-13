@echo off
schtasks /delete /tn "TormentNexus-CDP-Poster" /f 2>nul
schtasks /create /tn "TormentNexus-CDP-Poster" /tr "wscript.exe C:\Users\hyper\workspace\marketing_agent\cdp_poster_silent.vbs" /sc hourly /mo 2 /st 00:00 /ru "SYSTEM" /f
echo Done — Chrome stays alive, no taskkill, no popups ever.
pause
