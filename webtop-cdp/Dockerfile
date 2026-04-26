# === Stage 1: 编译 CDP 认证代理 (Go 静态二进制) ===
FROM golang:1.23-alpine AS builder
WORKDIR /build
COPY cdp-auth-proxy.go .
RUN go mod init cdp-auth-proxy && \
    CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -ldflags="-s -w" -o cdp-auth-proxy cdp-auth-proxy.go

# === Stage 2: 最终镜像 ===
FROM lscr.io/linuxserver/webtop:alpine-openbox

# 1. 安装 Chromium + xdpyinfo（用于检测 X server 就绪）+ dbus（扩展需要）
RUN apk add --no-cache chromium xdpyinfo dbus dbus-x11 font-noto-cjk && \
    mkdir -p /config/extensions /config/ssl

# 2. 注入 Chrome Wrapper（自动重启 + CDP 内存优化参数）
COPY chrome-wrapper.sh /usr/local/bin/chrome-wrapper.sh
RUN chmod +x /usr/local/bin/chrome-wrapper.sh

# 3. 注入 CDP 认证代理 (~2MB)
COPY --from=builder /build/cdp-auth-proxy /usr/local/bin/cdp-auth-proxy

# 4. 注入 s6 服务：CDP 代理 + Chromium 自动启动
COPY cdp-service/ /etc/s6-overlay/s6-rc.d/svc-cdp-proxy/
COPY chromium-service/ /etc/s6-overlay/s6-rc.d/svc-chromium/

# 5. 注册 s6 服务到 user bundle
RUN touch /etc/s6-overlay/s6-rc.d/user/contents.d/svc-cdp-proxy && \
    touch /etc/s6-overlay/s6-rc.d/user/contents.d/svc-chromium
