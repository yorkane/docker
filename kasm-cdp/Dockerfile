# 使用 Kasm 官方的 Chrome 镜像 (基于 Ubuntu)
FROM kasmweb/chrome:develop

USER root

# 1. 注入全局 Wrapper：使用 socat 代理 CDP 端口，确保 Chrome 正确监听并允许访问
RUN sed -i 's/archive.ubuntu.com/mirrors.aliyun.com/g' /etc/apt/sources.list && \
    sed -i 's/security.ubuntu.com/mirrors.aliyun.com/g' /etc/apt/sources.list && \
    apt-get update && apt-get install -y socat && rm -rf /var/lib/apt/lists/* && \
    mv /usr/bin/google-chrome /usr/bin/google-chrome-original && \
    echo '#!/bin/bash\nrm -f /home/kasm-user/chrome-data/SingletonLock\nif ! pgrep -x socat > /dev/null; then socat TCP-LISTEN:9222,fork,bind=0.0.0.0 TCP:127.0.0.1:19222 & fi\n/usr/bin/google-chrome-original --user-data-dir=/home/kasm-user/chrome-data --remote-debugging-port=19222 --remote-allow-origins="*" "$@"' > /usr/bin/google-chrome && \
    chmod +x /usr/bin/google-chrome

# 2. 允许 HTTP 访问，禁用密码认证，并禁止 Kasm 客户端动态修改分辨率
RUN sed -i 's/-sslOnly/-SecurityTypes None -DisableBasicAuth 1/g' /dockerstartup/vnc_startup.sh && \
    sed -i 's/ssl:/ssl:\n    require_ssl: false/g' /etc/kasmvnc/kasmvnc.yaml && \
    sed -i 's/allow_override_standard_vnc_server_settings: true/allow_override_standard_vnc_server_settings: false/g' /etc/kasmvnc/kasmvnc.yaml

# 3. 注入启动脚本：根据环境变量动态配置 KasmVNC 帧率和画质
#    支持的环境变量：
#      KASM_MAX_FRAME_RATE  - 最大帧率 (默认 30)
#      KASM_JPEG_QUALITY    - JPEG 压缩质量 0-9, 9 最高 (默认 7)
#      KASM_MAX_QUALITY     - 最大编码质量 0-9 (默认 8)
COPY kasm-config.sh /dockerstartup/kasm-config.sh
RUN chmod +x /dockerstartup/kasm-config.sh

# 切回 kasm 默认非特权用户
USER 1000

# 在启动时先执行配置脚本，再启动 VNC
ENTRYPOINT ["/bin/bash", "-c", "/dockerstartup/kasm-config.sh && /dockerstartup/vnc_startup.sh"]

# docker build -t kasm-cdp kasm-cdp/