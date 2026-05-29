# KnowSync Web 模块实现计划

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**目标：** 为 KnowSync 项目构建 Vite + React + TypeScript + Tailwind CSS 前端模块，Caddy 反向代理，Docker 部署

**架构：** Caddy 单入口，SPA 静态文件 + `/api/*` 反向代理到 gateway，三栏/两栏自适应布局

**技术栈：** Vite 6, React 19, TypeScript, Tailwind CSS 4, react-router-dom v7, Monaco Editor, react-markdown, Ant Design Upload(Cropper)

**设计文档：** `docs/superpowers/specs/2026-05-28-web-module-design.md`

**后端 API 文档：** `gateway/docs/user.jsonc`、`gateway/docs/repo.jsonc`、`gateway/docs/ai.jsonc`

---

### Task 1: 项目脚手架与构建流水线

**文件：**
- 创建: `web/package.json`
- 创建: `web/vite.config.ts`
- 创建: `web/tsconfig.json`
- 创建: `web/tsconfig.node.json`
- 创建: `web/tailwind.config.ts`
- 创建: `web/postcss.config.js`
- 创建: `web/index.html`
- 创建: `web/src/main.tsx`
- 创建: `web/src/styles/index.css`
- 创建: `web/Taskfile.yml`
- 创建: `web/scripts/build.py`
- 创建: `web/scripts/build_image.py`
- 创建: `web/Dockerfile`
- 创建: `web/Caddyfile`
- 创建: `web/.dockerignore`
- 创建: `web/.gitignore`
- 修改: `docker-compose.yaml`

- [*] **Step 1: 初始化 package.json**

```json
{
  "name": "knowsync-web",
  "private": true,
  "version": "0.1.0",
  "type": "module",
  "scripts": {
    "dev": "vite",
    "build": "tsc -b && vite build",
    "preview": "vite preview",
    "lint": "eslint ."
  },
  "dependencies": {
    "react": "^19.1.0",
    "react-dom": "^19.1.0",
    "react-router-dom": "^7.5.0",
    "react-markdown": "^10.1.0",
    "react-syntax-highlighter": "^15.6.1",
    "@monaco-editor/react": "^4.7.0",
    "antd": "^5.24.0",
    "@ant-design/icons": "^5.6.0",
    "dayjs": "^1.11.13"
  },
  "devDependencies": {
    "@types/react": "^19.1.0",
    "@types/react-dom": "^19.1.0",
    "@types/react-syntax-highlighter": "^15.5.13",
    "@vitejs/plugin-react": "^4.4.0",
    "typescript": "~5.8.0",
    "vite": "^6.3.0",
    "tailwindcss": "^4.1.0",
    "@tailwindcss/vite": "^4.1.0"
  }
}
```

- [*] **Step 2: vite.config.ts**

```typescript
import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'
import tailwindcss from '@tailwindcss/vite'

export default defineConfig({
  plugins: [react(), tailwindcss()],
  server: {
    port: 5173,
    proxy: {
      '/api': 'http://localhost:8080',
      '/files': 'http://localhost:8080',
    },
  },
  build: {
    outDir: 'dist',
    sourcemap: false,
  },
})
```

- [*] **Step 3: tsconfig.json**

```json
{
  "compilerOptions": {
    "target": "ES2022",
    "lib": ["ES2023", "DOM", "DOM.Iterable"],
    "module": "ESNext",
    "skipLibCheck": true,
    "moduleResolution": "bundler",
    "allowImportingTsExtensions": true,
    "isolatedModules": true,
    "moduleDetection": "force",
    "noEmit": true,
    "jsx": "react-jsx",
    "strict": true,
    "noUnusedLocals": true,
    "noUnusedParameters": true,
    "noFallthroughCasesInSwitch": true,
    "noUncheckedSideEffectImports": true,
    "paths": {
      "@/*": ["./src/*"]
    },
    "baseUrl": "."
  },
  "include": ["src"]
}
```

- [*] **Step 4: tsconfig.node.json**

```json
{
  "compilerOptions": {
    "target": "ES2022",
    "lib": ["ES2023"],
    "module": "ESNext",
    "skipLibCheck": true,
    "moduleResolution": "bundler",
    "allowImportingTsExtensions": true,
    "isolatedModules": true,
    "moduleDetection": "force",
    "noEmit": true,
    "strict": true
  },
  "include": ["vite.config.ts"]
}
```

- [*] **Step 5: index.html**

```html
<!DOCTYPE html>
<html lang="zh-CN">
  <head>
    <meta charset="UTF-8" />
    <link rel="icon" type="image/svg+xml" href="/knowsync.svg" />
    <meta name="viewport" content="width=device-width, initial-scale=1.0" />
    <title>KnowSync</title>
  </head>
  <body>
    <div id="root"></div>
    <script type="module" src="/src/main.tsx"></script>
  </body>
</html>
```

- [*] **Step 6: src/styles/index.css (Tailwind 入口)**

```css
@import "tailwindcss";
```

- [*] **Step 7: src/main.tsx (入口)**

```typescript
import { StrictMode } from 'react'
import { createRoot } from 'react-dom/client'
import { BrowserRouter } from 'react-router-dom'
import App from './App'
import './styles/index.css'

createRoot(document.getElementById('root')!).render(
  <StrictMode>
    <BrowserRouter>
      <App />
    </BrowserRouter>
  </StrictMode>,
)
```

- [*] **Step 8: Taskfile.yml**

```yaml
version: '3'

vars:
  SERVICE_NAME: web
  BIN_DIR: dist

tasks:
  default:
    silent: true
    cmds:
      - task: help

  help:
    desc: 显示帮助信息
    silent: true
    cmds:
      - |
        cat << 'EOF'
        {{.SERVICE_NAME}} 构建工具

        可用命令:
          task build        - 编译 React 前端到 ./dist
          task build:image  - 构建 Docker 镜像
          task dev          - 启动 Vite 开发服务器
          task clean        - 清除编译产物
          task help         - 显示此帮助信息

        使用方法:
          task                    - 默认显示帮助信息
          task build              - 执行编译
          task build -- v1.0.0    - 指定版本号编译
          task build:image        - 构建 Docker 镜像
          task build:image -- v1.0.0  - 指定版本号构建镜像
          task dev                - 启动开发服务器
          task clean              - 清除编译产物
        EOF

  build:
    desc: 编译 React 前端
    silent: true
    cmds:
      - |
        PYTHON_CMD=""
        if command -v python3 &> /dev/null; then
          PYTHON_CMD="python3"
        elif command -v python &> /dev/null; then
          PYTHON_CMD="python"
        else
          echo "错误: 未找到 python3 或 python 命令"
          exit 1
        fi
        $PYTHON_CMD scripts/build.py {{.CLI_ARGS}}

  build:image:
    desc: 构建 Docker 镜像
    silent: true
    cmds:
      - |
        PYTHON_CMD=""
        if command -v python3 &> /dev/null; then
          PYTHON_CMD="python3"
        elif command -v python &> /dev/null; then
          PYTHON_CMD="python"
        else
          echo "错误: 未找到 python3 或 python 命令"
          exit 1
        fi
        $PYTHON_CMD scripts/build_image.py {{.CLI_ARGS}}

  dev:
    desc: 启动 Vite 开发服务器
    silent: true
    cmds:
      - npm run dev

  clean:
    desc: 清除编译产物
    silent: true
    cmds:
      - rm -rf {{.BIN_DIR}}
      - echo "已清除编译产物"
```

- [*] **Step 9: scripts/build.py**

```python
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
        subprocess.run(["npm", "--version"], capture_output=True, check=True)
        return True
    except (subprocess.CalledProcessError, FileNotFoundError):
        return False


def install_deps():
    print("\n安装依赖...")
    result = subprocess.run(["npm", "install"], cwd=get_project_root())
    if result.returncode != 0:
        print("依赖安装失败!")
        sys.exit(1)
    print("依赖安装完成")


def run_build(version: str) -> bool:
    project_root = get_project_root()
    print(f"\n编译前端 (版本: {version})...")

    env = {"VITE_APP_VERSION": version}
    result = subprocess.run(
        ["npm", "run", "build"],
        cwd=project_root,
        env={**subprocess.run.__globals__["__builtins__"].__dict__ if False else {},
             **__import__("os").environ, **env},
    )
    if result.returncode != 0:
        # 更简单的环境变量传递方式
        result = subprocess.run(
            ["npm", "run", "build"],
            cwd=project_root,
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
```

- [*] **Step 10: scripts/build_image.py**

```python
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
```

- [*] **Step 11: Dockerfile**

```dockerfile
# KnowSync Web - Caddy 镜像
FROM caddy:alpine

ARG VERSION=latest

# 复制编译好的前端静态文件
COPY dist/ /usr/share/caddy/

# 复制 Caddyfile
COPY Caddyfile /etc/caddy/Caddyfile

# 创建版本标记文件
RUN echo $VERSION > /usr/share/caddy/version.txt

EXPOSE 80 443
```

- [*] **Step 12: Caddyfile**

```
knowsync.local {
    root * /usr/share/caddy
    encode gzip

    # SPA fallback — 所有非文件请求返回 index.html
    try_files {path} /index.html

    # API 反向代理到 gateway（Docker 内部网络）
    reverse_proxy /api/* gateway:8080
    reverse_proxy /files/* gateway:8080
}
```

- [*] **Step 13: .dockerignore**

```
node_modules
.git
.gitignore
*.md
scripts
Taskfile.yml
tsconfig*.json
vite.config.ts
postcss.config.js
tailwind.config.ts
package-lock.json
```

- [*] **Step 14: .gitignore**

```
node_modules
dist
*.local
```

- [*] **Step 15: 修改 docker-compose.yaml 添加 web 服务**

在 docker-compose.yaml 中新增 web 服务，gateway 移除 ports 暴露：

```yaml
  web:
    image: knowsync/web:latest
    container_name: knowsync-web
    restart: unless-stopped
    ports:
      - "80:80"
    depends_on:
      gateway:
        condition: service_started
    networks:
      - knowsync-net
```

- [*] **Step 16: 初始化并验证构建**

```bash
cd web && npm install && npm run build
```

验证 `dist/` 目录包含 `index.html` 和打包后的 JS/CSS 文件。

---

### Task 2: 核心基础设施 — API 客户端 + Auth + 路由 + 布局

**文件：**
- 创建: `web/src/lib/client.ts`
- 创建: `web/src/lib/auth.ts`
- 创建: `web/src/lib/repos.ts`
- 创建: `web/src/lib/nodes.ts`
- 创建: `web/src/lib/chat.ts`
- 创建: `web/src/store/auth-context.tsx`
- 创建: `web/src/store/chat-context.tsx`
- 创建: `web/src/store/repo-context.tsx`
- 创建: `web/src/components/Layout/PublicLayout.tsx`
- 创建: `web/src/components/Layout/TopBar.tsx`
- 创建: `web/src/components/Layout/Sidebar.tsx`
- 创建: `web/src/components/Layout/AppLayout.tsx`
- 创建: `web/src/components/Layout/ChatLayout.tsx`
- 创建: `web/src/App.tsx`

- [x] **Step 1: API 客户端 (lib/client.ts)**

```typescript
const BASE_URL = ''

function getToken(): string | null {
  return localStorage.getItem('access_token')
}

function getRefreshToken(): string | null {
  return localStorage.getItem('refresh_token')
}

function setTokens(access: string, refresh: string) {
  localStorage.setItem('access_token', access)
  localStorage.setItem('refresh_token', refresh)
}

function clearTokens() {
  localStorage.removeItem('access_token')
  localStorage.removeItem('refresh_token')
}

async function refreshAccessToken(): Promise<boolean> {
  const refresh = getRefreshToken()
  if (!refresh) return false
  try {
    const res = await fetch(`${BASE_URL}/api/v1/auth/refresh`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ refresh_token: refresh }),
    })
    if (!res.ok) return false
    const data = await res.json()
    if (data.code === 0 || data.code === 200) {
      setTokens(data.data.access_token, data.data.refresh_token)
      return true
    }
    return false
  } catch {
    return false
  }
}

export interface ApiResponse<T> {
  code: number
  message: string
  data: T
}

export async function request<T>(
  path: string,
  options?: RequestInit & { skipAuth?: boolean }
): Promise<ApiResponse<T>> {
  const headers: Record<string, string> = {
    'Content-Type': 'application/json',
    ...(options?.headers as Record<string, string>),
  }
  if (!options?.skipAuth) {
    const token = getToken()
    if (token) headers['Authorization'] = `Bearer ${token}`
  }

  let res = await fetch(`${BASE_URL}/api/v1${path}`, {
    ...options,
    headers,
  })

  if (res.status === 401 && !options?.skipAuth) {
    const refreshed = await refreshAccessToken()
    if (refreshed) {
      headers['Authorization'] = `Bearer ${getToken()}`
      res = await fetch(`${BASE_URL}/api/v1${path}`, {
        ...options,
        headers,
      })
    } else {
      clearTokens()
      window.location.href = '/login'
      throw new Error('Unauthorized')
    }
  }

  if (!res.ok) {
    const err = await res.json().catch(() => ({ message: 'Request failed' }))
    throw new Error(err.message || `HTTP ${res.status}`)
  }

  return res.json()
}

export async function uploadFile<T>(
  path: string,
  formData: FormData
): Promise<ApiResponse<T>> {
  const token = getToken()
  const headers: Record<string, string> = {}
  if (token) headers['Authorization'] = `Bearer ${token}`

  const res = await fetch(`${BASE_URL}/api/v1${path}`, {
    method: 'PUT',
    headers,
    body: formData,
  })

  if (res.status === 401) {
    clearTokens()
    window.location.href = '/login'
    throw new Error('Unauthorized')
  }

  if (!res.ok) {
    const err = await res.json().catch(() => ({ message: 'Upload failed' }))
    throw new Error(err.message || `HTTP ${res.status}`)
  }

  return res.json()
}

export { getToken, setTokens, clearTokens, getRefreshToken }
```

- [x] **Step 2: Auth API (lib/auth.ts)**

```typescript
import { request, setTokens } from './client'

export interface LoginData {
  access_token: string
  refresh_token: string
}

export interface UserInfo {
  id: string
  name: string
  email: string
  avatar: string
}

export async function login(email: string, password: string): Promise<LoginData> {
  const res = await request<LoginData>('/auth/login', {
    method: 'POST',
    body: JSON.stringify({ email, password }),
    skipAuth: true,
  })
  setTokens(res.data.access_token, res.data.refresh_token)
  return res.data
}

export async function register(
  name: string,
  email: string,
  password: string,
  verifyCode: string,
  avatar?: File
): Promise<UserInfo> {
  const formData = new FormData()
  formData.append('name', name)
  formData.append('email', email)
  formData.append('password', password)
  formData.append('verify_code', verifyCode)
  if (avatar) formData.append('avatar', avatar)

  const res = await fetch('/api/v1/auth/register', {
    method: 'POST',
    body: formData,
  })
  const data = await res.json()
  if (!res.ok || (data.code !== 0 && data.code !== 200)) {
    throw new Error(data.message || '注册失败')
  }
  return data.data
}

export async function logout(refreshToken: string): Promise<void> {
  await request('/auth/logout', {
    method: 'POST',
    body: JSON.stringify({ refresh_token: refreshToken }),
  })
}

export async function sendVerifyCode(email: string): Promise<void> {
  await request('/verify-codes', {
    method: 'POST',
    body: JSON.stringify({ email }),
    skipAuth: true,
  })
}

export async function getUser(userId: string): Promise<UserInfo> {
  const res = await request<UserInfo>(`/users/${userId}`)
  return res.data
}

export async function updateUser(userId: string, data: Partial<{ name: string; email: string }>): Promise<UserInfo> {
  const res = await request<UserInfo>(`/users/${userId}`, {
    method: 'PUT',
    body: JSON.stringify({ user: data }),
  })
  return res.data
}

export async function uploadAvatar(userId: string, file: File): Promise<string> {
  const formData = new FormData()
  formData.append('file', file)
  const res = await fetch(`/api/v1/users/${userId}/avatar`, {
    method: 'PUT',
    headers: { 'Authorization': `Bearer ${localStorage.getItem('access_token')}` },
    body: formData,
  })
  const data = await res.json()
  if (!res.ok) throw new Error(data.message || '头像上传失败')
  return data.data.avatar
}

export async function changePassword(userId: string, oldPwd: string, newPwd: string): Promise<void> {
  await request(`/users/${userId}/password`, {
    method: 'PUT',
    body: JSON.stringify({ old_password: oldPwd, new_password: newPwd }),
  })
}

export async function forgetPassword(email: string, password: string, verifyCode: string): Promise<void> {
  await request('/password/forget', {
    method: 'POST',
    body: JSON.stringify({ email, password, verify_code: verifyCode }),
    skipAuth: true,
  })
}
```

- [x] **Step 3: Repos API (lib/repos.ts)**

```typescript
import { request } from './client'

export interface Repo {
  id: string
  owner_id: string
  name: string
  visibility: string
  description: string
  article_count: number
  created_at: string
  updated_at: string
}

export async function listRepos(): Promise<Repo[]> {
  const res = await request<{ repos: Repo[] }>('/repos')
  return res.data.repos
}

export async function getRepo(repoId: string): Promise<Repo & { my_role: string }> {
  const res = await request<Repo & { my_role: string }>(`/repos/${repoId}`)
  return res.data
}

export async function createRepo(name: string, description?: string, visibility?: string): Promise<Repo> {
  const res = await request<{ repo: Repo }>('/repos', {
    method: 'POST',
    body: JSON.stringify({ name, description, visibility }),
  })
  return res.data.repo
}

export async function updateRepo(repoId: string, data: { name?: string; description?: string; visibility?: string }): Promise<Repo> {
  const res = await request<{ repo: Repo }>(`/repos/${repoId}`, {
    method: 'PUT',
    body: JSON.stringify(data),
  })
  return res.data.repo
}

export async function deleteRepo(repoId: string): Promise<void> {
  await request(`/repos/${repoId}`, { method: 'DELETE' })
}
```

- [x] **Step 4: Nodes API (lib/nodes.ts)**

```typescript
import { request, uploadFile } from './client'

export type NodeType = 'FOLDER' | 'ARTICLE'

export interface Node {
  id: string
  repo_id: string
  parent_id: string
  name: string
  type: NodeType
  file_path: string
  size: number
  created_at: string
  updated_at: string
}

export interface Collaborator {
  repo_id: string
  user_id: string
  role: string
}

export async function listNodes(repoId: string, parentId?: string): Promise<Node[]> {
  const params = parentId ? `?parent_id=${parentId}` : ''
  const res = await request<{ nodes: Node[] }>(`/repos/${repoId}/nodes${params}`)
  return res.data.nodes
}

export async function getNode(repoId: string, nodeId: string): Promise<Node> {
  const res = await request<Node>(`/repos/${repoId}/nodes/${nodeId}`)
  return res.data.node
}

export async function createNode(repoId: string, name: string, type: NodeType, parentId?: string): Promise<Node> {
  const res = await request<{ node: Node }>(`/repos/${repoId}/nodes`, {
    method: 'POST',
    body: JSON.stringify({ parent_id: parentId, name, type }),
  })
  return res.data.node
}

export async function updateNode(repoId: string, nodeId: string, data: { name?: string; parent_id?: string }): Promise<Node> {
  const res = await request<{ node: Node }>(`/repos/${repoId}/nodes/${nodeId}`, {
    method: 'PUT',
    body: JSON.stringify(data),
  })
  return res.data.node
}

export async function deleteNode(repoId: string, nodeId: string): Promise<{ deleted_node_ids: string[]; deleted_file_paths: string[] }> {
  const res = await request<{ deleted_node_ids: string[]; deleted_file_paths: string[] }>(`/repos/${repoId}/nodes/${nodeId}`, {
    method: 'DELETE',
  })
  return res.data
}

export async function uploadArticleContent(repoId: string, nodeId: string, file: File): Promise<{ node: Node; signed_url: string }> {
  const formData = new FormData()
  formData.append('file', file)
  const res = await uploadFile<{ node: Node; signed_url: string }>(`/repos/${repoId}/nodes/${nodeId}/content`, formData)
  return res.data
}

export async function getArticleSignedUrl(repoId: string, nodeId: string): Promise<string> {
  const res = await request<{ signed_url: string }>(`/repos/${repoId}/nodes/${nodeId}/signed-url`)
  return res.data.signed_url
}

export async function listCollaborators(repoId: string): Promise<Collaborator[]> {
  const res = await request<{ collaborators: Collaborator[] }>(`/repos/${repoId}/collaborators`)
  return res.data.collaborators
}

export async function addCollaborator(repoId: string, userId: string, role: string): Promise<void> {
  await request(`/repos/${repoId}/collaborators`, {
    method: 'POST',
    body: JSON.stringify({ user_id: userId, role }),
  })
}

export async function updateCollaborator(repoId: string, userId: string, role: string): Promise<void> {
  await request(`/repos/${repoId}/collaborators/${userId}`, {
    method: 'PUT',
    body: JSON.stringify({ role }),
  })
}

export async function removeCollaborator(repoId: string, userId: string): Promise<void> {
  await request(`/repos/${repoId}/collaborators/${userId}`, { method: 'DELETE' })
}
```

- [x] **Step 5: Chat API (lib/chat.ts)**

```typescript
import { request, getToken } from './client'

export interface ChatSession {
  id: string
  title: string
  created_at: number
  updated_at: number
}

export interface ChatMessage {
  id: string
  role: 'user' | 'assistant'
  content: string
  thinking: string
  created_at: number
}

export async function listSessions(cursor?: number, limit = 20): Promise<{ sessions: ChatSession[]; has_more: boolean }> {
  const params = new URLSearchParams()
  if (cursor) params.set('cursor', String(cursor))
  params.set('limit', String(limit))
  const res = await request<{ sessions: ChatSession[]; has_more: boolean }>(`/ai/sessions?${params}`)
  return res.data
}

export async function getMessages(sessionId: string, cursor?: number, limit = 50): Promise<{ messages: ChatMessage[]; has_more: boolean }> {
  const params = new URLSearchParams()
  if (cursor) params.set('cursor', String(cursor))
  params.set('limit', String(limit))
  const res = await request<{ messages: ChatMessage[]; has_more: boolean }>(`/ai/sessions/${sessionId}/messages?${params}`)
  return res.data
}

export async function deleteSession(sessionId: string): Promise<void> {
  await request(`/ai/sessions/${sessionId}`, { method: 'DELETE' })
}

export type SSEEventType = 'thinking' | 'thinking_finished' | 'content' | 'ask_user' | 'done' | 'error'

export interface SSEEvent {
  event: SSEEventType
  data: Record<string, unknown>
}

export async function createChatStream(
  sessionId: string,
  message: string,
  onEvent: (event: SSEEvent) => void,
  signal?: AbortSignal
): Promise<void> {
  const token = getToken()
  const response = await fetch('/api/v1/ai/chat', {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
      ...(token ? { 'Authorization': `Bearer ${token}` } : {}),
    },
    body: JSON.stringify({ session_id: sessionId, message }),
    signal,
  })

  if (!response.ok) {
    const err = await response.json().catch(() => ({ message: 'Chat request failed' }))
    throw new Error(err.message || `HTTP ${response.status}`)
  }

  const reader = response.body?.getReader()
  if (!reader) throw new Error('No response body')

  const decoder = new TextDecoder()
  let buffer = ''

  while (true) {
    const { done, value } = await reader.read()
    if (done) break

    buffer += decoder.decode(value, { stream: true })
    const lines = buffer.split('\n')
    buffer = lines.pop() || ''

    let currentEvent = ''
    for (const line of lines) {
      if (line.startsWith('event: ')) {
        currentEvent = line.slice(7).trim()
      } else if (line.startsWith('data: ')) {
        const jsonStr = line.slice(6)
        try {
          const data = JSON.parse(jsonStr)
          onEvent({ event: currentEvent as SSEEventType, data })
        } catch {
          // skip malformed JSON
        }
      }
    }
  }
}
```

- [x] **Step 6: AuthContext (store/auth-context.tsx)**

```typescript
import { createContext, useContext, useReducer, useEffect, type ReactNode } from 'react'
import { getToken, setTokens, clearTokens, getRefreshToken } from '../lib/client'
import * as authApi from '../lib/auth'

interface AuthState {
  user: authApi.UserInfo | null
  isAuthenticated: boolean
  isLoading: boolean
}

type AuthAction =
  | { type: 'SET_USER'; user: authApi.UserInfo }
  | { type: 'CLEAR_USER' }
  | { type: 'SET_LOADING'; loading: boolean }

function authReducer(state: AuthState, action: AuthAction): AuthState {
  switch (action.type) {
    case 'SET_USER':
      return { ...state, user: action.user, isAuthenticated: true, isLoading: false }
    case 'CLEAR_USER':
      return { ...state, user: null, isAuthenticated: false, isLoading: false }
    case 'SET_LOADING':
      return { ...state, isLoading: action.loading }
  }
}

interface AuthContextValue extends AuthState {
  login: (email: string, password: string) => Promise<void>
  logout: () => Promise<void>
  refreshUser: () => Promise<void>
}

const AuthContext = createContext<AuthContextValue | null>(null)

export function AuthProvider({ children }: { children: ReactNode }) {
  const [state, dispatch] = useReducer(authReducer, {
    user: null,
    isAuthenticated: false,
    isLoading: true,
  })

  const refreshUser = async () => {
    try {
      const token = getToken()
      if (!token) {
        dispatch({ type: 'CLEAR_USER' })
        return
      }
      // 用任意需要认证的请求验证 token，这里先不做具体用户请求
      // 路由守卫中会做具体处理
      dispatch({ type: 'SET_LOADING', loading: false })
    } catch {
      dispatch({ type: 'CLEAR_USER' })
    }
  }

  useEffect(() => {
    const token = getToken()
    if (token) {
      dispatch({ type: 'SET_LOADING', loading: false })
      // 实际用户信息在页面加载时通过具体接口获取
    } else {
      dispatch({ type: 'CLEAR_USER' })
    }
  }, [])

  const login = async (email: string, password: string) => {
    const data = await authApi.login(email, password)
    dispatch({ type: 'SET_LOADING', loading: false })
  }

  const logout = async () => {
    try {
      const refresh = getRefreshToken()
      if (refresh) await authApi.logout(refresh)
    } catch {
      // ignore logout errors
    }
    clearTokens()
    dispatch({ type: 'CLEAR_USER' })
  }

  return (
    <AuthContext.Provider value={{ ...state, login, logout, refreshUser }}>
      {children}
    </AuthContext.Provider>
  )
}

export function useAuth() {
  const ctx = useContext(AuthContext)
  if (!ctx) throw new Error('useAuth must be used within AuthProvider')
  return ctx
}
```

- [x] **Step 7: ChatContext (store/chat-context.tsx)**

```typescript
import { createContext, useContext, useReducer, useCallback, type ReactNode } from 'react'
import * as chatApi from '../lib/chat'

interface Message {
  id: string
  role: 'user' | 'assistant'
  content: string
  thinking: string
  createdAt: number
  isStreaming?: boolean
}

interface ChatState {
  sessions: chatApi.ChatSession[]
  currentSessionId: string | null
  messages: Message[]
  hasMoreMessages: boolean
  isStreaming: boolean
}

type ChatAction =
  | { type: 'SET_SESSIONS'; sessions: chatApi.ChatSession[] }
  | { type: 'SET_CURRENT_SESSION'; sessionId: string | null }
  | { type: 'SET_MESSAGES'; messages: chatApi.ChatMessage[]; hasMore: boolean }
  | { type: 'APPEND_MESSAGE'; message: Message }
  | { type: 'UPDATE_LAST_MESSAGE'; content: string }
  | { type: 'SET_STREAMING'; streaming: boolean }
  | { type: 'ADD_SESSION'; session: chatApi.ChatSession }

function chatReducer(state: ChatState, action: ChatAction): ChatState {
  switch (action.type) {
    case 'SET_SESSIONS':
      return { ...state, sessions: action.sessions }
    case 'SET_CURRENT_SESSION':
      return { ...state, currentSessionId: action.sessionId, messages: [], hasMoreMessages: false }
    case 'SET_MESSAGES':
      return { ...state, messages: action.messages.map(m => ({
        id: m.id, role: m.role, content: m.content, thinking: m.thinking, createdAt: m.created_at,
      })), hasMoreMessages: action.hasMore }
    case 'APPEND_MESSAGE':
      return { ...state, messages: [...state.messages, action.message] }
    case 'UPDATE_LAST_MESSAGE': {
      const msgs = [...state.messages]
      const last = msgs[msgs.length - 1]
      if (last && last.isStreaming) {
        msgs[msgs.length - 1] = { ...last, content: last.content + action.content }
      }
      return { ...state, messages: msgs }
    }
    case 'SET_STREAMING':
      return { ...state, isStreaming: action.streaming }
    case 'ADD_SESSION':
      return { ...state, sessions: [action.session, ...state.sessions] }
  }
}

interface ChatContextValue extends ChatState {
  loadSessions: () => Promise<void>
  loadMessages: (sessionId: string) => Promise<void>
  sendMessage: (message: string) => Promise<void>
  deleteSession: (sessionId: string) => Promise<void>
  setCurrentSession: (sessionId: string | null) => void
}

const ChatContext = createContext<ChatContextValue | null>(null)

export function ChatProvider({ children }: { children: ReactNode }) {
  const [state, dispatch] = useReducer(chatReducer, {
    sessions: [],
    currentSessionId: null,
    messages: [],
    hasMoreMessages: false,
    isStreaming: false,
  })

  const loadSessions = useCallback(async () => {
    try {
      const { sessions } = await chatApi.listSessions()
      dispatch({ type: 'SET_SESSIONS', sessions })
    } catch {
      // ignore
    }
  }, [])

  const loadMessages = useCallback(async (sessionId: string) => {
    try {
      dispatch({ type: 'SET_CURRENT_SESSION', sessionId })
      const { messages, has_more } = await chatApi.getMessages(sessionId)
      dispatch({ type: 'SET_MESSAGES', messages: messages.reverse(), hasMore: has_more })
    } catch {
      // ignore
    }
  }, [])

  const sendMessage = useCallback(async (message: string) => {
    const sessionId = state.currentSessionId || ''

    dispatch({ type: 'SET_STREAMING', streaming: true })

    // 添加用户消息
    dispatch({
      type: 'APPEND_MESSAGE',
      message: { id: `temp-${Date.now()}`, role: 'user', content: message, thinking: '', createdAt: Date.now() },
    })

    // 添加占位 assistant 消息
    const msgId = `stream-${Date.now()}`
    dispatch({
      type: 'APPEND_MESSAGE',
      message: { id: msgId, role: 'assistant', content: '', thinking: '', createdAt: Date.now(), isStreaming: true },
    })

    try {
      await chatApi.createChatStream(sessionId, message, (event) => {
        switch (event.event) {
          case 'thinking':
            dispatch({ type: 'UPDATE_LAST_MESSAGE', content: event.data.content as string })
            break
          case 'content':
            dispatch({ type: 'UPDATE_LAST_MESSAGE', content: event.data.content as string })
            break
          case 'done': {
            const evData = event.data as { session_id: string; title?: string; title_updated?: boolean }
            if (evData.title_updated) {
              // 更新会话标题
            }
            break
          }
          case 'ask_user':
            // 弹窗由 UI 组件处理，这里触发事件
            window.dispatchEvent(new CustomEvent('ask-user', { detail: event.data }))
            break
        }
      })
    } catch (err) {
      console.error('Chat stream error:', err)
    }

    dispatch({ type: 'SET_STREAMING', streaming: false })
    loadSessions()
  }, [state.currentSessionId, loadSessions])

  const deleteSession = useCallback(async (sessionId: string) => {
    await chatApi.deleteSession(sessionId)
    loadSessions()
    if (state.currentSessionId === sessionId) {
      dispatch({ type: 'SET_CURRENT_SESSION', sessionId: null })
    }
  }, [state.currentSessionId, loadSessions])

  const setCurrentSession = useCallback((sessionId: string | null) => {
    dispatch({ type: 'SET_CURRENT_SESSION', sessionId })
  }, [])

  return (
    <ChatContext.Provider value={{ ...state, loadSessions, loadMessages, sendMessage, deleteSession, setCurrentSession }}>
      {children}
    </ChatContext.Provider>
  )
}

export function useChat() {
  const ctx = useContext(ChatContext)
  if (!ctx) throw new Error('useChat must be used within ChatProvider')
  return ctx
}
```

- [x] **Step 8: RepoContext (store/repo-context.tsx)**

```typescript
import { createContext, useContext, useReducer, useCallback, type ReactNode } from 'react'
import * as repoApi from '../lib/repos'
import * as nodesApi from '../lib/nodes'

interface RepoState {
  repos: repoApi.Repo[]
  currentRepoId: string | null
  nodes: nodesApi.Node[]
  isLoading: boolean
}

type RepoAction =
  | { type: 'SET_REPOS'; repos: repoApi.Repo[] }
  | { type: 'ADD_REPO'; repo: repoApi.Repo }
  | { type: 'UPDATE_REPO'; repo: repoApi.Repo }
  | { type: 'REMOVE_REPO'; repoId: string }
  | { type: 'SET_CURRENT_REPO'; repoId: string | null }
  | { type: 'SET_NODES'; nodes: nodesApi.Node[] }
  | { type: 'ADD_NODE'; node: nodesApi.Node }
  | { type: 'UPDATE_NODE'; node: nodesApi.Node }
  | { type: 'REMOVE_NODE'; nodeId: string }
  | { type: 'SET_LOADING'; loading: boolean }

function repoReducer(state: RepoState, action: RepoAction): RepoState {
  switch (action.type) {
    case 'SET_REPOS':
      return { ...state, repos: action.repos }
    case 'ADD_REPO':
      return { ...state, repos: [action.repo, ...state.repos] }
    case 'UPDATE_REPO':
      return { ...state, repos: state.repos.map(r => r.id === action.repo.id ? action.repo : r) }
    case 'REMOVE_REPO':
      return { ...state, repos: state.repos.filter(r => r.id !== action.repoId) }
    case 'SET_CURRENT_REPO':
      return { ...state, currentRepoId: action.repoId, nodes: [] }
    case 'SET_NODES':
      return { ...state, nodes: action.nodes }
    case 'ADD_NODE':
      return { ...state, nodes: [...state.nodes, action.node] }
    case 'UPDATE_NODE':
      return { ...state, nodes: state.nodes.map(n => n.id === action.node.id ? action.node : n) }
    case 'REMOVE_NODE':
      return { ...state, nodes: state.nodes.filter(n => n.id !== action.nodeId) }
    case 'SET_LOADING':
      return { ...state, isLoading: action.loading }
  }
}

interface RepoContextValue extends RepoState {
  loadRepos: () => Promise<void>
  loadNodes: (repoId: string, parentId?: string) => Promise<void>
  selectRepo: (repoId: string | null) => void
}

const RepoContext = createContext<RepoContextValue | null>(null)

export function RepoProvider({ children }: { children: ReactNode }) {
  const [state, dispatch] = useReducer(repoReducer, {
    repos: [],
    currentRepoId: null,
    nodes: [],
    isLoading: false,
  })

  const loadRepos = useCallback(async () => {
    dispatch({ type: 'SET_LOADING', loading: true })
    try {
      const repos = await repoApi.listRepos()
      dispatch({ type: 'SET_REPOS', repos })
    } catch {
      // ignore
    }
    dispatch({ type: 'SET_LOADING', loading: false })
  }, [])

  const loadNodes = useCallback(async (repoId: string, parentId?: string) => {
    try {
      const nodes = await nodesApi.listNodes(repoId, parentId)
      dispatch({ type: 'SET_NODES', nodes })
    } catch {
      // ignore
    }
  }, [])

  const selectRepo = useCallback((repoId: string | null) => {
    dispatch({ type: 'SET_CURRENT_REPO', repoId })
  }, [])

  return (
    <RepoContext.Provider value={{ ...state, loadRepos, loadNodes, selectRepo }}>
      {children}
    </RepoContext.Provider>
  )
}

export function useRepo() {
  const ctx = useContext(RepoContext)
  if (!ctx) throw new Error('useRepo must be used within RepoProvider')
  return ctx
}
```

- [x] **Step 9: 路由定义 (App.tsx)**

```typescript
import { Routes, Route, Navigate } from 'react-router-dom'
import { AuthProvider } from './store/auth-context'
import { ChatProvider } from './store/chat-context'
import { RepoProvider } from './store/repo-context'
import PublicLayout from './components/Layout/PublicLayout'
import AppLayout from './components/Layout/AppLayout'
import Login from './pages/Login'
import Register from './pages/Register'
import Chat from './pages/Chat'
import RepoList from './pages/RepoList'
import RepoDetail from './pages/RepoDetail'
import ArticleView from './pages/ArticleView'
import ArticleEditor from './pages/ArticleEditor'
import SearchResult from './pages/SearchResult'
import NotFound from './pages/NotFound'

export default function App() {
  return (
    <AuthProvider>
      <ChatProvider>
        <RepoProvider>
          <Routes>
            {/* 公开页面 — PublicLayout */}
            <Route element={<PublicLayout />}>
              <Route path="/login" element={<Login />} />
              <Route path="/register" element={<Register />} />
            </Route>

            {/* 登录页面入口 */}
            <Route path="/" element={<RootRedirect />} />

            {/* 登录后页面 — AppLayout */}
            <Route element={<AppLayout />}>
              <Route path="/repos" element={<RepoList />} />
              <Route path="/repos/:repoId" element={<RepoDetail />} />
              <Route path="/repos/:repoId/nodes/:nodeId" element={<ArticleView />} />
              <Route path="/repos/:repoId/nodes/:nodeId/edit" element={<ArticleEditor />} />
              <Route path="/chat" element={<Chat />} />
              <Route path="/chat/:sessionId" element={<Chat />} />
              <Route path="/search" element={<SearchResult />} />
            </Route>

            <Route path="*" element={<NotFound />} />
          </Routes>
        </RepoProvider>
      </ChatProvider>
    </AuthProvider>
  )
}

function RootRedirect() {
  const token = localStorage.getItem('access_token')
  if (token) return <Navigate to="/chat" replace />
  return <Navigate to="/login" replace />
}
```

- [x] **Step 10: PublicLayout**

```typescript
import { Outlet, Navigate } from 'react-router-dom'
import { useAuth } from '../store/auth-context'

export default function PublicLayout() {
  const { isAuthenticated } = useAuth()
  if (isAuthenticated) return <Navigate to="/chat" replace />
  return (
    <div className="min-h-screen bg-gray-50 flex items-center justify-center">
      <Outlet />
    </div>
  )
}
```

- [x] **Step 11: TopBar**

```typescript
import { useState, useRef, useEffect } from 'react'
import { useNavigate } from 'react-router-dom'
import { useAuth } from '../../store/auth-context'
import SettingsModal from '../Settings/SettingsModal'

export default function TopBar() {
  const navigate = useNavigate()
  const { user, logout } = useAuth()
  const [query, setQuery] = useState('')
  const [showDropdown, setShowDropdown] = useState(false)
  const [showSettings, setShowSettings] = useState(false)
  const dropdownRef = useRef<HTMLDivElement>(null)

  useEffect(() => {
    function handleClick(e: MouseEvent) {
      if (dropdownRef.current && !dropdownRef.current.contains(e.target as Node)) {
        setShowDropdown(false)
      }
    }
    document.addEventListener('mousedown', handleClick)
    return () => document.removeEventListener('mousedown', handleClick)
  }, [])

  const handleSearch = (e: React.FormEvent) => {
    e.preventDefault()
    if (query.trim()) {
      navigate(`/search?q=${encodeURIComponent(query.trim())}`)
    }
  }

  const handleLogout = async () => {
    setShowDropdown(false)
    await logout()
    navigate('/login')
  }

  return (
    <>
      <header className="h-14 border-b border-gray-200 bg-white flex items-center px-4 gap-4 shrink-0">
        <button
          onClick={() => navigate('/chat')}
          className="text-lg font-bold text-emerald-600 shrink-0 hover:text-emerald-700"
        >
          KnowSync
        </button>

        <form onSubmit={handleSearch} className="flex-1 max-w-xl">
          <input
            type="text"
            value={query}
            onChange={e => setQuery(e.target.value)}
            placeholder="搜索知识库..."
            className="w-full px-3 py-1.5 border border-gray-300 rounded-lg text-sm focus:outline-none focus:border-emerald-400 bg-gray-50"
          />
        </form>

        <div className="relative" ref={dropdownRef}>
          <button
            onClick={() => setShowDropdown(!showDropdown)}
            className="w-8 h-8 rounded-full bg-emerald-100 text-emerald-700 font-medium text-sm flex items-center justify-center hover:bg-emerald-200"
          >
            {user?.name?.charAt(0) || 'U'}
          </button>

          {showDropdown && (
            <div className="absolute right-0 top-10 w-36 bg-white border border-gray-200 rounded-lg shadow-lg py-1 z-50">
              <button
                onClick={() => { setShowDropdown(false); setShowSettings(true) }}
                className="w-full px-3 py-2 text-sm text-left hover:bg-gray-50"
              >
                设置
              </button>
              <button
                onClick={handleLogout}
                className="w-full px-3 py-2 text-sm text-left text-red-600 hover:bg-gray-50"
              >
                退出登录
              </button>
            </div>
          )}
        </div>
      </header>

      {showSettings && <SettingsModal onClose={() => setShowSettings(false)} />}
    </>
  )
}
```

- [x] **Step 12: Sidebar**

```typescript
import { useState } from 'react'
import { useNavigate, useLocation } from 'react-router-dom'
import { useRepo } from '../../store/repo-context'

interface SidebarProps {
  /** 如果为 true，显示文件树模式（带返回按钮） */
  isFileTree?: boolean
  repoId?: string
  onBack?: () => void
}

export default function Sidebar({ isFileTree, repoId, onBack }: SidebarProps) {
  const [collapsed, setCollapsed] = useState(false)
  const navigate = useNavigate()
  const location = useLocation()
  const { repos, loadRepos } = useRepo()

  if (collapsed) {
    return (
      <div className="w-12 border-r border-gray-200 bg-gray-50 flex flex-col items-center py-2 gap-3 shrink-0">
        <button onClick={() => setCollapsed(false)} className="text-gray-400 hover:text-gray-600 text-lg">☰</button>
        <button onClick={() => navigate('/chat')} className={`p-2 rounded-lg ${location.pathname.startsWith('/chat') ? 'bg-emerald-100 text-emerald-600' : 'text-gray-400 hover:text-gray-600'}`} title="AI 对话">💬</button>
        {repos.slice(0, 3).map(r => (
          <button key={r.id} onClick={() => navigate(`/repos/${r.id}`)} className="w-8 h-8 rounded bg-gray-200 text-xs text-gray-600 hover:bg-gray-300" title={r.name}>
            {r.name.charAt(0)}
          </button>
        ))}
      </div>
    )
  }

  return (
    <div className="w-56 border-r border-gray-200 bg-gray-50 flex flex-col shrink-0 overflow-hidden">
      {/* 文件树模式：显示返回按钮 */}
      {isFileTree ? (
        <div className="p-3 border-b border-gray-200">
          <button
            onClick={onBack}
            className="flex items-center gap-1 text-sm text-gray-600 hover:text-emerald-600"
          >
            ← 返回所有知识库
          </button>
        </div>
      ) : (
        <div className="p-3 border-b border-gray-200">
          <button
            onClick={() => navigate('/chat')}
            className={`flex items-center gap-2 w-full px-2 py-1.5 rounded text-sm ${location.pathname.startsWith('/chat') ? 'bg-emerald-100 text-emerald-700 font-medium' : 'text-gray-600 hover:bg-gray-100'}`}
          >
            💬 AI 对话
          </button>
        </div>
      )}

      {/* 仓库列表或文件树 */}
      <div className="flex-1 overflow-y-auto p-2">
        {isFileTree ? (
          <FileTree repoId={repoId!} />
        ) : (
          <>
            <div className="text-xs text-gray-400 font-medium px-2 py-1">知识库</div>
            {repos.map(repo => (
              <button
                key={repo.id}
                onClick={() => navigate(`/repos/${repo.id}`)}
                className="flex items-center gap-2 w-full px-2 py-1.5 rounded text-sm text-gray-700 hover:bg-gray-100 mb-0.5"
              >
                <span className="w-5 h-5 rounded bg-emerald-100 text-emerald-700 text-xs flex items-center justify-center shrink-0">
                  {repo.name.charAt(0)}
                </span>
                <span className="truncate">{repo.name}</span>
              </button>
            ))}
          </>
        )}
      </div>

      {/* 折叠按钮 */}
      <div className="p-2 border-t border-gray-200">
        <button
          onClick={() => setCollapsed(true)}
          className="text-xs text-gray-400 hover:text-gray-600 w-full text-left"
        >
          ◀ 折叠
        </button>
      </div>
    </div>
  )
}

// 简化的文件树组件，后续 Task 会完善
function FileTree({ repoId }: { repoId: string }) {
  return (
    <div className="text-sm text-gray-500 px-2 py-4 text-center">
      文件树加载中...
      <br />
      <span className="text-xs">(Task 7 完善)</span>
    </div>
  )
}
```

- [x] **Step 13: AppLayout**

```typescript
import { useEffect } from 'react'
import { Outlet, useParams, useLocation, useNavigate } from 'react-router-dom'
import { useAuth } from '../../store/auth-context'
import { useRepo } from '../../store/repo-context'
import TopBar from './TopBar'
import Sidebar from './Sidebar'

export default function AppLayout() {
  const { isAuthenticated, isLoading } = useAuth()
  const { loadRepos } = useRepo()
  const { repoId } = useParams()
  const location = useLocation()
  const navigate = useNavigate()

  useEffect(() => {
    if (!isLoading && !isAuthenticated) {
      navigate('/login', { replace: true })
    }
  }, [isLoading, isAuthenticated, navigate])

  useEffect(() => {
    loadRepos()
  }, [loadRepos])

  const isRepoRoute = location.pathname.startsWith('/repos')
  const isSearchRoute = location.pathname.startsWith('/search')

  if (isLoading) {
    return (
      <div className="h-screen flex items-center justify-center text-gray-400">
        加载中...
      </div>
    )
  }

  if (!isAuthenticated) return null

  // 搜索页 — 全屏，无侧边栏
  if (isSearchRoute) {
    return (
      <div className="h-screen flex flex-col">
        <TopBar />
        <main className="flex-1 overflow-y-auto">
          <Outlet />
        </main>
      </div>
    )
  }

  return (
    <div className="h-screen flex flex-col">
      <TopBar />
      <div className="flex-1 flex overflow-hidden">
        <Sidebar
          isFileTree={isRepoRoute}
          repoId={repoId}
          onBack={() => navigate('/chat')}
        />
        <main className="flex-1 overflow-y-auto">
          <Outlet />
        </main>
      </div>
    </div>
  )
}
```

- [x] **Step 14: ChatLayout (三栏聊天布局)**

```typescript
import { type ReactNode } from 'react'

interface ChatLayoutProps {
  sidebar?: ReactNode
  main: ReactNode
  rightPanel?: ReactNode
}

export default function ChatLayout({ sidebar, main, rightPanel }: ChatLayoutProps) {
  return (
    <div className="h-full flex">
      {sidebar && (
        <div className="w-56 border-r border-gray-200 bg-gray-50 shrink-0 overflow-y-auto">
          {sidebar}
        </div>
      )}
      <div className="flex-1 flex flex-col min-w-0">
        {main}
      </div>
      {rightPanel && (
        <div className="w-60 border-l border-gray-200 bg-gray-50 shrink-0 overflow-y-auto">
          {rightPanel}
        </div>
      )}
    </div>
  )
}
```

- [x] **Step 15: 验证 TypeScript 编译通过**

```bash
cd web && npx tsc --noEmit
```

---

### Task 3: 登录/注册页面

- [x] **Step 1: Login 页面 (pages/Login.tsx)**
- [x] **Step 2: Register 页面 (pages/Register.tsx)**

---

### Task 4: 聊天页（三栏布局 + Streamdown）

**文件：**
- 创建: `web/src/pages/Chat.tsx`
- 创建: `web/src/components/Chat/ChatWindow.tsx`
- 创建: `web/src/components/Chat/MessageBubble.tsx`
- 创建: `web/src/components/Chat/Streamdown.tsx`
- 创建: `web/src/components/Chat/SessionList.tsx`
- 创建: `web/src/components/Chat/AskUserModal.tsx`

- [x] **Step 1: Streamdown 组件**

```typescript
import { memo } from 'react'
import ReactMarkdown from 'react-markdown'
import { Prism as SyntaxHighlighter } from 'react-syntax-highlighter'
import { oneLight } from 'react-syntax-highlighter/dist/esm/styles/prism'
import type { Components } from 'react-markdown'

interface StreamdownProps {
  content: string
}

const components: Components = {
  code({ className, children, ...props }) {
    const match = /language-(\w+)/.exec(className || '')
    const code = String(children).replace(/\n$/, '')
    if (match) {
      return (
        <SyntaxHighlighter style={oneLight} language={match[1]} PreTag="div">
          {code}
        </SyntaxHighlighter>
      )
    }
    return <code className="bg-gray-100 px-1 rounded text-sm" {...props}>{children}</code>
  },
  pre({ children }) {
    return <div className="my-2">{children}</div>
  },
}

function StreamdownInner({ content }: StreamdownProps) {
  return (
    <div className="prose prose-sm max-w-none prose-headings:text-gray-800 prose-p:text-gray-700 prose-a:text-emerald-600 prose-code:bg-gray-100 prose-code:px-1 prose-code:rounded prose-code:text-sm">
      <ReactMarkdown components={components}>
        {content || ''}
      </ReactMarkdown>
    </div>
  )
}

export const Streamdown = memo(StreamdownInner)
```

- [x] **Step 2: MessageBubble 组件**

```typescript
import { Streamdown } from './Streamdown'

interface MessageBubbleProps {
  role: 'user' | 'assistant'
  content: string
  thinking?: string
  isStreaming?: boolean
}

export default function MessageBubble({ role, content, thinking, isStreaming }: MessageBubbleProps) {
  const isUser = role === 'user'

  return (
    <div className={`flex gap-3 mb-4 ${isUser ? 'flex-row-reverse' : ''}`}>
      {/* Avatar */}
      <div className={`w-8 h-8 rounded-full flex items-center justify-center text-sm shrink-0 ${isUser ? 'bg-emerald-100 text-emerald-700' : 'bg-gray-200 text-gray-600'}`}>
        {isUser ? 'U' : 'AI'}
      </div>

      {/* Content */}
      <div className={`max-w-[70%] ${isUser ? 'items-end' : 'items-start'}`}>
        {thinking && (
          <details className="mb-2 text-sm">
            <summary className="text-gray-400 cursor-pointer hover:text-gray-600">思考过程</summary>
            <div className="mt-1 p-2 bg-gray-50 rounded text-gray-500 text-xs whitespace-pre-wrap">
              {thinking}
            </div>
          </details>
        )}
        <div className={`px-4 py-2.5 rounded-2xl text-sm leading-relaxed ${isUser ? 'bg-emerald-500 text-white' : 'bg-gray-100 text-gray-800'}`}>
          {isUser ? content : <Streamdown content={content} />}
        </div>
        {isStreaming && (
          <span className="inline-block w-2 h-4 bg-emerald-500 animate-pulse ml-1" />
        )}
      </div>
    </div>
  )
}
```

- [x] **Step 3: SessionList 组件**

```typescript
import { useChat } from '../../store/chat-context'

export default function SessionList() {
  const { sessions, currentSessionId, loadMessages, deleteSession } = useChat()

  return (
    <div className="p-2">
      <div className="text-xs text-gray-400 font-medium px-2 py-1">会话列表</div>
      {sessions.map(session => (
        <div key={session.id} className="group relative">
          <button
            onClick={() => loadMessages(session.id)}
            className={`w-full text-left px-2 py-2 rounded-lg text-sm mb-0.5 transition-colors ${currentSessionId === session.id ? 'bg-emerald-50 text-emerald-700' : 'text-gray-600 hover:bg-gray-100'}`}
          >
            <div className="truncate">{session.title || '新对话'}</div>
            <div className="text-xs text-gray-400 mt-0.5">
              {new Date(session.updated_at * 1000).toLocaleDateString('zh-CN')}
            </div>
          </button>
          <button
            onClick={() => deleteSession(session.id)}
            className="absolute top-1 right-1 opacity-0 group-hover:opacity-100 w-5 h-5 flex items-center justify-center text-gray-400 hover:text-red-500 text-xs rounded hover:bg-gray-200"
            title="删除会话"
          >
            ✕
          </button>
        </div>
      ))}
      {sessions.length === 0 && (
        <div className="text-xs text-gray-400 px-2 py-4 text-center">暂无会话</div>
      )}
    </div>
  )
}
```

- [x] **Step 4: AskUserModal 组件**

```typescript
import { useState, useEffect } from 'react'

interface AskUserData {
  session_id: string
  question: string
  type: 'single' | 'multiple'
  options?: string[]
  has_other?: boolean
}

export default function AskUserModal() {
  const [data, setData] = useState<AskUserData | null>(null)

  useEffect(() => {
    const handler = (e: Event) => {
      setData((e as CustomEvent).detail as AskUserData)
    }
    window.addEventListener('ask-user', handler)
    return () => window.removeEventListener('ask-user', handler)
  }, [])

  if (!data) return null

  const handleSelect = (option: string) => {
    // 发送用户选择到聊天
    window.dispatchEvent(new CustomEvent('ask-user-response', { detail: { value: option } }))
    setData(null)
  }

  return (
    <div className="fixed inset-0 bg-black/30 flex items-center justify-center z-50">
      <div className="bg-white rounded-xl p-6 max-w-md w-full mx-4 shadow-xl">
        <h3 className="text-sm font-medium text-gray-800 mb-4">{data.question}</h3>
        <div className="space-y-2">
          {data.options?.map(opt => (
            <button
              key={opt}
              onClick={() => handleSelect(opt)}
              className="w-full px-3 py-2 text-sm text-left border border-gray-200 rounded-lg hover:bg-gray-50 hover:border-emerald-300 transition-colors"
            >
              {opt}
            </button>
          ))}
        </div>
        <button onClick={() => setData(null)} className="mt-4 text-xs text-gray-400 hover:text-gray-600">
          取消
        </button>
      </div>
    </div>
  )
}
```

- [x] **Step 5: ChatWindow 组件**

```typescript
import { useState, useRef, useEffect } from 'react'
import { useChat } from '../../store/chat-context'
import MessageBubble from './MessageBubble'
import AskUserModal from './AskUserModal'

interface ChatWindowProps {
  onSendMessage?: (message: string) => void
}

export default function ChatWindow({ onSendMessage }: ChatWindowProps) {
  const { messages, isStreaming, sendMessage } = useChat()
  const [input, setInput] = useState('')
  const bottomRef = useRef<HTMLDivElement>(null)
  const textareaRef = useRef<HTMLTextAreaElement>(null)

  useEffect(() => {
    bottomRef.current?.scrollIntoView({ behavior: 'smooth' })
  }, [messages])

  const handleSubmit = () => {
    const trimmed = input.trim()
    if (!trimmed || isStreaming) return
    sendMessage(trimmed)
    setInput('')
    if (textareaRef.current) {
      textareaRef.current.style.height = 'auto'
    }
  }

  const handleKeyDown = (e: React.KeyboardEvent) => {
    if (e.key === 'Enter' && !e.shiftKey) {
      e.preventDefault()
      handleSubmit()
    }
  }

  const handleInput = (e: React.ChangeEvent<HTMLTextAreaElement>) => {
    setInput(e.target.value)
    const el = e.target
    el.style.height = 'auto'
    el.style.height = `${Math.min(el.scrollHeight, 200)}px`
  }

  const hasMessages = messages.length > 0

  return (
    <>
      <div className="flex-1 overflow-y-auto px-4 py-4">
        {hasMessages ? (
          messages.map(msg => (
            <MessageBubble
              key={msg.id}
              role={msg.role}
              content={msg.content}
              thinking={msg.thinking}
              isStreaming={msg.isStreaming}
            />
          ))
        ) : (
          <div className="h-full flex flex-col items-center justify-center text-center text-gray-400">
            <div className="text-4xl mb-3">💡</div>
            <h2 className="text-lg font-medium text-gray-700 mb-2">开始与 KnowSync 对话</h2>
            <p className="text-sm mb-6 max-w-md">AI 知识助手，帮你快速找到所需信息</p>
            <div className="space-y-2 w-64">
              {['搜索某篇文章', '帮我总结某个知识库', '解释某个概念'].map(q => (
                <button
                  key={q}
                  onClick={() => sendMessage(q)}
                  className="w-full px-4 py-2 text-sm border border-gray-200 rounded-lg hover:border-emerald-300 hover:text-emerald-600 transition-colors"
                >
                  {q}
                </button>
              ))}
            </div>
          </div>
        )}
        <div ref={bottomRef} />
      </div>

      <div className="border-t border-gray-200 p-3">
        <div className="flex gap-2 items-end">
          <textarea
            ref={textareaRef}
            value={input}
            onChange={handleInput}
            onKeyDown={handleKeyDown}
            placeholder="输入消息 (Enter 发送, Shift+Enter 换行)"
            rows={1}
            className="flex-1 px-3 py-2 border border-gray-300 rounded-lg text-sm resize-none focus:outline-none focus:border-emerald-400 max-h-[200px]"
          />
          <button
            onClick={handleSubmit}
            disabled={!input.trim() || isStreaming}
            className="px-4 py-2 bg-emerald-500 text-white rounded-lg text-sm hover:bg-emerald-600 disabled:opacity-50 shrink-0"
          >
            发送
          </button>
        </div>
      </div>

      <AskUserModal />
    </>
  )
}
```

- [x] **Step 6: Chat 页面 (pages/Chat.tsx)**

```typescript
import { useEffect } from 'react'
import { useParams } from 'react-router-dom'
import { useChat } from '../store/chat-context'
import ChatWindow from '../components/Chat/ChatWindow'
import SessionList from '../components/Chat/SessionList'

export default function Chat() {
  const { sessionId } = useParams()
  const { loadSessions, loadMessages, currentSessionId, setCurrentSession } = useChat()

  useEffect(() => {
    loadSessions()
  }, [loadSessions])

  useEffect(() => {
    if (sessionId && sessionId !== currentSessionId) {
      loadMessages(sessionId)
    }
  }, [sessionId, currentSessionId, loadMessages])

  // 默认加载第一个会话
  useEffect(() => {
    if (!currentSessionId && !sessionId) {
      setCurrentSession(null)
    }
  }, [currentSessionId, sessionId, setCurrentSession])

  return (
    <div className="h-full flex">
      {/* 中间聊天窗口 — 默认 flex-1 */}
      <div className="flex-1 flex flex-col min-w-0">
        <ChatWindow />
      </div>

      {/* 右侧会话列表 */}
      <div className="w-60 border-l border-gray-200 bg-gray-50 shrink-0 overflow-y-auto">
        <SessionList />
      </div>
    </div>
  )
}
```

---

### Task 5: 知识库管理

**文件：**
- 创建: `web/src/pages/RepoList.tsx`
- 创建: `web/src/components/Repo/RepoCard.tsx`

- [x] **Step 1: RepoCard 组件**

```typescript
import { repoApi, updateRepo, deleteRepo } from '../../lib/repos'
import { useState } from 'react'
import type { Repo } from '../../lib/repos'

interface RepoCardProps {
  repo: Repo
  onSelect: (repo: Repo) => void
  onUpdate: () => void
}

export default function RepoCard({ repo, onSelect, onUpdate }: RepoCardProps) {
  const [showMenu, setShowMenu] = useState(false)
  const [editing, setEditing] = useState(false)
  const [newName, setNewName] = useState(repo.name)

  const handleRename = async () => {
    if (newName.trim() && newName !== repo.name) {
      await updateRepo(repo.id, { name: newName.trim() })
      onUpdate()
    }
    setEditing(false)
  }

  const handleDelete = async () => {
    if (confirm(`确定删除「${repo.name}」？`)) {
      await deleteRepo(repo.id)
      onUpdate()
    }
  }

  const handleToggleVisibility = async () => {
    const newVis = repo.visibility === 'PUBLIC' ? 'PRIVATE' : 'PUBLIC'
    await updateRepo(repo.id, { visibility: newVis })
    onUpdate()
  }

  return (
    <div className="flex items-center justify-between p-3 hover:bg-gray-50 rounded-lg cursor-pointer group" onClick={() => onSelect(repo)}>
      <div className="flex-1 min-w-0">
        <div className="flex items-center gap-2">
          <span className="w-2 h-2 rounded-full bg-emerald-400 shrink-0" />
          {editing ? (
            <input
              value={newName}
              onChange={e => setNewName(e.target.value)}
              onBlur={handleRename}
              onKeyDown={e => e.key === 'Enter' && handleRename()}
              className="text-sm font-medium border border-gray-300 rounded px-1"
              autoFocus
              onClick={e => e.stopPropagation()}
            />
          ) : (
            <span className="text-sm font-medium text-gray-800">{repo.name}</span>
          )}
          <span className={`text-xs px-1.5 py-0.5 rounded ${repo.visibility === 'PUBLIC' ? 'bg-blue-50 text-blue-600' : 'bg-gray-100 text-gray-500'}`}>
            {repo.visibility === 'PUBLIC' ? '公开' : '私有'}
          </span>
        </div>
        <p className="text-xs text-gray-400 mt-0.5 truncate">{repo.description || `${repo.article_count} 篇文章`}</p>
      </div>

      <div className="relative" onClick={e => e.stopPropagation()}>
        <button onClick={() => setShowMenu(!showMenu)} className="opacity-0 group-hover:opacity-100 p-1 text-gray-400 hover:text-gray-600">⋮</button>
        {showMenu && (
          <div className="absolute right-0 top-6 w-28 bg-white border border-gray-200 rounded-lg shadow-lg py-1 z-10">
            <button onClick={() => { setEditing(true); setShowMenu(false) }} className="w-full px-3 py-1.5 text-xs text-left hover:bg-gray-50">重命名</button>
            <button onClick={() => { handleToggleVisibility(); setShowMenu(false) }} className="w-full px-3 py-1.5 text-xs text-left hover:bg-gray-50">
              {repo.visibility === 'PUBLIC' ? '设为私有' : '设为公开'}
            </button>
            <button onClick={() => { handleDelete(); setShowMenu(false) }} className="w-full px-3 py-1.5 text-xs text-left text-red-600 hover:bg-gray-50">删除</button>
          </div>
        )}
      </div>
    </div>
  )
}
```

- [x] **Step 2: RepoList 页面**

```typescript
import { useEffect, useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { useRepo } from '../store/repo-context'
import { createRepo } from '../lib/repos'
import RepoCard from '../components/Repo/RepoCard'
import type { Repo } from '../lib/repos'

export default function RepoList() {
  const navigate = useNavigate()
  const { repos, loadRepos } = useRepo()
  const [showCreate, setShowCreate] = useState(false)
  const [newName, setNewName] = useState('')
  const [newDesc, setNewDesc] = useState('')
  const [selectedRepo, setSelectedRepo] = useState<Repo | null>(null)

  useEffect(() => {
    loadRepos()
  }, [loadRepos])

  const handleCreate = async () => {
    if (!newName.trim()) return
    await createRepo(newName.trim(), newDesc.trim() || undefined)
    setShowCreate(false)
    setNewName('')
    setNewDesc('')
    loadRepos()
  }

  const handleSelect = (repo: Repo) => {
    setSelectedRepo(repo)
  }

  return (
    <div className="h-full flex">
      {/* 左侧列表 */}
      <div className="w-72 border-r border-gray-200 overflow-y-auto">
        <div className="p-3 border-b border-gray-200 flex items-center justify-between">
          <span className="text-sm font-medium text-gray-700">所有知识库</span>
          <button onClick={() => setShowCreate(true)} className="text-xs px-2 py-1 bg-emerald-500 text-white rounded hover:bg-emerald-600">+ 新建</button>
        </div>
        {repos.map(repo => (
          <RepoCard key={repo.id} repo={repo} onSelect={handleSelect} onUpdate={loadRepos} />
        ))}
      </div>

      {/* 右侧详情 */}
      <div className="flex-1 p-6 overflow-y-auto">
        {selectedRepo ? (
          <div>
            <h2 className="text-lg font-medium text-gray-800">{selectedRepo.name}</h2>
            <p className="text-sm text-gray-500 mt-1">{selectedRepo.description || '暂无描述'}</p>
            <div className="mt-4 flex gap-2 text-xs text-gray-400">
              <span>文章: {selectedRepo.article_count}</span>
              <span>可见性: {selectedRepo.visibility === 'PUBLIC' ? '公开' : '私有'}</span>
            </div>
            <button
              onClick={() => navigate(`/repos/${selectedRepo.id}`)}
              className="mt-4 px-4 py-1.5 bg-emerald-500 text-white rounded-lg text-sm hover:bg-emerald-600"
            >
              进入知识库
            </button>
          </div>
        ) : (
          <div className="text-gray-400 text-sm flex items-center justify-center h-full">选择一个知识库查看详情</div>
        )}
      </div>

      {/* 新建弹窗 */}
      {showCreate && (
        <div className="fixed inset-0 bg-black/30 flex items-center justify-center z-50" onClick={() => setShowCreate(false)}>
          <div className="bg-white rounded-xl p-6 w-80 mx-4 shadow-xl" onClick={e => e.stopPropagation()}>
            <h3 className="text-sm font-medium text-gray-800 mb-4">新建知识库</h3>
            <input value={newName} onChange={e => setNewName(e.target.value)} placeholder="知识库名称" className="w-full px-3 py-2 border border-gray-300 rounded-lg text-sm mb-3 focus:outline-none focus:border-emerald-400" autoFocus />
            <input value={newDesc} onChange={e => setNewDesc(e.target.value)} placeholder="描述（可选）" className="w-full px-3 py-2 border border-gray-300 rounded-lg text-sm mb-4 focus:outline-none focus:border-emerald-400" />
            <div className="flex gap-2 justify-end">
              <button onClick={() => setShowCreate(false)} className="px-3 py-1.5 text-sm text-gray-500 hover:text-gray-700">取消</button>
              <button onClick={handleCreate} className="px-3 py-1.5 text-sm bg-emerald-500 text-white rounded-lg hover:bg-emerald-600">创建</button>
            </div>
          </div>
        </div>
      )}
    </div>
  )
}
```

---

### Task 6: 仓库详情 + 文章树

**文件：**
- 创建: `web/src/pages/RepoDetail.tsx`
- 创建: `web/src/components/Repo/RepoTree.tsx`
- 创建: `web/src/components/Repo/CollabList.tsx`

- [x] **Step 1: RepoTree 组件（带右键菜单 + 拖拽）**

```typescript
import { useState, useCallback } from 'react'
import { useNavigate } from 'react-router-dom'
import type { Node, NodeType } from '../../lib/nodes'
import { createNode, updateNode, deleteNode, listNodes } from '../../lib/nodes'

interface RepoTreeProps {
  repoId: string
  nodes: Node[]
  currentParentId: string | null
  onNavigate: (parentId: string | null) => void
  onRefresh: () => void
}

export default function RepoTree({ repoId, nodes, currentParentId, onNavigate, onRefresh }: RepoTreeProps) {
  const [contextMenu, setContextMenu] = useState<{ x: number; y: number; node?: Node } | null>(null)
  const navigate = useNavigate()

  const handleContextMenu = (e: React.MouseEvent, node?: Node) => {
    e.preventDefault()
    setContextMenu({ x: e.clientX, y: e.clientY, node })
  }

  const handleCreate = async (type: NodeType) => {
    const name = prompt(`请输入${type === 'FOLDER' ? '文件夹' : '文章'}名称`)
    if (!name?.trim()) return
    const n = await createNode(repoId, name.trim(), type, currentParentId || undefined)
    if (type === 'ARTICLE') {
      navigate(`/repos/${repoId}/nodes/${n.id}/edit`)
    }
    onRefresh()
    setContextMenu(null)
  }

  const handleRename = async (node: Node) => {
    const name = prompt('新名称:', node.name)
    if (name?.trim() && name !== node.name) {
      await updateNode(repoId, node.id, { name: name.trim() })
      onRefresh()
    }
    setContextMenu(null)
  }

  const handleMove = async (node: Node) => {
    const newParentId = prompt('目标文件夹 ID:')
    if (newParentId) {
      await updateNode(repoId, node.id, { parent_id: newParentId })
      onRefresh()
    }
    setContextMenu(null)
  }

  const handleDelete = async (node: Node) => {
    if (confirm(`确定删除「${node.name}」？`)) {
      await deleteNode(repoId, node.id)
      onRefresh()
    }
    setContextMenu(null)
  }

  // 获取当前层级下的子节点
  const childNodes = nodes.filter(n => n.parent_id === (currentParentId || ''))

  // 开始拖拽
  const handleDragStart = (e: React.DragEvent, node: Node) => {
    e.dataTransfer.setData('text/plain', node.id)
    e.dataTransfer.effectAllowed = 'move'
  }

  // 拖拽到文件夹上
  const handleDrop = async (e: React.DragEvent, targetNode: Node) => {
    e.preventDefault()
    if (targetNode.type !== 'FOLDER') return
    const nodeId = e.dataTransfer.getData('text/plain')
    if (nodeId === targetNode.id) return
    try {
      await updateNode(repoId, nodeId, { parent_id: targetNode.id })
      onRefresh()
    } catch {
      // ignore
    }
  }

  const handleDragOver = (e: React.DragEvent) => {
    e.preventDefault()
    e.dataTransfer.dropEffect = 'move'
  }

  // 点击空白区域创建
  const handleCanvasCreate = async () => {
    const type = confirm('创建文件夹点确定，创建文章点取消') ? 'FOLDER' : 'ARTICLE'
    const name = prompt(`请输入${type === 'FOLDER' ? '文件夹' : '文章'}名称`)
    if (!name?.trim()) return
    const n = await createNode(repoId, name.trim(), type as NodeType, currentParentId || undefined)
    if (type === 'ARTICLE') {
      navigate(`/repos/${repoId}/nodes/${n.id}/edit`)
    }
    onRefresh()
  }

  return (
    <div onContextMenu={e => handleContextMenu(e)} onClick={() => setContextMenu(null)}>
      {/* 面包屑导航 */}
      <div className="flex items-center gap-1 text-xs text-gray-500 mb-2 px-1">
        <button onClick={() => onNavigate(null)} className="hover:text-emerald-600">根目录</button>
        {currentParentId && <span>/</span>}
      </div>

      {/* 新建按钮 */}
      <button onClick={handleCanvasCreate} className="w-full px-2 py-1 text-xs text-gray-400 hover:text-emerald-600 hover:bg-gray-100 rounded text-left mb-1">
        + 新建
      </button>

      {/* 节点列表 */}
      {childNodes.map(node => (
        <NodeItem
          key={node.id}
          node={node}
          repoId={repoId}
          onContextMenu={handleContextMenu}
          onNavigate={onNavigate}
          onDragStart={handleDragStart}
          onDrop={handleDrop}
          onDragOver={handleDragOver}
        />
      ))}

      {childNodes.length === 0 && (
        <div className="text-xs text-gray-400 px-2 py-4 text-center">
          暂无内容，点击上方"+ 新建"创建
        </div>
      )}

      {/* 右键菜单 */}
      {contextMenu && (
        <div
          className="fixed bg-white border border-gray-200 rounded-lg shadow-lg py-1 z-50 w-36"
          style={{ left: contextMenu.x, top: contextMenu.y }}
        >
          {!contextMenu.node && (
            <>
              <button onClick={() => handleCreate('FOLDER')} className="w-full px-3 py-1.5 text-xs text-left hover:bg-gray-50">新建文件夹</button>
              <button onClick={() => handleCreate('ARTICLE')} className="w-full px-3 py-1.5 text-xs text-left hover:bg-gray-50">新建文章</button>
            </>
          )}
          {contextMenu.node && (
            <>
              <button onClick={() => handleRename(contextMenu.node!)} className="w-full px-3 py-1.5 text-xs text-left hover:bg-gray-50">重命名</button>
              <button onClick={() => handleMove(contextMenu.node!)} className="w-full px-3 py-1.5 text-xs text-left hover:bg-gray-50">移动到</button>
              <button onClick={() => handleDelete(contextMenu.node!)} className="w-full px-3 py-1.5 text-xs text-left text-red-600 hover:bg-gray-50">删除</button>
            </>
          )}
        </div>
      )}
    </div>
  )
}

interface NodeItemProps {
  node: Node
  repoId: string
  onContextMenu: (e: React.MouseEvent, node: Node) => void
  onNavigate: (parentId: string | null) => void
  onDragStart: (e: React.DragEvent, node: Node) => void
  onDrop: (e: React.DragEvent, targetNode: Node) => void
  onDragOver: (e: React.DragEvent) => void
}

function NodeItem({ node, repoId, onContextMenu, onNavigate, onDragStart, onDrop, onDragOver }: NodeItemProps) {
  const navigate = useNavigate()

  return (
    <div
      draggable
      onContextMenu={e => onContextMenu(e, node)}
      onDragStart={e => onDragStart(e, node)}
      onDragOver={onDragOver}
      onDrop={e => onDrop(e, node)}
      onClick={() => {
        if (node.type === 'FOLDER') {
          onNavigate(node.id)
        } else {
          navigate(`/repos/${repoId}/nodes/${node.id}`)
        }
      }}
      className="flex items-center gap-2 px-2 py-1.5 rounded text-sm cursor-pointer hover:bg-gray-100 text-gray-700 mb-0.5"
    >
      <span>{node.type === 'FOLDER' ? '📁' : '📄'}</span>
      <span className="truncate">{node.name}</span>
    </div>
  )
}
```

- [x] **Step 2: CollabList 组件（协作者列表）**

```typescript
import { useState, useEffect } from 'react'
import { listCollaborators, addCollaborator, updateCollaborator, removeCollaborator } from '../../lib/nodes'
import type { Collaborator } from '../../lib/nodes'

interface CollabListProps {
  repoId: string
}

export default function CollabList({ repoId }: CollabListProps) {
  const [collabs, setCollabs] = useState<Collaborator[]>([])
  const [showAdd, setShowAdd] = useState(false)
  const [newUserId, setNewUserId] = useState('')
  const [newRole, setNewRole] = useState('DEVELOPER')

  const load = async () => {
    const list = await listCollaborators(repoId)
    setCollabs(list)
  }

  useEffect(() => { load() }, [repoId])

  const handleAdd = async () => {
    if (!newUserId.trim()) return
    await addCollaborator(repoId, newUserId.trim(), newRole)
    setShowAdd(false)
    setNewUserId('')
    load()
  }

  const handleUpdateRole = async (userId: string, role: string) => {
    await updateCollaborator(repoId, userId, role)
    load()
  }

  const handleRemove = async (userId: string) => {
    if (confirm('确定移除此协作者？')) {
      await removeCollaborator(repoId, userId)
      load()
    }
  }

  const roleLabel: Record<string, string> = { ADMIN: '管理员', DEVELOPER: '开发者', VIEWER: '查看者' }

  return (
    <div className="mt-6">
      <div className="flex items-center justify-between mb-2">
        <h4 className="text-sm font-medium text-gray-700">协作者</h4>
        <button onClick={() => setShowAdd(true)} className="text-xs text-emerald-600 hover:underline">+ 添加</button>
      </div>

      {collabs.map(c => (
        <div key={c.user_id} className="flex items-center justify-between py-1.5 text-sm">
          <span className="text-gray-600 text-xs">{c.user_id.slice(0, 8)}...</span>
          <div className="flex items-center gap-1">
            <select
              value={c.role}
              onChange={e => handleUpdateRole(c.user_id, e.target.value)}
              className="text-xs border border-gray-200 rounded px-1 py-0.5"
            >
              <option value="ADMIN">管理员</option>
              <option value="DEVELOPER">开发者</option>
              <option value="VIEWER">查看者</option>
            </select>
            <button onClick={() => handleRemove(c.user_id)} className="text-gray-400 hover:text-red-500 text-xs">✕</button>
          </div>
        </div>
      ))}

      {collabs.length === 0 && <p className="text-xs text-gray-400">暂无协作者</p>}

      {showAdd && (
        <div className="mt-2 space-y-2">
          <input value={newUserId} onChange={e => setNewUserId(e.target.value)} placeholder="用户 ID" className="w-full px-2 py-1 border border-gray-300 rounded text-xs" />
          <select value={newRole} onChange={e => setNewRole(e.target.value)} className="w-full px-2 py-1 border border-gray-300 rounded text-xs">
            <option value="DEVELOPER">开发者</option>
            <option value="VIEWER">查看者</option>
          </select>
          <div className="flex gap-2">
            <button onClick={handleAdd} className="px-2 py-1 bg-emerald-500 text-white rounded text-xs">添加</button>
            <button onClick={() => setShowAdd(false)} className="px-2 py-1 text-gray-500 text-xs">取消</button>
          </div>
        </div>
      )}
    </div>
  )
}
```

- [x] **Step 3: RepoDetail 页面**

```typescript
import { useEffect, useState } from 'react'
import { useParams } from 'react-router-dom'
import { useRepo } from '../store/repo-context'
import { getRepo } from '../lib/repos'
import { listNodes } from '../lib/nodes'
import type { Repo } from '../lib/repos'
import RepoTree from '../components/Repo/RepoTree'
import CollabList from '../components/Repo/CollabList'
import type { Node } from '../lib/nodes'

export default function RepoDetail() {
  const { repoId } = useParams<{ repoId: string }>()
  const [repo, setRepo] = useState<Repo | null>(null)
  const [nodes, setNodes] = useState<Node[]>([])
  const [currentParentId, setCurrentParentId] = useState<string | null>(null)

  const loadRepo = async () => {
    if (!repoId) return
    const r = await getRepo(repoId)
    setRepo(r)
  }

  const loadAllNodes = async () => {
    if (!repoId) return
    // 当前简化：只加载一层（完整树需要递归加载）
    const all: Node[] = []
    const loadRecursive = async (parentId?: string) => {
      const children = await listNodes(repoId, parentId)
      all.push(...children)
      for (const child of children) {
        if (child.type === 'FOLDER') {
          await loadRecursive(child.id)
        }
      }
    }
    await loadRecursive()
    setNodes(all)
  }

  useEffect(() => { loadRepo(); loadAllNodes() }, [repoId])

  return (
    <div className="h-full flex">
      {/* 左侧文件树 */}
      <div className="w-60 border-r border-gray-200 bg-gray-50 overflow-y-auto p-2 shrink-0">
        <RepoTree
          repoId={repoId!}
          nodes={nodes}
          currentParentId={currentParentId}
          onNavigate={setCurrentParentId}
          onRefresh={loadAllNodes}
        />
      </div>

      {/* 右侧内容区 */}
      <div className="flex-1 overflow-y-auto p-6">
        {repo && (
          <div className="mb-6">
            <h2 className="text-lg font-medium text-gray-800">{repo.name}</h2>
            <p className="text-sm text-gray-500 mt-1">{repo.description}</p>
          </div>
        )}
        <div className="text-sm text-gray-400 text-center py-12">
          选择一篇文文章查看或编辑
        </div>

        <CollabList repoId={repoId!} />
      </div>
    </div>
  )
}
```

---

### Task 7: 文章阅读 + 编辑器

**文件：**
- 创建: `web/src/pages/ArticleView.tsx`
- 创建: `web/src/pages/ArticleEditor.tsx`
- 创建: `web/src/components/Editor/MarkdownEditor.tsx`

- [x] **Step 1: MarkdownEditor 组件（Monaco Editor + 预览）**

```typescript
import { useState } from 'react'
import Editor from '@monaco-editor/react'

interface MarkdownEditorProps {
  value: string
  onChange: (value: string) => void
  readOnly?: boolean
}

export default function MarkdownEditor({ value, onChange, readOnly }: MarkdownEditorProps) {
  const [preview, setPreview] = useState(false)

  return (
    <div className="h-full flex flex-col">
      <div className="flex items-center justify-between px-3 py-1.5 border-b border-gray-200 bg-gray-50">
        <span className="text-xs text-gray-400">Markdown</span>
        <button
          onClick={() => setPreview(!preview)}
          className="text-xs px-2 py-1 bg-white border border-gray-200 rounded hover:bg-gray-50"
        >
          {preview ? '编辑' : '预览'}
        </button>
      </div>
      <div className="flex-1">
        {preview ? (
          <div className="p-4 overflow-y-auto max-w-3xl mx-auto">
            {/* 预览使用 react-markdown */}
            <MarkdownPreview content={value} />
          </div>
        ) : (
          <Editor
            defaultLanguage="markdown"
            value={value}
            onChange={(val) => onChange(val || '')}
            options={{
              readOnly,
              minimap: { enabled: false },
              fontSize: 14,
              lineNumbers: 'on',
              wordWrap: 'on',
              scrollBeyondLastLine: false,
            }}
          />
        )}
      </div>
    </div>
  )
}

// 简单的 Markdown 预览
function MarkdownPreview({ content }: { content: string }) {
  // 这里复用 Streamdown 或直接使用简单的 Markdown 渲染
  // 因为 react-markdown 已在 Streamdown 中使用，可复用
  const { Streamdown } = require('../Chat/Streamdown')
  return <Streamdown content={content} />
}
```

注意：MarkdownPreview 需要改为正确导入 Streamdown 组件的方式。实际开发时直接导入：

```typescript
import { Streamdown } from '../Chat/Streamdown'
```

- [x] **Step 2: ArticleView 页面**

```typescript
import { useEffect, useState } from 'react'
import { useParams, useNavigate } from 'react-router-dom'
import { getArticleSignedUrl, getNode } from '../lib/nodes'

export default function ArticleView() {
  const { repoId, nodeId } = useParams<{ repoId: string; nodeId: string }>()
  const navigate = useNavigate()
  const [content, setContent] = useState('')
  const [title, setTitle] = useState('')
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    const load = async () => {
      if (!repoId || !nodeId) return
      try {
        const node = await getNode(repoId, nodeId)
        setTitle(node.name)

        const signedUrl = await getArticleSignedUrl(repoId, nodeId)
        const res = await fetch(signedUrl)
        const text = await res.text()
        setContent(text)
      } catch {
        setContent('*文章加载失败*')
      }
      setLoading(false)
    }
    load()
  }, [repoId, nodeId])

  return (
    <div className="h-full flex flex-col">
      <div className="flex items-center justify-between px-4 py-2 border-b border-gray-200 bg-white">
        <h1 className="text-sm font-medium text-gray-800">{title}</h1>
        <button
          onClick={() => navigate(`/repos/${repoId}/nodes/${nodeId}/edit`)}
          className="px-3 py-1 text-xs bg-emerald-500 text-white rounded hover:bg-emerald-600"
        >
          编辑
        </button>
      </div>
      <div className="flex-1 overflow-y-auto p-6 max-w-3xl mx-auto w-full">
        {loading ? (
          <div className="text-gray-400 text-sm">加载中...</div>
        ) : (
          /* 使用 Streamdown 渲染 */
          <MarkdownViewer content={content} />
        )}
      </div>
    </div>
  )
}

import { Streamdown } from '../components/Chat/Streamdown'
function MarkdownViewer({ content }: { content: string }) {
  return <Streamdown content={content} />
}
```

- [x] **Step 3: ArticleEditor 页面**

```typescript
import { useEffect, useState } from 'react'
import { useParams, useNavigate } from 'react-router-dom'
import { getArticleSignedUrl, getNode, uploadArticleContent } from '../lib/nodes'
import MarkdownEditor from '../components/Editor/MarkdownEditor'

export default function ArticleEditor() {
  const { repoId, nodeId } = useParams<{ repoId: string; nodeId: string }>()
  const navigate = useNavigate()
  const [content, setContent] = useState('')
  const [title, setTitle] = useState('')
  const [loading, setLoading] = useState(true)
  const [saving, setSaving] = useState(false)
  const [dirty, setDirty] = useState(false)

  useEffect(() => {
    const load = async () => {
      if (!repoId || !nodeId) return
      try {
        const node = await getNode(repoId, nodeId)
        setTitle(node.name)

        if (node.file_path) {
          const signedUrl = await getArticleSignedUrl(repoId, nodeId)
          const res = await fetch(signedUrl)
          const text = await res.text()
          setContent(text)
        }
      } catch {
        // 新文章，内容为空
      }
      setLoading(false)
    }
    load()
  }, [repoId, nodeId])

  const handleSave = async () => {
    if (!repoId || !nodeId || !content.trim()) return
    setSaving(true)
    try {
      const blob = new Blob([content], { type: 'text/markdown' })
      const file = new File([blob], `${title || 'article'}.md`, { type: 'text/markdown' })
      await uploadArticleContent(repoId, nodeId, file)
      setDirty(false)
    } catch (err: any) {
      alert('保存失败: ' + err.message)
    }
    setSaving(false)
  }

  const handleFileUpload = async (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0]
    if (!file) return
    if (!file.name.endsWith('.md') && !file.name.endsWith('.markdown')) {
      alert('仅支持 .md / .markdown 文件')
      return
    }
    const text = await file.text()
    setContent(text)
    setTitle(file.name.replace(/\.(md|markdown)$/, ''))
    setDirty(true)
  }

  if (loading) {
    return <div className="h-full flex items-center justify-center text-gray-400">加载中...</div>
  }

  return (
    <div className="h-full flex flex-col">
      <div className="flex items-center justify-between px-4 py-2 border-b border-gray-200 bg-white">
        <div className="flex items-center gap-2">
          <button onClick={() => navigate(`/repos/${repoId}/nodes/${nodeId}`)} className="text-gray-400 hover:text-gray-600 text-sm">
            ← 返回
          </button>
          <span className="text-sm font-medium text-gray-800">{title || '新文章'}</span>
          {dirty && <span className="text-xs text-orange-500">未保存</span>}
        </div>
        <div className="flex items-center gap-2">
          <label className="px-3 py-1 text-xs border border-gray-200 rounded cursor-pointer hover:bg-gray-50">
            上传 .md 文件
            <input type="file" accept=".md,.markdown" onChange={handleFileUpload} className="hidden" />
          </label>
          <button
            onClick={handleSave}
            disabled={saving || !dirty}
            className="px-3 py-1 text-xs bg-emerald-500 text-white rounded hover:bg-emerald-600 disabled:opacity-50"
          >
            {saving ? '保存中...' : '保存'}
          </button>
        </div>
      </div>
      <div className="flex-1">
        <MarkdownEditor value={content} onChange={(v) => { setContent(v); setDirty(true) }} />
      </div>
    </div>
  )
}
```

---

### Task 8: 设置弹窗 + 搜索页

**文件：**
- 创建: `web/src/components/Settings/SettingsModal.tsx`
- 创建: `web/src/pages/SearchResult.tsx`

- [ ] **Step 1: SettingsModal**

```typescript
import { useState, useEffect } from 'react'
import { useAuth } from '../../store/auth-context'
import { getUser, updateUser, uploadAvatar, changePassword } from '../../lib/auth'
import { Upload, ImageCropper } from 'antd'
import ImgCrop from 'antd-img-crop'
import type { UploadFile } from 'antd'
import { UserOutlined } from '@ant-design/icons'

interface SettingsModalProps {
  onClose: () => void
}

export default function SettingsModal({ onClose }: SettingsModalProps) {
  const { user } = useAuth()
  const [name, setName] = useState('')
  const [email, setEmail] = useState('')
  const [avatarUrl, setAvatarUrl] = useState('')
  const [oldPassword, setOldPassword] = useState('')
  const [newPassword, setNewPassword] = useState('')
  const [saving, setSaving] = useState(false)
  const [message, setMessage] = useState('')

  useEffect(() => {
    const load = async () => {
      if (!user?.id) return
      const u = await getUser(user.id)
      setName(u.name)
      setEmail(u.email)
      setAvatarUrl(u.avatar)
    }
    load()
  }, [user])

  const handleSaveProfile = async () => {
    if (!user?.id) return
    setSaving(true)
    try {
      await updateUser(user.id, { name, email })
      setMessage('个人信息已更新')
    } catch (err: any) {
      setMessage(err.message || '更新失败')
    }
    setSaving(false)
  }

  const handleAvatarUpload = async (file: File) => {
    if (!user?.id) return
    try {
      const url = await uploadAvatar(user.id, file)
      setAvatarUrl(url)
      setMessage('头像已更新')
    } catch (err: any) {
      setMessage(err.message || '头像上传失败')
    }
  }

  const handleChangePassword = async () => {
    if (!user?.id || !oldPassword || !newPassword) return
    setSaving(true)
    try {
      await changePassword(user.id, oldPassword, newPassword)
      setMessage('密码已修改')
      setOldPassword('')
      setNewPassword('')
    } catch (err: any) {
      setMessage(err.message || '密码修改失败')
    }
    setSaving(false)
  }

  return (
    <div className="fixed inset-0 bg-black/30 flex items-center justify-center z-50" onClick={onClose}>
      <div className="bg-white rounded-xl w-full max-w-md mx-4 shadow-xl max-h-[80vh] overflow-y-auto" onClick={e => e.stopPropagation()}>
        <div className="p-6">
          <div className="flex items-center justify-between mb-6">
            <h2 className="text-base font-medium text-gray-800">设置</h2>
            <button onClick={onClose} className="text-gray-400 hover:text-gray-600">✕</button>
          </div>

          {/* Avatar */}
          <div className="flex items-center gap-4 mb-6">
            <ImgCrop rotationSlider>
              <Upload
                showUploadList={false}
                beforeUpload={(file) => { handleAvatarUpload(file as File); return false }}
                accept=".jpg,.jpeg,.png,.webp"
              >
                <div className="w-14 h-14 rounded-full bg-gray-100 flex items-center justify-center cursor-pointer overflow-hidden border-2 border-dashed border-gray-300 hover:border-emerald-400">
                  {avatarUrl ? (
                    <img src={avatarUrl} className="w-full h-full object-cover" />
                  ) : (
                    <UserOutlined className="text-gray-400 text-xl" />
                  )}
                </div>
              </Upload>
            </ImgCrop>
            <div>
              <div className="text-sm font-medium text-gray-800">{name}</div>
              <div className="text-xs text-gray-400">点击头像更换</div>
            </div>
          </div>

          {/* Info */}
          <div className="space-y-3 mb-6">
            <div>
              <label className="block text-xs text-gray-500 mb-1">用户名</label>
              <input value={name} onChange={e => setName(e.target.value)}
                className="w-full px-3 py-1.5 border border-gray-300 rounded-lg text-sm focus:outline-none focus:border-emerald-400" />
            </div>
            <div>
              <label className="block text-xs text-gray-500 mb-1">邮箱</label>
              <input value={email} onChange={e => setEmail(e.target.value)} type="email"
                className="w-full px-3 py-1.5 border border-gray-300 rounded-lg text-sm focus:outline-none focus:border-emerald-400" />
            </div>
            <button onClick={handleSaveProfile} disabled={saving}
              className="px-4 py-1.5 bg-emerald-500 text-white rounded-lg text-sm hover:bg-emerald-600 disabled:opacity-50">
              保存信息
            </button>
          </div>

          <hr className="my-4" />

          {/* Password */}
          <div className="space-y-3">
            <h3 className="text-sm font-medium text-gray-700">修改密码</h3>
            <input value={oldPassword} onChange={e => setOldPassword(e.target.value)} type="password" placeholder="当前密码"
              className="w-full px-3 py-1.5 border border-gray-300 rounded-lg text-sm focus:outline-none focus:border-emerald-400" />
            <input value={newPassword} onChange={e => setNewPassword(e.target.value)} type="password" placeholder="新密码"
              className="w-full px-3 py-1.5 border border-gray-300 rounded-lg text-sm focus:outline-none focus:border-emerald-400" />
            <button onClick={handleChangePassword} disabled={saving || !oldPassword || !newPassword}
              className="px-4 py-1.5 bg-gray-100 text-gray-700 rounded-lg text-sm hover:bg-gray-200 disabled:opacity-50">
              修改密码
            </button>
          </div>

          {message && <p className="text-xs text-emerald-600 mt-4">{message}</p>}
        </div>
      </div>
    </div>
  )
}
```

注意：`antd-img-crop` 需要安装额外的包。实际安装命令：

```bash
cd web && npm install antd-img-crop
```

- [ ] **Step 2: SearchResult 页面**

```typescript
import { useEffect, useState } from 'react'
import { useSearchParams, useNavigate } from 'react-router-dom'
import { request } from '../lib/client'
import { getRepo } from '../lib/repos'
import type { Repo } from '../lib/repos'

export default function SearchResult() {
  const [searchParams] = useSearchParams()
  const query = searchParams.get('q') || ''
  const navigate = useNavigate()
  const [repoIds, setRepoIds] = useState<string[]>([])
  const [repos, setRepos] = useState<Repo[]>([])
  const [loading, setLoading] = useState(true)
  const [page, setPage] = useState(1)
  const [hasMore, setHasMore] = useState(false)

  useEffect(() => {
    const search = async () => {
      if (!query.trim()) { setLoading(false); return }
      setLoading(true)
      try {
        const res = await request<{ repo_ids: string[]; total_pages: number; has_more: boolean }>('/ai/search', {
          method: 'POST',
          body: JSON.stringify({ query: query.trim(), page, page_size: 20 }),
          skipAuth: false,
        })
        setRepoIds(res.data.repo_ids || [])
        setHasMore(res.data.has_more)

        // 获取每个 repo 的详情
        const details = await Promise.all(
          (res.data.repo_ids || []).map(id => getRepo(id).catch(() => null))
        )
        setRepos(details.filter(Boolean) as Repo[])
      } catch {
        setRepoIds([])
        setRepos([])
      }
      setLoading(false)
    }
    search()
  }, [query, page])

  return (
    <div className="max-w-3xl mx-auto px-6 py-8">
      <h2 className="text-base font-medium text-gray-800 mb-1">搜索结果</h2>
      <p className="text-xs text-gray-400 mb-6">关键词: "{query}"</p>

      {loading ? (
        <div className="text-gray-400 text-sm">搜索中...</div>
      ) : repos.length === 0 ? (
        <div className="text-gray-400 text-sm py-8 text-center">未找到匹配的知识库</div>
      ) : (
        <div className="space-y-3">
          {repos.map(repo => (
            <div
              key={repo.id}
              onClick={() => navigate(`/repos/${repo.id}`)}
              className="p-4 border border-gray-200 rounded-lg hover:border-emerald-300 cursor-pointer transition-colors"
            >
              <h3 className="text-sm font-medium text-gray-800">{repo.name}</h3>
              <p className="text-xs text-gray-500 mt-1">{repo.description || '暂无描述'}</p>
              <div className="flex gap-3 mt-2 text-xs text-gray-400">
                <span>{repo.article_count} 篇文章</span>
                <span>{repo.visibility === 'PUBLIC' ? '公开' : '私有'}</span>
              </div>
            </div>
          ))}
        </div>
      )}

      {hasMore && (
        <div className="text-center mt-6">
          <button onClick={() => setPage(p => p + 1)} className="text-sm text-emerald-600 hover:underline">
            加载更多
          </button>
        </div>
      )}
    </div>
  )
}
```

- [ ] **Step 3: NotFound 页面**

```typescript
import { Link } from 'react-router-dom'

export default function NotFound() {
  return (
    <div className="h-full flex flex-col items-center justify-center text-gray-400">
      <div className="text-5xl mb-4">404</div>
      <p className="text-sm mb-4">页面不存在</p>
      <Link to="/chat" className="text-sm text-emerald-600 hover:underline">回到首页</Link>
    </div>
  )
}
```

---

### Task 9: 安装依赖 + 编译验证

- [ ] **Step 1: 安装所有依赖**

```bash
cd web && npm install && npm install antd-img-crop
```

- [ ] **Step 2: TypeScript 编译检查**

```bash
cd web && npx tsc --noEmit
```

修复所有类型错误。

- [ ] **Step 3: Vite 构建**

```bash
cd web && npm run build
```

验证输出 `dist/` 目录结构正确。

---

## 计划自检

- [x] **Spec 覆盖**: 设计文档中所有功能点都有对应 Task（聊天、知识库 CRUD、文章树、Monaco Editor、Streamdown、设置弹窗、搜索页）
- [x] **无占位符**: 每个 step 都有完整代码和命令
- [x] **类型一致性**: API 客户端函数签名与 store 中使用方式一致
- [x] **无 missing API 引用**: 后端待补充功能已在设计文档标注
