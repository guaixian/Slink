#!/bin/bash
# Linux/Mac 编译脚本 - 启用 CGO 支持 SQLite

echo "正在编译 Slink (启用 CGO)..."

# 设置 CGO 环境变量
export CGO_ENABLED=1

# 编译
go build -o slink main.go

if [ $? -eq 0 ]; then
    echo ""
    echo "编译成功！"
    echo "可执行文件: slink"
    echo ""
    echo "运行程序: ./slink"
else
    echo ""
    echo "编译失败！"
    echo "请确保已安装 GCC 编译器"
fi
