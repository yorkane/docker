#!/bin/bash
# Chrome Wrapper：Chrome 始终监听内部端口 19222，由代理管理外部访问
rm -f /home/kasm-user/chrome-data/SingletonLock

exec /usr/bin/google-chrome-original \
    --user-data-dir=/home/kasm-user/chrome-data \
    --remote-debugging-port=19222 \
    --remote-allow-origins="*" "$@"
