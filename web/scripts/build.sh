#!/usr/bin/env bash
# React 前端编译脚本
# 执行 npm run build 编译前端到 dist/ 目录
#
# 用法:
#   bash scripts/build.sh              # 使用默认版本 (latest)
#   bash scripts/build.sh v1.0.0       # 指定版本号

set -euo pipefail

DEFAULT_VERSION="latest"
SERVICE_NAME="web"

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"

VERSION="${1:-$DEFAULT_VERSION}"

echo "=================================================="
echo "${SERVICE_NAME} 前端编译"
echo "=================================================="
echo "版本: ${VERSION}"

# 检查 Node.js
if ! command -v node &> /dev/null; then
    echo ""
    echo "错误: 未找到 Node.js"
    exit 1
fi

# 检查 npm
if ! command -v npm &> /dev/null; then
    echo ""
    echo "错误: 未找到 npm"
    exit 1
fi

echo ""
echo "安装依赖..."
npm install
echo "依赖安装完成"

echo ""
echo "编译前端 (版本: ${VERSION})..."
VITE_APP_VERSION="$VERSION" npm run build

# 写入版本标记
DIST_DIR="$PROJECT_ROOT/dist"
VERSION_FILE="$DIST_DIR/version.txt"
mkdir -p "$DIST_DIR"
echo "$VERSION" > "$VERSION_FILE"

echo "编译成功! 输出目录: ${DIST_DIR}"

echo ""
echo "=================================================="
echo "编译完成!"
echo "=================================================="
