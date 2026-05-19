#!/usr/bin/env bash
# Blogo — API Client Generator
# 从 Swagger JSON 自动生成 TypeScript 类型和 API 方法
# 用法: bash scripts/generate-api-client.sh [web|admin|all]

set -e
ROOT="$(cd "$(dirname "$0")/.." && pwd)"
SWAGGER_JSON="$ROOT/docs/api/swagger.json"

if [ ! -f "$SWAGGER_JSON" ]; then
    echo "请先生成 Swagger 文档: cd blogo-server && make swagger"
    exit 1
fi

gen_for() {
    local target=$1
    local out_dir="$ROOT/$target/src/api/generated"
    mkdir -p "$out_dir"

    echo "=== 生成 $target TypeScript API 客户端 ==="

    npx openapi-typescript "$SWAGGER_JSON" \
        --output "$out_dir/schema.ts" \
        --export-type

    echo "  → $out_dir/schema.ts"
}

case "${1:-all}" in
    web)
        gen_for "blogo-web"
        ;;
    admin)
        gen_for "blogo-admin"
        ;;
    all)
        gen_for "blogo-web"
        gen_for "blogo-admin"
        ;;
    *)
        echo "用法: $0 [web|admin|all]"
        exit 1
        ;;
esac

echo "=== 完成 ==="
echo "导入方式: import type { paths } from '@/api/generated/schema'"
