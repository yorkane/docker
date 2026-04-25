#!/bin/bash
# Chrome Wrapper：根据 CDP_TOKEN 环境变量决定监听模式
rm -f /home/kasm-user/chrome-data/SingletonLock

if [ -n "$CDP_TOKEN" ]; then
    # Token 模式：Chrome 监听内部端口，由 cdp-auth-proxy 代理 9222
    exec /usr/bin/google-chrome-original \
        --user-data-dir=/home/kasm-user/chrome-data \
        --remote-debugging-port=19222 \
        --remote-allow-origins="*" "$@"
else
    # 无认证模式：Chrome 直接监听 0.0.0.0:9222，无代理开销
    exec /usr/bin/google-chrome-original \
        --user-data-dir=/home/kasm-user/chrome-data \
        --remote-debugging-port=9222 \
        --remote-debugging-address=0.0.0.0 \
        --remote-allow-origins="*" "$@"
fi
