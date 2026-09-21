# Blogo 工作交接说明

> 用途：跨会话交接。新会话开始时让助手先读本文件，再继续未完成的工作。
> 最后更新：2026-09-21

## 一、部署方式（重要）

- 服务器内存小，**不在服务器上编译**。流程是：本地编译 → `scp` 上传 → 服务器 `docker compose`。
- 后端二进制放进 `blogo-server/blogo`，由 `deploy/docker/Dockerfile.nobuild` 打进镜像 `blogo:latest`（服务名 `blogo-api`）。
- 配置 `configs/server/` 是 bind mount，改配置只需 `restart blogo-api`，**不用重建镜像**。
- 前端静态目录也是 bind mount：`blogo-web/dist`、`blogo-admin/dist`（nginx 服务）。
- 前端替换 dist 后必须清 Cloudflare 缓存（否则旧 index.html 引用的旧 js 已被删除 → 白屏）。
- 部署/回滚脚本（已写好并做过 dry-run 验证）：
  - `bash deploy/scripts/deploy.sh --dry-run` / `--with-config` / `--skip-frontend`
  - `bash deploy/scripts/rollback.sh`（列出备份）/ `<时间戳>`（全量回滚）/ `--backend-only` 等
  - 备份默认在 `$HOME/blogo-backups/<时间戳>/`（含旧二进制、旧镜像 tag、前后端 dist tar、配置、数据库 dump）

## 二、本地开发环境

```bash
# 后端 :8040（连本地 MySQL 3306 / Redis 6379，blogo-server/.env 指向本地库）
cd blogo-server && go run . start -d ../configs/server -c dev
cd blogo-web   && npm run dev      # :5173
cd blogo-admin && npm run dev      # :5174
```

- 后台账号：`admin`，口令见 `blogo-server/.env` 的 `ROOT_PASSWORD_HASH`（**每次启动都会把数据库里的 root 账号同步成该哈希**，所以在后台改密码会被重置为它）。
- 验证码答案存在 Redis：`redis-cli -n 1 GET "captcha:<captcha_id>"`（自动化测试用）。
- 本地库当前数据：1 篇文章、1 个标签「测试」（被该文章引用）、若干操作日志。

## 三、已完成并**本地实测通过**的工作

### 1. 图片接口越权修复（配置）
`configs/server/prod/middleware.yaml` 与 `dev/middleware.yaml`：不再跳过整个 `/api/v1/images/` 前缀，只保留 `"/api/v1/images/upload"` 自助上传；其余图片增删改走 Auth + Casbin。
- 实测：匿名 DELETE/PUT/批量删除/上传 → 401；普通用户删除 → 403；普通用户上传 → 200；公开读取（category、{id}/file）不受影响。

### 2. 标签（后台 b1 方案）
- 标签列表返回 `article_count`（引用文章数）；新增 `GET /api/v1/tags/{id}/references` 返回引用该标签的文章列表。
- 后台标签管理页：新增「引用」列（`N 篇` 可点 / `未引用`）、引用详情弹窗（标题可点跳文章编辑）、**删除失败会显示后端原因**（此前是 `catch {}` 静默吞掉）。
- 保留原策略：被文章引用时拒绝删除并返回「该标签被 N 篇文章引用，请先解除关联后再删除」。
- 实测：klist 引用数=1、references total=1、删除被引用标签 → 400 且原因正确。
- 已知边界：只统计**文章**引用，未统计 `project_tag`（标签仅被项目引用时仍可删，会留悬空关联）。

### 3. 首页文案 & 「关于」按钮后台可改
- 后端 `initDefaultSettings` 改为**按 key 幂等补齐**（原来是"设置表为空才初始化"，导致老站点永远拿不到新增配置项）。
- 新增设置项：`hero_title`、`hero_subtitle`（用 `|` 分行）、`about_button_label`。
- 前台 `HeroSection.tsx` 用 `useSettings()` 读取这三项，缺省回退到原硬编码文案。
- 种子页面标题与正文首行：`关于我` → `关于作者`。
- 实测：重启后三个 key 自动补齐；PUT 保存 → 200；公开接口读到新值；已改回默认值。
- **踩坑记录**：`GET /settings/all` 原先注册在受保护路由组里，靠各环境 auth 跳过列表放行；
  dev 配置里没有放行它 → 前台匿名请求 401 → `useSettings()` 取不到值 → 首页永远显示默认文案
  （表现为"后台改成功了、前台刷新没变化"）。现已把它**移到公开路由组**（`RegisterV1PublicRouters`），
  实测匿名可读，本地与线上表现一致。

### 3.1 关于页（/about）可配置
- 关于页原先只有标题和正文来自数据库，另外三处是硬编码：副标题、三张技术栈卡片、GitHub 链接。
- 现在全部来自后台「系统设置」：
  - `about_subtitle`：关于页副标题
  - `about_tech_stack`：技术栈卡片，格式 `标题|描述`，多张用 `;;` 分隔（图标按顺序轮换）
  - `github_url`：复用已有配置项
- 标题（`page.title`）与正文（`page.content`）走 后台 → 页面管理 → about。
- 实测：两个新 key 自动补齐、匿名可读、后台 PUT 200。
- 观感优化（用户反馈"突兀"）：删除了标题区下方那条 60px 渐变分隔线；内容为空时不再渲染空卡片；
  Email 按钮原先硬编码假地址 `admin@blogo.dev`，改为读 `contact_email`（留空则不显示该按钮）。
- 种子页面的关于内容已从占位文案改为正式的「关于作者」文案；本地已有站点也已同步更新为正式内容。
- 关于页正文改为**双入口**：`about_content`（系统设置，优先级高）→ 留空则回退到「页面管理 → about」的正文。
  实测：PUT 200、公开接口可读、清空（空值）也返回 200。
- 布局调整：社交按钮（GitHub/Email）从独立一行移入标题区（避免只剩一个按钮时空荡），并收紧竖向间距、
  去掉内容卡片多余的 marginTop。
- 设置项的值**允许留空**（`SettingForm.Value` 去掉 `binding:"required"`，否则清空邮箱/正文会 400）；
  后台设置编辑弹窗加宽到 760、Value 改为自适应多行文本框。

### 3.3 项目功能改造成「展示型」（本轮完成）

定位：项目 = 展示自己的作品（特点 + 技术栈 + 链接），不是内容管理。

- **后端**：`Project` 新增 `Highlights string`（text，每行一条特点）；`ProjectForm` 增加同名字段并**放开 `Content` 必填**。
  技术栈**复用现有 tags**（模型注释原本就是"技术栈标签"），无新表。生产 `auto_migrate: true` → 启动自动加列。
- **前台发布页 `PublishProject.tsx`**：整页重写为展示型（封面 / 名称 / Slug / 一句话简介 / 分类 / 项目状态 /
  技术栈（可选已有标签或输入新技术）/ 项目特点（动态多行）/ 仓库与演示链接 / 存草稿与发布）。
  去掉了可见性（含可见用户选择）、置顶、SEO 表单与 Markdown 正文编辑器；SEO 由标题、摘要、技术栈自动生成。
- **前台详情页 `ProjectDetail.tsx`**：整页重写为极简展示（封面 / 标题+状态 / 作者·日期·浏览量 / 简介 /
  技术栈标签 / 仓库与演示按钮 / 项目特点列表 / 作者本人的编辑入口）。
  移除评论区、项目历程、项目资源、点赞、收藏、目录与正文渲染。
  → 顺带消除了两个已知 bug：项目评论走错接口、`comment_count` 从不维护（相关页面已不再使用）。
- **后台**：删除 `ProjectEdit.tsx` 与 `/projects/new`、`/projects/:id` 路由；列表页去掉"新增项目"，
  行操作把"编辑"改成"查看"（打开前台详情页），保留批量发布/草稿/删除与置顶、精选。
- 验证：`go build` ✔、前端与后台 `npm run build` ✔；端到端创建测试项目（3 条特点 + 仓库地址）→ 列表与详情接口均正确返回。

**收尾已完成**：
- 前台项目卡片把"❤ 赞数"换成「仓库 / 演示」入口（`<span onClick>` + `stopPropagation`，避免嵌套在卡片 Link 里产生非法嵌套 `<a>`）；
- 删除无人使用的 `useProjects.ts` 中 like/favorite/timeline/resources 共 8 个 hooks，以及 `api/projects.ts` 里对应的 9 个接口函数；
- 删除后台 `store/api.ts` 中已无人使用的 createProject/updateProject 两个 mutation 及其导出；
- 删除后台 `store/api.ts` 中 timeline / resources 的 8 个端点（连同 tagTypes 里的 `ProjectTimeline` / `ProjectResources`），
  以及两个前端的 `ProjectTimeline` / `ProjectResource` / `ProjectLikeCountResult` 类型和 `Project.like_count` / `favorite_count` 字段；
- 前台列表页去掉"最多点赞"排序项与对应排序分支；首页精选项目卡片去掉点赞数（保留浏览量 + 仓库/演示图标）；
- 删除验证用的测试项目 `codex-demo-project`，本地项目表恢复为空（0 行）。
- 前端与后台 `npm run build` 均通过。

> 残留检查：两个前端里 `ProjectTimeline` / `ProjectResource` / `ProjectLikeCountResult` / `like_count` / `favorite_count` / `most_liked`
> 出现次数均为 0。后端仍保留 timeline / resources 接口与数据表（未删），只是前端不再使用。

### 3.2 后台"保存成功但实际没保存"的通用问题
RTK Query 的 mutation **默认不抛错**，必须 `.unwrap()` 才会进入 catch；此前多个页面写成
`await update(...)` → 失败也提示成功。已修复：`PageEdit`、`PageList`、`CategoryManage`、
`CommentManage`、`TagManage`（错误提示统一为 `err.data.error.detail`）。

### 4. 顺带修掉的两个真 bug
- `setting.dal.go` 有 **6 处**把 MySQL 保留字 `key` 裸写进 SQL（`Where("key = ?")`）→ `Error 1064`，导致**后台"系统设置"保存从来没成功过**。已全部加反引号。
- `PUT /settings/{key}` 曾要求请求体带 `key`，而后台表单不提交 key → 400。现改为以 URL 路径为准（`SettingForm.Key` 去掉 `required`）。

### 5. 其他已完成的改动（早前会话）
- Blogo 死代码清理：删除 6 个未引用组件/hook、若干死函数与死类型、移除 6 个未使用 npm 依赖（注意：`tslib` 必须保留，`echarts-for-react` 隐式依赖它）。
- 后台 4 个日志页（操作/登录/安全/审计）接入 `OperationLogTable`，替换原「功能开发中」占位页。
- 用户管理页角色显示修复：后端返回扁平 `role_code`，前端原先读嵌套 `role.code` → 角色显示不出来。
- 后端 bug：标签 ID 被当标签名（会创建以 ID 为名的垃圾标签）；时间筛选参数只认 RFC3339（前端传 `YYYY-MM-DD` → 400）。两者均已修复并提供 `resolveTagIDs`、`util.ParseFlexibleTime`。
- 访问量统计：新增 `POST /api/v1/statistics/visit`（前台每次路由切换上报），Redis 计数 + 按天 upsert 到 `statistics` 表；控制中心「访问量走势」因此有数据。

## 四、未完成 / 待决策

1. **尚未部署**：本轮（及之前的）改动需要上传 后端二进制 + `blogo-web/dist` + `blogo-admin/dist`；
   生产 `auto_migrate: true`，启动时会自动给 `project` 表加 `highlights` 列。部署用 `deploy/scripts/deploy.sh`（先 `--dry-run`），完成后清 Cloudflare 缓存。
2. **尚未提交**：工作区所有改动仍是未提交状态（且混有用户自己的改动），提交前不要 `git checkout .` / `git reset --hard`。
3. **通知接口存在越权风险（未修）**：`/api/v1/notifications` 同时位于 auth 与 casbin 跳过列表，且接口从请求参数取 `user_id`，
   匿名即可读写他人通知。修法：移出两个跳过列表 + 改为从上下文取用户 + 按通知 id 做归属校验（需用户确认后再动）。
4. **设置页不能新增 key**：后台「系统设置」只能编辑已有项，新增靠后端 seed 补齐。可按需加"新增设置"按钮。
5. **线上关于页标题**：本地 `page(slug=about)` 已改为"关于作者"，线上仍是旧值（种子只对新站生效）→ 部署后在 后台 → 页面管理 改标题即可。
6. **标签引用未含项目**：当前只统计文章引用（`article_tag`），未纳入 `project_tag`。
7. **响应耗时图**：控制中心「响应耗时走势」仍是 `Math.random()` 假数据，待决定：删除该 tab / 标注示例 / 真实采集（后者需给 `statistics` 加字段）。
8. **上传接口无文件时返回 500**（应为 400）；`blogo-web/src/components/MarkdownRenderer.tsx` 用了未声明的 `highlight.js` 传递依赖。
9. **nginx `index.html` 未加 no-cache**（可选）：加 `Cache-Control: no-cache` 可避免发版后"旧 HTML 引用已删除 JS"的白屏。

## 五、工作区状态与注意事项

- 所有改动**尚未提交**（工作区 dirty），且混有用户自己的改动（如 `deploy/compose/full-stack.yml` 的 healthcheck 调整）。
  → 回退前先确认，不要直接 `git checkout .` / `git reset --hard`。
- 新增未跟踪文件：`deploy/scripts/deploy.sh`、`deploy/scripts/rollback.sh`、`blogo-server/internal/mods/blog/biz/tag_resolve.go`、`blogo-server/pkg/util/datetime.go`、`blogo-admin/src/components/OperationLogTable.tsx`、`docs/HANDOVER.md`。
- 后端 `go build ./...`、前台/后台 `npm run build` 均通过；改动过注释清理（删除 220 行纯步骤复述注释，未动代码逻辑）。

## 六、建议的下一步

1. 把本次改动按主题提交（配置修复 / 标签 / 设置与首页文案 / 死代码清理），便于回滚与追溯。
2. 按第四节决策项逐条推进。
3. 部署时用 `deploy/scripts/deploy.sh`，先 `--dry-run` 再正式执行；上线后清 Cloudflare 缓存。
