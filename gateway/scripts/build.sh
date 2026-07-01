#!/usr/bin/env bash
# 跨平台编译脚本
# 将 Go 代码编译为 Linux 平台的二进制文件，输出到 bin/ 目录
#
# 用法:
#   bash scripts/build.sh              # 使用默认版本 (latest)，默认架构 (amd64)
#   bash scripts/build.sh v1.0.0       # 指定版本号，默认架构 (amd64)
#   bash scripts/build.sh --arm64      # 使用 arm64 架构
#   bash scripts/build.sh v1.0.0 --arm64  # 指定版本号和架构
#   bash scripts/build.sh --help       # 显示帮助

set -euo pipefail

SERVICE_NAME="gateway"
DEFAULT_VERSION="latest"
DEFAULT_ARCH="amd64"

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
BIN_DIR="$PROJECT_ROOT/bin"

VERSION="$DEFAULT_VERSION"
ARCH="$DEFAULT_ARCH"
CLEAN=false

usage() {
    cat << 'HELP'
gateway 跨平台编译脚本

用法:
  bash scripts/build.sh              # 使用默认版本 latest，默认架构 amd64
  bash scripts/build.sh v1.0.0       # 指定版本号 v1.0.0，默认架构 amd64
  bash scripts/build.sh --arm64      # 使用默认版本 latest，架构 arm64
  bash scripts/build.sh v1.0.0 --arm64  # 指定版本号和架构

选项:
  --arm64     编译 arm64 架构 (默认编译 amd64)
  --clean     编译前清理 bin 目录
  --help      显示帮助信息
HELP
    exit 0
}

# 解析参数
while [[ $# -gt 0 ]]; do
    case "$1" in
        --arm64)
            ARCH="arm64"
            shift
            ;;
        --clean)
            CLEAN=true
            shift
            ;;
        --help)
            usage
            ;;
        --*)
            echo "错误: 未知选项 $1"
            exit 1
            ;;
        *)
            VERSION="$1"
            shift
            ;;
    esac
done

BUILD_TIME=$(date +"%Y-%m-%d_%H:%M:%S")

echo "=================================================="
echo "${SERVICE_NAME} 跨平台编译"
echo "=================================================="
echo "版本: ${VERSION}"
echo "架构: ${ARCH}"

# 检查 Go
if ! command -v go &> /dev/null; then
    echo ""
    echo "错误: 未找到 Go 编译器"
    echo "请安装 Go: https://golang.org/dl/"
    exit 1
fi

GO_VERSION=$(go version)
echo "Go: ${GO_VERSION}"
echo "系统: $(uname -s) $(uname -m)"

# 清理
if $CLEAN; then
    if [ -d "$BIN_DIR" ]; then
        rm -f "$BIN_DIR"/*
        echo "已清理 ${BIN_DIR}"
    fi
fi

# 确定架构参数
case "$ARCH" in
    arm64) GOARCH="arm64" SUFFIX="linux-arm64" ;;
    amd64) GOARCH="amd64" SUFFIX="linux-amd64" ;;
    *)
        echo "错误: 不支持的架构 ${ARCH}"
        exit 1
        ;;
esac

mkdir -p "$BIN_DIR"

OUTPUT_NAME="${SERVICE_NAME}-${VERSION}-${SUFFIX}"
OUTPUT_PATH="${BIN_DIR}/${OUTPUT_NAME}"

LDFLAGS="-s -w -X main.version=${VERSION} -X main.buildTime=${BUILD_TIME}"

echo ""
echo "开始编译..."
echo "  目标: linux/${GOARCH}"
echo "  输出: ${OUTPUT_PATH}"

export GOOS=linux
export GOARCH="$GOARCH"
export CGO_ENABLED=0

if go build -ldflags "$LDFLAGS" -o "$OUTPUT_PATH" ./cmd; then
    echo "  成功!"
    SUCCESS=true
else
    echo "  失败!"
    SUCCESS=false
fi

echo ""
echo "=================================================="
if $SUCCESS; then
    echo "编译成功!"
else
    echo "编译失败!"
fi
echo "=================================================="

# 列出输出文件
if [ -d "$BIN_DIR" ]; then
    echo ""
    echo "输出文件:"
    for f in "$BIN_DIR"/${SERVICE_NAME}-${VERSION}-*; do
        if [ -f "$f" ]; then
            size=$(stat -c%s "$f" 2>/dev/null || stat -f%z "$f" 2>/dev/null)
            size_mb=$(echo "scale=2; $size / 1048576" | bc 2>/dev/null || echo "?")
            echo "  $(basename "$f") (${size_mb} MB)"
        fi
    done
fi

if ! $SUCCESS; then
    exit 1
fi
