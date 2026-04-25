# Webtop CDP — 轻量级 Chromium CDP 容器

基于 `lscr.io/linuxserver/webtop:alpine-openbox` 的超轻量 Chrome DevTools Protocol 自动化容器。  
专为 **2GB 内存**服务器优化，使用 Alpine + Openbox + Chromium 方案。

## 特性

- 🪶 **极致轻量**：Alpine + Openbox，运行内存 300-400MB
- 🌐 **Web 桌面**：通过浏览器访问远程桌面（端口 3001）
- 🔌 **CDP 协议**：完整 Chrome DevTools Protocol 支持
- 🔒 **Token 认证**：可选 CDP Token 认证（私网免认证）
- ♻️ **自动重启**：Chromium 崩溃/退出后 2 秒自动重启
- 🔧 **s6 进程管理**：通过 s6-overlay 管理所有服务

## 快速开始

### 使用 docker-compose（推荐）

```bash
# 拉取并启动
docker compose up -d webtop-cdp

# 查看日志
docker logs -f webtop-cdp
```

### 使用 docker run

```bash
docker run -d \
  --name webtop-cdp \
  -p 3001:3001 \
  -p 9227:9222 \
  -e PUID=1000 \
  -e PGID=1000 \
  -v ./webtop-data:/config \
  --shm-size=256m \
  --security-opt seccomp=unconfined \
  ghcr.io/yorkane/webtop-cdp:latest
```

## 访问方式

| 服务 | 地址 | 说明 |
|---|---|---|
| Web 桌面 | `https://host:3001` | Selkies WebRTC 桌面 |
| CDP 协议 | `http://host:9227/json/version` | Chrome DevTools Protocol |

## 环境变量

| 变量 | 默认值 | 说明 |
|---|---|---|
| `WEBTOP_CDP_TOKEN` | _(空)_ | CDP 认证 Token，空则不认证 |
| `WEBTOP_CDP_PORT` | `9227` | 宿主机 CDP 映射端口 |
| `WEBTOP_PASSWORD` | _(空)_ | Web 桌面密码，空则无认证 |
| `WEBTOP_RESOLUTION` | `1280x720` | VNC/桌面分辨率 |
| `PUID` | `1000` | 运行用户 UID |
| `PGID` | `1000` | 运行用户 GID |

## 内存优化

针对 2GB 服务器的优化措施：

| 优化项 | 参数 | 效果 |
|---|---|---|
| V8 堆限制 | `--max-old-space-size=512` | 限制 JS 内存 512MB |
| 渲染进程限制 | `--renderer-process-limit=2` | 最多 2 个渲染进程 |
| 磁盘缓存 | `--disk-cache-size=32MB` | 限制缓存 |
| 禁用 GPU | `--disable-gpu` | 无 GPU 环境 |
| 共享内存 | `--disable-dev-shm-usage` | 使用 /tmp |

## 与 kasm-cdp 对比

| 对比项 | kasm-cdp | webtop-cdp |
|---|---|---|
| 基础系统 | Ubuntu + KasmVNC | Alpine + Selkies |
| 浏览器 | Google Chrome | Chromium |
| 运行内存 | ~800MB+ | ~300-400MB |
| 镜像大小 | ~1.5GB | ~500MB |
| 适用场景 | 充足资源服务器 | 2GB 小内存服务器 |
