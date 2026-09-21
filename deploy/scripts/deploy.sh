#!/usr/bin/env bash
# ═══════════════════════════════════════════════════════════
# Blogo — 服务器端部署脚本
#
# 用法（在服务器上执行）：
#   bash deploy/scripts/deploy.sh --dry-run          # 只看会做什么，不做任何修改
#   bash deploy/scripts/deploy.sh                    # 部署后端 + 前端（含自动备份）
#   bash deploy/scripts/deploy.sh --with-config      # 顺带更新 configs/server 配置
#   bash deploy/scripts/deploy.sh --skip-db          # 跳过数据库备份（不建议）
#
# 前置条件：本地已把产物 scp 成下面这些名字（见 README 的部署说明）
#   blogo-server/blogo.new                  Linux 二进制
#   blogo-web/dist.new/                     前台 dist
#   blogo-admin/dist.new/                   后台 dist
#   configs/server/prod/middleware.yaml.new （仅 --with-config 时需要）
#
# 要点：
#   - 每步失败立即中止；后端健康检查不通过会自动回滚后端
#   - 所有备份写到 $HOME/blogo-backups/<时间戳>/
#   - 部署完请到 Cloudflare 控制台 Purge Everything（前端 dist 里的旧 hash 文件已被替换）
# ═══════════════════════════════════════════════════════════
set -euo pipefail

# ── 可覆盖的配置（用环境变量传入即可，例如 REPO_DIR=/srv/Blogo bash deploy.sh）──
REPO_DIR="${REPO_DIR:-$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)}"
COMPOSE_FILE="${COMPOSE_FILE:-deploy/compose/full-stack.yml}"
API_SERVICE="${API_SERVICE:-blogo-api}"
NGINX_SERVICE="${NGINX_SERVICE:-nginx}"
MYSQL_CONTAINER="${MYSQL_CONTAINER:-blogo-mysql}"
DB_NAME="${DB_NAME:-blogo}"
HEALTH_URL="${HEALTH_URL:-http://localhost:8040/health}"
BACKUP_ROOT="${BACKUP_ROOT:-$HOME/blogo-backups}"
KEEP_BACKUPS="${KEEP_BACKUPS:-5}"
HEALTH_RETRIES="${HEALTH_RETRIES:-30}"
HEALTH_INTERVAL="${HEALTH_INTERVAL:-2}"

# ── 开关 ──
WITH_CONFIG=0
SKIP_BACKEND=0
SKIP_FRONTEND=0
SKIP_DB=0
DRY_RUN=0

for arg in "$@"; do
  case "$arg" in
    --with-config)   WITH_CONFIG=1 ;;
    --skip-backend)  SKIP_BACKEND=1 ;;
    --skip-frontend) SKIP_FRONTEND=1 ;;
    --skip-db)       SKIP_DB=1 ;;
    --dry-run)       DRY_RUN=1 ;;
    -h|--help)       sed -n '2,20p' "${BASH_SOURCE[0]}"; exit 0 ;;
    *) echo "未知参数: $arg（可用: --with-config --skip-backend --skip-frontend --skip-db --dry-run）" >&2; exit 2 ;;
  esac
done

log()  { printf '\033[1;34m[deploy]\033[0m %s\n' "$*"; }
ok()   { printf '\033[1;32m  ✓\033[0m %s\n' "$*"; }
warn() { printf '\033[1;33m[warn]\033[0m %s\n' "$*"; }
err()  { printf '\033[1;31m[error]\033[0m %s\n' "$*" >&2; }
die()  { err "$*"; exit 1; }

# 统一命令执行入口：dry-run 时只打印
run() {
  if [ "$DRY_RUN" = "1" ]; then
    printf '  [dry-run] %s\n' "$*"
    return 0
  fi
  "$@"
}

http_code() { curl -s -o /dev/null -w '%{http_code}' --max-time 10 "$@" 2>/dev/null || echo "000"; }

wait_healthy() {
  local i
  for i in $(seq 1 "$HEALTH_RETRIES"); do
    if [ "$(http_code "$HEALTH_URL")" = "200" ]; then
      return 0
    fi
    sleep "$HEALTH_INTERVAL"
  done
  return 1
}

# ═══════════════ 0. 预检 ═══════════════
cd "$REPO_DIR" || die "找不到仓库目录: $REPO_DIR"
log "仓库目录: $REPO_DIR"

command -v docker >/dev/null 2>&1 || die "未安装 docker"
docker info >/dev/null 2>&1 || die "docker 守护进程不可用（需要 sudo 或检查服务状态）"

if docker compose version >/dev/null 2>&1; then
  COMPOSE=(docker compose)
elif command -v docker-compose >/dev/null 2>&1; then
  COMPOSE=(docker-compose)
else
  die "未找到 docker compose"
fi
log "compose 命令: ${COMPOSE[*]} -f $COMPOSE_FILE"

[ -f "$COMPOSE_FILE" ] || die "找不到 compose 文件: $COMPOSE_FILE"

HAVE_BACKEND=0; HAVE_FRONTEND=0; HAVE_CONFIG=0
if [ -f blogo-server/blogo.new ]; then HAVE_BACKEND=1; fi
if [ -d blogo-web/dist.new ] && [ -d blogo-admin/dist.new ]; then HAVE_FRONTEND=1; fi
if [ -f configs/server/prod/middleware.yaml.new ]; then HAVE_CONFIG=1; fi

if [ "$SKIP_BACKEND" = "1" ]; then HAVE_BACKEND=0; fi
if [ "$SKIP_FRONTEND" = "1" ]; then HAVE_FRONTEND=0; fi
if [ "$WITH_CONFIG" != "1" ]; then HAVE_CONFIG=0; fi

log "待部署: 后端=$HAVE_BACKEND 前端=$HAVE_FRONTEND 配置=$HAVE_CONFIG"
if [ "$HAVE_BACKEND" = "0" ] && [ "$HAVE_FRONTEND" = "0" ] && [ "$HAVE_CONFIG" = "0" ]; then
  die "没有找到任何待部署产物（blogo-server/blogo.new / *-web/dist.new / middleware.yaml.new），请先 scp"
fi
if [ "$HAVE_BACKEND" = "1" ]; then
  file blogo-server/blogo.new 2>/dev/null | grep -q 'ELF' || warn "blogo-server/blogo.new 看起来不是 Linux 二进制，确认一下再继续"
fi

# ═══════════════ 1. 备份 ═══════════════
TS="$(date +%Y%m%d-%H%M%S)"
BK="$BACKUP_ROOT/$TS"
log "创建备份目录: $BK"
run mkdir -p "$BK"

if [ "$HAVE_BACKEND" = "1" ]; then
  if [ -f blogo-server/blogo ]; then run cp -p blogo-server/blogo "$BK/blogo.prev"; fi
  run bash -c "docker image inspect blogo:latest >/dev/null 2>&1 && docker tag blogo:latest blogo:prev-$TS || true"
  run bash -c "docker image inspect blogo:latest >/dev/null 2>&1 && docker image inspect blogo:latest --format '{{.Id}} {{.Created}}' > '$BK/image-before.txt' || true"
fi

if [ "$HAVE_FRONTEND" = "1" ]; then
  if [ -d blogo-web/dist ];   then run tar czf "$BK/blogo-web-dist.tar.gz"   -C blogo-web dist; fi
  if [ -d blogo-admin/dist ]; then run tar czf "$BK/blogo-admin-dist.tar.gz" -C blogo-admin dist; fi
fi

if [ "$HAVE_CONFIG" = "1" ]; then
  run cp -p configs/server/prod/middleware.yaml "$BK/middleware.yaml.prev"
fi

if [ "$SKIP_DB" = "0" ]; then
  if docker ps --format '{{.Names}}' | grep -qx "$MYSQL_CONTAINER"; then
    log "备份数据库 $DB_NAME → $BK/$DB_NAME-$TS.sql"
    if [ "$DRY_RUN" = "0" ]; then
      docker exec "$MYSQL_CONTAINER" sh -c \
        "mysqldump -uroot -p\"\$MYSQL_ROOT_PASSWORD\" --single-transaction --quick --default-character-set=utf8mb4 --databases $DB_NAME" \
        > "$BK/$DB_NAME-$TS.sql" 2>"$BK/mysqldump.err" || die "mysqldump 失败，请看 $BK/mysqldump.err"
      [ -s "$BK/$DB_NAME-$TS.sql" ] || die "数据库备份文件为空，中止部署"
      ok "数据库备份完成: $(du -h "$BK/$DB_NAME-$TS.sql" | cut -f1)"
    else
      printf '  [dry-run] docker exec %s sh -c "mysqldump ... --databases %s" > %s/%s-%s.sql\n' \
        "$MYSQL_CONTAINER" "$DB_NAME" "$BK" "$DB_NAME" "$TS"
    fi
  else
    warn "容器 $MYSQL_CONTAINER 未运行，跳过数据库备份"
  fi
fi

if [ "$DRY_RUN" = "0" ]; then
  git rev-parse HEAD > "$BK/git-head.txt" 2>/dev/null || echo "unknown" > "$BK/git-head.txt"
  echo "$TS" > "$BACKUP_ROOT/latest"
  ok "备份完成: $BK"
fi

# ═══════════════ 2. 后端 ═══════════════
if [ "$HAVE_BACKEND" = "1" ]; then
  log "部署后端二进制…"
  run mv blogo-server/blogo.new blogo-server/blogo
  run "${COMPOSE[@]}" -f "$COMPOSE_FILE" build "$API_SERVICE"
  run "${COMPOSE[@]}" -f "$COMPOSE_FILE" up -d "$API_SERVICE"

  if [ "$DRY_RUN" = "0" ]; then
    log "等待健康检查: $HEALTH_URL"
    if wait_healthy; then
      ok "后端健康检查通过"
    else
      err "后端健康检查失败，开始自动回滚…"
      docker logs --tail 60 "$API_SERVICE" || true
      if [ -f "$BK/blogo.prev" ]; then
        cp -p "$BK/blogo.prev" blogo-server/blogo
        "${COMPOSE[@]}" -f "$COMPOSE_FILE" build "$API_SERVICE"
        "${COMPOSE[@]}" -f "$COMPOSE_FILE" up -d "$API_SERVICE"
        wait_healthy && ok "已回滚到旧二进制并恢复正常" || err "回滚后仍不健康，请手动排查"
      fi
      die "部署失败（后端不健康）"
    fi
  fi
fi

# ═══════════════ 3. 配置 ═══════════════
if [ "$HAVE_CONFIG" = "1" ]; then
  log "更新配置 configs/server/prod/middleware.yaml…"
  run mv configs/server/prod/middleware.yaml.new configs/server/prod/middleware.yaml
  run "${COMPOSE[@]}" -f "$COMPOSE_FILE" restart "$API_SERVICE"

  if [ "$DRY_RUN" = "0" ] && ! wait_healthy; then
    err "更新配置后健康检查失败，回滚配置…"
    cp -p "$BK/middleware.yaml.prev" configs/server/prod/middleware.yaml
    "${COMPOSE[@]}" -f "$COMPOSE_FILE" restart "$API_SERVICE"
    wait_healthy && ok "已回滚配置" || err "回滚配置后仍不健康"
    die "部署失败（配置导致后端不健康）"
  fi
fi

# ═══════════════ 4. 前端 ═══════════════
if [ "$HAVE_FRONTEND" = "1" ]; then
  log "替换前端 dist…"
  if [ -d blogo-web/dist.new ]; then
    run rm -rf blogo-web/dist
    run mv blogo-web/dist.new blogo-web/dist
  fi
  if [ -d blogo-admin/dist.new ]; then
    run rm -rf blogo-admin/dist
    run mv blogo-admin/dist.new blogo-admin/dist
  fi
  run "${COMPOSE[@]}" -f "$COMPOSE_FILE" restart "$NGINX_SERVICE"
  warn "别忘了到 Cloudflare 控制台 Purge Everything（否则用户可能拿到旧 index.html 引用已删除的旧 js）"
fi

# ═══════════════ 5. 冒烟测试 ═══════════════
if [ "$DRY_RUN" = "0" ]; then
  log "冒烟测试…"
  fail=0

  check() { # check <描述> <期望> <实际>
    if [ "$2" = "$3" ]; then ok "$1"; else err "$1（期望 $2，实际 $3）"; fail=1; fi
  }

  check "健康检查" 200 "$(http_code "$HEALTH_URL")"
  check "归档接口" 200 "$(http_code 'http://localhost:8040/api/v1/archives')"
  # bug5 回归防护：纯日期区间筛选曾经返回 400
  check "时间区间筛选(纯日期)" 200 "$(http_code 'http://localhost:8040/api/v1/articles?published_at_gte=2026-01-01&published_at_lte=2026-12-31')"
  check "标签筛选接口" 200 "$(http_code 'http://localhost:8040/api/v1/articles?tag_name=go')"
  check "访问统计上报" 200 "$(http_code -X POST 'http://localhost:8040/api/v1/statistics/visit')"

  if grep -q '"year"' <(curl -s --max-time 10 'http://localhost:8040/api/v1/archives'); then
    ok "归档返回了月份数据"
  else
    warn "归档接口没有返回月份数据（可能确实没有已发布文章）"
  fi

  if [ "$fail" = "1" ]; then
    warn "有检查项失败。后端健康则服务可用；如需整体回退执行："
    warn "  bash deploy/scripts/rollback.sh $TS"
  fi
fi

# ═══════════════ 6. 清理旧备份 ═══════════════
if [ "$DRY_RUN" = "0" ] && [ -d "$BACKUP_ROOT" ]; then
  old=$(ls -1dt "$BACKUP_ROOT"/*/ 2>/dev/null | tail -n +$((KEEP_BACKUPS + 1)) || true)
  if [ -n "$old" ]; then
    log "清理旧备份（保留最近 $KEEP_BACKUPS 份）"
    echo "$old" | while read -r d; do
      if [ -n "$d" ]; then run rm -rf "$d"; fi
    done
  fi
fi

# ═══════════════ 7. 汇总 ═══════════════
echo
log "部署完成"
echo "  备份目录   : $BK"
echo "  回滚命令   : bash deploy/scripts/rollback.sh $TS"
echo "  列出备份   : bash deploy/scripts/rollback.sh"
echo "  查看日志   : ${COMPOSE[*]} -f $COMPOSE_FILE logs --tail 100 $API_SERVICE"
echo
warn "前端有更新时：Cloudflare → Caching → Purge Everything，然后无痕窗口验证 page/blogo.cloud 与 admin.blogo.cloud"
