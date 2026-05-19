# Blogo

轻量级 SaaS 博客内容平台 — Go 后端 + React 前后台分离。

[![Go](https://img.shields.io/badge/Go-1.25-00ADD8?logo=go)](https://go.dev)
[![React](https://img.shields.io/badge/React-19-61DAFB?logo=react)](https://react.dev)
[![License](https://img.shields.io/badge/license-MIT-blue)](LICENSE)

---

## 功能

- **博客 CMS** — 文章 / 分类 / 标签 / 页面 / 友链 / 评论审核
- **用户系统** — 注册 / 登录 / 邮箱验证 / 个人中心 / 关注
- **RBAC 权限** — 用户 → 角色 → 菜单 → API（Casbin 策略引擎）
- **可见性控制** — 公开 / 私密 / 部分可见（指定用户）
- **贡献日历** — GitHub 风格的 Contribution Heatmap
- **统计面板** — 文章数 / 阅读量 / 评论 / 用户增长
- **操作审计** — 全链路操作日志 + 数据库日志
- **Prometheus 监控** — HTTP 指标 + 系统指标 + Grafana 仪表盘
- **R2 对象存储** — Cloudflare R2 图片上传（可选，回退本地）
- **SEO 友好** — Sitemap / Robots / SSR 就绪

## 技术栈

| 层 | 技术 |
|---|---|
| **后端** | Gin, GORM 2.0, Casbin 2.0, Google Wire (DI) |
| **数据库** | MySQL 8.0 (读写分离) + Redis 7 |
| **前台** | React 19, TypeScript, Vite, Ant Design 6, TailwindCSS, Zustand |
| **后台** | React 19, TypeScript, Vite, Ant Design 6, Redux Toolkit, ECharts |
| **监控** | Prometheus + Grafana (预置仪表盘) |
| **部署** | Docker / Docker Compose / Nginx 反向代理 |

## 项目结构

```
Blogo/
├── blogo-server/           Go 后端 API 服务
│   ├── cmd/                启动命令 (start/stop/version)
│   ├── configs/            配置文件 (dev/prod 环境)
│   │   ├── dev/            开发环境
│   │   └── prod/           生产环境
│   ├── internal/           内部实现
│   │   ├── bootstrap/      启动引导 (HTTP/日志/中间件)
│   │   ├── config/         配置加载 (Viper + 环境变量)
│   │   ├── mods/           业务模块
│   │   │   ├── blog/       博客 CMS (API → BIZ → DAL → Schema)
│   │   │   └── rbac/       用户 RBAC (登录/角色/菜单/权限)
│   │   ├── utility/prom/   Prometheus 监控
│   │   └── wirex/          Wire 依赖注入
│   ├── pkg/                可复用公共库
│   ├── docs/               Swagger API 文档
│   ├── migrations/         数据库迁移 SQL
│   └── main.go             入口
│
├── blogo-admin/            React 管理后台 (Ant Design 6)
│   └── src/pages/          仪表盘 / 文章 / 评论 / 用户 / 角色 / 菜单 / 设置
│
├── blog-web/               React 用户前台 (TailwindCSS + Ant Design 6)
│   └── src/pages/          首页 / 文章 / 归档 / 分类 / 标签 / 用户 / 发布
│
├── deploy/                 部署配置
│   ├── docker/             Dockerfile
│   ├── compose/            docker-compose (全栈 / 监控)
│   ├── nginx/              Nginx 反向代理
│   └── grafana/            Grafana 仪表盘
│
├── logs/                   日志 (运行时)
├── storage/                存储 (运行时)
├── scripts/                开发/构建脚本
├── Makefile                工程化命令入口
├── README.md
└── LICENSE
```

## 本地运行

### 前置条件

- **Go** 1.25+
- **Node.js** 22+
- **MySQL** 8.0+
- **Redis** 7+
- **Air** (可选，热重载) — `go install github.com/air-verse/air@latest`

### 1. 环境变量

```bash
cd blogo-server
cp .env.example .env
# 编辑 .env 填入数据库密码、JWT 密钥、Redis 密码等
```

关键环境变量：

| 变量 | 说明 |
|---|---|
| `DB_DSN` | MySQL 连接串 |
| `REDIS_PASSWORD` | Redis 密码 |
| `ROOT_PASSWORD_HASH` | 管理员 bcrypt 哈希 |
| `JWT_SIGNING_KEY` | JWT 签名密钥 (>=32字符) |
| `SITE_URL` | 站点 URL |
| `PROMETHEUS_USERNAME` / `PROMETHEUS_PASSWORD` | 监控 Basic Auth |

### 2. 启动开发服务

```bash
# 方式一: Makefile (推荐)
make dev

# 方式二: 分步启动
cd blogo-server && air              # 后端 :8040
cd blog-web && npm run dev          # 前台 :5173
cd blogo-admin && npm run dev       # 后台 :5174
```

### 3. 访问

| 服务 | 地址 | 默认账号 |
|---|---|---|
| API | http://localhost:8040 | — |
| Swagger | http://localhost:8040/swagger/index.html | — |
| 前台 | http://localhost:5173 | — |
| 后台 | http://localhost:5174 | admin / abc-123 |

## Docker 部署

```bash
# 1. 构建镜像
docker build -t blogo-server:latest -f deploy/docker/Dockerfile blogo-server

# 2. 创建网络
docker network create blogo-net

# 3. 启动全栈服务 (MySQL + Redis + Server + Nginx)
docker-compose -f deploy/compose/full-stack.yml up -d

# 4. 启动监控 (Prometheus + Grafana)
docker-compose -f deploy/compose/monitoring.yml up -d
```

## 常用命令

```bash
make help          # 查看所有命令
make dev           # 启动全部开发服务
make server        # 仅启动后端
make web           # 仅启动前台
make admin         # 仅启动后台
make build         # 构建全部
make docker        # 构建 Docker 镜像
make clean         # 清理构建产物
make wire          # 生成 Wire DI 代码
make swagger       # 生成 Swagger 文档
```

## 配置文件

| 文件 | 说明 |
|---|---|
| `configs/dev/server.yaml` | 开发环境主配置 |
| `configs/prod/server.yaml` | 生产环境主配置 |
| `configs/rbac_model.conf` | Casbin RBAC 模型 |
| `configs/gen_rbac_policy.csv` | 自动生成的 RBAC 策略 |
| `.env` | 环境变量（不入库） |

配置加载顺序: `struct 默认值` → `YAML 文件` → `.env 环境变量` → `代码强制覆盖`

## API 文档

启动后端后访问: http://localhost:8040/swagger/index.html

```bash
# 生成/更新 Swagger 文档
make swagger
```

## License

MIT — 详见 [LICENSE](LICENSE)
