#!/usr/bin/env python3
"""
React 前端编译脚本
执行 npm run build 编译前端到 dist/ 目录

用法:
  python3 scripts/build.py              # 使用默认版本 (latest)
  python3 scripts/build.py v1.0.0       # 指定版本号
  python3 scripts/build.py --help       # 显示帮助
"""

import argparse
import subprocess
import sys
from pathlib import Path


DEFAULT_VERSION = "latest"
SERVICE_NAME = "web"


def get_project_root() -> Path:
    return Path(__file__).parent.parent


def check_node() -> bool:
    try:
        subprocess.run(["node", "--version"], capture_output=True, check=True)
        return True
    except (subprocess.CalledProcessError, FileNotFoundError):
        return False


def check_npm() -> bool:
    try:
        # 使用 shell=True 确保能找到 npm 命令（解决 Windows 环境变量问题）
        subprocess.run("npm --version", shell=True, capture_output=True, check=True)
        return True
    except (subprocess.CalledProcessError, FileNotFoundError):
        return False


def install_deps():
    print("\n安装依赖...")
    # 使用 shell=True 确保能找到 npm 命令（解决 Windows 环境变量问题）
    result = subprocess.run("npm install", shell=True, cwd=get_project_root())
    if result.returncode != 0:
        print("依赖安装失败!")
        sys.exit(1)
    print("依赖安装完成")


def run_build(version: str) -> bool:
    project_root = get_project_root()
    print(f"\n编译前端 (版本: {version})...")

    env = {"VITE_APP_VERSION": version}
    # 使用 shell=True 确保能找到 npm 命令（解决 Windows 环境变量问题）
    result = subprocess.run(
        "npm run build",
        shell=True,
        cwd=project_root,
        env={**__import__("os").environ, **env},
    )

    if result.returncode != 0:
        print("编译失败!")
        return False

    # 写入版本标记
    dist_dir = project_root / "dist"
    version_file = dist_dir / "version.txt"
    dist_dir.mkdir(exist_ok=True)
    version_file.write_text(version)

    print(f"编译成功! 输出目录: {dist_dir}")
    return True


def main():
    parser = argparse.ArgumentParser(
        description=f"{SERVICE_NAME} React 前端编译脚本",
    )
    parser.add_argument("version", nargs="?", default=DEFAULT_VERSION, help=f"版本号 (默认: {DEFAULT_VERSION})")

    args = parser.parse_args()

    print(f"{'=' * 50}")
    print(f"{SERVICE_NAME} 前端编译")
    print(f"{'=' * 50}")
    print(f"版本: {args.version}")

    if not check_node():
        print("\n错误: 未找到 Node.js")
        sys.exit(1)

    if not check_npm():
        print("\n错误: 未找到 npm")
        sys.exit(1)

    install_deps()

    if not run_build(args.version):
        sys.exit(1)

    print(f"\n{'=' * 50}")
    print(f"编译完成!")
    print(f"{'=' * 50}")


if __name__ == "__main__":
    main()
