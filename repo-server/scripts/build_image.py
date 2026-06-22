#!/usr/bin/env python3
"""
Docker 镜像构建脚本
使用已编译的二进制文件构建 Docker 镜像

注意: 需要先运行 build.py 编译二进制文件

用法:
  python scripts/build_image.py              # 使用默认版本 (latest)
  python scripts/build_image.py v1.0.0       # 指定版本号
  python scripts/build_image.py --help       # 显示帮助
"""

import argparse
import subprocess
import sys
from pathlib import Path


# 镜像配置
SERVICE_NAME = "repo-server"
IMAGE_PREFIX = "knowsync"
DEFAULT_VERSION = "latest"
DEFAULT_PLATFORM = "linux/amd64"


def get_project_root() -> Path:
    """获取项目根目录"""
    return Path(__file__).parent.parent


def check_docker() -> bool:
    """检查是否安装了 Docker"""
    try:
        subprocess.run(
            ["docker", "--version"],
            capture_output=True,
            check=True
        )
        return True
    except (subprocess.CalledProcessError, FileNotFoundError):
        return False


def check_binary_exists(version: str, platform: str) -> bool:
    """检查二进制文件是否存在"""
    project_root = get_project_root()
    
    # 确定使用哪个二进制文件
    if "arm64" in platform or "aarch64" in platform:
        binary_suffix = "linux-arm64"
    else:
        binary_suffix = "linux-amd64"
    
    binary_name = f"{SERVICE_NAME}-{version}-{binary_suffix}"
    binary_path = project_root / "bin" / binary_name
    
    if not binary_path.exists():
        print(f"\n错误: 未找到二进制文件 {binary_path}")
        print(f"请先运行: python scripts/build.py {version}")
        if "arm64" in platform:
            print(f"或运行: python scripts/build.py {version} --arm64")
        return False
    
    return True


def build_docker_image(version: str, platform: str) -> bool:
    """构建 Docker 镜像"""
    project_root = get_project_root()
    image_name = f"{IMAGE_PREFIX}/{SERVICE_NAME}:{version}"
    
    print(f"\n步骤 2: 构建 Docker 镜像...")
    print(f"  镜像: {image_name}")
    print(f"  平台: {platform}")
    
    # 确定使用哪个二进制文件
    if "arm64" in platform or "aarch64" in platform:
        binary_suffix = "linux-arm64"
    else:
        binary_suffix = "linux-amd64"
    
    binary_name = f"{SERVICE_NAME}-{version}-{binary_suffix}"
    binary_path = project_root / "bin" / binary_name
    
    if not binary_path.exists():
        print(f"  错误: 未找到二进制文件 {binary_path}")
        return False
    
    # 构建 Docker 镜像
    cmd = [
        "docker", "build",
        "--platform", platform,
        "--build-arg", f"BINARY_NAME={binary_name}",
        "--build-arg", f"SERVICE_NAME={SERVICE_NAME}",
        "-t", image_name,
        "-f", str(project_root / "Dockerfile"),
        str(project_root)
    ]
    
    print(f"  命令: {' '.join(cmd)}")
    
    try:
        result = subprocess.run(cmd, check=True)
        print(f"  镜像构建成功!")
        return True
    except subprocess.CalledProcessError as e:
        print(f"  镜像构建失败!")
        return False


def tag_latest(version: str) -> bool:
    """为镜像添加 latest 标签"""
    if version == "latest":
        return True
    
    image_name = f"{IMAGE_PREFIX}/{SERVICE_NAME}:{version}"
    latest_name = f"{IMAGE_PREFIX}/{SERVICE_NAME}:latest"
    
    print(f"\n步骤 3: 添加 latest 标签...")
    
    cmd = ["docker", "tag", image_name, latest_name]
    
    try:
        subprocess.run(cmd, check=True)
        print(f"  已添加标签: {latest_name}")
        return True
    except subprocess.CalledProcessError:
        print(f"  添加标签失败!")
        return False


def main():
    parser = argparse.ArgumentParser(
        description=f"{SERVICE_NAME} Docker 镜像构建脚本",
        formatter_class=argparse.RawDescriptionHelpFormatter,
        epilog="""
示例:
  python scripts/build_image.py              # 使用默认版本 latest
  python scripts/build_image.py v1.0.0       # 指定版本号 v1.0.0
  python scripts/build_image.py --platform linux/arm64 v1.0.0  # 指定平台

注意: 需要先运行 build.py 编译二进制文件
        """
    )
    parser.add_argument(
        "version",
        nargs="?",
        default=DEFAULT_VERSION,
        help=f"版本号 (默认: {DEFAULT_VERSION})"
    )
    parser.add_argument(
        "--platform",
        default=DEFAULT_PLATFORM,
        help=f"目标平台 (默认: {DEFAULT_PLATFORM})"
    )
    parser.add_argument(
        "--no-latest",
        action="store_true",
        help="不添加 latest 标签"
    )
    
    args = parser.parse_args()
    version = args.version
    
    print(f"=" * 50)
    print(f"{SERVICE_NAME} Docker 镜像构建")
    print(f"=" * 50)
    print(f"版本: {version}")
    print(f"平台: {args.platform}")
    
    # 检查 Docker 环境
    if not check_docker():
        print("\n错误: 未找到 Docker")
        print("请安装 Docker: https://docs.docker.com/get-docker/")
        sys.exit(1)
    
    # 检查二进制文件是否存在
    if not check_binary_exists(version, args.platform):
        sys.exit(1)
    
    # 构建 Docker 镜像
    if not build_docker_image(version, args.platform):
        sys.exit(1)
    
    # 添加 latest 标签
    if not args.no_latest:
        tag_latest(version)
    
    # 完成
    print(f"\n{'=' * 50}")
    print(f"构建完成!")
    print(f"{'=' * 50}")
    print(f"\n镜像列表:")
    print(f"  {IMAGE_PREFIX}/{SERVICE_NAME}:{version}")
    if not args.no_latest and version != "latest":
        print(f"  {IMAGE_PREFIX}/{SERVICE_NAME}:latest")
    print(f"\n运行命令:")
    print(f"  docker run -p 8080:8080 {IMAGE_PREFIX}/{SERVICE_NAME}:{version}")


if __name__ == "__main__":
    main()
