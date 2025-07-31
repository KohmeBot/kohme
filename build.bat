@echo off
set project=kohme

REM 整理依赖
go mod tidy
if errorlevel 1 (
    exit /b 1
    pause
)

REM 运行插件
go run ./cmd/plugin
if errorlevel 1 (
    exit /b 1
    pause
)

REM 编译项目
go build -ldflags "-s -w" -o %project%.exe ./cmd/bot
if errorlevel 1 (
    exit /b 1
    pause
)

echo Build success: %project%.exe
pause
