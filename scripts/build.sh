#!/usr/bin/env bash
# Blogo — 生产构建脚本 (Linux/macOS)
# 用法: bash scripts/build.sh
set -e

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT"

VERSION="${1:-v1.0.0}"
echo "=== Blogo 生产构建 (${VERSION}) ==="

# Go 后端
echo "[1/3] 构建 Go 后端..."
cd blogo-server
go build -trimpath -ldflags "-w -s -X main.VERSION=${VERSION}" -o blogo-server .
cd ..

# 用户前台
echo "[2/3] 构建用户前台..."
cd blog-web
npm ci && npm run build
cd ..

# 管理后台
echo "[3/3] 构建管理后台..."
cd blogo-admin
npm ci && npm run build
cd ..

echo "=== 构建完成 ==="
echo "后端:   blogo-server/blogo-server"
echo "前台:   blog-web/dist/"
echo "后台:   blogo-admin/dist/"
