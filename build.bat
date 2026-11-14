@echo off
setlocal
echo Updating submodules if configured...
git submodule update --init --recursive 2>nul
if errorlevel 1 echo Submodule update skipped - no submodules configured.

set VERSION=dev
if exist VERSION set /p VERSION=<VERSION
if "%VERSION%"=="dev" if exist VERSION.md set /p VERSION=<VERSION.md

REM NOTE: Do NOT run `go test ./internal/gitcheck/...` here - its test suite runs
REM `git reset` against the active working tree. Keep build.bat aligned with build.sh.
echo Building Marketing Agent v%VERSION%...
go build -trimpath -ldflags "-X main.version=%VERSION%" -o bin\marketing_agent.exe .\cmd\marketing_agent
if errorlevel 1 (
    echo Build failed!
    exit /b 1
)
echo Build successful - bin\marketing_agent.exe v%VERSION%
