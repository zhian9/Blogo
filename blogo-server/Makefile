# ═══════════════════════════════════════════════════════
# blogo-server Makefile — 在 blogo-server/ 目录内使用
#
# 常用:
#   make start    开发运行 (go run)
#   make build    编译二进制
#   make serve    编译 + 前台运行
#   make clean    清理
#
# 项目级命令请从根目录运行: cd .. && make server
# ═══════════════════════════════════════════════════════

.PHONY: start build build-linux wire swagger openapi clean serve serve-d stop

# ========== 构建参数 ==========
APP             := goAdmin
SERVER_BIN      := $(APP)

GIT_DESCRIBE    := $(shell git describe --tags --always --dirty 2>/dev/null || echo v1.0.0)
GIT_HASH        := $(shell git rev-parse --short HEAD 2>/dev/null || echo unknown)

RELEASE_VERSION := v1.0.0
RELEASE_TAG     := $(GIT_DESCRIBE)

LDFLAGS         := -w -s \
	-X 'main.VERSION=$(RELEASE_TAG)' \
	-X 'main.GIT_HASH=$(GIT_HASH)' \
	-X 'main.BUILD_TIME=$(shell date -u '+%Y-%m-%d %H:%M UTC')' \
	-X 'github.com/zhian9/blogo-server/cmd.gitHash=$(GIT_HASH)' \
	-X 'github.com/zhian9/blogo-server/cmd.buildTime=$(shell date -u '+%Y-%m-%d %H:%M UTC')'

# ========== 运行配置 ==========
CONFIG_DIR      := ../configs/server
CONFIG_FILES    := dev
STATIC_DIR      := ../blogo-admin/dist
START_ARGS      := -d $(CONFIG_DIR) -c $(CONFIG_FILES) -s $(STATIC_DIR)

# ========== 默认目标 ==========
all: start

# ========== 开发与构建目标 ==========

start:
	go run -ldflags "$(LDFLAGS)" main.go start $(START_ARGS)

build:
	go build -ldflags "$(LDFLAGS)" -o $(SERVER_BIN)

build-linux:
	CGO_ENABLED=1 GOOS=linux GOARCH=amd64 \
	CC="zig cc -target x86_64-linux-musl" \
	CXX="zig c++ -target x86_64-linux-musl" \
	CGO_CFLAGS="-D_LARGEFILE64_SOURCE" \
	go build -ldflags "-extldflags '-static' $(LDFLAGS)" -o $(SERVER_BIN)_linux_amd64.static

# ========== 代码生成 ==========

wire:
	wire gen ./internal/wirex

swagger:
	swag init --parseDependency --generalInfo ./main.go --output ../docs/api

openapi:
	rm -rf ../docs/api/openapi
	docker run --rm -v ${PWD}:/local openapitools/openapi-generator-cli generate \
		-i /local/../docs/api/swagger.yaml \
		-g openapi \
		-o /local/../docs/api/openapi
	@echo "OpenAPI v3 规范已生成: ../docs/api/openapi/openapi.yaml"

# ========== 清理与运行 ==========

clean:
	rm -rf data $(SERVER_BIN) $(SERVER_BIN)_linux_amd64.static

serve: build
	./$(SERVER_BIN) start $(START_ARGS)

serve-d: build
	./$(SERVER_BIN) start $(START_ARGS) --daemon

stop:
	./$(SERVER_BIN) stop
