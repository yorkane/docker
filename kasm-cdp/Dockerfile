# 使用 Kasm 官方的 Chrome 镜像 (基于 Ubuntu)
FROM kasmweb/chrome:develop

USER root

# 1. 注入 Chrome Wrapper：确保 Chrome 使用内部 CDP 端口启动
#    外部 CDP 访问由 cdp-auth-proxy.py 统一代理和认证
RUN sed -i 's/archive.ubuntu.com/mirrors.aliyun.com/g' /etc/apt/sources.list && \
    sed -i 's/security.ubuntu.com/mirrors.aliyun.com/g' /etc/apt/sources.list && \
    apt-get update && rm -rf /var/lib/apt/lists/* && \
    mv /usr/bin/google-chrome /usr/bin/google-chrome-original && \
    echo '#!/bin/bash\nrm -f /home/kasm-user/chrome-data/SingletonLock\n/usr/bin/google-chrome-original --user-data-dir=/home/kasm-user/chrome-data --remote-debugging-port=19222 --remote-allow-origins="*" "$@"' > /usr/bin/google-chrome && \
    chmod +x /usr/bin/google-chrome

# 2. 允许 HTTP 访问，禁用密码认证，并禁止 Kasm 客户端动态修改分辨率
RUN sed -i 's/-sslOnly/-SecurityTypes None -DisableBasicAuth 1/g' /dockerstartup/vnc_startup.sh && \
    sed -i 's/ssl:/ssl:\n    require_ssl: false/g' /etc/kasmvnc/kasmvnc.yaml && \
    sed -i 's/allow_override_standard_vnc_server_settings: true/allow_override_standard_vnc_server_settings: false/g' /etc/kasmvnc/kasmvnc.yaml

# 3. 注入启动脚本和 CDP 认证代理
#    环境变量：
#      KASM_MAX_FRAME_RATE  - 最大帧率 (默认 30)
#      KASM_JPEG_QUALITY    - JPEG 压缩质量 0-9, 9 最高 (默认 7)
#      KASM_MAX_QUALITY     - 最大编码质量 0-9 (默认 8)
#      CDP_TOKEN            - CDP 认证 Token (未设置则免认证)
COPY kasm-config.sh /dockerstartup/kasm-config.sh
COPY cdp-auth-proxy.py /dockerstartup/cdp-auth-proxy.py
RUN chmod +x /dockerstartup/kasm-config.sh

# 切回 kasm 默认非特权用户
USER 1000

# 启动顺序：配置 VNC → 启动 CDP 认证代理(后台) → 启动 VNC
ENTRYPOINT ["/bin/bash", "-c", "/dockerstartup/kasm-config.sh && python3 /dockerstartup/cdp-auth-proxy.py & /dockerstartup/vnc_startup.sh"]