#!/usr/bin/env bash
# Proto 文件编译脚本
# 将 proto/ 目录下的所有 .proto 文件编译生成 Go 代码到 pkg/pb/ 目录
#
# 用法:
#   bash build_proto.sh                    # 使用默认值
#   bash build_proto.sh proto pkg/pb       # 指定目录

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"

PROTO_DIR="${1:-proto}"
OUTPUT_DIR="${2:-pkg/pb}"

PROTO_DIR_ABS="${PROJECT_ROOT}/${PROTO_DIR}"
OUTPUT_DIR_ABS="${PROJECT_ROOT}/${OUTPUT_DIR}"

echo "=================================================="
echo "KnowSync Proto 编译脚本"
echo "=================================================="
echo ""
echo "Proto 目录: ${PROTO_DIR}"
echo "输出目录: ${OUTPUT_DIR}"
echo ""

# 检查 protoc
if ! command -v protoc &> /dev/null; then
    echo "错误: 未找到 protoc 编译器"
    echo ""
    echo "请安装 Protocol Buffers 编译器:"
    echo ""
    echo "macOS:"
    echo "  brew install protobuf"
    echo ""
    echo "Ubuntu/Debian:"
    echo "  apt-get install -y protobuf-compiler"
    echo ""
    echo "Windows (使用 chocolatey):"
    echo "  choco install protoc"
    echo ""
    echo "或者从官方下载: https://github.com/protocolbuffers/protobuf/releases"
    exit 1
fi

# 检查 Go 插件
if ! command -v protoc-gen-go &> /dev/null; then
    echo "错误: 未找到 protoc-gen-go 插件"
    echo ""
    echo "请安装 Go Protocol Buffers 插件:"
    echo ""
    echo "  go install google.golang.org/protobuf/cmd/protoc-gen-go@latest"
    echo "  go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest"
    echo ""
    echo "安装后请确保 \$GOPATH/bin 或 \$GOBIN 在 PATH 环境变量中"
    exit 1
fi

# 检查 proto 目录
if [ ! -d "$PROTO_DIR_ABS" ]; then
    echo "错误: proto 目录不存在: ${PROTO_DIR_ABS}"
    exit 1
fi

# 创建输出目录
mkdir -p "$OUTPUT_DIR_ABS"

# 查找 proto 文件
PROTO_FILES=()
while IFS= read -r -d '' file; do
    PROTO_FILES+=("$file")
done < <(find "$PROTO_DIR_ABS" -maxdepth 1 -name "*.proto" -print0 | sort -z)

if [ ${#PROTO_FILES[@]} -eq 0 ]; then
    echo "警告: 在 ${PROTO_DIR_ABS} 中未找到 .proto 文件"
    exit 0
fi

echo "找到 ${#PROTO_FILES[@]} 个 proto 文件:"
for f in "${PROTO_FILES[@]}"; do
    echo "  - $(basename "$f")"
done
echo ""

# 编译 proto 文件
SUCCESS_COUNT=0
FAIL_COUNT=0

for proto_file in "${PROTO_FILES[@]}"; do
    filename=$(basename "$proto_file")
    echo -n "正在编译: ${filename} ... "
    if protoc \
        --proto_path="$PROTO_DIR_ABS" \
        --go_out="$OUTPUT_DIR_ABS" \
        --go_opt=paths=source_relative \
        --go-grpc_out="$OUTPUT_DIR_ABS" \
        --go-grpc_opt=paths=source_relative \
        "$proto_file"; then
        echo "✓ 成功"
        SUCCESS_COUNT=$((SUCCESS_COUNT + 1))
    else
        echo "✗ 失败"
        FAIL_COUNT=$((FAIL_COUNT + 1))
    fi
done

echo ""
echo "=================================================="
echo "编译完成: 成功 ${SUCCESS_COUNT} 个, 失败 ${FAIL_COUNT} 个"
echo "输出目录: ${OUTPUT_DIR_ABS}"
echo "=================================================="

if [ "$FAIL_COUNT" -gt 0 ]; then
    exit 1
fi
