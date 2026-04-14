@echo off

set "SCRIPT_DIR=%~dp0"
cd /d "%SCRIPT_DIR%"

call "%SCRIPT_DIR%go_compile_game.bat"

call "%SCRIPT_DIR%go_compile_gateway.bat"

call "%SCRIPT_DIR%go_compile_login.bat"