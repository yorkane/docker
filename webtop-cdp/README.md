# Webtop CDP — 轻量级 Chromium CDP 容器

基于 `lscr.io/linuxserver/webtop:alpine-openbox` 的超轻量 Chrome DevTools Protocol 自动化容器。  
专为 **2GB 内存**服务器优化，使用 Alpine + Openbox + Chromium 方案。

## 特性

- 🪶 **极致轻量**：Alpine + Openbox，运行内存 300-400MB
- 🌐 **Web 桌面**：通过浏览器访问远程桌面（端口 3001）
- 🔌 **CDP 协议**：完整 Chrome DevTools Protocol 支持（TLS 加密）
- 🔒 **路径 Token**：Token 嵌入 URL 路径，直接兼容 Playwright 远程连接
- ♻️ **自动重启**：Chromium 崩溃/退出后 2 秒自动重启
- 🔧 **s6 进程管理**：通过 s6-overlay 管理所有服务
- 🔐 **自签名证书**：启动时自动生成 TLS 证书，也可挂载自有证书

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
  -e CDP_TOKEN=my-secret-token \
  -v ./webtop-data:/config \
  --shm-size=256m \
  --security-opt seccomp=unconfined \
  ghcr.io/yorkane/webtop-cdp:latest
```

## 访问方式

| 服务 | 地址 | 说明 |
|---|---|---|
| Web 桌面 | `https://host:3001` | Selkies WebRTC 桌面 |
| CDP 协议 | `https://host:9227/<token>/json/version` | Chrome DevTools Protocol (TLS) |
| Playwright | `wss://host:9227/<token>/` | `connectOverCDP` 远程连接 |

## Playwright 远程连接

```javascript
const { chromium } = require('playwright');

const browser = await chromium.connectOverCDP({
  endpointURL: 'https://host:9227/my-secret-token/',
});

const context = browser.contexts()[0];
const page = context.pages()[0] || await context.newPage();
await page.goto('https://example.com');
console.log(await page.title());
await browser.close();
```

> **自签名证书**：需设置 `NODE_TLS_REJECT_UNAUTHORIZED=0` 或在连接选项中 `ignoreHTTPSErrors: true`。

## 环境变量

| 变量 | 默认值 | 说明 |
|---|---|---|
| `WEBTOP_CDP_TOKEN` | _(空)_ | CDP Token，嵌入 URL 路径。空则无需 Token |
| `WEBTOP_CDP_PORT` | `9227` | 宿主机 CDP 映射端口 |
| `WEBTOP_CDP_TLS_CERT` | `/config/ssl/cdp-cert.pem` | TLS 证书路径（不存在则自动生成） |
| `WEBTOP_CDP_TLS_KEY` | `/config/ssl/cdp-key.pem` | TLS 密钥路径 |
| `WEBTOP_PASSWORD` | _(空)_ | Web 桌面密码，空则无认证 |
| `CHROME_EXTENSIONS` | _(空)_ | 预加载扩展路径（逗号分隔），或放入 `/config/extensions/` 自动加载 |
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
| CDP 传输 | HTTP (明文) | HTTPS/WSS (TLS) |
| 认证方式 | Query/Bearer Token | 路径前缀 Token |
| 运行内存 | ~800MB+ | ~300-400MB |
| 镜像大小 | ~1.5GB | ~500MB |
| 适用场景 | 充足资源服务器 | 2GB 小内存服务器 |
