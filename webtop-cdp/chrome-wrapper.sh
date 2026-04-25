#!/usr/bin/with-contenv bash
# Chromium Wrapper：自动重启 + CDP + 内存优化
# 适用于 2GB 内存服务器

CHROME_BIN=$(which chromium-browser 2>/dev/null || which chromium 2>/dev/null)

if [ -z "$CHROME_BIN" ]; then
    echo "❌ Chromium 未找到"
    exit 1
fi

# 确保数据目录存在
mkdir -p /config/chromium-data

# 等待 X server 就绪
echo "⏳ 等待桌面环境就绪..."
until xdpyinfo -display "${DISPLAY:-:1}" >/dev/null 2>&1; do
    sleep 1
done
echo "✅ 桌面环境就绪 (DISPLAY=${DISPLAY:-:1})"

while true; do
    echo "🚀 启动 Chromium (CDP :19222)..."
    $CHROME_BIN \
        --no-sandbox \
        --disable-gpu \
        --disable-software-rasterizer \
        --remote-debugging-port=19222 \
        --remote-debugging-address=127.0.0.1 \
        --remote-allow-origins="*" \
        --user-data-dir=/config/chromium-data \
        --disable-dev-shm-usage \
        --disable-background-networking \
        --disable-default-apps \
        --disable-extensions \
        --disable-sync \
        --disable-translate \
        --no-first-run \
        --no-default-browser-check \
        --disable-backgrounding-occluded-windows \
        --renderer-process-limit=2 \
        --js-flags="--max-old-space-size=512" \
        --disk-cache-size=33554432 \
        --disable-breakpad \
        --lang=zh-CN \
        about:blank 2>&1
    EXIT_CODE=$?
    echo "⚠️ Chromium 已退出 (code=$EXIT_CODE)，2 秒后自动重启..."
    sleep 2
done
