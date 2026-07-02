#!/usr/bin/env bash
# Docker 镜像构建脚本
# 使用已编译的前端文件构建 Docker 镜像（Caddy + 静态文件）
#
# 注意: 需要先运行 build.sh 编译前端
#
# 用法:
#   bash scripts/build_image.sh              # 使用默认版本 (latest)
#   bash scripts/build_image.sh v1.0.0       # 指定版本号

set -euo pipefail

SERVICE_NAME="web"
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
web Docker 镜像构建脚本

用法:
  bash scripts/build_image.sh              # 使用默认版本 latest
  bash scripts/build_image.sh v1.0.0       # 指定版本号 v1.0.0
  bash scripts/build_image.sh --platform linux/arm64 v1.0.0  # 指定平台

选项:
  --platform   目标平台 (默认: linux/amd64)
  --no-latest  不添加 latest 标签

注意: 需要先运行 build.sh 编译前端
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

# 检查 dist 目录
DIST_DIR="${PROJECT_ROOT}/dist"
if [ ! -d "$DIST_DIR" ]; then
    echo ""
    echo "错误: 未找到 dist/ 目录"
    echo "请先运行: bash scripts/build.sh"
    exit 1
fi

# 构建镜像
IMAGE_NAME="${IMAGE_PREFIX}/${SERVICE_NAME}:${VERSION}"

# 删除旧的同名镜像，防止悬空镜像
if docker image inspect "$IMAGE_NAME" &> /dev/null; then
    echo ""
    echo "删除旧镜像: ${IMAGE_NAME}..."
    docker rmi "$IMAGE_NAME" || true
fi

echo ""
echo "构建 Docker 镜像..."
echo "  镜像: ${IMAGE_NAME}"
echo "  平台: ${PLATFORM}"
echo "  命令: docker build --platform ${PLATFORM} --build-arg VERSION=${VERSION} -t ${IMAGE_NAME} -f ${PROJECT_ROOT}/Dockerfile ${PROJECT_ROOT}"

docker build \
    --platform "$PLATFORM" \
    --build-arg "VERSION=${VERSION}" \
    -t "$IMAGE_NAME" \
    -f "${PROJECT_ROOT}/Dockerfile" \
    "$PROJECT_ROOT"

echo "  镜像构建成功!"

# 添加 latest 标签
if ! $NO_LATEST && [ "$VERSION" != "latest" ]; then
    LATEST_NAME="${IMAGE_PREFIX}/${SERVICE_NAME}:latest"
    echo ""
    echo "添加 latest 标签..."

    # 删除旧的 latest 镜像，防止悬空镜像
    if docker image inspect "$LATEST_NAME" &> /dev/null; then
        echo "  删除旧镜像: ${LATEST_NAME}..."
        docker rmi "$LATEST_NAME" || true
    fi

    docker tag "$IMAGE_NAME" "$LATEST_NAME"
    echo "  已添加标签: ${LATEST_NAME}"
fi

echo ""
echo "=================================================="
echo "构建完成!"
echo "=================================================="
echo ""
echo "镜像: ${IMAGE_NAME}"
if ! $NO_LATEST && [ "$VERSION" != "latest" ]; then
    echo "      ${IMAGE_PREFIX}/${SERVICE_NAME}:latest"
fi
echo ""
echo "运行: docker run -p 80:80 ${IMAGE_NAME}"
