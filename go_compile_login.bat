@echo off
echo 构建 Linux 版本游戏服务器...

cd /d "%~dp0login"

set CGO_ENABLED=0
set GOOS=linux
set GOARCH=amd64

go build -o "%~dp0bin/server/login_server" main.go

echo 构建完成！文件已输出到 g7/bin/server/login_server