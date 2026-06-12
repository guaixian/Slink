@echo off
REM Windows 运行脚本 - 启用 CGO 支持 SQLite

echo ========================================
echo 正在启动 Slink (开发模式)
echo ========================================
echo.

REM 设置 CGO 环境变量
set CGO_ENABLED=1

REM 显示CGO状态
echo CGO_ENABLED=%CGO_ENABLED%
echo.

REM 检查是否安装了GCC
where gcc >nul 2>&1
if %ERRORLEVEL% NEQ 0 (
    echo [警告] 未检测到 GCC 编译器！
    echo SQLite 需要 GCC 支持。
    echo 请安装 MinGW-w64: https://sourceforge.net/projects/mingw-w64/
    echo.
    echo 或者使用 MySQL/PostgreSQL 数据库（不需要 GCC）
    echo.
    pause
    exit /b 1
)

echo [信息] 检测到 GCC 编译器
echo.

REM 运行
echo 正在启动服务器...
echo.
go run main.go

if %ERRORLEVEL% NEQ 0 (
    echo.
    echo [错误] 启动失败！
    echo.
    pause
    exit /b 1
)

pause
