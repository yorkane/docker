# Kasm Chrome with CDP (kasm-cdp)

基于 `kasmweb/chrome:develop` 的 Docker 镜像，暴露 Chrome DevTools Protocol (CDP) 端口，支持通过 Puppeteer、Playwright 等自动化工具远程控制浏览器，同时提供 VNC Web 桌面用于可视化调试。

## 功能特性

- **CDP 远程调试**：Chrome DevTools Protocol 通过 `9222` 端口暴露
- **VNC Web 桌面**：通过 `6901` 端口在浏览器中访问桌面
- **CDP Token 认证**：外部/局域网 IP 访问 CDP 端口需 Token 认证，容器间通信免认证
- **可调画质**：通过环境变量动态配置 VNC 帧率和图像质量

## CDP 认证机制

当设置了 `CDP_TOKEN` 环境变量时，CDP 认证代理会启用 Token 校验：

| 来源 IP | 认证要求 |
|---|---|
| `127.0.0.0/8` (localhost) | ✅ 免认证 |
| `10.0.0.0/8` (Docker/内网) | ✅ 免认证 |
| `172.16.0.0/12` (Docker bridge) | ✅ 免认证 |
| `192.168.0.0/16` (局域网) | ✅ 免认证 |
| 其他外部 IP | 🔒 需要 Token |

**认证方式（二选一）：**

```bash
# 方式一：URL 参数
curl http://<host>:9222/json/version?token=<your-token>

# 方式二：Authorization Header
curl -H "Authorization: Bearer <your-token>" http://<host>:9222/json/version

# WebSocket 连接也支持 token 参数
ws://<host>:9222/devtools/browser/<id>?token=<your-token>
```

> **注意**：未设置 `CDP_TOKEN` 时，所有请求均放行，行为与之前版本一致。

## 环境变量

| 变量 | 默认值 | 说明 |
|---|---|---|
| `CDP_TOKEN` | _(空)_ | CDP 认证 Token，为空则不启用认证 |
| `VNC_PW` | _(必填)_ | VNC 密码（Kasm 基础镜像要求） |
| `KASM_MAX_FRAME_RATE` | `30` | VNC 最大帧率 |
| `KASM_JPEG_QUALITY` | `7` | JPEG 压缩质量 (0-9，9 最高) |
| `KASM_MAX_QUALITY` | `8` | 最大编码质量 (0-9) |

## 使用方法

### 1. 构建镜像（可选，CI 自动构建）

```bash
docker build -t ghcr.io/yorkane/kasm-cdp:latest kasm-cdp/
```

### 2. 推送镜像（可选，CI 自动推送）

```bash
docker push ghcr.io/yorkane/kasm-cdp:latest
```

> 提交到 `kasm-cdp/` 目录的代码变更会自动触发 GitHub Actions 构建并推送镜像。

### 3. 启动容器

**基础启动（无 CDP 认证）：**

```bash
docker run -d \
  --name kasm-cdp \
  --shm-size=512m \
  -p 6901:6901 \
  -p 9222:9222 \
  -e VNC_PW=password \
  ghcr.io/yorkane/kasm-cdp:latest
```

**启用 CDP Token 认证（推荐生产环境使用）：**

```bash
docker run -d \
  --name kasm-cdp \
  --shm-size=512m \
  -p 6901:6901 \
  -p 9222:9222 \
  -e VNC_PW=password \
  -e CDP_TOKEN=my-secret-token-123 \
  ghcr.io/yorkane/kasm-cdp:latest
```

**自定义画质：**

```bash
docker run -d \
  --name kasm-cdp \
  --shm-size=512m \
  -p 6901:6901 \
  -p 9222:9222 \
  -e VNC_PW=password \
  -e CDP_TOKEN=my-secret-token-123 \
  -e KASM_MAX_FRAME_RATE=60 \
  -e KASM_JPEG_QUALITY=9 \
  -e KASM_MAX_QUALITY=9 \
  ghcr.io/yorkane/kasm-cdp:latest
```

### 4. 验证服务

```bash
# 测试 VNC 桌面
curl -sk -o /dev/null -w "HTTP %{http_code}\n" https://localhost:6901

# 测试 CDP（无认证模式）
curl http://localhost:9222/json/version

# 测试 CDP（Token 认证模式）
curl "http://localhost:9222/json/version?token=my-secret-token-123"
curl -H "Authorization: Bearer my-secret-token-123" http://localhost:9222/json/version
```

## 端口说明

| 端口 | 协议 | 用途 |
|---|---|---|
| `6901` | HTTPS/WSS | KasmVNC Web 桌面 |
| `9222` | HTTP/WS | Chrome DevTools Protocol (CDP) |

## 架构

```
外部请求 → :9222 [cdp-auth-proxy.py] → Token校验 → :19222 [Chrome CDP]
                                     ↗
容器内/Docker网络 → :9222 → 直接放行 → :19222 [Chrome CDP]
```
