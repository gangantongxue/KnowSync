#!/usr/bin/env python3
"""
Proto 文件编译脚本
将 proto/ 目录下的所有 .proto 文件编译生成 Go 代码到 pkg/pb/ 目录

用法:
  python3 build_proto.py                    # 使用默认值
  python3 build_proto.py proto pkg/pb       # 指定目录
  python3 build_proto.py --help             # 显示帮助
"""

import argparse
import os
import subprocess
import sys
from pathlib import Path


def get_project_root() -> Path:
    """获取项目根目录（脚本所在目录的父目录）"""
    return Path(__file__).parent.parent


def check_protoc() -> bool:
    """检查是否安装了 protoc"""
    try:
        subprocess.run(
            ["protoc", "--version"],
            capture_output=True,
            check=True
        )
        return True
    except (subprocess.CalledProcessError, FileNotFoundError):
        return False


def install_protoc_instructions():
    """输出安装 protoc 的指引"""
    print("错误: 未找到 protoc 编译器")
    print("")
    print("请安装 Protocol Buffers 编译器:")
    print("")
    print("macOS:")
    print("  brew install protobuf")
    print("")
    print("Ubuntu/Debian:")
    print("  apt-get install -y protobuf-compiler")
    print("")
    print("Windows (使用 chocolatey):")
    print("  choco install protoc")
    print("")
    print("或者从官方下载: https://github.com/protocolbuffers/protobuf/releases")
    sys.exit(1)


def check_go_plugins() -> bool:
    """检查是否安装了 Go 插件"""
    try:
        subprocess.run(
            ["protoc-gen-go", "--version"],
            capture_output=True,
            check=True
        )
        return True
    except (subprocess.CalledProcessError, FileNotFoundError):
        return False


def install_go_plugins_instructions():
    """输出安装 Go 插件的指引"""
    print("错误: 未找到 protoc-gen-go 插件")
    print("")
    print("请安装 Go Protocol Buffers 插件:")
    print("")
    print("  go install google.golang.org/protobuf/cmd/protoc-gen-go@latest")
    print("")
    print("安装后请确保 $GOPATH/bin 或 $GOBIN 在 PATH 环境变量中")
    sys.exit(1)


def find_proto_files(proto_dir: Path) -> list[Path]:
    """查找所有 proto 文件"""
    return sorted(proto_dir.glob("*.proto"))


def compile_proto(proto_file: Path, proto_dir: Path, output_dir: Path) -> bool:
    """编译单个 proto 文件"""
    cmd = [
        "protoc",
        f"--proto_path={proto_dir}",
        f"--go_out={output_dir}",
        f"--go_opt=paths=source_relative",
        str(proto_file)
    ]

    try:
        result = subprocess.run(
            cmd,
            capture_output=True,
            text=True,
            check=True
        )
        return True
    except subprocess.CalledProcessError as e:
        print(f"编译失败: {proto_file.name}")
        print(f"错误信息: {e.stderr}")
        return False


def parse_args():
    """解析命令行参数"""
    parser = argparse.ArgumentParser(
        description="编译 proto 文件生成 Go 代码",
        formatter_class=argparse.RawDescriptionHelpFormatter,
        epilog="""
示例:
  %(prog)s                    # 使用默认目录
  %(prog)s proto pkg/pb       # 指定 proto 和输出目录
        """
    )
    parser.add_argument(
        "proto_dir",
        nargs="?",
        default="proto",
        help="proto 文件所在目录 (默认: proto)"
    )
    parser.add_argument(
        "output_dir",
        nargs="?",
        default="pkg/pb",
        help="Go 代码输出目录 (默认: pkg/pb)"
    )
    return parser.parse_args()


def main():
    """主函数"""
    args = parse_args()

    print("=" * 50)
    print("KnowSync Proto 编译脚本")
    print("=" * 50)
    print("")

    # 检查依赖
    if not check_protoc():
        install_protoc_instructions()

    if not check_go_plugins():
        install_go_plugins_instructions()

    # 获取路径
    project_root = get_project_root()
    proto_dir = project_root / args.proto_dir
    output_dir = project_root / args.output_dir

    print(f"Proto 目录: {proto_dir.relative_to(project_root)}")
    print(f"输出目录: {output_dir.relative_to(project_root)}")
    print("")

    # 检查 proto 目录
    if not proto_dir.exists():
        print(f"错误: proto 目录不存在: {proto_dir}")
        sys.exit(1)

    # 创建输出目录
    output_dir.mkdir(parents=True, exist_ok=True)

    # 查找 proto 文件
    proto_files = find_proto_files(proto_dir)

    if not proto_files:
        print(f"警告: 在 {proto_dir} 中未找到 .proto 文件")
        sys.exit(0)

    print(f"找到 {len(proto_files)} 个 proto 文件:")
    for f in proto_files:
        print(f"  - {f.name}")
    print("")

    # 编译 proto 文件
    success_count = 0
    fail_count = 0

    for proto_file in proto_files:
        print(f"正在编译: {proto_file.name} ...", end=" ")
        if compile_proto(proto_file, proto_dir, output_dir):
            print("✓ 成功")
            success_count += 1
        else:
            print("✗ 失败")
            fail_count += 1

    print("")
    print("=" * 50)
    print(f"编译完成: 成功 {success_count} 个, 失败 {fail_count} 个")
    print(f"输出目录: {output_dir}")
    print("=" * 50)

    if fail_count > 0:
        sys.exit(1)


if __name__ == "__main__":
    main()
