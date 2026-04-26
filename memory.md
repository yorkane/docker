# 项目 Memory（Agent 上下文）

> 本文件供 AI Agent 在每次对话开始时读取，快速理解项目结构与约定。

## 仓库概览

- **仓库**: `yorkane/docker` — Docker 镜像工程集合
- **远程**: `https://github.com/yorkane/docker`
- **默认分支**: `main`
- **镜像注册表**: `ghcr.io/yorkane/`
- **语言**: Shell / Go / Dockerfile / YAML

## 核心定位

本仓库包含多个独立 Docker 镜像项目，每个子目录是一个完整的镜像工程。所有镜像均面向 **Chrome DevTools Protocol (CDP) 远程浏览器自动化** 场景，为 Playwright / Puppeteer 等自动化框架提供可远程控制的浏览器容器。

## 目录结构

```
/code/docker/
├── .github/workflows/       # GitHub Actions CI/CD（每目录一个 workflow）
│   ├── kasm-cdp.yml          # kasm-cdp 自动构建
│   └── webtop-cdp.yml        # webtop-cdp 自动构建
├── kasm-cdp/                 # 完整版 CDP 容器（Ubuntu + KasmVNC + Chrome）
│   ├── Dockerfile            # 多阶段构建: golang → kasmweb/chrome:develop
│   ├── cdp-auth-proxy.go     # CDP 认证代理（Go）
│   ├── chrome-wrapper.sh     # Chrome 自动重启包装脚本
│   ├── kasm-config.sh        # VNC 认证/画质动态配置
│   ├── docker-compose.yml    # 独立 compose
│   ├── .env.example          # 环境变量模板
│   └── README.md
├── webtop-cdp/               # 轻量版 CDP 容器（Alpine + Openbox + Chromium）
│   ├── Dockerfile            # 多阶段构建: golang → linuxserver/webtop:alpine-openbox
│   ├── cdp-auth-proxy.go     # CDP 认证代理（Go，与 kasm-cdp 共用同一份代码）
│   ├── chrome-wrapper.sh     # Chromium 自动重启 + 内存优化参数
│   ├── cdp-service/          # s6-overlay 服务定义：CDP 代理
│   │   ├── run
│   │   └── type
│   ├── chromium-service/     # s6-overlay 服务定义：Chromium 浏览器
│   │   ├── run               # 含 X 等待、D-Bus 启动、下载目录配置
│   │   └── type
│   └── README.md
├── docker-compose.yml        # 根级 compose（统一编排入口）
├── .env.example              # 根级环境变量模板（含 Selkies 画质/音频参数）
├── .env                      # 实际环境变量（已 gitignore）
├── .gitignore
├── README.md                 # 项目总览文档
├── LICENSE                   # Apache 2.0
└── memory.md                 # 本文件
```

## 镜像对比

| 维度 | `kasm-cdp` | `webtop-cdp` |
|---|---|---|
| 镜像地址 | `ghcr.io/yorkane/kasm-cdp:latest` | `ghcr.io/yorkane/webtop-cdp:latest` |
| 基础镜像 | `kasmweb/chrome:develop` | `lscr.io/linuxserver/webtop:alpine-openbox` |
| 操作系统 | Ubuntu | Alpine |
| 浏览器 | Google Chrome | Chromium |
| 桌面方案 | KasmVNC (HTTP :6901) | Selkies WebRTC (HTTPS :3001) |
| 进程管理 | Kasm 内置 + bash | s6-overlay |
| 运行内存 | ~800MB+ | ~300-400MB |
| 镜像大小 | ~1.5GB | ~500MB |
| 适用场景 | 资源充足服务器 (4GB+) | 小内存服务器 (2GB) |
| 画质控制 | KasmVNC YAML 配置 | Selkies 环境变量 |

## 共享核心组件

### CDP 认证代理 (`cdp-auth-proxy.go`)

两个镜像共用同一份 Go 源码（分别存储在各自目录中）。

- **监听**: `0.0.0.0:9222`（容器内端口）
- **上游**: `127.0.0.1:19222`（Chrome/Chromium 实际 CDP 端口）
- **认证逻辑**:
  - 环境变量 `CDP_TOKEN` 未设置 → 全部放行
  - 私有网段 (127/8, 10/8, 172.16/12, 192.168/16) → 免认证
  - 外部 IP → 需 `?token=xxx` 或 `Authorization: Bearer xxx`
- **技术**: Go 静态编译、TCP splice 零拷贝、goroutine-per-connection

### Chrome Wrapper

自动重启循环脚本，Chrome 退出后 2 秒自动重拉。webtop-cdp 版本额外包含内存优化参数：
- `--renderer-process-limit=2`
- `--js-flags="--max-old-space-size=512"`
- `--disk-cache-size=33554432`
- `--disable-gpu` / `--disable-dev-shm-usage`
- 支持 `CHROME_EXTENSIONS` 环境变量或 `/config/extensions/` 自动加载

## CI/CD 约定

1. **每个子目录一个 Workflow**: `.github/workflows/<目录名>.yml`
2. **触发条件**: push 到 `main` 分支且对应目录或 workflow 文件有变更
3. **构建流程**: `actions/checkout@v4` → `docker/login-action@v3 (ghcr.io)` → `docker/build-push-action@v6`
4. **Tag 策略**: 无代码 tag 时始终构建 `latest`；有 tag 应对应版本号
5. **Dockerfile 修改即触发**: 任何对子目录内文件的修改（包括 Dockerfile、脚本、配置）都会触发重新构建

## 开发约定

### 代码管理
- **必须使用 `github-mcp` 工具管理代码**（commit、push、PR 等通过 MCP 完成）
- **每个子目录独立**: 配备独立 `README.md`、`Dockerfile`，可能有独立 `docker-compose.yml`
- **根 `docker-compose.yml`**: 统一编排入口，引用根 `.env` 文件

### 新增镜像项目流程
1. 创建新子目录 `/<image-name>/`
2. 编写 `Dockerfile` + `README.md`
3. 创建 `.github/workflows/<image-name>.yml`（复制已有模板修改 `context` 和 `tags`）
4. 在根 `README.md` 的镜像列表中添加条目
5. 如需统一编排，在根 `docker-compose.yml` 中添加 service
6. push 触发 GitHub Actions 自动构建镜像到 `ghcr.io/yorkane/<image-name>:latest`

### 环境变量命名
- Kasm 系列: `KASM_` 前缀（`KASM_VNC_PW`, `KASM_CDP_TOKEN`, `KASM_CDP_PORT`）
- Webtop 系列: `WEBTOP_` 前缀（`WEBTOP_USER`, `WEBTOP_PASSWORD`, `WEBTOP_CDP_TOKEN`）
- Selkies 画质: `SELKIES_` 前缀（`SELKIES_FRAMERATE`, `SELKIES_ENCODER`, `SELKIES_JPEG_QUALITY`）
- 容器内部: 无前缀（`CDP_TOKEN`, `VNC_PW`, `DISPLAY`）

## 端口约定

| 服务 | 容器内部 | kasm-cdp 宿主机默认 | webtop-cdp 宿主机默认 |
|---|---|---|---|
| Web 桌面 | 6901 (kasm) / 3001 (webtop) | 6901 | 3001 |
| CDP 代理 | 9222 | 9226 | 9227 |
| Chrome 实际 CDP | 19222 | — (仅容器内) | — (仅容器内) |

## 基础设施上下文（用户环境）

> 以下为用户常用基础设施，新项目如需可直接使用：

| 服务 | 地址 | 认证 |
|---|---|---|
| Redis | `192.168.1.30:6379` | 密码 `Wasu3.14` |
| PostgreSQL | `192.168.1.31` | 用户 `postgres` / 密码 `Wasu@3.14` |
| VLM 模型 (OpenAI 协议) | `vllm235-1\|2\|3.ai-t.wtvdev.com` | 3 实例 |
| OCR 服务 | `ocr-cpu.ai-t.wtvdev.com` | 文档 `/docs` |
| ASR 服务 | `asr217.ai-t.wtvdev.com` | 文档 `/docs` |
| 共享磁盘 | `/onas/share1\|2\|3` | 3 目录 |

## 常见操作速查

```bash
# 启动轻量版
docker compose up -d webtop-cdp

# 启动完整版
docker compose up -d chrome-cdp

# 验证 CDP 可用性
curl http://localhost:9227/json/version              # webtop-cdp
curl http://localhost:9226/json/version              # kasm-cdp

# 带 Token 认证
curl "http://host:9227/json/version?token=<token>"

# 查看日志
docker logs -f webtop-cdp
docker logs -f kasm-cdp
```

## 注意事项

1. **`cdp-auth-proxy.go` 在两个目录中各有一份副本**，修改时需同步更新（或后续考虑提取为共享构建上下文）
2. **`.env` 文件已 gitignore**，只有 `.env.example` 被跟踪
3. **`webtop-data/` 和 `chrome-data/`** 是容器 volume 挂载目录，已 gitignore
4. **Dockerfile 采用多阶段构建**：Stage 1 用 `golang:1.23-alpine` 编译 Go 代理，Stage 2 为最终运行镜像
5. **webtop-cdp 使用 s6-overlay 管理进程**，kasm-cdp 使用 bash entrypoint + 后台进程
6. **CJK 字体**已在 webtop-cdp 中安装 (`font-noto-cjk`)，kasm-cdp 基于 Ubuntu 自带
7. **扩展加载**: webtop-cdp 支持 `CHROME_EXTENSIONS` 环境变量或 `/config/extensions/` 目录自动加载
