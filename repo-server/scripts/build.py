#!/usr/bin/env python3
"""
跨平台编译脚本
将 Go 代码编译为 Linux 平台的二进制文件，输出到 bin/ 目录

用法:
  python scripts/build.py              # 使用默认版本 (latest)，默认架构 (amd64)
  python scripts/build.py v1.0.0       # 指定版本号，默认架构 (amd64)
  python scripts/build.py --arm64      # 使用 arm64 架构
  python scripts/build.py v1.0.0 --arm64  # 指定版本号和架构
  python scripts/build.py --help       # 显示帮助
"""

import argparse
import os
import platform
import subprocess
import sys
from pathlib import Path


# 编译配置
DEFAULT_VERSION = "latest"
DEFAULT_ARCH = "amd64"
SERVICE_NAME = "repo-server"


# 架构映射
ARCH_TARGETS = {
    "amd64": {"goos": "linux", "goarch": "amd64", "suffix": "linux-amd64"},
    "arm64": {"goos": "linux", "goarch": "arm64", "suffix": "linux-arm64"},
}


def get_project_root() -> Path:
    """获取项目根目录"""
    return Path(__file__).parent.parent


def get_bin_dir() -> Path:
    """获取二进制输出目录"""
    return get_project_root() / "bin"


def check_go() -> bool:
    """检查是否安装了 Go"""
    try:
        subprocess.run(
            ["go", "version"],
            capture_output=True,
            check=True
        )
        return True
    except (subprocess.CalledProcessError, FileNotFoundError):
        return False


def get_go_version() -> str:
    """获取 Go 版本"""
    result = subprocess.run(
        ["go", "version"],
        capture_output=True,
        text=True,
        check=True
    )
    return result.stdout.strip()


def clean_bin_dir():
    """清理 bin 目录"""
    bin_dir = get_bin_dir()
    if bin_dir.exists():
        for f in bin_dir.glob("*"):
            if f.is_file():
                f.unlink()
        print(f"已清理 {bin_dir}")


def build_binary(target: dict, version: str) -> bool:
    """编译单个目标平台的二进制文件"""
    bin_dir = get_bin_dir()
    bin_dir.mkdir(exist_ok=True)
    
    output_name = f"{SERVICE_NAME}-{version}-{target['suffix']}"
    output_path = bin_dir / output_name
    
    # 构建 ldflags
    ldflags = f"-s -w -X main.version={version} -X main.buildTime={get_build_time()}"
    
    env = os.environ.copy()
    env["GOOS"] = target["goos"]
    env["GOARCH"] = target["goarch"]
    env["CGO_ENABLED"] = "0"
    
    cmd = [
        "go", "build",
        "-ldflags", ldflags,
        "-o", str(output_path),
        "./cmd"
    ]
    
    print(f"\n编译 {target['goos']}/{target['goarch']}...")
    print(f"  输出: {output_path}")
    
    try:
        result = subprocess.run(
            cmd,
            env=env,
            capture_output=True,
            text=True,
            check=True
        )
        print(f"  成功!")
        return True
    except subprocess.CalledProcessError as e:
        print(f"  失败!")
        print(f"  错误: {e.stderr}")
        return False


def get_build_time() -> str:
    """获取构建时间"""
    from datetime import datetime
    return datetime.now().strftime("%Y-%m-%d_%H:%M:%S")


def main():
    parser = argparse.ArgumentParser(
        description=f"{SERVICE_NAME} 跨平台编译脚本",
        formatter_class=argparse.RawDescriptionHelpFormatter,
        epilog="""
示例:
  python scripts/build.py              # 使用默认版本 latest，默认架构 amd64
  python scripts/build.py v1.0.0       # 指定版本号 v1.0.0，默认架构 amd64
  python scripts/build.py --arm64      # 使用默认版本 latest，架构 arm64
  python scripts/build.py v1.0.0 --arm64  # 指定版本号和架构
        """
    )
    parser.add_argument(
        "version",
        nargs="?",
        default=DEFAULT_VERSION,
        help=f"版本号 (默认: {DEFAULT_VERSION})"
    )
    parser.add_argument(
        "--arm64",
        action="store_true",
        help="编译 arm64 架构 (默认编译 amd64)"
    )
    parser.add_argument(
        "--clean",
        action="store_true",
        help="编译前清理 bin 目录"
    )
    
    args = parser.parse_args()
    version = args.version
    arch = "arm64" if args.arm64 else DEFAULT_ARCH
    
    print(f"=" * 50)
    print(f"{SERVICE_NAME} 跨平台编译")
    print(f"=" * 50)
    print(f"版本: {version}")
    print(f"架构: {arch}")
    print(f"Go: {get_go_version()}")
    print(f"系统: {platform.system()} {platform.machine()}")
    
    # 检查 Go 环境
    if not check_go():
        print("\n错误: 未找到 Go 编译器")
        print("请安装 Go: https://golang.org/dl/")
        sys.exit(1)
    
    # 清理
    if args.clean:
        clean_bin_dir()
    
    # 编译
    print(f"\n开始编译...")
    target = ARCH_TARGETS[arch]
    success = build_binary(target, version)
    
    # 结果
    print(f"\n{'=' * 50}")
    if success:
        print(f"编译成功!")
    else:
        print(f"编译失败!")
    print(f"{'=' * 50}")
    
    # 列出输出文件
    bin_dir = get_bin_dir()
    if bin_dir.exists():
        print(f"\n输出文件:")
        for f in sorted(bin_dir.glob(f"{SERVICE_NAME}-{version}-*")):
            size = f.stat().st_size
            size_mb = size / (1024 * 1024)
            print(f"  {f.name} ({size_mb:.2f} MB)")
    
    if not success:
        sys.exit(1)


if __name__ == "__main__":
    main()
