@echo off
setlocal enabledelayedexpansion

set project=kohme

REM go mod tidy
go mod tidy
IF ERRORLEVEL 1 (
    exit /b 1
)

REM go generate
go generate
IF ERRORLEVEL 1 (
    exit /b 1
)

REM go build
set CGO_ENABLED=1
go build -ldflags "-s -w" -o %project% ./cmd/bot
IF ERRORLEVEL 1 (
    exit /b 1
)

echo build success %project%
