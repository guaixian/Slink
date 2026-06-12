@echo off
REM Windows 编译脚本 - 启用 CGO 支持 SQLite

echo 正在编译 Slink (启用 CGO)...

REM 设置 CGO 环境变量
set CGO_ENABLED=1

REM 编译
go build -o slink.exe main.go

if %ERRORLEVEL% EQU 0 (
    echo.
    echo 编译成功！
    echo 可执行文件: slink.exe
    echo.
    echo 运行程序: slink.exe
) else (
    echo.
    echo 编译失败！
    echo 请确保已安装 GCC 编译器 (MinGW)
    echo 下载地址: https://sourceforge.net/projects/mingw-w64/
)

pause
