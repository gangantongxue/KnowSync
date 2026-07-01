#!/usr/bin/env bash
# Docker 镜像构建脚本
# 使用已编译的二进制文件构建 Docker 镜像
#
# 注意: 需要先运行 build.sh 编译二进制文件
#
# 用法:
#   bash scripts/build_image.sh              # 使用默认版本 (latest)
#   bash scripts/build_image.sh v1.0.0       # 指定版本号
#   bash scripts/build_image.sh --platform linux/arm64 v1.0.0  # 指定平台

set -euo pipefail

SERVICE_NAME="ai-server"
IMAGE_PREFIX="knowsync"
DEFAULT_VERSION="latest"
DEFAULT_PLATFORM="linux/amd64"

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"

VERSION="$DEFAULT_VERSION"
PLATFORM="$DEFAULT_PLATFORM"
NO_LATEST=false

# 解析参数
while [[ $# -gt 0 ]]; do
    case "$1" in
        --platform)
            PLATFORM="$2"
            shift 2
            ;;
        --no-latest)
            NO_LATEST=true
            shift
            ;;
        --help)
            cat << 'HELP'
ai-server Docker 镜像构建脚本

用法:
  bash scripts/build_image.sh              # 使用默认版本 latest
  bash scripts/build_image.sh v1.0.0       # 指定版本号 v1.0.0
  bash scripts/build_image.sh --platform linux/arm64 v1.0.0  # 指定平台

选项:
  --platform   目标平台 (默认: linux/amd64)
  --no-latest  不添加 latest 标签

注意: 需要先运行 build.sh 编译二进制文件
HELP
            exit 0
            ;;
        *)
            VERSION="$1"
            shift
            ;;
    esac
done

echo "=================================================="
echo "${SERVICE_NAME} Docker 镜像构建"
echo "=================================================="
echo "版本: ${VERSION}"
echo "平台: ${PLATFORM}"

# 检查 Docker
if ! command -v docker &> /dev/null; then
    echo ""
    echo "错误: 未找到 Docker"
    echo "请安装 Docker: https://docs.docker.com/get-docker/"
    exit 1
fi

# 确定二进制文件后缀
if echo "$PLATFORM" | grep -qE "arm64|aarch64"; then
    BINARY_SUFFIX="linux-arm64"
else
    BINARY_SUFFIX="linux-amd64"
fi

BINARY_NAME="${SERVICE_NAME}-${VERSION}-${BINARY_SUFFIX}"
BINARY_PATH="${PROJECT_ROOT}/bin/${BINARY_NAME}"

# 检查二进制文件
if [ ! -f "$BINARY_PATH" ]; then
    echo ""
    echo "错误: 未找到二进制文件 ${BINARY_PATH}"
    echo "请先运行: bash scripts/build.sh ${VERSION}"
    if echo "$PLATFORM" | grep -qE "arm64|aarch64"; then
        echo "或运行: bash scripts/build.sh ${VERSION} --arm64"
    fi
    exit 1
fi

# 构建镜像
IMAGE_NAME="${IMAGE_PREFIX}/${SERVICE_NAME}:${VERSION}"

echo ""
echo "构建 Docker 镜像..."
echo "  镜像: ${IMAGE_NAME}"
echo "  平台: ${PLATFORM}"
echo "  命令: docker build --platform ${PLATFORM} --build-arg BINARY_NAME=${BINARY_NAME} --build-arg SERVICE_NAME=${SERVICE_NAME} -t ${IMAGE_NAME} -f ${PROJECT_ROOT}/Dockerfile ${PROJECT_ROOT}"

docker build \
    --platform "$PLATFORM" \
    --build-arg "BINARY_NAME=${BINARY_NAME}" \
    --build-arg "SERVICE_NAME=${SERVICE_NAME}" \
    -t "$IMAGE_NAME" \
    -f "${PROJECT_ROOT}/Dockerfile" \
    "$PROJECT_ROOT"

echo "  镜像构建成功!"

# 添加 latest 标签
if ! $NO_LATEST && [ "$VERSION" != "latest" ]; then
    LATEST_NAME="${IMAGE_PREFIX}/${SERVICE_NAME}:latest"
    echo ""
    echo "添加 latest 标签..."
    docker tag "$IMAGE_NAME" "$LATEST_NAME"
    echo "  已添加标签: ${LATEST_NAME}"
fi

echo ""
echo "=================================================="
echo "构建完成!"
echo "=================================================="
echo ""
echo "镜像列表:"
echo "  ${IMAGE_NAME}"
if ! $NO_LATEST && [ "$VERSION" != "latest" ]; then
    echo "  ${IMAGE_PREFIX}/${SERVICE_NAME}:latest"
fi
echo ""
echo "运行命令:"
echo "  docker run -p 8080:8080 ${IMAGE_NAME}"
