# ═══════════════════════════════════════════════════════
# Blogo — 生产级多阶段 Dockerfile
# ═══════════════════════════════════════════════════════
# 构建:
#   docker build -t blogo:latest -f deploy/docker/Dockerfile .
# 运行:
#   docker run -d --name blogo -p 8040:8040 --env-file blogo-server/.env blogo:latest
# ═══════════════════════════════════════════════════════

# ── Stage 1: 编译 ────────────────────────────────────
FROM golang:1.25-alpine AS builder

ARG VERSION=v1.0.0
ARG GOPROXY=https://goproxy.cn,direct

RUN apk add --no-cache git ca-certificates tzdata

WORKDIR /src

# 缓存依赖层（仅 go.mod / go.sum 变化时重建）
COPY blogo-server/go.mod blogo-server/go.sum ./
RUN go mod download

# 复制源码并编译
COPY blogo-server/ ./
RUN CGO_ENABLED=0 go build -trimpath \
    -ldflags "-w -s -X main.VERSION=${VERSION}" \
    -o /bin/blogo .

# ── Stage 2: 运行 ────────────────────────────────────
FROM alpine:3.21

RUN apk add --no-cache ca-certificates tzdata curl \
    && rm -rf /var/cache/apk/*

ENV TZ=Asia/Shanghai

# 非 root 用户
RUN addgroup -S appgroup && adduser -S appuser -G appgroup

# 创建运行时目录
RUN mkdir -p /app/configs /app/storage/uploads /app/logs \
    && chown -R appuser:appgroup /app

WORKDIR /app

# 从构建阶段复制二进制
COPY --from=builder /bin/blogo /usr/local/bin/blogo

# 从项目根复制配置
COPY configs/server/ /app/configs/

# 复制静态文件目录（如有前端产物）
# COPY blogo-admin/dist/ /app/static/

USER appuser

EXPOSE 8040

STOPSIGNAL SIGTERM

HEALTHCHECK --interval=30s --timeout=3s --start-period=10s --retries=3 \
    CMD curl -sf http://localhost:8040/health || exit 1

ENTRYPOINT ["blogo", "start", "-d", "/app/configs", "-c", "prod"]
