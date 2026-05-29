#!/usr/bin/env python3
"""
Docker 镜像构建脚本
先编译前端，然后构建 Docker 镜像（Caddy + 静态文件）

用法:
  python3 scripts/build_image.py              # 使用默认版本 (latest)
  python3 scripts/build_image.py v1.0.0       # 指定版本号
  python3 scripts/build_image.py --help       # 显示帮助
"""

import argparse
import subprocess
import sys
from pathlib import Path


SERVICE_NAME = "web"
IMAGE_PREFIX = "knowsync"
DEFAULT_VERSION = "latest"
DEFAULT_PLATFORM = "linux/amd64"


def get_project_root() -> Path:
    return Path(__file__).parent.parent


def check_docker() -> bool:
    try:
        subprocess.run(["docker", "--version"], capture_output=True, check=True)
        return True
    except (subprocess.CalledProcessError, FileNotFoundError):
        return False


def run_build_script(version: str) -> bool:
    build_script = get_project_root() / "scripts" / "build.py"
    print(f"\n步骤 1: 编译前端...")
    try:
        subprocess.run(["python3", str(build_script), version], check=True)
        return True
    except subprocess.CalledProcessError:
        print("编译失败!")
        return False


def build_docker_image(version: str, platform: str) -> bool:
    project_root = get_project_root()
    image_name = f"{IMAGE_PREFIX}/{SERVICE_NAME}:{version}"

    print(f"\n步骤 2: 构建 Docker 镜像...")
    print(f"  镜像: {image_name}")
    print(f"  平台: {platform}")

    if not (project_root / "dist").exists():
        print(f"  错误: 未找到 dist/ 目录，请先编译")
        return False

    cmd = [
        "docker", "build",
        "--platform", platform,
        "--build-arg", f"VERSION={version}",
        "-t", image_name,
        "-f", str(project_root / "Dockerfile"),
        str(project_root)
    ]

    print(f"  命令: {' '.join(cmd)}")

    try:
        subprocess.run(cmd, check=True)
        print(f"  镜像构建成功!")
        return True
    except subprocess.CalledProcessError:
        print(f"  镜像构建失败!")
        return False


def tag_latest(version: str) -> bool:
    if version == "latest":
        return True
    image_name = f"{IMAGE_PREFIX}/{SERVICE_NAME}:{version}"
    latest_name = f"{IMAGE_PREFIX}/{SERVICE_NAME}:latest"
    print(f"\n步骤 3: 添加 latest 标签...")
    try:
        subprocess.run(["docker", "tag", image_name, latest_name], check=True)
        print(f"  已添加标签: {latest_name}")
        return True
    except subprocess.CalledProcessError:
        print(f"  添加标签失败!")
        return False


def main():
    parser = argparse.ArgumentParser(description=f"{SERVICE_NAME} Docker 镜像构建脚本")
    parser.add_argument("version", nargs="?", default=DEFAULT_VERSION, help=f"版本号 (默认: {DEFAULT_VERSION})")
    parser.add_argument("--platform", default=DEFAULT_PLATFORM, help=f"目标平台 (默认: {DEFAULT_PLATFORM})")
    parser.add_argument("--no-latest", action="store_true", help="不添加 latest 标签")

    args = parser.parse_args()

    print(f"{'=' * 50}")
    print(f"{SERVICE_NAME} Docker 镜像构建")
    print(f"{'=' * 50}")
    print(f"版本: {args.version}")
    print(f"平台: {args.platform}")

    if not check_docker():
        print("\n错误: 未找到 Docker")
        sys.exit(1)

    if not run_build_script(args.version):
        sys.exit(1)

    if not build_docker_image(args.version, args.platform):
        sys.exit(1)

    if not args.no_latest:
        tag_latest(args.version)

    print(f"\n{'=' * 50}")
    print(f"构建完成!")
    print(f"{'=' * 50}")
    print(f"\n镜像: {IMAGE_PREFIX}/{SERVICE_NAME}:{args.version}")
    if not args.no_latest and args.version != "latest":
        print(f"      {IMAGE_PREFIX}/{SERVICE_NAME}:latest")
    print(f"\n运行: docker run -p 80:80 {IMAGE_PREFIX}/{SERVICE_NAME}:{args.version}")


if __name__ == "__main__":
    main()
