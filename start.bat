@echo off
REM 快速启动脚本 - 不检查GCC

echo 正在启动 Slink...

set CGO_ENABLED=1
go run main.go

pause
