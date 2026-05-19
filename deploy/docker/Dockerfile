# SPDX-License-Identifier: MIT
# Copyright (c) 2026-present 李星云 (lxy911)
# Project: Blogo — https://github.com/zhian9/Blogo
#
# 多阶段构建 Dockerfile
# 用法:
#   docker build -t blogo-server:latest -f deploy/docker/Dockerfile blogo-server
#   docker run -d --name blogo-server -p 8040:8040 --env-file blogo-server/.env blogo-server:latest

# ═══════════════════════════════════════════════════════
# Stage 1: 构建
# ═══════════════════════════════════════════════════════
FROM golang:1.25-alpine AS builder

ARG APP=Blogo
ARG VERSION=v1.0.0

RUN apk add --no-cache gcc musl-dev

ENV CGO_CFLAGS="-D_LARGEFILE64_SOURCE"
ENV GOPROXY=https://goproxy.cn,direct

WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=1 go build -trimpath \
    -ldflags "-w -s -X main.VERSION=${VERSION}" \
    -o /bin/${APP} .

# ═══════════════════════════════════════════════════════
# Stage 2: 运行
# ═══════════════════════════════════════════════════════
FROM alpine:3.21

RUN apk --no-cache add ca-certificates tzdata \
    && rm -rf /var/cache/apk/*

ENV TZ=Asia/Shanghai

RUN adduser -D -s /bin/sh appuser

WORKDIR /app

COPY --from=builder /bin/${APP} /usr/local/bin/blogo-server
COPY --from=builder /src/configs ./configs

USER appuser

EXPOSE 8040

HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget -qO- http://localhost:8040/health || exit 1

ENTRYPOINT ["blogo-server", "start", "-d", "./configs", "-c", "dev"]
