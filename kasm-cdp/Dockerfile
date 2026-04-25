# === Stage 1: 编译 CDP 认证代理 (Go 静态二进制) ===
FROM golang:1.23-alpine AS builder
WORKDIR /build
COPY cdp-auth-proxy.go .
RUN go mod init cdp-auth-proxy && \
    CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -ldflags="-s -w" -o cdp-auth-proxy cdp-auth-proxy.go

# === Stage 2: 最终镜像 ===
FROM kasmweb/chrome:develop

USER root

# 1. Chrome Wrapper：根据 CDP_TOKEN 决定监听模式
#    有 Token → Chrome :19222 + 代理 :9222（认证）
#    无 Token → Chrome 直接 0.0.0.0:9222（零开销）
RUN sed -i 's/archive.ubuntu.com/mirrors.aliyun.com/g' /etc/apt/sources.list && \
    sed -i 's/security.ubuntu.com/mirrors.aliyun.com/g' /etc/apt/sources.list && \
    apt-get update && rm -rf /var/lib/apt/lists/* && \
    mv /usr/bin/google-chrome /usr/bin/google-chrome-original
COPY chrome-wrapper.sh /usr/bin/google-chrome
RUN chmod +x /usr/bin/google-chrome

# 2. VNC 配置：禁用 SSL、禁用密码、禁止客户端改分辨率
RUN sed -i 's/-sslOnly/-SecurityTypes None -DisableBasicAuth 1/g' /dockerstartup/vnc_startup.sh && \
    sed -i 's/ssl:/ssl:\n    require_ssl: false/g' /etc/kasmvnc/kasmvnc.yaml && \
    sed -i 's/allow_override_standard_vnc_server_settings: true/allow_override_standard_vnc_server_settings: false/g' /etc/kasmvnc/kasmvnc.yaml

# 3. 注入 Go 代理二进制 (~2MB) 和配置脚本
#    环境变量：
#      CDP_TOKEN            - CDP 认证 Token (未设置则不启动代理，直接放行)
#      KASM_MAX_FRAME_RATE  - 最大帧率 (默认 30)
#      KASM_JPEG_QUALITY    - JPEG 压缩质量 0-9 (默认 7)
#      KASM_MAX_QUALITY     - 最大编码质量 0-9 (默认 8)
COPY --from=builder /build/cdp-auth-proxy /usr/local/bin/cdp-auth-proxy
COPY kasm-config.sh /dockerstartup/kasm-config.sh
RUN chmod +x /dockerstartup/kasm-config.sh

# 切回非特权用户
USER 1000

# 启动顺序：VNC 配置 → CDP 代理(后台) → VNC 桌面
# 无 CDP_TOKEN 时代理为纯转发模式（零认证开销，splice 零拷贝）
ENTRYPOINT ["/bin/bash", "-c", "/dockerstartup/kasm-config.sh && cdp-auth-proxy & /dockerstartup/vnc_startup.sh"]