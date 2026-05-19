# ═══════════════════════════════════════════════════════
# Blogo — Monorepo 工程化 Makefile
# ═══════════════════════════════════════════════════════
# 用法:
#   make dev          启动全部开发服务（后端 + 前台 + 后台）
#   make server       仅启动 Go 后端 (Air 热重载)
#   make web          仅启动用户前台 (Vite HMR)
#   make admin        仅启动管理后台 (Vite HMR)
#   make build        构建全部
#   make docker       构建 Docker 镜像
#   make clean        清理所有构建产物
# ═══════════════════════════════════════════════════════

.PHONY: dev server web admin build build-server build-web build-admin docker clean help

# ── 路径定义 ──────────────────────────────────────────
SERVER_DIR  := blogo-server
WEB_DIR     := blogo-web
ADMIN_DIR   := blogo-admin
DOCKER_DIR  := deploy/docker

# ── 环境检测 ──────────────────────────────────────────
# Windows 下 npm 用 cmd /c 启动，避免路径问题
NPM_RUN     := npm run
ifeq ($(OS),Windows_NT)
	NPM_RUN := cmd /c npm run
endif

# ═══════════════════════════════════════════════════════
# 开发服务
# ═══════════════════════════════════════════════════════

# 启动全部服务（并行）
dev:
	@echo "=== 启动全部开发服务 ==="
	@echo "后端:    http://localhost:8040"
	@echo "前台:    http://localhost:5173"
	@echo "后台:    http://localhost:5174"
	@echo ""
	cd $(SERVER_DIR) && air & \
	cd $(WEB_DIR) && npm run dev & \
	cd $(ADMIN_DIR) && npm run dev & \
	wait

# 仅后端（Go + Air 热重载）
server:
	@echo "=== 启动 Go 后端 (Air) ==="
	cd $(SERVER_DIR) && air

# 仅用户前台（React + Vite）
web:
	@echo "=== 启动用户前台 (Vite) ==="
	cd $(WEB_DIR) && $(NPM_RUN) dev

# 仅管理后台（React + Vite）
admin:
	@echo "=== 启动管理后台 (Vite) ==="
	cd $(ADMIN_DIR) && $(NPM_RUN) dev

# ═══════════════════════════════════════════════════════
# 构建
# ═══════════════════════════════════════════════════════

build: build-server build-web build-admin
	@echo "=== 全部构建完成 ==="

build-server:
	@echo "=== 构建 Go 后端 ==="
	cd $(SERVER_DIR) && $(MAKE) build

build-web:
	@echo "=== 构建用户前台 ==="
	cd $(WEB_DIR) && npm run build

build-admin:
	@echo "=== 构建管理后台 ==="
	cd $(ADMIN_DIR) && npm run build

# ═══════════════════════════════════════════════════════
# Docker
# ═══════════════════════════════════════════════════════

docker:
	@echo "=== 构建 Docker 镜像 ==="
	docker build -t blogo-server:latest -f $(DOCKER_DIR)/Dockerfile $(SERVER_DIR)

# ═══════════════════════════════════════════════════════
# 代码生成
# ═══════════════════════════════════════════════════════

wire:
	@echo "=== 生成 Wire DI 代码 ==="
	cd $(SERVER_DIR) && $(MAKE) wire

swagger:
	@echo "=== 生成 Swagger 文档 ==="
	cd $(SERVER_DIR) && $(MAKE) swagger

api-client:
	@echo "=== 生成 TypeScript API 客户端 ==="
	bash scripts/generate-api-client.sh all

# ═══════════════════════════════════════════════════════
# 清理
# ═══════════════════════════════════════════════════════

clean:
	@echo "=== 清理构建产物 ==="
	cd $(SERVER_DIR) && $(MAKE) clean
	rm -rf $(WEB_DIR)/dist
	rm -rf $(ADMIN_DIR)/dist
	find logs -name '*.log' -delete 2>/dev/null || true
	rm -rf storage/temp/*
	@echo "=== 清理完成 ==="

# ═══════════════════════════════════════════════════════
# 帮助
# ═══════════════════════════════════════════════════════

help:
	@echo "Blogo 工程化命令:"
	@echo ""
	@echo "  开发:"
	@echo "    make dev          启动全部开发服务"
	@echo "    make server       仅启动 Go 后端"
	@echo "    make web          仅启动用户前台"
	@echo "    make admin        仅启动管理后台"
	@echo ""
	@echo "  构建:"
	@echo "    make build        构建全部"
	@echo "    make build-server 构建 Go 后端"
	@echo "    make build-web    构建用户前台"
	@echo "    make build-admin  构建管理后台"
	@echo ""
	@echo "  Docker:"
	@echo "    make docker       构建 Docker 镜像"
	@echo ""
	@echo "  代码生成:"
	@echo "    make wire         生成 Wire DI 代码"
	@echo "    make swagger      生成 Swagger / OpenAPI 文档"
	@echo "    make api-client   生成 TypeScript API 客户端"
	@echo ""
	@echo "  工具:"
	@echo "    make clean        清理所有构建产物"
	@echo "    make help         显示此帮助"
