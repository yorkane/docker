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

# 1. Chrome Wrapper：CDP 监听内部端口 19222，由代理统一管理 9222
#    同时移除 apt 缓存，不安装额外依赖
RUN sed -i 's/archive.ubuntu.com/mirrors.aliyun.com/g' /etc/apt/sources.list && \
    sed -i 's/security.ubuntu.com/mirrors.aliyun.com/g' /etc/apt/sources.list && \
    mv /usr/bin/google-chrome /usr/bin/google-chrome-original && \
    printf '#!/bin/bash\nrm -f /home/kasm-user/chrome-data/SingletonLock\n/usr/bin/google-chrome-original --user-data-dir=/home/kasm-user/chrome-data --remote-debugging-port=19222 --remote-allow-origins="*" "$@"\n' > /usr/bin/google-chrome && \
    chmod +x /usr/bin/google-chrome

# 2. VNC 配置：禁用 SSL、禁用密码、禁止客户端改分辨率
RUN sed -i 's/-sslOnly/-SecurityTypes None -DisableBasicAuth 1/g' /dockerstartup/vnc_startup.sh && \
    sed -i 's/ssl:/ssl:\n    require_ssl: false/g' /etc/kasmvnc/kasmvnc.yaml && \
    sed -i 's/allow_override_standard_vnc_server_settings: true/allow_override_standard_vnc_server_settings: false/g' /etc/kasmvnc/kasmvnc.yaml

# 3. 注入 Go 代理二进制 (~3MB) 和配置脚本
#    环境变量：
#      CDP_TOKEN            - CDP 认证 Token (未设置则免认证)
#      KASM_MAX_FRAME_RATE  - 最大帧率 (默认 30)
#      KASM_JPEG_QUALITY    - JPEG 压缩质量 0-9 (默认 7)
#      KASM_MAX_QUALITY     - 最大编码质量 0-9 (默认 8)
COPY --from=builder /build/cdp-auth-proxy /usr/local/bin/cdp-auth-proxy
COPY kasm-config.sh /dockerstartup/kasm-config.sh
RUN chmod +x /dockerstartup/kasm-config.sh

# 切回非特权用户
USER 1000

# 启动顺序：VNC 配置 → CDP 认证代理(后台) → VNC 桌面
ENTRYPOINT ["/bin/bash", "-c", "/dockerstartup/kasm-config.sh && cdp-auth-proxy & /dockerstartup/vnc_startup.sh"]