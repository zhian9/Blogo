@echo off
REM Blogo — 开发环境一键启动 (Windows)
REM 用法: scripts\dev.bat
setlocal

echo ============================================
echo    Blogo — 开发环境一键启动
echo    后端: http://localhost:8040
echo    前台: http://localhost:5173
echo    后台: http://localhost:5174
echo ============================================

echo.
echo [1/3] 启动 Go 后端...
start "Blogo-Server" cmd /c "cd blogo-server && air"

echo [2/3] 启动用户前台...
start "Blogo-Web" cmd /c "cd blog-web && npm run dev"

echo [3/3] 启动管理后台...
start "Blogo-Admin" cmd /c "cd blogo-admin && npm run dev"

echo.
echo 所有服务已在新窗口中启动。
echo 关闭各窗口即可停止对应服务。

endlocal
