# Copyright 2025 lxy911(李星云) <lxyaa911@gmail.com>. All rights reserved.
# Use of this source code is governed by an MIT-style license
# that can be found in the LICENSE file.
# Project repository: https://github.com/zhian9/go-admin

# 构建阶段
FROM golang:alpine AS builder

ARG APP=goAdmin
ARG VERSION=v1.0.0

# 安装编译依赖
RUN apk add --no-cache gcc musl-dev sqlite-dev

# 启用大文件支持
ENV CGO_CFLAGS="-D_LARGEFILE64_SOURCE"
# 设置 Go 代理（无多余空格！）
ENV GOPROXY=https://goproxy.cn,direct

WORKDIR /src
COPY . .

# 构建二进制
RUN go build -trimpath -ldflags "-w -s -X main.VERSION=${VERSION}" -o /bin/${APP} .

# 运行阶段
FROM alpine:latest

# 安装运行依赖
RUN apk --no-cache add ca-certificates tzdata \
    && rm -rf /var/cache/apk/*

# 设置时区
ENV TZ=Asia/Shanghai

# 创建非 root 用户
RUN adduser -D -s /bin/sh appuser

WORKDIR /app

# 复制二进制
COPY --from=builder /bin/goAdmin /usr/local/bin/goAdmin

# 复制配置文件、数据目录和静态资源（根据你的项目结构）
COPY --from=builder /src/configs ./configs
COPY --from=builder /src/data ./data
# 如果有前端构建产物，请取消注释以下行
# COPY --from=builder /src/dist ./dist

# 切换到非 root 用户
USER appuser

EXPOSE 8040

# 启动命令
ENTRYPOINT ["goAdmin", "start", "-d", "./configs", "-c", "prod", "-s", "./dist"]