#!/bin/bash
# Chrome Wrapper：自动重启 + CDP 内部端口监听
# Chrome 退出后自动重启，间隔 2 秒防止快速循环

while true; do
    rm -f /home/kasm-user/chrome-data/SingletonLock
    /usr/bin/google-chrome-original \
        --user-data-dir=/home/kasm-user/chrome-data \
        --remote-debugging-port=19222 \
        --remote-allow-origins="*" \
        --lang=${CHROME_LANGUAGE:-zh-CN} "$@"
    echo "⚠️ Chrome 已退出 (code=$?)，2 秒后自动重启..."
    sleep 2
done
