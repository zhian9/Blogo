#!/usr/bin/env bash
# ═══════════════════════════════════════════════════════════
# Blogo — 回滚脚本（配合 deploy.sh 使用）
#
# 用法：
#   bash deploy/scripts/rollback.sh                        # 列出所有可回滚的备份
#   bash deploy/scripts/rollback.sh 20260921-010203        # 全量回滚（后端+前端+配置）
#   bash deploy/scripts/rollback.sh 20260921-010203 --backend-only
#   bash deploy/scripts/rollback.sh 20260921-010203 --frontend-only
#   bash deploy/scripts/rollback.sh 20260921-010203 --config-only
#   bash deploy/scripts/rollback.sh 20260921-010203 --with-db --yes   # 连数据库一起恢复（危险）
#   ... 任何组合都可以加 --dry-run 先看一遍
#
# 说明：只回滚磁盘上的文件和镜像，不会自动删数据；
#       --with-db 会用备份 SQL 覆盖当前数据库，需要二次确认。
# ═══════════════════════════════════════════════════════════
set -euo pipefail

REPO_DIR="${REPO_DIR:-$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)}"
COMPOSE_FILE="${COMPOSE_FILE:-deploy/compose/full-stack.yml}"
API_SERVICE="${API_SERVICE:-blogo-api}"
NGINX_SERVICE="${NGINX_SERVICE:-nginx}"
MYSQL_CONTAINER="${MYSQL_CONTAINER:-blogo-mysql}"
HEALTH_URL="${HEALTH_URL:-http://localhost:8040/health}"
BACKUP_ROOT="${BACKUP_ROOT:-$HOME/blogo-backups}"
HEALTH_RETRIES="${HEALTH_RETRIES:-30}"
HEALTH_INTERVAL="${HEALTH_INTERVAL:-2}"

TS=""
ONLY=""
WITH_DB=0
DRY_RUN=0
ASSUME_YES=0

for arg in "$@"; do
  case "$arg" in
    --backend-only)  ONLY="backend" ;;
    --frontend-only) ONLY="frontend" ;;
    --config-only)   ONLY="config" ;;
    --with-db)       WITH_DB=1 ;;
    --dry-run)       DRY_RUN=1 ;;
    --yes|-y)        ASSUME_YES=1 ;;
    -h|--help)       sed -n '2,18p' "${BASH_SOURCE[0]}"; exit 0 ;;
    -*)
      echo "未知参数: $arg" >&2; exit 2 ;;
    *)
      [ -z "$TS" ] || { echo "只能指定一个备份时间戳" >&2; exit 2; }
      TS="$arg" ;;
  esac
done

log()  { printf '\033[1;34m[rollback]\033[0m %s\n' "$*"; }
ok()   { printf '\033[1;32m  ✓\033[0m %s\n' "$*"; }
warn() { printf '\033[1;33m[warn]\033[0m %s\n' "$*"; }
err()  { printf '\033[1;31m[error]\033[0m %s\n' "$*" >&2; }
die()  { err "$*"; exit 1; }

run() {
  if [ "$DRY_RUN" = "1" ]; then
    printf '  [dry-run] %s\n' "$*"; return 0
  fi
  "$@"
}

http_code() { curl -s -o /dev/null -w '%{http_code}' --max-time 10 "$@" 2>/dev/null || echo "000"; }

wait_healthy() {
  local i
  for i in $(seq 1 "$HEALTH_RETRIES"); do
    [ "$(http_code "$HEALTH_URL")" = "200" ] && return 0
    sleep "$HEALTH_INTERVAL"
  done
  return 1
}

# ── 不带参数：列出备份 ──
if [ -z "$TS" ]; then
  log "备份目录: $BACKUP_ROOT"
  if [ -f "$BACKUP_ROOT/latest" ]; then
    echo "  最近一次部署备份: $(cat "$BACKUP_ROOT/latest")"
  fi
  echo
  if [ -d "$BACKUP_ROOT" ]; then
    printf '%-20s %-10s %s\n' "时间戳" "大小" "内容"
    for d in $(ls -1dt "$BACKUP_ROOT"/*/ 2>/dev/null); do
      name="$(basename "$d")"
      size="$(du -sh "$d" 2>/dev/null | cut -f1)"
      items=""
      if [ -f "$d/blogo.prev" ];              then items="$items 二进制"; fi
      if [ -f "$d/blogo-web-dist.tar.gz" ];   then items="$items web"; fi
      if [ -f "$d/blogo-admin-dist.tar.gz" ]; then items="$items admin"; fi
      if [ -f "$d/middleware.yaml.prev" ];    then items="$items 配置"; fi
      if ls "$d"/*.sql >/dev/null 2>&1;       then items="$items 数据库"; fi
      printf '%-20s %-10s %s\n' "$name" "$size" "$items"
    done
  else
    echo "  （还没有任何备份）"
  fi
  echo
  echo "回滚示例: bash deploy/scripts/rollback.sh <时间戳>"
  exit 0
fi

# ── 回滚 ──
BK="$BACKUP_ROOT/$TS"
[ -d "$BK" ] || die "找不到备份目录: $BK（用不带参数的方式列出可用备份）"

cd "$REPO_DIR" || die "找不到仓库目录: $REPO_DIR"
command -v docker >/dev/null 2>&1 || die "未安装 docker"
if docker compose version >/dev/null 2>&1; then
  COMPOSE=(docker compose)
else
  COMPOSE=(docker-compose)
fi

log "回滚目标: $BK"
[ -n "$ONLY" ] && log "范围: $ONLY" || log "范围: 全部（后端 + 前端 + 配置）"

# 1. 后端
if [ -z "$ONLY" ] || [ "$ONLY" = "backend" ]; then
  if [ -f "$BK/blogo.prev" ]; then
    log "回滚后端二进制…"
    run cp -p "$BK/blogo.prev" blogo-server/blogo
    run "${COMPOSE[@]}" -f "$COMPOSE_FILE" build "$API_SERVICE"
    run "${COMPOSE[@]}" -f "$COMPOSE_FILE" up -d "$API_SERVICE"
    if [ "$DRY_RUN" = "0" ]; then
      wait_healthy && ok "后端已恢复并健康" || { docker logs --tail 60 "$API_SERVICE" || true; die "回滚后后端仍不健康，请手动排查"; }
    fi
  else
    warn "该备份里没有 blogo.prev（当时可能没部署后端），跳过"
  fi
fi

# 2. 前端
if [ -z "$ONLY" ] || [ "$ONLY" = "frontend" ]; then
  if [ -f "$BK/blogo-web-dist.tar.gz" ] || [ -f "$BK/blogo-admin-dist.tar.gz" ]; then
    log "回滚前端静态文件…"
    if [ -f "$BK/blogo-web-dist.tar.gz" ]; then
      run rm -rf blogo-web/dist
      run tar xzf "$BK/blogo-web-dist.tar.gz" -C blogo-web
      if [ "$DRY_RUN" = "0" ]; then ok "已恢复 blogo-web/dist"; fi
    fi
    if [ -f "$BK/blogo-admin-dist.tar.gz" ]; then
      run rm -rf blogo-admin/dist
      run tar xzf "$BK/blogo-admin-dist.tar.gz" -C blogo-admin
      if [ "$DRY_RUN" = "0" ]; then ok "已恢复 blogo-admin/dist"; fi
    fi
    run "${COMPOSE[@]}" -f "$COMPOSE_FILE" restart "$NGINX_SERVICE"
    warn "记得再清一次 Cloudflare 缓存"
  else
    warn "该备份里没有前端 dist，跳过"
  fi
fi

# 3. 配置
if [ -z "$ONLY" ] || [ "$ONLY" = "config" ]; then
  if [ -f "$BK/middleware.yaml.prev" ]; then
    log "回滚配置 middleware.yaml…"
    run cp -p "$BK/middleware.yaml.prev" configs/server/prod/middleware.yaml
    run "${COMPOSE[@]}" -f "$COMPOSE_FILE" restart "$API_SERVICE"
    if [ "$DRY_RUN" = "0" ]; then
      wait_healthy && ok "配置已恢复并健康" || warn "配置回滚后健康检查未通过，请查看日志"
    fi
  else
    warn "该备份里没有配置备份，跳过"
  fi
fi

# 4. 数据库（可选，危险）
if [ "$WITH_DB" = "1" ]; then
  dump="$(ls -1 "$BK"/*.sql 2>/dev/null | head -n1 || true)"
  [ -n "$dump" ] || die "该备份里没有 .sql 文件"
  warn "即将用 $dump 覆盖当前数据库！这会丢失备份之后产生的所有数据。"
  if [ "$ASSUME_YES" != "1" ] && [ "$DRY_RUN" = "0" ]; then
    printf '确认继续请输入 yes: '
    read -r answer
    [ "$answer" = "yes" ] || die "已取消"
  fi
  if [ "$DRY_RUN" = "0" ]; then
    docker exec -i "$MYSQL_CONTAINER" sh -c "mysql -uroot -p\"\$MYSQL_ROOT_PASSWORD\"" < "$dump" \
      || die "数据库恢复失败，请手动处理"
    ok "数据库已从备份恢复"
    wait_healthy && ok "服务健康" || warn "服务健康检查未通过，请看日志"
  else
    printf '  [dry-run] mysql < %s\n' "$dump"
  fi
fi

echo
log "回滚完成（备份仍保留在 $BK，可重复执行）"
