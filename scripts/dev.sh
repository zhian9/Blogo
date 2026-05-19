#!/usr/bin/env bash
# Blogo — 开发环境一键启动 (Linux/macOS)
# 用法: bash scripts/dev.sh
set -e

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT"

echo "╔════════════════════════════════════════════╗"
echo "║   Blogo — 开发环境一键启动                ║"
echo "║   后端: http://localhost:8040             ║"
echo "║   前台: http://localhost:5173             ║"
echo "║   后台: http://localhost:5174             ║"
echo "╚════════════════════════════════════════════╝"

# 启动后端
echo "[1/3] 启动 Go 后端..."
cd blogo-server
air &
SERVER_PID=$!
cd ..

# 等待后端就绪
sleep 2

# 启动前台
echo "[2/3] 启动用户前台..."
cd blog-web
npm run dev &
WEB_PID=$!
cd ..

# 启动后台
echo "[3/3] 启动管理后台..."
cd blogo-admin
npm run dev &
ADMIN_PID=$!
cd ..

echo ""
echo "所有服务已启动，按 Ctrl+C 停止..."

# 捕获退出信号并清理子进程
cleanup() {
    echo "正在停止所有服务..."
    kill $SERVER_PID $WEB_PID $ADMIN_PID 2>/dev/null
    wait
    echo "已停止。"
}

trap cleanup EXIT INT TERM
wait
