@echo off
echo 构建 Linux 版本游戏服务器...

:: 切换到 game 目录（源码目录）
cd /d "%~dp0game"

:: 设置环境变量（Windows CMD 专用）
set CGO_ENABLED=0
set GOOS=linux
set GOARCH=amd64

:: 编译 → 输出到上级 bin 文件夹
go build -o "%~dp0bin/server/game_server" main.go

echo 构建完成！文件已输出到 g7/bin/server/game_server
