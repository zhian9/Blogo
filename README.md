# Blogo

> **简体中文** | [English](README.en.md)

一个自建并已上线运行的**博客内容平台（CMS）**：Go 后端 + React 双端（管理后台 / 用户前台）。

[![Go](https://img.shields.io/badge/Go-1.25-00ADD8?logo=go)](https://go.dev)
[![React](https://img.shields.io/badge/React-19-61DAFB?logo=react)](https://react.dev)
[![MySQL](https://img.shields.io/badge/MySQL-8.0-4479A1?logo=mysql)](https://www.mysql.com)
[![License](https://img.shields.io/badge/license-MIT-blue)](LICENSE)

> 前台：<https://blogo.cloud>　后台：<https://admin.blogo.cloud>
> （自建单站点部署，非多租户）

---

## 这是什么

Blogo 是一个从零实现的博客内容平台：后端提供内容管理、用户与权限、评论审核、访问统计与操作审计等能力，前端分为**管理后台**（面向作者/管理员）与**用户前台**（面向读者）两个独立应用。

设计目标是「一个人也能长期维护的内容平台」：单进程部署、后台可配置、权限与内容可见性由数据驱动，无需改代码即可调整站点文案与运营策略。

规模参考：后端约 **2.4 万行** Go 代码、REST 路由 **150+**、数据表 **25** 张；两个前端合计约 **1.5 万行** TypeScript。

## 功能特性

**内容管理**

- **文章**：Markdown 写作与预渲染、分类 / 标签 / 置顶、草稿与发布、浏览量统计
- **项目展示**：项目卡片（封面、一句话简介、技术栈、项目特点、仓库与演示链接），面向「作品集」场景
- **配套内容**：分类、标签、独立页面、友情链接、图片资源
- **评论**：支持文章与项目、游客评论、审核机制（通过 / 驳回 / 删除）

**用户与权限**

- 注册（邮箱激活）、登录（图形验证码）、个人中心、关注 / 粉丝、收藏与点赞
- **RBAC 权限**：用户 → 角色 → 菜单 → API 资源，权限策略由数据库映射自动生成并支持热更新
- **内容可见性**：文章 / 项目支持 公开 / 私密 / 指定用户可见 三级策略

**运营与可观测**

- **访问统计**：PV、独立访客、独立 IP，控制中心以图表展示最近 1 / 7 / 30 天趋势
- **操作审计**：记录操作人、来源 IP、模块、动作、请求路径、状态码与错误详情，支持多条件检索
- **监控**：Prometheus 指标（请求量、延迟直方图）+ Grafana 仪表盘
- **站点配置**：首页文案、关于页文案与技术栈、联系邮箱等均可在后台修改，前台刷新即生效

**基础设施能力**

- 图片上传：Cloudflare R2 对象存储（未配置时自动回退本地磁盘）
- 邮件：邮箱激活、订阅与通知，异步 Mail Worker 投递
- 安全：JWT 认证（Redis 存证，支持主动失效）、接口限流（按 IP / 按用户）、异步操作审计

## 技术栈

| 层 | 技术 |
|---|---|
| **后端** | Go 1.25、Gin、GORM 2.0、Casbin 2.0、Google Wire（依赖注入）、Viper（配置）、zap（日志） |
| **存储** | MySQL 8.0、Redis 7（缓存 / 验证码 / 令牌存证 / 限流 / 访问计数） |
| **管理后台** | React 19、TypeScript、Vite、Ant Design 6、Redux Toolkit + RTK Query、ECharts |
| **用户前台** | React 19、TypeScript、Vite、Ant Design 6、TailwindCSS、Zustand、TanStack Query |
| **部署与监控** | Docker Compose、Nginx、Cloudflare、Prometheus、Grafana |

## 架构说明

**模块化单体（非微服务）**：一个 Gin 进程内划分两个业务模块 —— `rbac`（认证 / 用户 / 角色 / 菜单 / 权限）与 `blog`（内容 / 互动 / 统计），模块之间通过接口与依赖注入解耦。

**分层**：`API`（HTTP 入参出参）→ `BIZ`（业务编排）→ `DAL`（数据访问）→ `schema`（模型与表单），依赖由 Google Wire 在启动时装配（`make wire` 重新生成）。

**鉴权链路**：JWT（Redis 存证，支持登出与主动失效）→ 拦截器解析并注入身份到 `context` → Casbin 按角色逐条鉴权。业务层只从 `context` 取身份，不信任请求参数中的用户标识。

**权限策略热更新**：角色-菜单-资源的映射关系在数据库中维护，服务启动时生成 Casbin 策略文件；菜单 / 角色变更时通过 Redis 信号触发重载，配合定时比对实现免重启生效。

**性能与一致性取舍**

- 列表的可见性过滤下沉到 SQL（`公开 OR 作者本人 OR 指定可见`），「指定用户」场景批量校验，避免逐条查询产生 N+1
- 访问统计先在 Redis 做当日计数（Set 幂等去重 + TTL），再按天幂等写回统计表，避免「每次访问写库」
- 统计查询按日期升序返回并补齐缺失日期，图表时间轴连续，前端不做补洞
- 站点配置与单页内容属于后台高频修改项，前端采用「短缓存 + 窗口聚焦刷新」，保证「后台改完、前台刷新即见」

## 项目结构

```
Blogo/
├── blogo-server/               Go 后端
│   ├── cmd/                    CLI 入口（cobra：start / stop / version）
│   ├── internal/
│   │   ├── bootstrap/          启动引导（HTTP 服务、中间件装配、优雅关闭）
│   │   ├── config/             配置加载（YAML + .env 环境变量展开）
│   │   ├── mods/               业务模块
│   │   │   ├── rbac/           认证、用户、角色、菜单、操作日志
│   │   │   └── blog/           文章、项目、评论、统计、图片、订阅等
│   │   ├── utility/prom/       Prometheus 指标
│   │   └── wirex/              Wire 依赖注入（wire_gen.go 生成）
│   └── pkg/                    可复用库（cachex / jwtx / logging / middleware / ossx …）
├── blogo-admin/                React 管理后台（:5174）
├── blogo-web/                  React 用户前台（:5173）
├── configs/server/             后端配置（dev / prod / test + Casbin 模型与策略）
├── deploy/
│   ├── compose/                docker compose（全栈 / 监控）
│   ├── docker/                 Dockerfile（可编译）与 Dockerfile.nobuild（免编译）
│   ├── nginx/                  反向代理与站点配置
│   ├── prometheus/ grafana/    监控配置与仪表盘
│   └── scripts/                服务器端 deploy.sh / rollback.sh
├── docs/                       API 文档（Swagger）与项目说明
├── scripts/                    本地开发与构建脚本
└── Makefile                    工程化命令入口
```

## 快速开始

### 前置条件

- Go 1.25+
- Node.js 20+（推荐 22）
- MySQL 8.0、Redis 7
- （可选）Air 热重载：`go install github.com/air-verse/air@latest`

### 1. 准备配置

```bash
cd blogo-server
cp .env.example .env
# 按需修改：数据库连接、Redis 密码、JWT 签名密钥、管理员密码哈希等
```

### 2. 初始化数据库

只需创建空库（例如 `blogo`）并保证 `.env` 中的 `DB_DSN` 指向它。服务首次启动时会**自动建表**并初始化种子数据：管理员账号、基础角色、菜单树、默认页面与站点配置（配置项 `auto_migrate: true`）。

### 3. 启动（三个终端）

```bash
# 后端 :8040
make server              # 等价于 cd blogo-server && air

# 用户前台 :5173
make web

# 管理后台 :5174
make admin
```

> 不使用 Air 时，可用 `cd blogo-server && go run . start -d ../configs/server -c dev` 直接启动。

### 4. 访问

| 服务 | 地址 | 说明 |
|---|---|---|
| 后端 API | http://localhost:8040 | `/health` 为健康检查 |
| 用户前台 | http://localhost:5173 | Vite 已配置 `/api`、`/uploads` 代理到 8040 |
| 管理后台 | http://localhost:5174 | 账号为 `.env` 中 `ROOT_USERNAME`，口令对应 `ROOT_PASSWORD_HASH` |
| Swagger | http://localhost:8040/swagger/index.html | **默认关闭**，将配置里的 `disable_swagger` 改为 `false` 后可用 |

## 部署

线上部署在配置较低的云服务器上，**编译在本地完成、服务器只负责运行**，链路如下：

```bash
# 1. 本地交叉编译 Linux 静态二进制
cd blogo-server
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
  go build -trimpath -ldflags "-w -s -X main.VERSION=v1.0.0" -o blogo .

# 2. 上传产物（二进制 + 两个前端 dist），先传成 .new 再切换，避免半成品上线
scp blogo-server/blogo  user@server:/path/Blogo/blogo-server/blogo.new
scp -r blogo-web/dist   user@server:/path/Blogo/blogo-web/dist.new
scp -r blogo-admin/dist user@server:/path/Blogo/blogo-admin/dist.new

# 3. 服务器上执行部署脚本（自动备份 + 健康检查 + 失败自动回滚）
cd /path/Blogo
bash deploy/scripts/deploy.sh --dry-run    # 先演练
bash deploy/scripts/deploy.sh              # 正式部署
```

相关说明：

- **配置与前端静态文件都是 bind mount**：`configs/server` 挂进容器、`dist` 挂进 Nginx，因此改配置或发版**不需要重建镜像**（只有二进制变更才需要 `docker compose build`）
- 镜像使用 `deploy/docker/Dockerfile.nobuild`（只拷贝本地编译好的二进制，服务器不编译）
- **回滚**：`bash deploy/scripts/rollback.sh` 可列出备份并一键回滚（备份包含旧二进制、旧镜像 tag、前端 dist、配置与数据库 dump）
- 线上由 Nginx + Cloudflare + HTTPS 提供访问，前台 / 后台 / 监控按域名拆分，监控面板带 Basic Auth
- 前端发版后建议清理 Cloudflare 缓存（`vite build` 会替换带哈希的静态资源）

## 常用命令

| 命令 | 说明 |
|---|---|
| `make server` / `make web` / `make admin` | 分别启动后端（Air 热重载）、用户前台、管理后台 |
| `make build` | 构建后端 + 两个前端（`build-server` / `build-web` / `build-admin` 可单独执行） |
| `make docker` | 构建后端 Docker 镜像 |
| `make wire` | 重新生成 Wire 依赖注入代码 |
| `make swagger` | 生成 Swagger 文档 |
| `make clean` | 清理构建产物 |
| `make help` | 查看全部命令 |

## 配置说明

| 路径 | 说明 |
|---|---|
| `configs/server/dev/`、`prod/`、`test/` | 各环境主配置（服务、存储、中间件、日志） |
| `configs/server/rbac_model.conf` | Casbin RBAC 模型定义 |
| `configs/server/gen_rbac_policy.csv` | 由角色-菜单-资源自动生成的策略文件 |
| `blogo-server/.env` | 环境变量（**不入库**，见 `.env.example`） |

| 关键环境变量 | 说明 |
|---|---|
| `DB_DSN` | MySQL 连接串 |
| `REDIS_PASSWORD` | Redis 密码 |
| `ROOT_USERNAME` / `ROOT_PASSWORD_HASH` | 管理员账号与 bcrypt 口令哈希（每次启动会同步到数据库） |
| `JWT_SIGNING_KEY` | JWT 签名密钥（至少 32 字符） |
| `SITE_URL` / `CORS_ALLOW_ORIGINS` | 站点地址与跨域白名单 |
| `R2_*` | Cloudflare R2 对象存储（留空则使用本地存储） |

配置加载顺序：结构体默认值 → YAML 文件 → `.env` 环境变量 → 代码强制覆盖。

## 已知不足与后续计划

保持诚实，这些是目前明确知道、尚未完成的部分：

- **没有单元测试**：仅有零散的手工回归（如部署脚本中的接口冒烟检查），后续优先补权限、内容可见性、统计计数等关键路径的测试
- **通知接口权限待收紧**：`/api/v1/notifications` 目前在鉴权放行列表内且从请求参数取用户标识，需要改为从令牌身份取用户并做归属校验
- **控制中心的「响应耗时」图表尚未接入真实数据**（当前为占位数据），需要补真实的延迟采集
- **单站点部署**：当前为单租户自建部署，未做多租户隔离与计费；若产品化需要引入租户模型、租户级配置与配额体系

## 许可

MIT — 详见 [LICENSE](LICENSE)
