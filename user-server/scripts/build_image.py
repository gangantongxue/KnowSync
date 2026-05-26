#!/usr/bin/env python3
"""
Docker 镜像构建脚本
先编译二进制文件，然后构建 Docker 镜像

用法:
  python3 scripts/build_image.py              # 使用默认版本 (latest)
  python3 scripts/build_image.py v1.0.0       # 指定版本号
  python3 scripts/build_image.py --help       # 显示帮助
"""

import argparse
import subprocess
import sys
from pathlib import Path


# 镜像配置
SERVICE_NAME = "user-server"
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


def run_build_script(version: str) -> bool:
    """运行编译脚本"""
    build_script = get_project_root() / "scripts" / "build.py"
    
    print(f"\n步骤 1: 编译二进制文件...")
    try:
        result = subprocess.run(
            ["python3", str(build_script), version],
            check=True
        )
        return True
    except subprocess.CalledProcessError:
        print("编译失败!")
        return False


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
  python3 scripts/build_image.py              # 使用默认版本 latest
  python3 scripts/build_image.py v1.0.0       # 指定版本号 v1.0.0
  python3 scripts/build_image.py --platform linux/arm64 v1.0.0  # 指定平台
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
    
    # 步骤 1: 编译
    if not run_build_script(version):
        sys.exit(1)
    
    # 步骤 2: 构建镜像
    if not build_docker_image(version, args.platform):
        sys.exit(1)
    
    # 步骤 3: 添加 latest 标签
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
