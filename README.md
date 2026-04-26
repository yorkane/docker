# Docker 镜像集合

本仓库包含多个 Docker 镜像工程，每个子目录是一个独立的镜像项目。  
所有镜像通过 GitHub Actions 自动构建并推送至 `ghcr.io/yorkane/`。

## 镜像列表

| 镜像 | 基础系统 | 说明 | 运行内存 |
|---|---|---|---|
| [kasm-cdp](./kasm-cdp/) | Ubuntu + KasmVNC | Chrome CDP 自动化容器（功能完整版） | ~800MB |
| [webtop-cdp](./webtop-cdp/) | Alpine + Openbox | Chromium CDP 轻量容器（2GB 服务器优化） | ~190MB |

## 快速开始

```bash
# 配置环境变量
cp .env.example .env
# 编辑 .env 设置 Token、密码等

# 启动轻量版（推荐 2GB 服务器）
docker compose up -d webtop-cdp

# 或启动完整版（4GB+ 服务器）
docker compose up -d chrome-cdp
```

## 环境变量

详见 [.env.example](./.env.example) 文件。

### Kasm CDP 变量（KASM_ 前缀）

| 变量 | 默认值 | 说明 |
|---|---|---|
| `KASM_VNC_PW` | `disabled` | VNC 密码，`disabled` 禁用认证 |
| `KASM_CDP_TOKEN` | _(空)_ | CDP Token，空则不认证 |
| `KASM_CDP_PORT` | `9226` | CDP 外部端口 |

### Webtop CDP 变量（WEBTOP_ 前缀）

| 变量 | 默认值 | 说明 |
|---|---|---|
| `WEBTOP_USER` | _(空)_ | Web 桌面用户名，空则无认证 |
| `WEBTOP_PASSWORD` | _(空)_ | Web 桌面密码，空则无认证 |
| `WEBTOP_CDP_TOKEN` | _(空)_ | CDP Token，空则不认证 |
| `WEBTOP_CDP_PORT` | `9227` | CDP 外部端口 |

## 访问方式

| 服务 | Kasm CDP | Webtop CDP |
|---|---|---|
| Web 桌面 | `http://host:6901` | `https://host:3001` |
| CDP 协议 | `http://host:9226/json/version` | `https://host:9227/<token>/json/version` |
| Playwright | — | `wss://host:9227/<token>/` |

## CI/CD

每次 push 到 `main` 分支时，GitHub Actions 自动构建对应目录的镜像：
- `ghcr.io/yorkane/kasm-cdp:latest`
- `ghcr.io/yorkane/webtop-cdp:latest`
